package domain

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
	HookType HookType
	Executed bool
	Success  bool
	Failures []HookFailure
}

// HookFailure represents details of a failed hook command
type HookFailure struct {
	Command  string
	ExitCode int
	Output   string
}
