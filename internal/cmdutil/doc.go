// Package cmdutil holds the composition seams shared by every command:
//
//   - Factory: lazy function fields + sync.OnceValue-cached config + Init().
//   - ExitCodeFor: maps errors to ExitOK / ExitError / ExitUsage (0/1/2).
//   - Global flag helpers (--output / --quiet / --verbose).
//   - HookRunner interface declaration (consumer-side, shared by all
//     commands that trigger hooks).
//
// cmdutil never imports cmd/. cmd/ imports cmdutil.
package cmdutil
