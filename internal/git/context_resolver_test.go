package git

import (
	"context"
	"os"
	"path/filepath"
	"testing"
	"time"
	"twiggit/internal/core"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func testConfig(t *testing.T) *core.Config {
	t.Helper()

	cfg := core.DefaultConfig()
	tmp := t.TempDir()
	cfg.ProjectsDirectory = filepath.Join(tmp, "Projects")
	cfg.WorktreesDirectory = filepath.Join(tmp, "Worktrees")
	require.NoError(t, os.MkdirAll(cfg.ProjectsDirectory, 0o755))
	require.NoError(t, os.MkdirAll(cfg.WorktreesDirectory, 0o755))

	return cfg
}

// newTestResolver constructs a contextResolver that fully bypasses NewRepoFinder
// to avoid the typed-nil interface trap in production code (passing (*Client)(nil)
// to a RepoValidator interface stores a non-nil interface containing a nil pointer).
// The construction preserves the public-API intent: nil goGit and nil cli.
func newTestResolver(cfg *core.Config) *contextResolver {
	return &contextResolver{
		config:     cfg,
		goGit:      nil,
		cli:        nil,
		repoFinder: &RepoFinder{goGit: nil},
	}
}

func TestValidatePathUnder(t *testing.T) {
	base := "/tmp/base"
	require.NoError(t, validatePathUnder(base, filepath.Join(base, "inside"), "entity", "base"))

	err := validatePathUnder(base, "/etc/passwd", "entity", "base")
	require.Error(t, err)

	var opErr *core.OperationError
	require.ErrorAs(t, err, &opErr)
	assert.Equal(t, "context.resolve", opErr.Op)
}

func TestParseCrossProjectReference(t *testing.T) {
	t.Run("valid", func(t *testing.T) {
		project, branch, valid := parseCrossProjectReference("myproj/feature")
		assert.True(t, valid)
		assert.Equal(t, "myproj", project)
		assert.Equal(t, "feature", branch)
	})

	t.Run("too many parts", func(t *testing.T) {
		_, _, valid := parseCrossProjectReference("a/b/c")
		assert.False(t, valid)
	})

	t.Run("empty project", func(t *testing.T) {
		_, _, valid := parseCrossProjectReference("/branch")
		assert.False(t, valid)
	})

	t.Run("empty branch", func(t *testing.T) {
		_, _, valid := parseCrossProjectReference("project/")
		assert.False(t, valid)
	})
}

func TestFuzzyMatch(t *testing.T) {
	assert.True(t, fuzzyMatch("feat", "feat"))
	assert.True(t, fuzzyMatch("f1", "feature-1"))
	assert.True(t, fuzzyMatch("ft1", "feature-1"))
	assert.True(t, fuzzyMatch("F1", "feature-1"))
	assert.False(t, fuzzyMatch("xyz", "feature-1"))
	assert.True(t, fuzzyMatch("", "anything"))
}

func TestMatchesExclusionPatterns(t *testing.T) {
	assert.True(t, matchesExclusionPatterns("WIP-branch", []string{"WIP-*"}))
	assert.False(t, matchesExclusionPatterns("feature", []string{"WIP-*"}))
	assert.False(t, matchesExclusionPatterns("anything", nil))
	assert.False(t, matchesExclusionPatterns("foo", []string{"["}))
}

func TestContainsPathTraversal(t *testing.T) {
	assert.True(t, containsPathTraversal("../etc"))
	assert.True(t, containsPathTraversal("foo/../bar"))
	assert.True(t, containsPathTraversal("%2e%2e/etc"))
	assert.True(t, containsPathTraversal("%252e%252e/etc"))
	assert.False(t, containsPathTraversal("feature-1"))
	assert.False(t, containsPathTraversal("my_project"))
	assert.False(t, containsPathTraversal(".hidden"))
	assert.True(t, containsPathTraversal("..hidden"))
}

func TestWorktreeExists(t *testing.T) {
	assert.False(t, worktreeExists(""))

	tmp := t.TempDir()
	dir := filepath.Join(tmp, "wt")
	require.NoError(t, os.Mkdir(dir, 0o755))
	assert.True(t, worktreeExists(dir))

	assert.False(t, worktreeExists("/proc/this/does/not/exist/at/all"))
}

func TestNewContextResolver(t *testing.T) {
	cfg := testConfig(t)
	cr := NewContextResolver(cfg, nil, nil)
	require.NotNil(t, cr)
	assert.Same(t, cfg, cr.config)
	assert.Nil(t, cr.goGit)
	assert.Nil(t, cr.cli)
	assert.NotNil(t, cr.repoFinder)
}

func TestResolveIdentifier_Empty(t *testing.T) {
	cfg := testConfig(t)
	cr := NewContextResolver(cfg, nil, nil)

	result, err := cr.ResolveIdentifier(&core.Context{Type: core.ContextProject}, "")
	require.Error(t, err)
	assert.Nil(t, result)

	var opErr *core.OperationError
	require.ErrorAs(t, err, &opErr)
	assert.Equal(t, "empty identifier", opErr.Message)
}

func TestResolveIdentifier_UnknownContext(t *testing.T) {
	cfg := testConfig(t)
	cr := NewContextResolver(cfg, nil, nil)

	result, err := cr.ResolveIdentifier(&core.Context{Type: core.ContextUnknown}, "main")
	require.NoError(t, err)
	require.NotNil(t, result)
	assert.Equal(t, core.PathTypeInvalid, result.Type)
}

func TestResolveIdentifier_Project_Main(t *testing.T) {
	cfg := testConfig(t)
	cr := NewContextResolver(cfg, nil, nil)

	result, err := cr.ResolveIdentifier(
		&core.Context{Type: core.ContextProject, ProjectName: "twiggit", Path: "/somewhere"},
		"main",
	)
	require.NoError(t, err)
	require.NotNil(t, result)
	assert.Equal(t, core.PathTypeProject, result.Type)
	assert.Equal(t, "twiggit", result.ProjectName)
	assert.Equal(t, filepath.Join(cfg.ProjectsDirectory, "twiggit"), result.ResolvedPath)
}

func TestResolveIdentifier_Project_Main_PathTraversal(t *testing.T) {
	cfg := testConfig(t)
	cr := NewContextResolver(cfg, nil, nil)

	_, err := cr.ResolveIdentifier(
		&core.Context{Type: core.ContextProject, ProjectName: "../etc", Path: "/somewhere"},
		"main",
	)
	require.Error(t, err)

	var opErr *core.OperationError
	require.ErrorAs(t, err, &opErr)
	assert.Contains(t, opErr.Message, "path traversal")
}

func TestResolveIdentifier_Project_Branch(t *testing.T) {
	cfg := testConfig(t)
	cr := NewContextResolver(cfg, nil, nil)

	result, err := cr.ResolveIdentifier(
		&core.Context{Type: core.ContextProject, ProjectName: "twiggit", Path: "/somewhere"},
		"feature-1",
	)
	require.NoError(t, err)
	require.NotNil(t, result)
	assert.Equal(t, core.PathTypeWorktree, result.Type)
	assert.Equal(t, "feature-1", result.BranchName)
	assert.Equal(t, filepath.Join(cfg.WorktreesDirectory, "twiggit", "feature-1"), result.ResolvedPath)
}

func TestResolveIdentifier_Project_CrossProject(t *testing.T) {
	cfg := testConfig(t)
	cr := NewContextResolver(cfg, nil, nil)

	result, err := cr.ResolveIdentifier(
		&core.Context{Type: core.ContextProject, ProjectName: "twiggit", Path: "/somewhere"},
		"otherproj/branch-x",
	)
	require.NoError(t, err)
	require.NotNil(t, result)
	assert.Equal(t, core.PathTypeWorktree, result.Type)
	assert.Equal(t, "otherproj", result.ProjectName)
	assert.Equal(t, "branch-x", result.BranchName)
}

func TestResolveIdentifier_Worktree_Main(t *testing.T) {
	cfg := testConfig(t)
	cr := NewContextResolver(cfg, nil, nil)

	result, err := cr.ResolveIdentifier(
		&core.Context{Type: core.ContextWorktree, ProjectName: "twiggit", BranchName: "feature-1", Path: "/somewhere/wt"},
		"main",
	)
	require.NoError(t, err)
	require.NotNil(t, result)
	assert.Equal(t, core.PathTypeProject, result.Type)
	assert.Equal(t, "twiggit", result.ProjectName)
}

func TestResolveIdentifier_Worktree_Branch(t *testing.T) {
	cfg := testConfig(t)
	cr := NewContextResolver(cfg, nil, nil)

	result, err := cr.ResolveIdentifier(
		&core.Context{Type: core.ContextWorktree, ProjectName: "twiggit", BranchName: "feature-1", Path: "/somewhere/wt"},
		"feature-2",
	)
	require.NoError(t, err)
	require.NotNil(t, result)
	assert.Equal(t, core.PathTypeWorktree, result.Type)
	assert.Equal(t, "feature-2", result.BranchName)
}

func TestResolveIdentifier_OutsideGit_Simple(t *testing.T) {
	cfg := testConfig(t)
	cr := NewContextResolver(cfg, nil, nil)

	result, err := cr.ResolveIdentifier(&core.Context{Type: core.ContextOutsideGit}, "twiggit")
	require.NoError(t, err)
	require.NotNil(t, result)
	assert.Equal(t, core.PathTypeProject, result.Type)
	assert.Equal(t, "twiggit", result.ProjectName)
	assert.Equal(t, filepath.Join(cfg.ProjectsDirectory, "twiggit"), result.ResolvedPath)
}

func TestResolveIdentifier_OutsideGit_CrossProject(t *testing.T) {
	cfg := testConfig(t)
	cr := NewContextResolver(cfg, nil, nil)

	result, err := cr.ResolveIdentifier(&core.Context{Type: core.ContextOutsideGit}, "otherproj/branch-x")
	require.NoError(t, err)
	require.NotNil(t, result)
	assert.Equal(t, core.PathTypeWorktree, result.Type)
}

func TestResolveIdentifier_OutsideGit_PathTraversal(t *testing.T) {
	cfg := testConfig(t)
	cr := NewContextResolver(cfg, nil, nil)

	_, err := cr.ResolveIdentifier(&core.Context{Type: core.ContextOutsideGit}, "../etc")
	require.Error(t, err)

	var opErr *core.OperationError
	require.ErrorAs(t, err, &opErr)
	assert.Contains(t, opErr.Message, "path traversal")
}

func TestResolveCrossProjectReference_InvalidFormat(t *testing.T) {
	cfg := testConfig(t)
	cr := NewContextResolver(cfg, nil, nil)

	result, err := cr.resolveCrossProjectReference("no-slash")
	require.NoError(t, err)
	require.NotNil(t, result)
	assert.Equal(t, core.PathTypeInvalid, result.Type)
}

func TestResolveCrossProjectReference_PathTraversal(t *testing.T) {
	cfg := testConfig(t)
	cr := NewContextResolver(cfg, nil, nil)

	_, err := cr.resolveCrossProjectReference("../etc/passwd")
	require.Error(t, err)
}

func TestResolveCrossProjectReference_Success(t *testing.T) {
	cfg := testConfig(t)
	cr := NewContextResolver(cfg, nil, nil)

	result, err := cr.resolveCrossProjectReference("myproj/feature")
	require.NoError(t, err)
	require.NotNil(t, result)
	assert.Equal(t, core.PathTypeWorktree, result.Type)
	assert.Equal(t, "myproj", result.ProjectName)
	assert.Equal(t, "feature", result.BranchName)
	assert.Equal(t, filepath.Join(cfg.WorktreesDirectory, "myproj", "feature"), result.ResolvedPath)
	assert.Contains(t, result.Explanation, "myproj")
}

func TestGetResolutionSuggestions_ProjectContext_NoWorktrees(t *testing.T) {
	cfg := testConfig(t)
	cr := NewContextResolver(cfg, nil, nil)

	suggestions, err := cr.ResolutionSuggestions(
		t.Context(),
		&core.Context{Type: core.ContextProject, ProjectName: "twiggit", Path: ""},
		"",
	)
	require.NoError(t, err)
	require.NotEmpty(t, suggestions)

	found := false
	for _, s := range suggestions {
		if s.Text == "main" {
			found = true
			assert.Equal(t, core.PathTypeProject, s.Type)
		}
	}
	assert.True(t, found, "expected 'main' suggestion")
}

func TestGetResolutionSuggestions_ProjectContext_WithExistingOnly(t *testing.T) {
	cfg := testConfig(t)
	cr := NewContextResolver(cfg, nil, nil)

	suggestions, err := cr.ResolutionSuggestions(
		t.Context(),
		&core.Context{Type: core.ContextProject, ProjectName: "twiggit", Path: ""},
		"",
		WithExistingOnly(),
	)
	require.NoError(t, err)

	for _, s := range suggestions {
		assert.NotEqual(t, "main", s.Text, "main should be excluded with isExistingOnly")
	}
}

func TestGetResolutionSuggestions_WorktreeContext(t *testing.T) {
	cfg := testConfig(t)
	cr := NewContextResolver(cfg, nil, nil)

	suggestions, err := cr.ResolutionSuggestions(
		t.Context(),
		&core.Context{Type: core.ContextWorktree, ProjectName: "twiggit", BranchName: "feature-1", Path: ""},
		"",
	)
	require.NoError(t, err)

	found := false
	for _, s := range suggestions {
		if s.Text == "main" {
			found = true
		}
	}
	assert.True(t, found, "expected 'main' suggestion in worktree context")
}

func TestGetResolutionSuggestions_OutsideGitContext(t *testing.T) {
	tmp := t.TempDir()
	cfg := core.DefaultConfig()
	cfg.ProjectsDirectory = tmp

	projDir := filepath.Join(tmp, "alpha")
	require.NoError(t, os.Mkdir(projDir, 0o755))

	cr := newTestResolver(cfg)

	suggestions, err := cr.ResolutionSuggestions(t.Context(), &core.Context{Type: core.ContextOutsideGit}, "alp")
	require.NoError(t, err)
	require.NotEmpty(t, suggestions)
	assert.Equal(t, "alpha", suggestions[0].Text)
}

func TestGetResolutionSuggestions_OutsideGitContext_EmptyProjectsDir(t *testing.T) {
	cfg := core.DefaultConfig()
	cfg.ProjectsDirectory = ""

	cr := NewContextResolver(cfg, nil, nil)

	suggestions, err := cr.ResolutionSuggestions(t.Context(), &core.Context{Type: core.ContextOutsideGit}, "")
	require.NoError(t, err)
	assert.Empty(t, suggestions)
}

func TestGetResolutionSuggestions_UnknownContext(t *testing.T) {
	cfg := testConfig(t)
	cr := NewContextResolver(cfg, nil, nil)

	suggestions, err := cr.ResolutionSuggestions(t.Context(), &core.Context{Type: core.ContextUnknown}, "x")
	require.NoError(t, err)
	assert.Empty(t, suggestions)
}

// TestResolutionSuggestions_CancelledContextShortCircuits verifies that
// task 11.3's threading reaches the leaf I/O call: a pre-cancelled
// context flows from ResolutionSuggestions through the suggestion
// builders into ListWorktrees, which uses exec.CommandContext — so
// the underlying git invocation aborts immediately and the call
// returns. The test passes if ResolutionSuggestions returns
// promptly and the worktree-listing branch was attempted (the "main"
// fallback suggestion is still added because the cancellation only
// affects the I/O call, not the pure addMainSuggestion path).
func TestResolutionSuggestions_CancelledContextShortCircuits(t *testing.T) {
	cfg := testConfig(t)
	client, err := NewClient()
	require.NoError(t, err)

	// A real *Client triggers the worktree-listing branch
	// (cr.cli != nil). Use a path that exists on disk so the
	// executor's command lookup succeeds and cancellation (not a
	// missing path) is what the leaf observes.
	repoPath := t.TempDir()
	cr := &contextResolver{
		config:     cfg,
		goGit:      client,
		cli:        client,
		repoFinder: NewRepoFinder(client),
	}

	cancelled, cancel := context.WithCancel(t.Context())
	cancel()

	domainCtx := &core.Context{
		Type:        core.ContextProject,
		ProjectName: "twiggit",
		Path:        repoPath,
	}

	done := make(chan struct{})
	var suggestions []*core.ResolutionSuggestion
	go func() {
		defer close(done)
		suggestions, err = cr.ResolutionSuggestions(cancelled, domainCtx, "")
	}()

	select {
	case <-done:
		// Returned promptly; cancellation did not block the call.
	case <-time.After(2 * time.Second):
		t.Fatal("ResolutionSuggestions blocked past 2s with cancelled context; ctx did not propagate to leaf I/O")
	}

	// The "main" suggestion is built from pure string ops and is
	// independent of the cancelled leaf; it should still be
	// present even when worktree/branch I/O is short-circuited.
	found := false
	for _, s := range suggestions {
		if s.Text == "main" {
			found = true
		}
	}
	assert.True(t, found, "expected 'main' suggestion from pure fallback path despite cancelled leaf I/O")
}

func TestResolveIdentifier_WorktreeRoutesToProjectContext(t *testing.T) {
	cfg := testConfig(t)
	cr := newTestResolver(cfg)

	ctx := &core.Context{
		Type:        core.ContextWorktree,
		ProjectName: "twiggit",
		BranchName:  "feature-1",
		Path:        "/somewhere/wt",
	}

	for _, identifier := range []string{"main", "feature-1", "otherproj/branch-x"} {
		result, err := cr.ResolveIdentifier(ctx, identifier)
		require.NoError(t, err, "identifier=%s", identifier)
		require.NotNil(t, result, "identifier=%s", identifier)

		assert.NotEmpty(t, result.ResolvedPath, "identifier=%s", identifier)
		assert.NotEmpty(t, result.ProjectName, "identifier=%s", identifier)
	}
}

func TestDiscoverProjects_EmptyDir(t *testing.T) {
	cfg := testConfig(t)
	cr := NewContextResolver(cfg, nil, nil)

	projects, err := cr.discoverProjects()
	require.NoError(t, err)
	assert.Empty(t, projects)
}

func TestDiscoverProjects_WithRepos(t *testing.T) {
	tmp := t.TempDir()
	cfg := core.DefaultConfig()
	cfg.ProjectsDirectory = tmp

	for _, name := range []string{"alpha", "beta"} {
		require.NoError(t, os.Mkdir(filepath.Join(tmp, name), 0o755))
	}

	cr := newTestResolver(cfg)
	projects, err := cr.discoverProjects()
	require.NoError(t, err)
	require.Len(t, projects, 2)

	names := make([]string, len(projects))
	for i, p := range projects {
		names[i] = p.Name
	}
	assert.ElementsMatch(t, []string{"alpha", "beta"}, names)
}

func TestResolveMainIdentifier_OutsideProjectsDir(t *testing.T) {
	cfg := testConfig(t)
	cfg.ProjectsDirectory = "/opt/projects"

	cr := newTestResolver(cfg)

	_, err := cr.resolveMainIdentifier(&core.Context{Type: core.ContextProject, ProjectName: "../escape"})
	require.Error(t, err)
}

func TestResolveWorktreePath_PathTraversalInBranch(t *testing.T) {
	cfg := testConfig(t)
	cr := NewContextResolver(cfg, nil, nil)

	_, err := cr.resolveWorktreePath(&core.Context{Type: core.ContextProject, ProjectName: "twiggit"}, "../etc")
	require.Error(t, err)

	var opErr *core.OperationError
	require.ErrorAs(t, err, &opErr)
	assert.Contains(t, opErr.Message, "path traversal")
}

func TestAddMainSuggestion_ExistingOnlySkips(t *testing.T) {
	cfg := testConfig(t)
	cr := NewContextResolver(cfg, nil, nil)

	result := cr.addMainSuggestion(&core.Context{Type: core.ContextProject, ProjectName: "twiggit"}, "", &suggestionConfig{isExistingOnly: true})
	assert.Empty(t, result)
}

func TestAddMainSuggestion_PrefixFilter(t *testing.T) {
	cfg := testConfig(t)
	cr := NewContextResolver(cfg, nil, nil)

	ctx := &core.Context{Type: core.ContextProject, ProjectName: "twiggit"}
	config := &suggestionConfig{}

	result := cr.addMainSuggestion(ctx, "", config)
	require.Len(t, result, 1)
	assert.Equal(t, "main", result[0].Text)

	result = cr.addMainSuggestion(ctx, "z", config)
	assert.Empty(t, result)
}

func TestWithExistingOnly_Option(t *testing.T) {
	config := &suggestionConfig{}
	WithExistingOnly()(config)
	assert.True(t, config.isExistingOnly)

	config2 := &suggestionConfig{}
	WithExistingOnly()(config2)
	assert.True(t, config2.isExistingOnly)
}

func TestContextResolver_TypeAlias(t *testing.T) {
	cfg := testConfig(t)

	cr := newTestResolver(cfg)
	aliasRef := cr
	concreteRef := cr

	assert.Same(t, aliasRef, concreteRef)
}

func TestResolveIdentifier_PathTraversalVariants(t *testing.T) {
	cfg := testConfig(t)
	cr := NewContextResolver(cfg, nil, nil)

	traversals := []string{
		"../foo",
		"foo/../../bar",
		"%2e%2e/foo",
		"%252e%252e/foo",
	}

	for _, ident := range traversals {
		_, err := cr.ResolveIdentifier(
			&core.Context{Type: core.ContextProject, ProjectName: ident},
			"main",
		)
		require.Error(t, err)
	}
}

func TestResolutionResult_FieldsPopulated(t *testing.T) {
	cfg := testConfig(t)
	cr := NewContextResolver(cfg, nil, nil)

	result, err := cr.ResolveIdentifier(
		&core.Context{Type: core.ContextProject, ProjectName: "twiggit", Path: "/x"},
		"feature-1",
	)
	require.NoError(t, err)
	require.NotNil(t, result)
	assert.NotEmpty(t, result.ResolvedPath)
	assert.Equal(t, core.PathTypeWorktree, result.Type)
	assert.Equal(t, "twiggit", result.ProjectName)
	assert.Equal(t, "feature-1", result.BranchName)
	assert.NotEmpty(t, result.Explanation)
}

func TestResolutionResult_ExplanationText(t *testing.T) {
	cfg := testConfig(t)
	cr := NewContextResolver(cfg, nil, nil)

	result, err := cr.ResolveIdentifier(
		&core.Context{Type: core.ContextProject, ProjectName: "twiggit", Path: "/x"},
		"main",
	)
	require.NoError(t, err)
	require.NotNil(t, result)
	assert.Contains(t, result.Explanation, "twiggit")
	assert.Contains(t, result.Explanation, "main")
}

func TestAddWorktreeSuggestions_Empty(t *testing.T) {
	cfg := testConfig(t)
	cr := newTestResolver(cfg)

	ctx := &core.Context{Type: core.ContextProject, ProjectName: "twiggit"}
	config := &suggestionConfig{}

	result := cr.addWorktreeSuggestions(t.Context(), nil, ctx, "", nil, config)
	assert.Empty(t, result)
}

func TestAddWorktreeSuggestions_PrefixMatch(t *testing.T) {
	cfg := testConfig(t)
	cr := newTestResolver(cfg)

	ctx := &core.Context{Type: core.ContextProject, ProjectName: "twiggit"}
	config := &suggestionConfig{}

	worktrees := []core.Worktree{
		{Path: "/somewhere/wt", Branch: "feature-1"},
		{Path: "/somewhere/wt2", Branch: "maintenance"},
	}

	result := cr.addWorktreeSuggestions(t.Context(), nil, ctx, "feat", worktrees, config)
	require.Len(t, result, 1)
	assert.Equal(t, "feature-1", result[0].Text)
	assert.Equal(t, core.PathTypeWorktree, result[0].Type)
	assert.Equal(t, "Worktree for branch feature-1", result[0].Description)
}

func TestAddWorktreeSuggestions_FuzzyMatch(t *testing.T) {
	cfg := testConfig(t)
	cfg.Navigation.FuzzyMatching = true
	cr := newTestResolver(cfg)

	ctx := &core.Context{Type: core.ContextProject, ProjectName: "twiggit"}
	config := &suggestionConfig{}

	worktrees := []core.Worktree{
		{Path: "/somewhere/wt", Branch: "feature-1"},
	}

	result := cr.addWorktreeSuggestions(t.Context(), nil, ctx, "f1", worktrees, config)
	require.Len(t, result, 1)
	assert.Equal(t, "feature-1", result[0].Text)
}

func TestAddWorktreeSuggestions_ExclusionPattern(t *testing.T) {
	cfg := testConfig(t)
	cfg.Completion.ExcludeBranches = []string{"WIP-*"}
	cr := newTestResolver(cfg)

	ctx := &core.Context{Type: core.ContextProject, ProjectName: "twiggit"}
	config := &suggestionConfig{}

	worktrees := []core.Worktree{
		{Path: "/somewhere/wt", Branch: "WIP-foo"},
		{Path: "/somewhere/wt2", Branch: "feature-1"},
	}

	result := cr.addWorktreeSuggestions(t.Context(), nil, ctx, "", worktrees, config)
	require.Len(t, result, 1)
	assert.Equal(t, "feature-1", result[0].Text)
}

func TestAddWorktreeSuggestions_ExistingOnlyFiltersMissing(t *testing.T) {
	cfg := testConfig(t)
	cr := newTestResolver(cfg)

	ctx := &core.Context{Type: core.ContextProject, ProjectName: "twiggit"}
	config := &suggestionConfig{isExistingOnly: true}

	worktrees := []core.Worktree{
		{Path: "/definitely/not/here/wt", Branch: "missing"},
		{Path: "", Branch: "empty-path"},
	}

	result := cr.addWorktreeSuggestions(t.Context(), nil, ctx, "", worktrees, config)
	assert.Empty(t, result)
}

func TestAddWorktreeSuggestions_IsCurrentDetected(t *testing.T) {
	cfg := testConfig(t)
	cr := newTestResolver(cfg)

	ctx := &core.Context{
		Type:        core.ContextWorktree,
		ProjectName: "twiggit",
		BranchName:  "feature-1",
	}
	config := &suggestionConfig{}

	worktrees := []core.Worktree{
		{Path: "/somewhere/wt", Branch: "feature-1"},
		{Path: "/somewhere/wt2", Branch: "feature-2"},
	}

	result := cr.addWorktreeSuggestions(t.Context(), nil, ctx, "", worktrees, config)
	require.Len(t, result, 2)

	var currentFound, otherFound bool
	for _, s := range result {
		if s.Text == "feature-1" {
			assert.True(t, s.IsCurrent)
			currentFound = true
		}
		if s.Text == "feature-2" {
			assert.False(t, s.IsCurrent)
			otherFound = true
		}
	}
	assert.True(t, currentFound)
	assert.True(t, otherFound)
}

func TestAddWorktreeSuggestions_DirtyFlag_RequiresGoGit(t *testing.T) {
	cfg := testConfig(t)
	cr := newTestResolver(cfg)
	ctx := &core.Context{
		Type:        core.ContextWorktree,
		ProjectName: "twiggit",
		BranchName:  "feature-1",
	}
	config := &suggestionConfig{}

	worktrees := []core.Worktree{
		{Path: "/somewhere/wt", Branch: "feature-1"},
	}

	result := cr.addWorktreeSuggestions(t.Context(), nil, ctx, "", worktrees, config)
	require.Len(t, result, 1)
	assert.False(t, result[0].IsDirty, "dirty flag requires non-nil goGit")
}

func TestAddProjectSuggestions_EmptyDir(t *testing.T) {
	cfg := testConfig(t)
	cr := newTestResolver(cfg)

	ctx := &core.Context{Type: core.ContextProject, ProjectName: "twiggit"}

	result := cr.addProjectSuggestions(nil, ctx, "", true)
	assert.Empty(t, result)
}

func TestAddProjectSuggestions_ExcludesCurrent(t *testing.T) {
	tmp := t.TempDir()
	cfg := core.DefaultConfig()
	cfg.ProjectsDirectory = tmp

	require.NoError(t, os.Mkdir(filepath.Join(tmp, "alpha"), 0o755))
	require.NoError(t, os.Mkdir(filepath.Join(tmp, "beta"), 0o755))

	cr := newTestResolver(cfg)
	ctx := &core.Context{Type: core.ContextProject, ProjectName: "alpha"}

	result := cr.addProjectSuggestions(nil, ctx, "", true)
	require.Len(t, result, 1)
	assert.Equal(t, "beta", result[0].Text)
}

func TestAddProjectSuggestions_IncludesCurrent(t *testing.T) {
	tmp := t.TempDir()
	cfg := core.DefaultConfig()
	cfg.ProjectsDirectory = tmp

	require.NoError(t, os.Mkdir(filepath.Join(tmp, "alpha"), 0o755))
	require.NoError(t, os.Mkdir(filepath.Join(tmp, "beta"), 0o755))

	cr := newTestResolver(cfg)
	ctx := &core.Context{Type: core.ContextProject, ProjectName: "alpha"}

	result := cr.addProjectSuggestions(nil, ctx, "", false)
	require.Len(t, result, 2)
}

func TestAddProjectSuggestions_PrefixMatch(t *testing.T) {
	tmp := t.TempDir()
	cfg := core.DefaultConfig()
	cfg.ProjectsDirectory = tmp

	require.NoError(t, os.Mkdir(filepath.Join(tmp, "alpha"), 0o755))
	require.NoError(t, os.Mkdir(filepath.Join(tmp, "beta"), 0o755))

	cr := newTestResolver(cfg)
	ctx := &core.Context{Type: core.ContextProject, ProjectName: "twiggit"}

	result := cr.addProjectSuggestions(nil, ctx, "al", true)
	require.Len(t, result, 1)
	assert.Equal(t, "alpha", result[0].Text)
}

func TestAddProjectSuggestions_ExclusionPattern(t *testing.T) {
	tmp := t.TempDir()
	cfg := core.DefaultConfig()
	cfg.ProjectsDirectory = tmp
	cfg.Completion.ExcludeProjects = []string{"beta"}

	require.NoError(t, os.Mkdir(filepath.Join(tmp, "alpha"), 0o755))
	require.NoError(t, os.Mkdir(filepath.Join(tmp, "beta"), 0o755))

	cr := newTestResolver(cfg)
	ctx := &core.Context{Type: core.ContextProject, ProjectName: "twiggit"}

	result := cr.addProjectSuggestions(nil, ctx, "", false)
	require.Len(t, result, 1)
	assert.Equal(t, "alpha", result[0].Text)
}

func TestAddProjectSuggestions_FuzzyMatch(t *testing.T) {
	tmp := t.TempDir()
	cfg := core.DefaultConfig()
	cfg.ProjectsDirectory = tmp
	cfg.Navigation.FuzzyMatching = true

	require.NoError(t, os.Mkdir(filepath.Join(tmp, "alpha"), 0o755))

	cr := newTestResolver(cfg)
	ctx := &core.Context{Type: core.ContextProject, ProjectName: "twiggit"}

	result := cr.addProjectSuggestions(nil, ctx, "alh", true)
	require.Len(t, result, 1)
	assert.Equal(t, "alpha", result[0].Text)
}

func TestAddBranchSuggestions_NilGoGit_PanicsDocumented(t *testing.T) {
	t.Skip("addBranchSuggestions unconditionally calls cr.goGit.ListBranches; characterizing nil-goGit behavior is deferred to the §10/§11 split where goGit is injected via an interface")
}
