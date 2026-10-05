package core

import (
	"strconv"
	"strings"
)

// Demoted git / config / context error constructors. Each previously
// concrete error type collapses to OperationError; Op names the source.

// NewGitRepositoryError demoted: returns *OperationError.
func NewGitRepositoryError(path, message string, err error) *OperationError {
	return &OperationError{
		Op:      "git.repository",
		Entity:  path,
		Message: message,
		Cause:   err,
	}
}

// NewGitWorktreeError demoted: returns *OperationError.
func NewGitWorktreeError(worktreePath, branchName, message string, err error) *OperationError {
	oe := &OperationError{
		Op:      "git.worktree",
		Entity:  worktreePath,
		Message: message,
		Cause:   err,
	}
	if branchName != "" {
		oe.Field = "branch:" + branchName
	}
	return oe
}

// NewGitCommandError demoted: returns *OperationError. Legacy fields
// (Command, Args, ExitCode, Stdout, Stderr) collapse into the cause
// chain; this constructor retains the legacy signature for source
// compatibility but the rich context is no longer carried on the
// OperationError directly.
func NewGitCommandError(command string, args []string, exitCode int, _, _, message string, err error) *OperationError {
	msg := command
	if len(args) > 0 {
		msg += " " + strings.Join(args, " ")
	}
	msg += " (exit " + strconv.Itoa(exitCode) + "): " + message
	return &OperationError{
		Op:      "git.command",
		Message: msg,
		Cause:   err,
	}
}

// NewConfigError demoted: returns *OperationError.
func NewConfigError(path, message string, err error) *OperationError {
	return &OperationError{
		Op:      "config.load",
		Entity:  path,
		Message: message,
		Cause:   err,
	}
}

// NewContextDetectionError demoted: returns *OperationError.
func NewContextDetectionError(path, message string, err error) *OperationError {
	return &OperationError{
		Op:      "context.detect",
		Entity:  path,
		Message: message,
		Cause:   err,
	}
}

// intToStr removed: strconv.Itoa handles the same byte output for the
// 0..255 exit-code range used by this CLI. See task 2.4 in
// naming-refactor-modernize.
