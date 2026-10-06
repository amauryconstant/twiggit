package core

// RebaseOutcome classifies the per-worktree result of a rebase invocation.
// Int iota per enum-unknown-zero rule with RebaseOutcomeUnknown at 0.
type RebaseOutcome int

const (
	// RebaseOutcomeUnknown is the zero value of RebaseOutcome. It marks a
	// RebasedWorktree whose rebase has not yet been classified.
	RebaseOutcomeUnknown RebaseOutcome = iota
	// RebaseOutcomeClean marks a rebase that completed without conflicts.
	RebaseOutcomeClean
	// RebaseOutcomeConflicted marks a rebase that stopped mid-flight because
	// git reported a conflict. The worktree remains in mid-rebase state and
	// the caller is expected to drive --continue or --abort.
	RebaseOutcomeConflicted
	// RebaseOutcomeAborted marks a rebase that was rolled back, either
	// explicitly via RebaseOutcome or implicitly because no rebase was in
	// progress when Continue was invoked.
	RebaseOutcomeAborted
	// RebaseOutcomeNothingToDo marks a rebase that detected the branch
	// already contained the base's tip. The worktree is unchanged.
	RebaseOutcomeNothingToDo
)

// String returns the kebab-case wire form used in test assertions and
// human-readable summaries. The mapping is total: every variant maps to a
// distinct string, and the unknown sentinel maps to "unknown" so renderers
// can distinguish unrecognised values from a missing field.
func (o RebaseOutcome) String() string {
	switch o {
	case RebaseOutcomeClean:
		return "clean"
	case RebaseOutcomeConflicted:
		return "conflicted"
	case RebaseOutcomeAborted:
		return "aborted"
	case RebaseOutcomeNothingToDo:
		return "nothing-to-do"
	default:
		return "unknown"
	}
}

// RebaseRequest carries every input the cmd layer resolves before the
// rebase walk starts. Empty optional fields indicate "not set"; the cmd
// layer fills them in (worktree path, project name, branch name, onto
// base) so the helpers never have to re-derive them from the context.
type RebaseRequest struct {
	ProjectName   string
	BranchName    string
	WorktreePath  string
	OntoBranch    string
	IsForce       bool
	IsFetch       bool
	IsAll         bool
	IsContinue    bool
	IsAbort       bool
	IsSetBase     bool
	NewBaseBranch string
}

// RebasedWorktree records the per-worktree result of a rebase walk entry.
// Error carries the wrapped adapter error when the worktree's rebase
// failed; cmd layer reads the field directly without string-keyed access.
type RebasedWorktree struct {
	ProjectName  string
	BranchName   string
	WorktreePath string
	TrackedBase  string
	Outcome      RebaseOutcome
	SkipReason   string
	Error        error
}

// RebaseResult aggregates a rebase walk. NavigationPath is non-empty only
// for single-target rebase successes, mirroring PruneWorktreesResult so
// the shell wrapper can cd into the rebased worktree via stdout.
type RebaseResult struct {
	RebasedWorktrees []*RebasedWorktree
	SkippedWorktrees []*RebasedWorktree
	TotalRebased     int
	TotalSkipped     int
	TotalConflicts   int
	NavigationPath   string
}

// SyncRequest carries every input the cmd layer resolves before the
// sync walk starts. Empty Remote resolves to Config.Sync.DefaultRemote
// or "origin".
type SyncRequest struct {
	ProjectName string
	BranchName  string
	Remote      string
	IsAll       bool
	IsFetchOnly bool
	IsRebase    bool
}

// SyncedBranch records the before- and after-tip of a synced branch.
// OldTip and NewTip are commit sha strings; identical values indicate
// the branch was already up to date.
type SyncedBranch struct {
	ProjectName string
	BranchName  string
	RemoteName  string
	OldTip      string
	NewTip      string
}

// SyncResult aggregates a sync walk. RebasedBranches is populated only
// when SyncRequest.IsRebase is true.
type SyncResult struct {
	SyncedBranches  []*SyncedBranch
	RebasedBranches []*RebasedWorktree
	TotalSynced     int
	TotalRebased    int
	TotalConflicts  int
	NavigationPath  string
}
