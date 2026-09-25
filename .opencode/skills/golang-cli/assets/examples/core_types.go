// Path: internal/core/worktree.go — Functional Core: stdlib + samber/lo, no I/O
package core

import (
	"strings"

	"github.com/samber/lo"
)

// BranchName is a value object: construction validates once, so every command
// taking a branch name receives a valid one.
type BranchName struct{ value string }

func NewBranchName(raw string) (BranchName, error) {
	if err := branchRules.Validate(raw); err != nil {
		return BranchName{}, err
	}
	return BranchName{value: raw}, nil
}

func (b BranchName) String() string { return b.value }

var branchRules = NewPipeline(
	func(s string) error {
		if s == "" {
			return &ValidationError{Field: "branch", Value: s, Message: "must not be empty"}
		}
		return nil
	},
	func(s string) error {
		if s == "HEAD" || strings.ContainsAny(s, " ~^:?*[\\") {
			return &ValidationError{Field: "branch", Value: s, Message: "not a valid git branch name",
				Suggestions: []string{"use letters, digits, '-', '_' and '/'"}}
		}
		return nil
	},
)

type Worktree struct {
	Name     string `json:"name"`
	Branch   string `json:"branch"`
	Path     string `json:"path"`
	Archived bool   `json:"archived"`
}

// Project is a type with behavior: its methods encode business rules.
type Project struct {
	Name      string
	Worktrees []Worktree
}

func (p *Project) ActiveWorktrees() []Worktree {
	return lo.Filter(p.Worktrees, func(w Worktree, _ int) bool { return !w.Archived })
}

// CanDelete is a rule spanning several values. It takes the protected branch
// list, not the config struct, so the core never depends on config loading.
func CanDelete(p *Project, wt Worktree, protected []string) error {
	if len(p.ActiveWorktrees()) <= 1 {
		return &OperationError{Op: "delete", Message: "cannot delete the last active worktree"}
	}
	if lo.Contains(protected, wt.Branch) {
		return &ValidationError{Field: "branch", Value: wt.Branch, Message: "branch is protected"}
	}
	return nil
}
