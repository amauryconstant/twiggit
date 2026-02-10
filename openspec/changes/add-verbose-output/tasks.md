## 1. Foundation

- [ ] 1.1 Add persistent `--verbose` flag to cmd/root.go using PersistentFlags().CountP()
- [ ] 1.2 Create cmd/util.go file with package cmd declaration
- [ ] 1.3 Implement logv() function in cmd/util.go with level checking, indentation, and stderr output

## 2. Command Layer Implementation

- [ ] 2.1 Update cmd/create.go to use logv() for level 1 "Creating worktree" message
- [ ] 2.2 Update cmd/create.go to use logv() for level 2 detailed parameters (from branch, to path, in repo, creating parent dir)
- [ ] 2.3 Remove all fmt.Fprintf DEBUG statements from cmd/create.go
- [ ] 2.4 Update cmd/delete.go to use logv() for level 1 "Deleting worktree" message
- [ ] 2.5 Update cmd/delete.go to use logv() for level 2 detailed parameters (project, branch, force)
- [ ] 2.6 Update cmd/list.go to use logv() for level 1 "Listing worktrees" message
- [ ] 2.7 Update cmd/list.go to use logv() for level 2 detailed parameters (project, repository, including main worktree)
- [ ] 2.8 Update cmd/cd.go to use logv() for level 1 "Navigating to worktree" message
- [ ] 2.9 Update cmd/cd.go to use logv() for level 2 detailed parameters (worktree path, resolved project)
- [ ] 2.10 Update cmd/setup-shell.go to use logv() for level 1 "Setting up shell wrapper" message
- [ ] 2.11 Update cmd/setup-shell.go to use logv() for level 2 detailed parameters (shell type, config file path)

## 3. Service Layer Cleanup

- [ ] 3.1 Remove all fmt.Fprintf(os.Stderr, "DEBUG: ...") statements from internal/services/worktree_service.go
- [ ] 3.2 Run grep -r "fmt.Fprintf.*os\.Stderr.*DEBUG" internal/services/ to verify no other debug output exists

## 4. Testing

- [ ] 4.1 Manually test twiggit create with -v flag (level 1 output only)
- [ ] 4.2 Manually test twiggit create with -vv flag (level 1 + level 2 output)
- [ ] 4.3 Manually test twiggit create with no verbose flag (normal output only)
- [ ] 4.4 Manually test twiggit delete with -v and -vv flags
- [ ] 4.5 Manually test twiggit list with -v and -vv flags
- [ ] 4.6 Manually test twiggit cd with -v and -vv flags
- [ ] 4.7 Manually test twiggit setup-shell with -v and -vv flags
- [ ] 4.8 Verify pipe behavior: run `twiggit create <branch> -vv | cat` to confirm verbose output goes to stderr and normal output goes to stdout
- [ ] 4.9 Verify no color codes in verbose output
- [ ] 4.10 Verify no "DEBUG:" or "[VERBOSE]" prefixes in output
- [ ] 4.11 Run E2E tests to ensure no regressions
- [ ] 4.12 Run mise run check, mise run lint:fix, mise run test to verify all changes pass validation

## 5. Documentation

- [ ] 5.1 Add verbose output guidelines to cmd/AGENTS.md
- [ ] 5.2 Document logv() function usage in cmd/AGENTS.md with examples
- [ ] 5.3 Add guidance on when to use level 1 vs level 2 in cmd/AGENTS.md
