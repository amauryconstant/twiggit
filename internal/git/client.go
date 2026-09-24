// Package git — composite GitClient construction.
//
// Client is the canonical git I/O surface for downstream consumers.
// It composes a reader (read-side go-git operations) and a cliClient
// (write-side git CLI operations) and exposes their methods through
// embedded promotion. Callers that need only one half can use the
// half-specific constructors (e.g. NewCLIClient for tests).
package git

import (
	"fmt"

	"github.com/go-git/go-git/v5"
	lru "github.com/hashicorp/golang-lru/v2"

	"twiggit/internal/core"
)

// defaultCacheSize is the size used when WithCacheSize is omitted.
const defaultCacheSize = 25

// ClientOption configures NewClient via the functional-options pattern.
// Options are applied in the order passed; later options override earlier
// ones (so WithCacheDisabled after WithCacheSize disables caching entirely
// but keeps the size for re-enabling).
type ClientOption func(*clientConfig)

type clientConfig struct {
	cacheSize    int
	cacheEnabled bool
}

// WithCacheSize sets the LRU cache capacity to n. Panics-equivalent
// behavior: n must be > 0; invalid values are rejected silently and
// the default size is used. Callers that need strict validation should
// call with a known-positive value.
func WithCacheSize(n int) ClientOption {
	return func(c *clientConfig) {
		if n > 0 {
			c.cacheSize = n
		}
	}
}

// WithCacheDisabled turns off the LRU cache. OpenRepository still
// functions (and returns Repository handles) but each call hits the
// underlying go-git PlainOpen. Useful for tests that need to observe
// fresh repository state.
func WithCacheDisabled() ClientOption {
	return func(c *clientConfig) {
		c.cacheEnabled = false
	}
}

// Client is the composite git I/O surface. It embeds *reader and
// *cliClient so read methods (OpenRepository, ListBranches, ...) and
// write methods (CreateWorktree, DeleteWorktree, ...) are promoted to
// the top-level type. cmd/ consumes *Client directly through the
// cmdutil.Factory; no role-interface wrapper is required.
type Client struct {
	*reader
	*cliClient
}

// NewClient constructs the composite git Client with the supplied
// options. Cache defaults to enabled, size 25. Use WithCacheSize and
// WithCacheDisabled to override.
//
// NewClient also wires the cliClient's write-side executor to a
// process-backed CommandExecutor with defaultCLITimeout. Without
// this wiring the cliClient would panic on first write call (its
// executor field would be nil); the previous slice-9 cut removed
// the service-layer glue that constructed both halves separately
// and exposed this requirement on the production path.
//
// NewClient is the canonical entry point for the git package and is
// the lazy field on cmdutil.Factory.GitClient.
func NewClient(opts ...ClientOption) (*Client, error) {
	cfg := clientConfig{
		cacheSize:    defaultCacheSize,
		cacheEnabled: true,
	}
	for _, opt := range opts {
		opt(&cfg)
	}

	cache, err := defaultGoGitCacheFactory(cfg.cacheSize)
	if err != nil {
		return nil, fmt.Errorf("create git LRU cache: %w", err)
	}

	executor := NewCommandExecutor(defaultCLITimeout)

	return &Client{
		reader: &reader{
			cache:        cache,
			cacheEnabled: cfg.cacheEnabled,
		},
		cliClient: &cliClient{
			executor:       executor,
			defaultTimeout: defaultCLITimeout,
		},
	}, nil
}

// newClientWithCacheFactory is the test seam for cache-allocator failure
// coverage. Production callers should use NewClient.
func newClientWithCacheFactory(size int, enabled bool, factory goGitCacheFactory) (*Client, error) {
	cache, err := factory(size)
	if err != nil {
		return nil, err
	}

	executor := NewCommandExecutor(defaultCLITimeout)

	return &Client{
		reader: &reader{
			cache:        cache,
			cacheEnabled: enabled,
		},
		cliClient: &cliClient{
			executor:       executor,
			defaultTimeout: defaultCLITimeout,
		},
	}, nil
}

// compile-time guard so the lru import is used even if all callers go
// through the public Client surface.
var _ = lru.New[string, *git.Repository]

// compile-time role satisfaction. The composite *Client embeds *reader and
// *cliClient, so its method set is the union of both halves; these checks
// guard all 14 role methods at once against drift on the unexported concretes.
var (
	_ core.RepositoryOpener = (*Client)(nil)
	_ core.BranchReader     = (*Client)(nil)
	_ core.RepositoryReader = (*Client)(nil)
	_ core.RemoteReader     = (*Client)(nil)
	_ core.WorktreeWriter   = (*Client)(nil)
	_ core.BranchWriter     = (*Client)(nil)
)
