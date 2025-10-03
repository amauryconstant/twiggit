package cmd

import (
	"testing"

	"github.com/spf13/cobra"
	"github.com/stretchr/testify/assert"
	"twiggit/internal/domain"
)

func TestCommandInterfaces_ContractCompliance(t *testing.T) {
	testCases := []struct {
		name        string
		command     *cobra.Command
		expectError bool
		setupFunc   func() *cobra.Command
	}{
		{
			name:        "list command interface compliance",
			setupFunc:   setupListCommand,
			expectError: false,
		},
		{
			name:        "create command interface compliance",
			setupFunc:   setupCreateCommand,
			expectError: false,
		},
		{
			name:        "delete command interface compliance",
			setupFunc:   setupDeleteCommand,
			expectError: false,
		},
		{
			name:        "cd command interface compliance",
			setupFunc:   setupCDCommand,
			expectError: false,
		},
	}

	for _, tc := range testCases {
		t.Run(tc.name, func(t *testing.T) {
			cmd := tc.setupFunc()
			assert.NotNil(t, cmd)
			assert.NotEmpty(t, cmd.Use)
			assert.NotEmpty(t, cmd.Short)
			assert.NotNil(t, cmd.RunE)
		})
	}
}

// setupListCommand creates a list command for testing
func setupListCommand() *cobra.Command {
	config := &CommandConfig{
		Services: &ServiceContainer{
			WorktreeService:   nil, // Will be mocked in actual tests
			ProjectService:    nil,
			NavigationService: nil,
			ContextService:    nil,
		},
		Config: domain.DefaultConfig(),
	}
	return NewListCommand(config)
}

// setupCreateCommand creates a create command for testing
func setupCreateCommand() *cobra.Command {
	config := &CommandConfig{
		Services: &ServiceContainer{
			WorktreeService:   nil, // Will be mocked in actual tests
			ProjectService:    nil,
			NavigationService: nil,
			ContextService:    nil,
		},
		Config: domain.DefaultConfig(),
	}
	return NewCreateCommand(config)
}

// setupDeleteCommand creates a delete command for testing
func setupDeleteCommand() *cobra.Command {
	config := &CommandConfig{
		Services: &ServiceContainer{
			WorktreeService:   nil, // Will be mocked in actual tests
			ProjectService:    nil,
			NavigationService: nil,
			ContextService:    nil,
		},
		Config: domain.DefaultConfig(),
	}
	return NewDeleteCommand(config)
}

// setupCDCommand creates a cd command for testing
func setupCDCommand() *cobra.Command {
	config := &CommandConfig{
		Services: &ServiceContainer{
			WorktreeService:   nil, // Will be mocked in actual tests
			ProjectService:    nil,
			NavigationService: nil,
			ContextService:    nil,
		},
		Config: domain.DefaultConfig(),
	}
	return NewCDCommand(config)
}
