package cmd

import (
	"testing"

	"github.com/stretchr/testify/assert"
)

func TestRootCommand_BasicProperties(t *testing.T) {
	assert.Equal(t, "twiggit", rootCmd.Use)
	assert.Equal(t, "Pragmatic git worktree management tool", rootCmd.Short)
	assert.Contains(t, rootCmd.Long, "git worktrees")
}
