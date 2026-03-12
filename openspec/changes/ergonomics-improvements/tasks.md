## 1. Command Aliases

- [ ] 1.1 Add `Aliases: []string{"ls"}` to list command in `cmd/list.go`
- [ ] 1.2 Add `Aliases: []string{"rm"}` to delete command in `cmd/delete.go`

## 2. Auto-Confirmation Flag for Prune

- [ ] 2.1 Add `--yes/-y` boolean flag to prune command in `cmd/prune.go`
- [ ] 2.2 Modify `executePrune` to accept yes flag parameter
- [ ] 2.3 Update bulk confirmation logic to skip prompt when yes flag is set
- [ ] 2.4 Update Long description to document `--yes` vs `--force` distinction

## 3. Short Flag for List --all

- [ ] 3.1 Change `--all` flag registration from `BoolVar` to `BoolVarP` with short form `-a` in `cmd/list.go`

## 4. Fix Help Text Duplication

- [ ] 4.1 Remove "Flags:" section from Long description in `cmd/create.go`
- [ ] 4.2 Add "Examples:" section to Long description in `cmd/create.go`
- [ ] 4.3 Remove "Flags:" section from Long description in `cmd/init.go`

## 5. Add Examples to Commands

- [ ] 5.1 Add "Examples:" section to Long description in `cmd/list.go`
- [ ] 5.2 Add "Examples:" section to Long description in `cmd/delete.go`

## 6. Testing

- [ ] 6.1 Add E2E test: `twiggit ls` behaves like `twiggit list`
- [ ] 6.2 Add E2E test: `twiggit rm <target>` behaves like `twiggit delete <target>`
- [ ] 6.3 Add E2E test: `twiggit prune --all --yes` skips confirmation prompt
- [ ] 6.4 Add E2E test: `twiggit list -a` behaves like `twiggit list --all`
- [ ] 6.5 Verify aliases appear in help text output
