# Spec Delta

## MODIFIED Requirements

### Requirement: --help groups subcommands

The root `twiggit` command SHALL assign subcommands to groups via
`cobra.Command.AddGroup` so `--help` output is scannable. Groups
SHALL be registered before the `AddCommand` calls that reference
them; cobra does not retroactively assign groups.

| Group ID | Members |
|---|---|
| core | list, create, delete, prune, rebase, sync |
| navigation | cd |
| setup | init |
| meta | version, completion |

#### Scenario: --help shows four groups

- **WHEN** the user runs `twiggit --help`
- **THEN** the output contains the four group headers `Core:`,
  `Navigation:`, `Setup:`, `Meta:` with the respective subcommands
  under each
- **AND** the `Core:` group SHALL include `rebase` and `sync` as
  members

#### Scenario: Group registered before subcommand

- **WHEN** `cmd.AddGroup` is called after `cmd.AddCommand` for a
  subcommand
- **THEN** that subcommand does NOT appear under the group in
  `--help` output (cobra does not retroactively assign groups)