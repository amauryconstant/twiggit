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

	"twiggit/internal/config"
	"twiggit/internal/core"
	"twiggit/internal/git"
	"twiggit/internal/iostreams"
	"twiggit/internal/version"
)

// Factory is the composition seam shared by every cmd/*.go. Each lazy
// field is a function so callers invoke it at first use rather than
// constructing every dependency eagerly. The Config field is the
// canonical sync.OnceValue cache: every other field that depends on
// config (GitClient, Logger) reaches it through f.Config(), so the
// config file is parsed exactly once per binary invocation.
//
// Construction populates IOStreams, Context, AppVersion, and
// Executable eagerly (they are cheap and never fail). The expensive
// fields (Config, GitClient, Logger) are sync.OnceValue /
// sync.OnceFunc wrappers that run on first call. Init touches each
// lazy field so initialization failures surface before any command
// body executes.
//
// Customizing for tests: assign new functions to the fields before the
// first call (or before Init). Once cached, the result is pinned for
// the Factory's lifetime.
type Factory struct {
	// IOStreams is the terminal I/O surface. Populated by NewFactory
	// from iostreams.System(); tests may swap in iostreams.Test().
	IOStreams *iostreams.IOStreams

	// Context is the application context. main.go sets it to the
	// signal.NotifyContext result so SIGINT / SIGTERM cancel every
	// long-running command; tests pass t.Context(). Subcommand RunE
	// closures read this through opts.Ctx.
	Context context.Context

	// AppVersion is the build-time injected version string.
	AppVersion string

	// Executable is the basename of the running binary (twiggit).
	Executable string

	// Config returns the loaded config. sync.OnceValue-cached; the
	// first call reads and parses the config file, every later call
	// returns the same *core.Config pointer. A load error pins for
	// the Factory's lifetime (subsequent calls return the same error).
	Config func() (*core.Config, error)

	// GitClient returns the composite git client. sync.OnceValue-cached;
	// the closure first reads f.Config() so config load errors surface
	// here rather than as a confused git-construction failure.
	GitClient func() (*git.Client, error)

	// Per-role lazy fields. Each returns the cached *git.Client typed as
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

// NewFactory returns a Factory wired with the system IOStreams, the
// build-time version, and lazy fields for the three expensive
// dependencies. Every lazy field is wrapped in sync.OnceValue /
// sync.OnceFunc so the underlying work runs at most once per Factory.
//
// NewFactory never errors: all failures are deferred to the first call
// to a lazy field (or to Init). Tests that need to short-circuit a
// failure assign a replacement function before calling the field.
//
// Context is set to context.Background() by default; main.go
// replaces it with the signal.NotifyContext result so SIGINT /
// SIGTERM propagate to long-running commands. Tests pass
// t.Context() or a cancellable context directly.
func NewFactory() *Factory {
	f := &Factory{
		IOStreams:  iostreams.System(),
		Context:    context.Background(),
		AppVersion: version.Version,
		Executable: executableName(),
	}

	f.Config = sync.OnceValues(func() (*core.Config, error) {
		manager := config.NewManager()
		return manager.Load()
	})

	f.GitClient = sync.OnceValues(func() (*git.Client, error) {
		// Touch Config so config load errors surface here too.
		if _, err := f.Config(); err != nil {
			return nil, err
		}
		return git.NewClient()
	})

	f.RepoOpener = func() (core.RepositoryOpener, error) {
		c, err := f.GitClient()
		if err != nil {
			return nil, err
		}
		return c, nil
	}
	f.BranchReader = func() (core.BranchReader, error) {
		c, err := f.GitClient()
		if err != nil {
			return nil, err
		}
		return c, nil
	}
	f.RepositoryReader = func() (core.RepositoryReader, error) {
		c, err := f.GitClient()
		if err != nil {
			return nil, err
		}
		return c, nil
	}
	f.RemoteReader = func() (core.RemoteReader, error) {
		c, err := f.GitClient()
		if err != nil {
			return nil, err
		}
		return c, nil
	}
	f.WorktreeWriter = func() (core.WorktreeWriter, error) {
		c, err := f.GitClient()
		if err != nil {
			return nil, err
		}
		return c, nil
	}
	f.BranchWriter = func() (core.BranchWriter, error) {
		c, err := f.GitClient()
		if err != nil {
			return nil, err
		}
		return c, nil
	}

	f.Logger = sync.OnceValue(func() *slog.Logger {
		return slog.New(slog.NewTextHandler(io.Discard, &slog.HandlerOptions{Level: slog.LevelDebug}))
	})

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
