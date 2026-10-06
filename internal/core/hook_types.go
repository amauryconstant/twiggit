package core

// HookType represents the type of hook being executed.
// Int iota per enum-unknown-zero rule with HookTypeUnknown at 0.
type HookType int

const (
	// HookTypeUnknown is the zero value of HookType, used when no
	// specific hook type has been identified.
	HookTypeUnknown HookType = iota
	// HookTypePostCreate is the hook executed after worktree creation.
	HookTypePostCreate
	// HookTypePreRebase is the hook executed before a rebase starts.
	// Non-zero exit aborts the rebase.
	HookTypePreRebase
	// HookTypePostRebase is the hook executed after a clean rebase.
	// Non-zero exit is a warning; the rebase outcome is preserved.
	HookTypePostRebase
	// HookTypePostSync is the hook executed after a sync walk. Non-zero
	// exit is a warning; the sync outcome is preserved.
	HookTypePostSync
)

// HookPostCreate is a transitional alias for HookTypePostCreate kept so
// test/integration callers compile during the migration window. New
// code SHALL use HookTypePostCreate; the next phase agent removes this
// alias after the integration test migrates to the new name.
var HookPostCreate = HookTypePostCreate

// String returns the kebab-case wire form used in config TOML and JSON.
// HookTypeUnknown formats as "unknown" so renderers can distinguish
// unrecognised values from a missing field.
func (h HookType) String() string {
	switch h {
	case HookTypePostCreate:
		return "post-create"
	case HookTypePreRebase:
		return "pre-rebase"
	case HookTypePostRebase:
		return "post-rebase"
	case HookTypePostSync:
		return "post-sync"
	default:
		return "unknown"
	}
}

// HookConfig represents the hooks section of .twiggit.toml. Each event
// is a slice so multiple hook definitions can chain on a single event.
type HookConfig struct {
	PostCreate []HookDefinition `toml:"post-create" koanf:"post-create"`
	PreRebase  []HookDefinition `toml:"pre-rebase" koanf:"pre-rebase"`
	PostRebase []HookDefinition `toml:"post-rebase" koanf:"post-rebase"`
	PostSync   []HookDefinition `toml:"post-sync" koanf:"post-sync"`
}

// HookDefinition represents a single hook's configuration. Command is a
// single string shell-parsed by the runtime; future work may add Commands
// []string for multi-line scripts (deferred; single-string keeps TOML
// round-trip simple).
type HookDefinition struct {
	Command          string `toml:"command" koanf:"command"`
	WorkingDirectory string `toml:"working_directory,omitempty" koanf:"working_directory"`
	TimeoutSeconds   int    `toml:"timeout_seconds,omitempty" koanf:"timeout_seconds"`
}

// HookResult represents the result of hook execution
type HookResult struct {
	HookType     HookType
	HasExecuted  bool
	IsSuccessful bool
	Failures     []HookFailure
}

// HookFailure represents details of a failed hook command. Output
// retains the original field name so cmd/create.go's
// `failure.Output` rendering continues to compile; Error carries
// the underlying executor error for chain walking; TimedOut marks
// a context.DeadlineExceeded path so callers can render a
// distinct timeout hint.
type HookFailure struct {
	Command  string
	ExitCode int
	Output   string
	Error    error
	TimedOut bool
}

// HookRunRequest carries the context needed to execute hooks. Pure data —
// the consumer-side interface (cmdutil.HookRunner) takes this type; the
// implementation lives in internal/git/hook_runner.go. The six optional
// rebase/sync fields default to empty and are surfaced to the hook
// process only when the field's lifecycle stage matches the HookType
// and the field is non-empty.
type HookRunRequest struct {
	HookType       HookType
	WorktreePath   string
	ProjectName    string
	BranchName     string
	SourceBranch   string
	MainRepoPath   string
	ConfigFilePath string
	RebaseBase     string
	RebaseOldTip   string
	RebaseNewTip   string
	RebaseResult   string
	SyncRemote     string
	SyncBranch     string
}
