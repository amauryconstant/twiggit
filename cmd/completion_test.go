package cmd

import (
	"testing"
	"time"

	"github.com/stretchr/testify/assert"

	"twiggit/internal/core"
	"twiggit/internal/infrastructure"
)

func TestCompletion_GetCompletionTimeout(t *testing.T) {
	tests := []struct {
		name     string
		config   *core.Config
		expected time.Duration
	}{
		{
			name:     "nil config returns default",
			config:   nil,
			expected: 500 * time.Millisecond,
		},
		{
			name:     "empty timeout returns default",
			config:   &core.Config{},
			expected: 500 * time.Millisecond,
		},
		{
			name: "valid timeout from config",
			config: &core.Config{
				Completion: core.CompletionConfig{Timeout: "1s"},
			},
			expected: time.Second,
		},
		{
			name: "invalid timeout returns default",
			config: &core.Config{
				Completion: core.CompletionConfig{Timeout: "invalid"},
			},
			expected: 500 * time.Millisecond,
		},
		{
			name: "custom timeout 250ms",
			config: &core.Config{
				Completion: core.CompletionConfig{Timeout: "250ms"},
			},
			expected: 250 * time.Millisecond,
		},
		{
			name: "custom timeout 2s",
			config: &core.Config{
				Completion: core.CompletionConfig{Timeout: "2s"},
			},
			expected: 2 * time.Second,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			result := getCompletionTimeout(tt.config)
			assert.Equal(t, tt.expected, result)
		})
	}
}

func TestCompletion_SuggestionsToCarapaceAction_EmptySuggestions(t *testing.T) {
	action := suggestionsToCarapaceAction([]*core.ResolutionSuggestion{}, "main")

	assert.NotNil(t, action)
}

func TestCompletion_SuggestionsToCarapaceAction_NilSuggestions(t *testing.T) {
	action := suggestionsToCarapaceAction(nil, "main")

	assert.NotNil(t, action)
}

func TestCompletion_SuggestionsToCarapaceAction_WithSuggestions(t *testing.T) {
	suggestions := []*core.ResolutionSuggestion{
		{Text: "main", Description: "Project root directory"},
		{Text: "feature-1", Description: "Worktree for branch feature-1"},
		{Text: "develop", Description: "Branch develop (create worktree)"},
	}

	action := suggestionsToCarapaceAction(suggestions, "main")

	assert.NotNil(t, action)
}

func TestCompletion_SuggestionsToCarapaceAction_SingleSuggestion(t *testing.T) {
	suggestions := []*core.ResolutionSuggestion{
		{Text: "main", Description: "Project root directory", Type: core.PathTypeProject, ProjectName: "test-project"},
	}

	action := suggestionsToCarapaceAction(suggestions, "main")

	assert.NotNil(t, action)
}

func TestCompletion_ActionWorktreeTarget_ReturnsAction(t *testing.T) {
	config := &CommandConfig{
		Services: &ServiceContainer{},
		Config:   &core.Config{},
	}
	action := actionWorktreeTarget(config)

	assert.NotNil(t, action)
}

func TestCompletion_ActionBranches_ReturnsAction(t *testing.T) {
	config := &CommandConfig{
		Services: &ServiceContainer{},
		Config:   &core.Config{},
	}
	action := actionBranches(config)

	assert.NotNil(t, action)
}

func TestCompletion_ActionBranchesForProject_ReturnsAction(t *testing.T) {
	config := &CommandConfig{
		Services: &ServiceContainer{},
		Config:   &core.Config{},
	}
	action := actionBranchesForProject("myproject", config)

	assert.NotNil(t, action)
}

func TestCompletion_ActionWorktreeTarget_WithExistingOnly(t *testing.T) {
	config := &CommandConfig{
		Services: &ServiceContainer{},
		Config:   &core.Config{},
	}
	action := actionWorktreeTarget(config, infrastructure.WithExistingOnly())

	assert.NotNil(t, action)
}

func TestCompletion_ActionWorktreeTarget_WithNilConfig(t *testing.T) {
	nilConfig := &CommandConfig{
		Services: &ServiceContainer{},
		Config:   nil,
	}

	action := actionWorktreeTarget(nilConfig)

	assert.NotNil(t, action)
}

func TestCompletion_GetCompletionTimeout_NilConfigSafety(t *testing.T) {
	result := getCompletionTimeout(nil)

	assert.Equal(t, 500*time.Millisecond, result)
}
