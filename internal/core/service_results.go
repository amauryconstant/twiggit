package core

import (
	"reflect"
	"time"
)

// Result represents a generic result type following the Result/Either pattern.
// Success tracks the (value, error) tuple shape so IsSuccess / IsError dispatch
// on the flag without re-comparing Error == nil; the Value / Error fields stay
// the canonical access points for downstream code.
type Result[T any] struct {
	Value   T
	Success bool
	Error   error
}

// NewResult creates a new successful result.
// When T is a slice type, the underlying backing array is cloned to
// prevent callers from mutating internal state through the returned struct.
func NewResult[T any](value T) Result[T] {
	return Result[T]{Value: cloneSlice(value), Success: true, Error: nil}
}

// NewResultOr constructs a Result from a (value, error) pair. Success tracks
// err == nil so callers do not have to repeat the comparison. Named with the
// Rust Result::Ok_or shape; the single-arg NewResult / NewErrResult constructors
// remain the canonical success / failure constructors per core-types spec.
func NewResultOr[T any](value T, err error) Result[T] {
	return Result[T]{Value: cloneSlice(value), Success: err == nil, Error: err}
}

// NewErrResult creates a new error result
func NewErrResult[T any](err error) Result[T] {
	var zero T
	return Result[T]{Value: zero, Success: false, Error: err}
}

// cloneSlice returns a copy of the slice when value is a slice type;
// non-slice values pass through untouched. Defensive copy avoids the
// backing-array-aliasing trap when the caller mutates the slice after
// the result has been stored.
func cloneSlice[T any](value T) T {
	v := reflect.ValueOf(value)
	if !v.IsValid() || v.Kind() != reflect.Slice {
		return value
	}
	if v.IsNil() {
		return value
	}
	dst := reflect.MakeSlice(v.Type(), v.Len(), v.Cap())
	reflect.Copy(dst, v)
	return dst.Interface().(T) //nolint:errcheck // reflect.Copy guarantees non-nil dst; type assertion to T cannot fail
}

// IsSuccess returns true if the result is successful
func (r Result[T]) IsSuccess() bool {
	return r.Success
}

// IsError returns true if the result contains an error
func (r Result[T]) IsError() bool {
	return !r.Success
}

// WorktreeStatus represents the diagnostic view of a single worktree:
// its repository status, ahead/behind counts against the resolved
// base, merge readiness, dirty state, last-commit date, and a
// best-effort skip flag. Populated by the *git.Client.ReadWorktreeStatus
// adapter (cross-half method on the composite) and consumed by
// cmd/status and cmd/delete.
type WorktreeStatus struct {
	Worktree              *Worktree         `json:"worktree,omitempty"`
	ProjectName           string            `json:"project,omitempty"`
	RepositoryStatus      *RepositoryStatus `json:"repository_status,omitempty"`
	LastChecked           time.Time         `json:"last_checked"`
	IsClean               bool              `json:"is_clean"`
	HasUncommittedChanges bool              `json:"has_uncommitted_changes"`
	Base                  string            `json:"base"`
	IsMerged              bool              `json:"is_merged"`
	IsStale               bool              `json:"is_stale"`
	LastCommitDate        time.Time         `json:"last_commit_date"`
	IsSkipped             bool              `json:"is_skipped"`
	SkipReason            string            `json:"skip_reason,omitempty"`
}

// Dirty reports whether the worktree has uncommitted changes.
// Aliased to HasUncommittedChanges for cmd-layer readability.
func (s *WorktreeStatus) Dirty() bool { return s.HasUncommittedChanges }

// ComputeIsStale returns true when either the behind-count exceeds
// cfg.Status.StaleBehind (when > 0) or the time since LastCommitDate
// exceeds cfg.Status.StaleDays * 24h (when > 0 and LastCommitDate is
// not the zero time). When both thresholds are zero the heuristic is
// disabled. The cmd layer calls this per row after the adapter
// returns so the IsStale field carries the per-invocation value.
func (s *WorktreeStatus) ComputeIsStale(cfg *Config) bool {
	if cfg == nil {
		return false
	}
	if cfg.Status.StaleBehind > 0 && s.RepositoryStatus != nil && s.RepositoryStatus.Behind >= cfg.Status.StaleBehind {
		return true
	}
	if cfg.Status.StaleDays > 0 && !s.LastCommitDate.IsZero() {
		age := time.Since(s.LastCommitDate)
		if age >= time.Duration(cfg.Status.StaleDays)*24*time.Hour {
			return true
		}
	}
	return false
}

// ProjectInfo represents comprehensive project information
type ProjectInfo struct {
	Name          string
	Path          string
	GitRepoPath   string
	Worktrees     []*Worktree
	Branches      []*Branch
	Remotes       []*Remote
	DefaultBranch string
	IsBare        bool
	LastModified  time.Time
}

// ProjectSummary represents lightweight project information without expensive git data
type ProjectSummary struct {
	Name        string
	Path        string
	GitRepoPath string
}

// CreateWorktreeResult represents the result of a worktree creation operation
type CreateWorktreeResult struct {
	Worktree   *Worktree
	HookResult *HookResult
}
