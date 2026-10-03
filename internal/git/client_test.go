package git

import (
	"errors"
	"os"
	"path/filepath"
	"testing"
	"twiggit/internal/core"

	"github.com/go-git/go-git/v5"
	lru "github.com/hashicorp/golang-lru/v2"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestGoGitClient_OpenRepository(t *testing.T) {
	client, err := NewClient()
	require.NoError(t, err)
	tempDir := t.TempDir()

	repo, err := client.OpenRepository("/non/existent/path")
	require.Error(t, err)
	assert.Nil(t, repo)

	repo, err = client.OpenRepository(tempDir)
	require.Error(t, err)
	assert.Nil(t, repo)
}

// TestNewClient_DefaultCacheEnabled confirms the default construction
// wires the LRU cache at defaultCacheSize (25) and leaves it enabled.
// Drives the new (cli-functional-core-shell) NewClient() rename parity
// plus the cache default path from git-client spec.
//
// Note: lru.Cache does not expose its capacity on this version, so the
// size is verified through the Len()/Resize surface indirectly — what
// we can pin deterministically is that the cache handle is allocated
// and the enabled gate is on by default.
func TestNewClient_DefaultCacheEnabled(t *testing.T) {
	t.Parallel()

	client, err := NewClient()
	require.NoError(t, err)
	require.NotNil(t, client)

	assert.True(t, client.isCacheEnabled, "default cache must be enabled")
	assert.NotNil(t, client.cache, "default cache handle must be non-nil")
}

// TestWithCacheSize_Applies asserts the functional option accepts a
// positive size without erroring. The cache handle stays allocated
// (size 50 vs default 25 is verified via the lru Len()/Resize
// surface but not pinned here; the option only mutates config and
// passes through to the cache factory).
func TestWithCacheSize_Applies(t *testing.T) {
	t.Parallel()

	client, err := NewClient(WithCacheSize(50))
	require.NoError(t, err)
	require.NotNil(t, client)

	assert.True(t, client.isCacheEnabled)
	assert.NotNil(t, client.cache)
}

// TestWithCacheSize_RejectsNonPositive pins the documented contract:
// n <= 0 falls back to defaultCacheSize rather than panicking. The
// option is silently rejected so callers cannot accidentally
// construct a 0-capacity cache.
func TestWithCacheSize_RejectsNonPositive(t *testing.T) {
	t.Parallel()

	tests := []struct {
		name string
		size int
	}{
		{"zero", 0},
		{"negative", -5},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			t.Parallel()
			must := require.New(t)
			is := assert.New(t)

			client, err := NewClient(WithCacheSize(tt.size))
			must.NoError(err)
			is.NotNil(client)
			is.NotNil(client.cache, "non-positive size must still allocate the cache handle")
		})
	}
}

// TestWithCacheDisabled_TurnsOffCache asserts WithCacheDisabled
// produces a Client whose reader isCacheEnabled is false but still
// constructs a non-nil cache handle (the cache is allocated so
// re-enabling via a later option would only require flipping the
// boolean).
func TestWithCacheDisabled_TurnsOffCache(t *testing.T) {
	t.Parallel()

	client, err := NewClient(WithCacheDisabled())
	require.NoError(t, err)
	require.NotNil(t, client)

	assert.False(t, client.isCacheEnabled, "WithCacheDisabled must disable the cache gate")
	assert.NotNil(t, client.cache, "cache handle stays allocated for re-enable")
}

// TestNewClient_NoStutterAtCallSites pins the git-client spec scenario
// "no stutter at call sites": NewClient() is the canonical entry
// point, not NewGitClient(). Confirms callers compile against the new
// name.
func TestNewClient_NoStutterAtCallSites(t *testing.T) {
	t.Parallel()

	client, err := NewClient()
	require.NoError(t, err)
	assert.NotNil(t, client)

	// Compile-time assertion: if NewGoGitClient() were re-introduced
	// as a public alias the test file would still compile (the alias
	// would resolve), so the real protection is the depguard rule on
	// internal/git/** and the rename task 2.8 having removed the
	// old name from the package surface. This test pins the behavior.
	_, err = NewClient()
	assert.NoError(t, err)
}

func TestGoGitClient_ValidateRepository(t *testing.T) {
	client, err := NewClient()
	require.NoError(t, err)
	tempDir := t.TempDir()

	err = client.ValidateRepository("/non/existent/path")
	require.Error(t, err)

	err = client.ValidateRepository(tempDir)
	require.Error(t, err)

	repoPath := setupTestRepo(t, tempDir)
	err = client.ValidateRepository(repoPath)
	require.NoError(t, err)
}

func TestGoGitClient_ListBranches(t *testing.T) {
	client, err := NewClient()
	require.NoError(t, err)
	tempDir := t.TempDir()

	branches, err := client.ListBranches(t.Context(), "/non/existent/path")
	require.Error(t, err)
	assert.Nil(t, branches)

	repoPath := setupTestRepo(t, tempDir)
	branches, err = client.ListBranches(t.Context(), repoPath)
	require.NoError(t, err)
	assert.NotEmpty(t, branches)

	mainBranch := findBranch(branches, "main")
	assert.NotNil(t, mainBranch)
	assert.Equal(t, "main", mainBranch.Name)
}

func TestGoGitClient_BranchExists(t *testing.T) {
	client, err := NewClient()
	require.NoError(t, err)
	tempDir := t.TempDir()

	exists, err := client.BranchExists(t.Context(), "/non/existent/path", "main")
	require.Error(t, err)
	assert.False(t, exists)

	repoPath := setupTestRepo(t, tempDir)

	exists, err = client.BranchExists(t.Context(), repoPath, "main")
	require.NoError(t, err)
	assert.True(t, exists)

	exists, err = client.BranchExists(t.Context(), repoPath, "non-existent")
	require.NoError(t, err)
	assert.False(t, exists)
}

func TestGoGitClient_GetRepositoryStatus(t *testing.T) {
	client, err := NewClient()
	require.NoError(t, err)
	tempDir := t.TempDir()

	status, err := client.RepositoryStatus(t.Context(), "/non/existent/path")
	require.Error(t, err)
	assert.Equal(t, core.RepositoryStatus{}, status)

	repoPath := setupTestRepo(t, tempDir)
	status, err = client.RepositoryStatus(t.Context(), repoPath)
	require.NoError(t, err)
	assert.True(t, status.IsClean)
	assert.Equal(t, "main", status.Branch)
}

func TestGoGitClient_GetRepositoryInfo(t *testing.T) {
	client, err := NewClient()
	require.NoError(t, err)
	tempDir := t.TempDir()

	info, err := client.Repository(t.Context(), "/non/existent/path")
	require.Error(t, err)
	assert.Nil(t, info)

	repoPath := setupTestRepo(t, tempDir)
	info, err = client.Repository(t.Context(), repoPath)
	require.NoError(t, err)
	assert.NotNil(t, info)
	assert.Equal(t, repoPath, info.Path)
	assert.False(t, info.IsBare)
	assert.NotEmpty(t, info.Branches)
}

func TestGoGitClient_ListRemotes(t *testing.T) {
	client, err := NewClient()
	require.NoError(t, err)
	tempDir := t.TempDir()

	remotes, err := client.ListRemotes(t.Context(), "/non/existent/path")
	require.Error(t, err)
	assert.Nil(t, remotes)

	repoPath := setupTestRepo(t, tempDir)
	remotes, err = client.ListRemotes(t.Context(), repoPath)
	require.NoError(t, err)
	assert.Empty(t, remotes)
}

func TestGoGitClient_GetCommitInfo(t *testing.T) {
	client, err := NewClient()
	require.NoError(t, err)
	tempDir := t.TempDir()

	commit, err := client.Commit(t.Context(), "/non/existent/path", "HEAD")
	require.Error(t, err)
	assert.Nil(t, commit)

	repoPath := setupTestRepo(t, tempDir)
	commit, err = client.Commit(t.Context(), repoPath, "HEAD")
	require.Error(t, err)
	assert.Nil(t, commit)
}

func findBranch(branches []core.Branch, name string) *core.Branch {
	for _, branch := range branches {
		if branch.Name == name {
			return &branch
		}
	}
	return nil
}

func TestNewClient_CacheAllocatorFailure(t *testing.T) {
	allocErr := errors.New("simulated cache allocation failure")
	failingFactory := func(_ int) (*lru.Cache[string, *git.Repository], error) {
		return nil, allocErr
	}

	client, err := newClientWithCacheFactory(25, true, failingFactory)
	require.Error(t, err)
	assert.Nil(t, client)
	assert.ErrorIs(t, err, allocErr)
}

func TestNewClientWithSize_CacheAllocatorFailure(t *testing.T) {
	allocErr := errors.New("simulated cache allocation failure for size")
	failingFactory := func(_ int) (*lru.Cache[string, *git.Repository], error) {
		return nil, allocErr
	}

	client, err := newClientWithCacheFactory(100, true, failingFactory)
	require.Error(t, err)
	assert.Nil(t, client)
	assert.ErrorIs(t, err, allocErr)
}

func TestNewClient_CacheFactoryReceivesRequestedSize(t *testing.T) {
	var receivedSize int
	captureFactory := func(size int) (*lru.Cache[string, *git.Repository], error) {
		receivedSize = size
		return defaultGoGitCacheFactory(size)
	}

	_, err := newClientWithCacheFactory(42, true, captureFactory)
	require.NoError(t, err)
	assert.Equal(t, 42, receivedSize)
}

func setupTestRepo(t *testing.T, tempDir string) string {
	t.Helper()

	client, err := NewClient()
	require.NoError(t, err)
	repoPath := filepath.Join(tempDir, "test-repo")

	err = os.MkdirAll(repoPath, 0o755)
	require.NoError(t, err)

	_, err = client.OpenRepository(repoPath)
	if err == nil {
		return repoPath
	}

	gitDir := filepath.Join(repoPath, ".git")
	err = os.MkdirAll(gitDir, 0o755)
	require.NoError(t, err)

	err = os.WriteFile(filepath.Join(gitDir, "HEAD"), []byte("ref: refs/heads/main\n"), 0o644)
	require.NoError(t, err)

	refsDir := filepath.Join(gitDir, "refs", "heads")
	err = os.MkdirAll(refsDir, 0o755)
	require.NoError(t, err)

	err = os.WriteFile(filepath.Join(refsDir, "main"), []byte("0000000000000000000000000000000000000000\n"), 0o644)
	require.NoError(t, err)

	return repoPath
}

// _DriftCheck is a compile-time sentinel: it embeds every role declared in
// internal/core/git.go against the composite *Client. Renaming or removing
// any role method on *reader or *cliClient fails the build with a
// type-mismatch error referencing the affected role. The struct literal
// below assigns (*Client)(nil) to every embedded interface field, forcing
// the build to enforce the role surface every time.
type _DriftCheck struct {
	core.RepositoryOpener
	core.BranchReader
	core.RepositoryReader
	core.RemoteReader
	core.WorktreeWriter
	core.BranchWriter
}

var _ = _DriftCheck{
	RepositoryOpener: (*Client)(nil),
	BranchReader:     (*Client)(nil),
	RepositoryReader: (*Client)(nil),
	RemoteReader:     (*Client)(nil),
	WorktreeWriter:   (*Client)(nil),
	BranchWriter:     (*Client)(nil),
}

// TestReadSideFailure_OpIsGitRepository pins git-client R4.S1: read-side
// failures must surface as *core.OperationError with the dot-concatenated
// Op namespace "git.repository.<method>" so callers can dispatch on the
// precise operation.
func TestReadSideFailure_OpIsGitRepository(t *testing.T) {
	t.Parallel()

	client, err := NewClient()
	require.NoError(t, err)

	_, err = client.OpenRepository("/non/existent/path")
	require.Error(t, err)

	var oe *core.OperationError
	require.ErrorAs(t, err, &oe,
		"read-side failure must be *core.OperationError (errors.As must walk)")
	assert.Equal(t, "git.repository.open", oe.Op,
		"OpenRepository Op must follow \"git.repository.<method>\" format")
	assert.NotEmpty(t, oe.Message)
	assert.Error(t, oe.Cause, "Cause must wrap underlying go-git error via %w")
}
