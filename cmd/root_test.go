package cmd

import (
	"testing"

	"github.com/stretchr/testify/assert"
)

func TestRootCommand_BasicProperties(t *testing.T) {
	assert.Equal(t, "twiggit", rootCmd.Use)
	assert.Equal(t, "A pragmatic tool for managing git worktrees", rootCmd.Short)
	assert.Contains(t, rootCmd.Long, "git worktrees")
}
