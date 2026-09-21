package domain

import "errors"

// Sentinel catalog for typed error-chain dispatch.
//
// All sentinels are package-level vars of type error, with messages of
// the form "domain: <resource> <state>". Callers identify sentinels
// exclusively through errors.Is; string equality on the message is
// never a contract. The canonical list and exact messages are owned by
// the domain-typed-errors spec.
var (
	ErrGitRepoNotFound       = errors.New("domain: git repository not found")
	ErrWorktreeNotFound      = errors.New("domain: worktree not found")
	ErrProjectNotFound       = errors.New("domain: project not found")
	ErrResolutionNotFound    = errors.New("domain: resolution target not found")
	ErrShellAlreadyInstalled = errors.New("domain: shell wrapper already installed")
	ErrShellNotInstalled     = errors.New("domain: shell wrapper not installed")
	ErrInvalidShellType      = errors.New("domain: invalid shell type")
	ErrShellInferenceFailed  = errors.New("domain: could not infer shell type")
	ErrShellDetectionFailed  = errors.New("domain: shell detection failed")
	ErrWrapperGeneration     = errors.New("domain: wrapper generation failed")
	ErrWrapperInstallation   = errors.New("domain: wrapper installation failed")
	ErrConfigFileNotFound    = errors.New("domain: config file not found")
)
