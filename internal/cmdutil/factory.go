package cmdutil

import (
	"context"
	"errors"
	"fmt"
	"io"
	"log/slog"
	"os"
	"path/filepath"
	"sync"
	"twiggit/internal/core"
	"twiggit/internal/iostreams"
)

// FactoryOption mutates a Factory during construction. Functional options
// keep NewFactory's signature stable as Factory dependencies grow.
type FactoryOption func(*Factory)

// Client is the type of the composite git client exposed via Factory.GitClient.
// Declared as `any` so `*git.Client` (in internal/git) satisfies it implicitly
// without cmdutil importing internal/git — the depguard narrow relies on
// this. Callers narrow via the per-role fields (RepoOpener, BranchReader,
// ...) which already type as core.* interfaces, or cast to `*git.Client`
// directly inside cmd/ which keeps the concrete type because cmd/ still
// imports internal/git.
type Client = any

// Sentinel errors returned by the default lazy fields when their
// corresponding functional option is not supplied. Wrapped with %w by Init
// so `errors.Is(err, cmdutil.ErrNoConfigLoader)` walks the chain.
var (
	ErrNoConfigLoader = errors.New("cmdutil: no config loader wired")
	ErrNoGitClient    = errors.New("cmdutil: no git client wired")
)

// Factory is the composition seam shared by every cmd/*.go. Each lazy
// field is a function so callers invoke it at first use rather than
// constructing every dependency eagerly. The Config field is the
// canonical sync.OnceValue cache: every other field that depends on
// config (GitClient, Logger) reaches it through f.Config(), so the
// config file is parsed exactly once per binary invocation.
//
// Construction populates IOStreams, Context, and Executable eagerly
// (they are cheap and never fail). The expensive fields (Config,
// GitClient, Logger) are sync.OnceValue / sync.OnceFunc wrappers that
// run on first call. Init touches each lazy field so initialization
// failures surface before any command body executes.
//
// Customizing for tests: supply functional options via NewFactory(opts...)
// before the first call (or before Init). Once cached, the result is
// pinned for the Factory's lifetime.
type Factory struct {
	// IOStreams is the terminal I/O surface. Populated by NewFactory
	// from iostreams.System(); tests may swap in iostreams.Test().
	IOStreams *iostreams.IOStreams

	// Context is the application context. main.go sets it to the
	// signal.NotifyContext result so SIGINT / SIGTERM cancel every
	// long-running command; tests pass t.Context(). Subcommand RunE
	// closures read this through opts.Ctx.
	Context context.Context

	// AppVersion is the build-time injected version string. main.go
	// supplies it via WithVersion(version.Version); the default no-arg
	// NewFactory leaves it empty so factory_test.go literals compile
	// unchanged.
	AppVersion string

	// Executable is the basename of the running binary (twiggit).
	Executable string

	// Config returns the loaded config. sync.OnceValue-cached; the
	// first call reads and parses the config file, every later call
	// returns the same *core.Config pointer. A load error pins for
	// the Factory's lifetime (subsequent calls return the same error).
	// Default is a sentinel-returning function; wire with WithConfigLoader.
	Config func() (*core.Config, error)

	// GitClient returns the composite git client. sync.OnceValue-cached;
	// the closure first reads f.Config() so config load errors surface
	// here rather than as a confused git-construction failure.
	// Default is a sentinel-returning function; wire with WithGitClientFactory.
	GitClient func() (Client, error)

	// Per-role lazy fields. Each returns the cached composite typed as
	// the requested role (Go interface satisfaction via embedded promotion
	// on the composite). Bodies route through f.GitClient() so the
	// sync.OnceValues cache on the composite is the only cache; per-role
	// fields are additive narrowing for future read-only or write-only
	// commands. See core-git spec "Factory per-role fields are additive
	// to composite".
	RepoOpener       func() (core.RepositoryOpener, error)
	BranchReader     func() (core.BranchReader, error)
	RepositoryReader func() (core.RepositoryReader, error)
	RemoteReader     func() (core.RemoteReader, error)
	WorktreeWriter   func() (core.WorktreeWriter, error)
	BranchWriter     func() (core.BranchWriter, error)

	// Logger returns the slog channel used for TWIGGIT_DEBUG output.
	// sync.OnceFunc-cached; first call wires a text handler that
	// discards writes so debug logs do not leak into the user stream.
	Logger func() *slog.Logger

	// GlobalOptions is the backing struct for the persistent
	// --output / --quiet / --verbose flags defined on the root
	// cobra command. Shared by reference with every subcommand so
	// the bound cobra flag pointers mutate one struct rather than
	// being copied per subcommand.
	//
	// main.go populates this via NewRootCommand before cmd.Execute
	// runs; tests construct Factory literals and assign a fresh
	// pointer themselves.
	GlobalOptions *GlobalOptions
}

// WithVersion sets Factory.AppVersion. main.go passes version.Version.
func WithVersion(v string) FactoryOption {
	return func(f *Factory) { f.AppVersion = v }
}

// WithConfigLoader wires Factory.Config with the supplied constructor.
// The option wraps the constructor in sync.OnceValues so the cache
// contract (one parse per Factory lifetime) is preserved regardless of
// whether the caller pre-wraps or not.
func WithConfigLoader(load func() (*core.Config, error)) FactoryOption {
	return func(f *Factory) { f.Config = sync.OnceValues(load) }
}

// WithGitClientFactory wires Factory.GitClient with the supplied
// constructor. Same sync.OnceValues wrapping as WithConfigLoader.
func WithGitClientFactory(build func() (Client, error)) FactoryOption {
	return func(f *Factory) { f.GitClient = sync.OnceValues(build) }
}

// NewFactory returns a Factory wired with the system IOStreams and
// eager fields. Lazy fields default to sentinel-returning functions so
// the test seam (factory_test.go literals with no args) keeps compiling;
// main.go passes WithVersion, WithConfigLoader, and WithGitClientFactory
// to wire the real constructors.
//
// NewFactory never errors: all failures are deferred to the first call
// to a lazy field (or to Init). Tests that need to short-circuit a
// failure assign a replacement function before calling the field.
//
// Context is set to context.Background() by default; main.go
// replaces it with the signal.NotifyContext result so SIGINT /
// SIGTERM propagate to long-running commands. Tests pass
// t.Context() or a cancellable context directly.
func NewFactory(opts ...FactoryOption) *Factory {
	f := &Factory{
		IOStreams:  iostreams.System(),
		Context:    context.Background(),
		Executable: executableName(),
	}

	// Default lazy fields return sentinels if invoked without wiring.
	f.Config = sync.OnceValues(func() (*core.Config, error) {
		return nil, ErrNoConfigLoader
	})
	f.GitClient = sync.OnceValues(func() (Client, error) {
		return nil, ErrNoGitClient
	})

	f.RepoOpener = func() (core.RepositoryOpener, error) {
		c, err := f.GitClient()
		if err != nil {
			return nil, err
		}
		// Type assertion: f.GitClient returns Client (any); the wired
		// implementation is always *git.Client which satisfies
		// core.RepositoryOpener via embedded promotion. The assertion
		// cannot fail under the wired path; tests inject f.GitClient
		// directly so they control the asserted type.
		client, _ := c.(core.RepositoryOpener)
		return client, nil
	}
	f.BranchReader = func() (core.BranchReader, error) {
		c, err := f.GitClient()
		if err != nil {
			return nil, err
		}
		client, _ := c.(core.BranchReader)
		return client, nil
	}
	f.RepositoryReader = func() (core.RepositoryReader, error) {
		c, err := f.GitClient()
		if err != nil {
			return nil, err
		}
		client, _ := c.(core.RepositoryReader)
		return client, nil
	}
	f.RemoteReader = func() (core.RemoteReader, error) {
		c, err := f.GitClient()
		if err != nil {
			return nil, err
		}
		client, _ := c.(core.RemoteReader)
		return client, nil
	}
	f.WorktreeWriter = func() (core.WorktreeWriter, error) {
		c, err := f.GitClient()
		if err != nil {
			return nil, err
		}
		client, _ := c.(core.WorktreeWriter)
		return client, nil
	}
	f.BranchWriter = func() (core.BranchWriter, error) {
		c, err := f.GitClient()
		if err != nil {
			return nil, err
		}
		client, _ := c.(core.BranchWriter)
		return client, nil
	}

	f.Logger = sync.OnceValue(func() *slog.Logger {
		return slog.New(slog.NewTextHandler(io.Discard, &slog.HandlerOptions{Level: slog.LevelDebug}))
	})

	for _, opt := range opts {
		opt(f)
	}

	return f
}

// Init touches every lazy field once and returns the joined set of
// initialization errors. main.go calls Init before Execute so any
// broken config or git-client construction failure surfaces as a clean
// exit instead of a half-started command.
//
// Per the single-handling rule, Init collects every lazy-field error
// with errors.Join rather than returning on the first one; callers
// see the full diagnostic in one pass.
func (f *Factory) Init() error {
	var errs []error

	if _, err := f.Config(); err != nil {
		errs = append(errs, fmt.Errorf("cmdutil: config init: %w", err))
	}
	if _, err := f.GitClient(); err != nil {
		errs = append(errs, fmt.Errorf("cmdutil: git client init: %w", err))
	}
	if _, err := f.RepoOpener(); err != nil {
		errs = append(errs, fmt.Errorf("cmdutil: repo opener init: %w", err))
	}
	if _, err := f.BranchReader(); err != nil {
		errs = append(errs, fmt.Errorf("cmdutil: branch reader init: %w", err))
	}
	if _, err := f.RepositoryReader(); err != nil {
		errs = append(errs, fmt.Errorf("cmdutil: repository reader init: %w", err))
	}
	if _, err := f.RemoteReader(); err != nil {
		errs = append(errs, fmt.Errorf("cmdutil: remote reader init: %w", err))
	}
	if _, err := f.WorktreeWriter(); err != nil {
		errs = append(errs, fmt.Errorf("cmdutil: worktree writer init: %w", err))
	}
	if _, err := f.BranchWriter(); err != nil {
		errs = append(errs, fmt.Errorf("cmdutil: branch writer init: %w", err))
	}
	if f.Logger() == nil {
		errs = append(errs, errors.New("cmdutil: logger init: returned nil"))
	}

	return errors.Join(errs...)
}

// executableName returns filepath.Base(os.Args[0]) or "twiggit" when
// os.Args[0] is empty (some test runners clear it). The fallback
// guarantees AppVersion callers never see an empty Executable.
func executableName() string {
	if len(os.Args) == 0 || os.Args[0] == "" {
		return "twiggit"
	}
	return filepath.Base(os.Args[0])
}
