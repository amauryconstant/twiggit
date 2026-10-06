package core

import (
	"errors"
	"testing"
)

func TestRebaseOutcome_String(t *testing.T) {
	t.Parallel()

	tests := []struct {
		name     string
		outcome  RebaseOutcome
		expected string
	}{
		{"unknown", RebaseOutcomeUnknown, "unknown"},
		{"clean", RebaseOutcomeClean, "clean"},
		{"conflicted", RebaseOutcomeConflicted, "conflicted"},
		{"aborted", RebaseOutcomeAborted, "aborted"},
		{"nothing-to-do", RebaseOutcomeNothingToDo, "nothing-to-do"},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			t.Parallel()
			if got := tt.outcome.String(); got != tt.expected {
				t.Fatalf("RebaseOutcome(%d).String() = %q, want %q", tt.outcome, got, tt.expected)
			}
		})
	}
}

func TestRebaseOutcome_ZeroIsUnknown(t *testing.T) {
	t.Parallel()

	var r RebaseOutcome
	if r != RebaseOutcomeUnknown {
		t.Fatalf("zero RebaseOutcome should equal RebaseOutcomeUnknown, got %d", r)
	}
}

func TestRebasedWorktree_ZeroValueSafe(t *testing.T) {
	t.Parallel()

	var w RebasedWorktree
	if w.ProjectName != "" || w.BranchName != "" || w.WorktreePath != "" || w.TrackedBase != "" || w.SkipReason != "" {
		t.Fatalf("RebasedWorktree zero value must be empty strings, got %+v", w)
	}
	if w.Outcome != RebaseOutcomeUnknown {
		t.Fatalf("RebasedWorktree zero Outcome must be Unknown, got %d", w.Outcome)
	}
	if w.Error != nil {
		t.Fatalf("RebasedWorktree zero Error must be nil, got %v", w.Error)
	}
}

func TestRebaseResult_ZeroValueSafe(t *testing.T) {
	t.Parallel()

	var r RebaseResult
	if r.RebasedWorktrees != nil || r.SkippedWorktrees != nil {
		t.Fatalf("RebaseResult slices must be nil when zero, got %+v", r)
	}
	if r.TotalRebased != 0 || r.TotalSkipped != 0 || r.TotalConflicts != 0 || r.NavigationPath != "" {
		t.Fatalf("RebaseResult counters must be zero, got %+v", r)
	}
}

func TestRebaseRequest_FieldsReadable(t *testing.T) {
	t.Parallel()

	req := RebaseRequest{
		ProjectName:   "twiggit",
		BranchName:    "feature",
		WorktreePath:  "/tmp/wt",
		OntoBranch:    "main",
		IsForce:       true,
		IsFetch:       true,
		IsAll:         false,
		IsContinue:    false,
		IsAbort:       false,
		IsSetBase:     true,
		NewBaseBranch: "develop",
	}
	if req.ProjectName != "twiggit" || req.BranchName != "feature" || req.WorktreePath != "/tmp/wt" {
		t.Fatalf("RebaseRequest basic fields not readable: %+v", req)
	}
	if !req.IsForce || !req.IsFetch || !req.IsSetBase {
		t.Fatalf("RebaseRequest bool fields not readable: %+v", req)
	}
	if req.NewBaseBranch != "develop" {
		t.Fatalf("RebaseRequest.NewBaseBranch = %q, want %q", req.NewBaseBranch, "develop")
	}
}

func TestSyncRequest_FieldsReadable(t *testing.T) {
	t.Parallel()

	req := SyncRequest{
		ProjectName: "twiggit",
		BranchName:  "main",
		Remote:      "upstream",
		IsAll:       true,
		IsFetchOnly: false,
		IsRebase:    true,
	}
	if req.Remote != "upstream" || !req.IsAll || !req.IsRebase {
		t.Fatalf("SyncRequest fields not readable: %+v", req)
	}
}

func TestSyncedBranch_FieldsReadable(t *testing.T) {
	t.Parallel()

	b := SyncedBranch{
		ProjectName: "twiggit",
		BranchName:  "main",
		RemoteName:  "origin",
		OldTip:      "abc123",
		NewTip:      "def456",
	}
	if b.OldTip == b.NewTip {
		t.Fatalf("OldTip and NewTip should differ for this case: %+v", b)
	}
}

func TestSyncResult_ZeroValueSafe(t *testing.T) {
	t.Parallel()

	var r SyncResult
	if r.SyncedBranches != nil || r.RebasedBranches != nil {
		t.Fatalf("SyncResult slices must be nil when zero, got %+v", r)
	}
	if r.TotalSynced != 0 || r.TotalRebased != 0 || r.TotalConflicts != 0 {
		t.Fatalf("SyncResult counters must be zero, got %+v", r)
	}
}

// TestSyncResult_FetchOnlyLeavesRebasedBranchesEmpty covers the spec
// scenario "--fetch-only leaves RebasedBranches empty": even when a
// sync populates SyncedBranches, --fetch-only must keep the rebase
// slice empty. The previous zero-value-safe test only covered the
// fully empty case.
func TestSyncResult_FetchOnlyLeavesRebasedBranchesEmpty(t *testing.T) {
	t.Parallel()

	r := SyncResult{
		SyncedBranches: []*SyncedBranch{
			{ProjectName: "p", BranchName: "main", RemoteName: "origin", OldTip: "abc", NewTip: "def"},
		},
		TotalSynced: 1,
	}
	if len(r.SyncedBranches) != 1 {
		t.Fatalf("SyncedBranches must hold 1 entry, got %d", len(r.SyncedBranches))
	}
	if len(r.RebasedBranches) != 0 || r.TotalRebased != 0 {
		t.Fatalf("--fetch-only must keep RebasedBranches empty: got %+v", r)
	}
}

func TestSentinels_Rebase(t *testing.T) {
	t.Parallel()

	sentinels := map[string]error{
		"ErrRebaseConflict":   ErrRebaseConflict,
		"ErrRebaseInProgress": ErrRebaseInProgress,
		"ErrBaseNotSet":       ErrBaseNotSet,
	}

	for name, s := range sentinels {
		t.Run(name, func(t *testing.T) {
			t.Parallel()
			msg := s.Error()
			if msg == "" {
				t.Fatalf("%s message must be non-empty", name)
			}
			if msg[0] < 'a' || msg[0] > 'z' {
				t.Fatalf("%s message must start lowercase, got %q", name, msg)
			}
			if msg[len(msg)-1] == '.' || msg[len(msg)-1] == '!' || msg[len(msg)-1] == '?' {
				t.Fatalf("%s message must not end with punctuation, got %q", name, msg)
			}
		})
	}
}

func TestSentinels_RebaseDistinct(t *testing.T) {
	t.Parallel()

	if errors.Is(ErrRebaseConflict, ErrRebaseInProgress) {
		t.Fatalf("ErrRebaseConflict must not match ErrRebaseInProgress")
	}
	if errors.Is(ErrRebaseConflict, ErrBaseNotSet) {
		t.Fatalf("ErrRebaseConflict must not match ErrBaseNotSet")
	}
	if errors.Is(ErrRebaseInProgress, ErrBaseNotSet) {
		t.Fatalf("ErrRebaseInProgress must not match ErrBaseNotSet")
	}
}

func TestOperationError_Is_Rebase(t *testing.T) {
	t.Parallel()

	conflict := &OperationError{Op: "rebase.worktree", Cause: ErrRebaseConflict}
	if !errors.Is(conflict, ErrRebaseConflict) {
		t.Fatalf("Op=rebase.worktree must walk to ErrRebaseConflict via Is")
	}

	abort := &OperationError{Op: "rebase.continue", Cause: ErrRebaseInProgress}
	if !errors.Is(abort, ErrRebaseInProgress) {
		t.Fatalf("Op=rebase.continue must walk to ErrRebaseInProgress via Is")
	}

	base := &OperationError{Op: "base.tracked", Cause: ErrBaseNotSet}
	if !errors.Is(base, ErrBaseNotSet) {
		t.Fatalf("Op=base.tracked must walk to ErrBaseNotSet via Is")
	}

	// Plain Op=rebase (no subop) should also match.
	plain := &OperationError{Op: "rebase", Cause: ErrRebaseConflict}
	if !errors.Is(plain, ErrRebaseConflict) {
		t.Fatalf("Op=rebase must walk to ErrRebaseConflict via Is")
	}

	// Negative: not a rebase op with no rebase cause must not match.
	other := &OperationError{Op: "worktree.create", Cause: errors.New("plain")}
	if errors.Is(other, ErrRebaseConflict) {
		t.Fatalf("Op=worktree.create with non-rebase cause must not walk to ErrRebaseConflict")
	}
}
