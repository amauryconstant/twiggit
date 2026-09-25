// Path: internal/cmdutil/hooks.go
package cmdutil

import (
	"github.com/spf13/cobra"

	"github.com/you/myapp/internal/core"
)

// RunParentPreRun runs the nearest ancestor hook above owner, the command
// that defines the calling hook. Walk from owner, not from c (the leaf):
// starting at c finds owner's own hook and recurses.
//
//	admin.PersistentPreRunE = func(c *cobra.Command, args []string) error {
//		if err := cmdutil.RunParentPreRun(admin, c, args); err != nil { // root: logging
//			return err
//		}
//		return checkAdminToken(c, args)
//	}
func RunParentPreRun(owner, c *cobra.Command, args []string) error {
	for p := owner.Parent(); p != nil; p = p.Parent() {
		if p.PersistentPreRunE != nil {
			return p.PersistentPreRunE(c, args)
		}
	}
	return nil
}

// AddPreRun stacks mw in front of the hook already set on cmd. It composes
// only hooks on that same command and does not fix parent shadowing.
func AddPreRun(cmd *cobra.Command, mw func(*cobra.Command, []string) error) {
	next := cmd.PersistentPreRunE
	cmd.PersistentPreRunE = func(c *cobra.Command, args []string) error {
		if err := mw(c, args); err != nil {
			return err
		}
		if next != nil {
			return next(c, args)
		}
		return nil
	}
}

// RequireAuth skips commands annotated as authless and returns an error with
// the next step when the token is missing. Wire it with
// AddPreRun(root, RequireAuth(loadToken)).
func RequireAuth(loadToken func() (string, error)) func(*cobra.Command, []string) error {
	return func(c *cobra.Command, _ []string) error {
		if c.Annotations["auth"] == "skip" { // set on version, completion, auth login
			return nil
		}
		if tok, err := loadToken(); err != nil || tok == "" {
			return &core.OperationError{Op: "auth", Message: "not logged in", Cause: err,
				Suggestions: []string{"run 'myapp auth login'"}}
		}
		return nil
	}
}
