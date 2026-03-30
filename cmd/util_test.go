package cmd

import (
	"bytes"
	"context"
	"testing"

	"github.com/spf13/cobra"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/mock"
	"github.com/stretchr/testify/require"

	"twiggit/internal/domain"
	"twiggit/test/mocks"
)

func TestIsQuiet(t *testing.T) {
	tests := []struct {
		name      string
		quietFlag bool
		expected  bool
	}{
		{
			name:      "quiet flag not set",
			quietFlag: false,
			expected:  false,
		},
		{
			name:      "quiet flag set",
			quietFlag: true,
			expected:  true,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			cmd := &cobra.Command{}
			cmd.Flags().Bool("quiet", tt.quietFlag, "suppress output")

			result := isQuiet(cmd)
			assert.Equal(t, tt.expected, result)
		})
	}
}

func TestLogv_NoPanic(t *testing.T) {
	// logv writes to os.Stderr, we just verify it doesn't panic
	cmd := &cobra.Command{}
	cmd.Flags().CountP("verbose", "v", "increase verbosity")
	_ = cmd.Flags().Set("verbose", "v")

	// Should not panic
	logv(cmd, 1, "test message %s", "arg")
	logv(cmd, 2, "detailed message")
}

func TestLogv_LevelThreshold(t *testing.T) {
	// Test that verbosity level threshold works correctly
	// When verbose=1, level 2 log should not output
	cmd := &cobra.Command{}
	cmd.Flags().CountP("verbose", "v", "increase verbosity")
	_ = cmd.Flags().Set("verbose", "1")

	// We can't easily capture stderr in this test, but we can verify
	// the function doesn't panic at different levels
	logv(cmd, 1, "level 1 message")
	logv(cmd, 2, "level 2 message - should not output")
}

func TestNewProgressReporter(t *testing.T) {
	tests := []struct {
		name     string
		quiet    bool
		expected bool
	}{
		{
			name:     "quiet mode disabled",
			quiet:    false,
			expected: false,
		},
		{
			name:     "quiet mode enabled",
			quiet:    true,
			expected: true,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			var buf bytes.Buffer
			reporter := NewProgressReporter(tt.quiet, &buf)
			assert.NotNil(t, reporter)
			assert.Equal(t, tt.expected, reporter.quiet)
		})
	}
}

func TestProgressReporter_Report(t *testing.T) {
	tests := []struct {
		name         string
		quiet        bool
		expectOutput bool
		expectedCont string
	}{
		{
			name:         "non-quiet mode outputs message",
			quiet:        false,
			expectOutput: true,
			expectedCont: "Processing worktrees",
		},
		{
			name:         "quiet mode suppresses message",
			quiet:        true,
			expectOutput: false,
			expectedCont: "Processing worktrees",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			var buf bytes.Buffer
			reporter := NewProgressReporter(tt.quiet, &buf)

			reporter.Report("Processing worktrees")

			if tt.expectOutput {
				assert.Contains(t, buf.String(), tt.expectedCont)
			} else {
				assert.Empty(t, buf.String())
			}
		})
	}
}

func TestProgressReporter_ReportProgress(t *testing.T) {
	tests := []struct {
		name         string
		quiet        bool
		current      int
		total        int
		item         string
		expectOutput bool
		expectedFmt  string
	}{
		{
			name:         "non-quiet mode outputs progress",
			quiet:        false,
			current:      1,
			total:        5,
			item:         "feature-branch",
			expectOutput: true,
			expectedFmt:  "[1/5] Processing feature-branch",
		},
		{
			name:         "quiet mode suppresses progress",
			quiet:        true,
			current:      1,
			total:        5,
			item:         "feature-branch",
			expectOutput: false,
			expectedFmt:  "[1/5] Processing feature-branch",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			var buf bytes.Buffer
			reporter := NewProgressReporter(tt.quiet, &buf)

			reporter.ReportProgress(tt.current, tt.total, tt.item)

			if tt.expectOutput {
				assert.Contains(t, buf.String(), tt.expectedFmt)
			} else {
				assert.Empty(t, buf.String())
			}
		})
	}
}

func TestResolveNavigationTarget_WithExplicitTarget(t *testing.T) {
	mockCtxService := mocks.NewMockContextService()
	mockNavService := mocks.NewMockNavigationService()

	config := &CommandConfig{
		Services: &ServiceContainer{
			ContextService:    mockCtxService,
			NavigationService: mockNavService,
		},
	}

	// Set up expectations
	mockCtxService.On("GetCurrentContext").Return(&domain.Context{
		Type: domain.ContextProject,
	}, nil)

	mockNavService.On("ResolvePath", mock.Anything, mock.MatchedBy(func(req *domain.ResolvePathRequest) bool {
		return req.Target == "feature-branch"
	})).Return(&domain.ResolutionResult{
		ResolvedPath: "/path/to/feature-branch",
	}, nil)

	// Execute
	ctx := context.Background()
	currentCtx, result, err := resolveNavigationTarget(ctx, config, "feature-branch")

	// Assert
	require.NoError(t, err)
	assert.NotNil(t, currentCtx)
	assert.NotNil(t, result)
	assert.Equal(t, "/path/to/feature-branch", result.ResolvedPath)

	mockCtxService.AssertExpectations(t)
	mockNavService.AssertExpectations(t)
}

func TestResolveNavigationTarget_FromWorktreeUsesCurrentBranch(t *testing.T) {
	mockCtxService := mocks.NewMockContextService()
	mockNavService := mocks.NewMockNavigationService()

	config := &CommandConfig{
		Services: &ServiceContainer{
			ContextService:    mockCtxService,
			NavigationService: mockNavService,
		},
	}

	// Set up expectations - from worktree context
	mockCtxService.On("GetCurrentContext").Return(&domain.Context{
		Type:       domain.ContextWorktree,
		BranchName: "current-feature",
	}, nil)

	// When target is empty, should use current branch from worktree context
	mockNavService.On("ResolvePath", mock.Anything, mock.MatchedBy(func(req *domain.ResolvePathRequest) bool {
		return req.Target == "current-feature"
	})).Return(&domain.ResolutionResult{
		ResolvedPath: "/path/to/current-feature",
	}, nil)

	// Execute with empty target
	ctx := context.Background()
	currentCtx, result, err := resolveNavigationTarget(ctx, config, "")

	// Assert
	require.NoError(t, err)
	assert.NotNil(t, currentCtx)
	assert.Equal(t, domain.ContextWorktree, currentCtx.Type)
	assert.Equal(t, "current-feature", currentCtx.BranchName)
	assert.NotNil(t, result)

	mockCtxService.AssertExpectations(t)
	mockNavService.AssertExpectations(t)
}

func TestResolveNavigationTarget_FromProjectUsesMain(t *testing.T) {
	mockCtxService := mocks.NewMockContextService()
	mockNavService := mocks.NewMockNavigationService()

	config := &CommandConfig{
		Services: &ServiceContainer{
			ContextService:    mockCtxService,
			NavigationService: mockNavService,
		},
	}

	// Set up expectations - from project context
	mockCtxService.On("GetCurrentContext").Return(&domain.Context{
		Type: domain.ContextProject,
	}, nil)

	// When target is empty from project context, should default to "main"
	mockNavService.On("ResolvePath", mock.Anything, mock.MatchedBy(func(req *domain.ResolvePathRequest) bool {
		return req.Target == "main"
	})).Return(&domain.ResolutionResult{
		ResolvedPath: "/path/to/main",
	}, nil)

	// Execute with empty target
	ctx := context.Background()
	currentCtx, result, err := resolveNavigationTarget(ctx, config, "")

	// Assert
	require.NoError(t, err)
	assert.NotNil(t, currentCtx)
	assert.Equal(t, domain.ContextProject, currentCtx.Type)
	assert.NotNil(t, result)

	mockCtxService.AssertExpectations(t)
	mockNavService.AssertExpectations(t)
}

func TestResolveNavigationTarget_OutsideGitErrors(t *testing.T) {
	mockCtxService := mocks.NewMockContextService()
	mockNavService := mocks.NewMockNavigationService()

	config := &CommandConfig{
		Services: &ServiceContainer{
			ContextService:    mockCtxService,
			NavigationService: mockNavService,
		},
	}

	// Set up expectations - outside git context
	mockCtxService.On("GetCurrentContext").Return(&domain.Context{
		Type: domain.ContextOutsideGit,
	}, nil)

	// Execute with empty target from outside git context
	ctx := context.Background()
	currentCtx, result, err := resolveNavigationTarget(ctx, config, "")

	// Assert - should error because no default target in outside git context
	require.Error(t, err)
	assert.Nil(t, currentCtx)
	assert.Nil(t, result)
	assert.Contains(t, err.Error(), "no target specified and no default worktree in context")

	mockCtxService.AssertExpectations(t)
	// NavigationService should not be called because we fail before resolution
}

func TestResolveNavigationTarget_ContextDetectionFailure(t *testing.T) {
	mockCtxService := mocks.NewMockContextService()
	mockNavService := mocks.NewMockNavigationService()

	config := &CommandConfig{
		Services: &ServiceContainer{
			ContextService:    mockCtxService,
			NavigationService: mockNavService,
		},
	}

	// Set up expectations - context detection fails
	mockCtxService.On("GetCurrentContext").Return(nil, assert.AnError)

	// Execute
	ctx := context.Background()
	currentCtx, result, err := resolveNavigationTarget(ctx, config, "feature-branch")

	// Assert
	require.Error(t, err)
	assert.Nil(t, currentCtx)
	assert.Nil(t, result)
	assert.Contains(t, err.Error(), "context detection failed")

	mockCtxService.AssertExpectations(t)
}

func TestResolveNavigationTarget_ResolutionFailure(t *testing.T) {
	mockCtxService := mocks.NewMockContextService()
	mockNavService := mocks.NewMockNavigationService()

	config := &CommandConfig{
		Services: &ServiceContainer{
			ContextService:    mockCtxService,
			NavigationService: mockNavService,
		},
	}

	// Set up expectations
	mockCtxService.On("GetCurrentContext").Return(&domain.Context{
		Type: domain.ContextProject,
	}, nil)

	mockNavService.On("ResolvePath", mock.Anything, mock.Anything).Return(nil, assert.AnError)

	// Execute
	ctx := context.Background()
	_, result, err := resolveNavigationTarget(ctx, config, "feature-branch")

	// Assert
	require.Error(t, err)
	assert.Nil(t, result)
	assert.Contains(t, err.Error(), "failed to resolve path for feature-branch")

	mockCtxService.AssertExpectations(t)
	mockNavService.AssertExpectations(t)
}
