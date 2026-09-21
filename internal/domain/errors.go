package domain

import (
	"errors"
	"fmt"
	"strings"
)

// ContextDetectionError represents context detection errors
type ContextDetectionError struct {
	Path    string
	Err     error
	Message string
}

func (e *ContextDetectionError) Error() string {
	return fmt.Sprintf("context detection failed for %s: %s", e.Path, e.Message)
}

func (e *ContextDetectionError) Unwrap() error {
	return e.Err
}

// NewContextDetectionError creates a new context detection error
func NewContextDetectionError(path, message string, err error) *ContextDetectionError {
	return &ContextDetectionError{
		Path:    path,
		Err:     err,
		Message: message,
	}
}

// GitRepositoryError represents git repository operation errors
type GitRepositoryError struct {
	Path    string
	Message string
	Err     error
}

func (e *GitRepositoryError) Error() string {
	return fmt.Sprintf("git repository operation failed for %s: %s", e.Path, e.Message)
}

func (e *GitRepositoryError) Unwrap() error {
	return e.Err
}

// Is reports whether the wrapped sentinel matches.
func (e *GitRepositoryError) Is(target error) bool {
	return target == ErrGitRepoNotFound
}

// NewGitRepositoryError creates a new git repository error
func NewGitRepositoryError(path, message string, err error) *GitRepositoryError {
	return &GitRepositoryError{
		Path:    path,
		Err:     err,
		Message: message,
	}
}

// GitWorktreeError represents git worktree operation errors
type GitWorktreeError struct {
	WorktreePath string
	BranchName   string
	Message      string
	Err          error
}

func (e *GitWorktreeError) Error() string {
	return e.formatErrorMessage()
}

// formatErrorMessage formats the worktree error message with appropriate details
func (e *GitWorktreeError) formatErrorMessage() string {
	baseMsg := e.formatBaseMessage()

	// Include cause details if available
	if causeDetails := e.getCauseDetails(); causeDetails != "" {
		baseMsg += "\ncaused by: " + causeDetails
	}

	return baseMsg
}

// formatBaseMessage creates the base error message without cause details
func (e *GitWorktreeError) formatBaseMessage() string {
	if e.BranchName != "" {
		return fmt.Sprintf("git worktree operation failed for %s (branch: %s): %s", e.WorktreePath, e.BranchName, e.Message)
	}
	return fmt.Sprintf("git worktree operation failed for %s: %s", e.WorktreePath, e.Message)
}

// getCauseDetails extracts useful information from the cause error
func (e *GitWorktreeError) getCauseDetails() string {
	if e.Err == nil {
		return ""
	}

	// If the cause is a GitCommandError, include its details for better debugging
	gitCmdErr := &GitCommandError{}
	if errors.As(e.Err, &gitCmdErr) {
		return gitCmdErr.Error()
	}

	// For other error types, just return the error message
	return e.Err.Error()
}

func (e *GitWorktreeError) Unwrap() error {
	return e.Err
}

// Is reports whether the wrapped sentinel matches.
func (e *GitWorktreeError) Is(target error) bool {
	return target == ErrWorktreeNotFound
}

// NewGitWorktreeError creates a new git worktree error
func NewGitWorktreeError(worktreePath, branchName, message string, err error) *GitWorktreeError {
	return &GitWorktreeError{
		WorktreePath: worktreePath,
		BranchName:   branchName,
		Message:      message,
		Err:          err,
	}
}

// GitCommandError represents git command execution errors
type GitCommandError struct {
	Command  string
	Args     []string
	ExitCode int
	Stdout   string
	Stderr   string
	Message  string
	Err      error
}

func (e *GitCommandError) Error() string {
	return e.formatErrorMessage()
}

// formatErrorMessage formats the git command error message with appropriate details
func (e *GitCommandError) formatErrorMessage() string {
	baseMsg := fmt.Sprintf("git command failed: %s %v (exit code %d): %s", e.Command, e.Args, e.ExitCode, e.Message)

	// Include stderr if it contains useful information
	if e.hasUsefulStderr() {
		baseMsg += "\nstderr: " + e.Stderr
	}

	return baseMsg
}

// hasUsefulStderr checks if stderr contains information worth displaying
func (e *GitCommandError) hasUsefulStderr() bool {
	return e.Stderr != "" && !containsOnlyWhitespace(e.Stderr)
}

// containsOnlyWhitespace checks if a string contains only whitespace characters
func containsOnlyWhitespace(s string) bool {
	return strings.TrimSpace(s) == ""
}

func (e *GitCommandError) Unwrap() error {
	return e.Err
}

// NewGitCommandError creates a new git command error
func NewGitCommandError(command string, args []string, exitCode int, stdout, stderr, message string, err error) *GitCommandError {
	return &GitCommandError{
		Command:  command,
		Args:     args,
		ExitCode: exitCode,
		Stdout:   stdout,
		Stderr:   stderr,
		Message:  message,
		Err:      err,
	}
}

// ConfigError represents configuration-related errors
type ConfigError struct {
	Path    string
	Message string
	Err     error
}

func (e *ConfigError) Error() string {
	return fmt.Sprintf("config error for %s: %s", e.Path, e.Message)
}

func (e *ConfigError) Unwrap() error {
	return e.Err
}

// NewConfigError creates a new config error
func NewConfigError(path, message string, err error) *ConfigError {
	return &ConfigError{
		Path:    path,
		Message: message,
		Err:     err,
	}
}
