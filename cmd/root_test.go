package cmd

import (
	"testing"

	"github.com/carapace-sh/carapace"
	"github.com/stretchr/testify/assert"
)

func TestRootCommand_BasicProperties(t *testing.T) {
	config := &CommandConfig{
		Services: &ServiceContainer{},
		Config:   nil,
	}
	rootCmd := NewRootCommand(config)

	assert.Equal(t, "twiggit", rootCmd.Use)
	assert.Equal(t, "A pragmatic tool for managing git worktrees", rootCmd.Short)
	assert.Contains(t, rootCmd.Long, "git worktrees")
}

func TestRootCommand_CarapaceValidation(t *testing.T) {
	config := &CommandConfig{
		Services: &ServiceContainer{},
		Config:   nil,
	}
	NewRootCommand(config)

	carapace.Test(t)
}
