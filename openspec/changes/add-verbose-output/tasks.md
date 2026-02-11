## 1. Foundation

- [x] 1.1 Add persistent `--verbose` flag to cmd/root.go using PersistentFlags().CountP()
- [x] 1.2 Create cmd/util.go file with package cmd declaration
- [x] 1.3 Implement logv() function in cmd/util.go with level checking, indentation, and stderr output

## 2. Command Layer Implementation

- [x] 2.1 Update cmd/create.go to use logv() for level 1 "Creating worktree" message
- [x] 2.2 Update cmd/create.go to use logv() for level 2 detailed parameters (from branch, to path, in repo, creating parent dir)
- [x] 2.3 Remove all fmt.Fprintf DEBUG statements from cmd/create.go
- [x] 2.4 Update cmd/delete.go to use logv() for level 1 "Deleting worktree" message
- [x] 2.5 Update cmd/delete.go to use logv() for level 2 detailed parameters (project, branch, force)
- [x] 2.6 Update cmd/list.go to use logv() for level 1 "Listing worktrees" message
- [x] 2.7 Update cmd/list.go to use logv() for level 2 detailed parameters (project, repository, including main worktree)
- [x] 2.8 Update cmd/cd.go to use logv() for level 1 "Navigating to worktree" message
- [x] 2.9 Update cmd/cd.go to use logv() for level 2 detailed parameters (worktree path, resolved project)
- [x] 2.10 Update cmd/setup-shell.go to use logv() for level 1 "Setting up shell wrapper" message
- [x] 2.11 Update cmd/setup-shell.go to use logv() for level 2 detailed parameters (shell type, config file path)

## 3. Service Layer Cleanup

- [x] 3.1 Remove all fmt.Fprintf(os.Stderr, "DEBUG: ...") statements from internal/services/worktree_service.go
- [x] 3.2 Run grep -r "fmt.Fprintf.*os\.Stderr.*DEBUG" internal/services/ to verify no other debug output exists

## 4. Testing

- [x] 4.1 Manually test twiggit create with -v flag (level 1 output only)
- [x] 4.2 Manually test twiggit create with -vv flag (level 1 + level 2 output)
- [x] 4.3 Manually test twiggit create with no verbose flag (normal output only)
- [x] 4.4 Manually test twiggit delete with -v and -vv flags
- [x] 4.5 Manually test twiggit list with -v and -vv flags
- [x] 4.6 Manually test twiggit cd with -v and -vv flags
- [x] 4.7 Manually test twiggit setup-shell with -v and -vv flags
- [x] 4.8 Verify pipe behavior: run `twiggit create <branch> -vv | cat` to confirm verbose output goes to stderr and normal output goes to stdout
- [x] 4.9 Verify no color codes in verbose output
- [x] 4.10 Verify no "DEBUG:" or "[VERBOSE]" prefixes in output
- [x] 4.11 Run E2E tests to ensure no regressions
- [x] 4.12 Run mise run check, mise run lint:fix, mise run test to verify all changes pass validation

## 5. Documentation

- [x] 5.1 Add verbose output guidelines to cmd/AGENTS.md
- [x] 5.2 Document logv() function usage in cmd/AGENTS.md with examples
- [x] 5.3 Add guidance on when to use level 1 vs level 2 in cmd/AGENTS.md
