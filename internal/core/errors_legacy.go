package core

import "strings"

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
func NewGitCommandError(command string, args []string, exitCode int, stdout, stderr, message string, err error) *OperationError {
	msg := command
	if len(args) > 0 {
		msg += " " + strings.Join(args, " ")
	}
	msg += " (exit " + intToStr(exitCode) + "): " + message
	return &OperationError{
		Op:      "git.command",
		Message: msg,
		Cause:   err,
	}
}

// NewConfigError demoted: returns *OperationError.
func NewConfigError(path, message string, err error) *OperationError {
	return &OperationError{
		Op:      "config",
		Entity:  path,
		Message: message,
		Cause:   err,
	}
}

// NewContextDetectionError demoted: returns *OperationError.
func NewContextDetectionError(path, message string, err error) *OperationError {
	return &OperationError{
		Op:      "context.detection",
		Entity:  path,
		Message: message,
		Cause:   err,
	}
}

// intToStr formats a non-negative int without importing strconv. Exit
// codes in this CLI are 0..255.
func intToStr(i int) string {
	if i == 0 {
		return "0"
	}
	if i < 0 {
		return "-" + intToStr(-i)
	}
	var buf [20]byte
	pos := len(buf)
	for i > 0 {
		pos--
		buf[pos] = byte('0' + i%10)
		i /= 10
	}
	return string(buf[pos:])
}
