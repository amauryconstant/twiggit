package core

// HookType represents the type of hook being executed
type HookType string

const (
	// HookPostCreate is the hook executed after worktree creation
	HookPostCreate HookType = "post-create"
)

// HookConfig represents the hooks section of .twiggit.toml
type HookConfig struct {
	PostCreate *HookDefinition `toml:"post-create" koanf:"post-create"`
}

// HookDefinition represents a single hook's configuration
type HookDefinition struct {
	Commands []string `toml:"commands" koanf:"commands"`
}

// HookResult represents the result of hook execution
type HookResult struct {
	HookType     HookType
	HasExecuted  bool
	IsSuccessful bool
	Failures     []HookFailure
}

// HookFailure represents details of a failed hook command
type HookFailure struct {
	Command  string
	ExitCode int
	Output   string
}

// HookRunRequest carries the context needed to execute hooks. Pure data —
// the consumer-side interface (cmdutil.HookRunner) takes this type; the
// implementation lives in internal/git/hook_runner.go.
type HookRunRequest struct {
	HookType       HookType
	WorktreePath   string
	ProjectName    string
	BranchName     string
	SourceBranch   string
	MainRepoPath   string
	ConfigFilePath string
}
