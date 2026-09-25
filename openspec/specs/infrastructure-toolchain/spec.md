# Capability: Toolchain Ownership and Verify Composition

## Purpose

Pin the contract that separates code-only dependencies from tool
binaries, and lock the ordering of every maintenance gate invoked by
`mise run verify`. Cross-cuts `infrastructure-release` (CI) and the
developer workflow (pre-commit, local verify).

## Requirements

### Requirement: Tool ownership separation

Tool binaries SHALL be pinned exclusively in `.mise/config.toml`
`[tools]`. The `go.mod` file SHALL carry only code dependencies —
packages imported by Go source — and SHALL NOT contain a `tool (...)`
block. Adding a tool dependency to `go.mod tool` is a spec violation.

The tool set governed by this rule includes, at minimum:
govulncheck, gocover-cobertura, gocovmerge, gopls, the ginkgo CLI,
golangci-lint, goreleaser, and pre-commit.

No task, pre-commit hook, or CI script SHALL invoke a tool via
`go run <module>@<version>` (network fetch); tool binaries SHALL come
from the mise PATH shim, fall back to the `Tool not found` exit, and
never silently fetch a version.

#### Scenario: `go.mod` carries no tool block

- **WHEN** a fresh checkout is inspected
- **THEN** `go.mod` SHALL contain no `tool (...)` block
- **AND** every entry in the `require` blocks SHALL be a package
  imported from a `.go` source file in the module

#### Scenario: Tool binaries resolve via mise

- **WHEN** a developer runs `mise install`
- **THEN** every binary referenced from a `[tasks.*]` `run` entry
  SHALL be reachable as a bare command on `$PATH`
- **AND** running the same task without `mise install` SHALL fail with
  `command not found` rather than fetching a substitute

#### Scenario: No `go run ...@latest` in tasks or pre-commit

- **WHEN** `.mise/config.toml` and `.pre-commit-config.yaml` are
  scanned for tool invocations
- **THEN** no entry SHALL match the `go run .*@` form

### Requirement: `verify` task composition

`mise run verify` SHALL chain, in this order:

1. `format` — rewrites files in place using `gofmt -w .`
2. `lint:fix` — auto-fixes resolvable lint violations via
   `golangci-lint run --fix`
3. `lint:gated` — read-only verification: `gopls check **/*.go`
   followed by `golangci-lint run`
4. `vuln:check` — scans for known vulnerabilities via
   `govulncheck ./...`
5. `test` — runs unit, integration, race, and e2e suites
6. `build` — produces the `bin/twiggit` binary

Each step SHALL exit non-zero on failure; subsequent steps SHALL NOT
run after a failure. Steps 1 and 2 mutate the working tree; step 3
SHALL be read-only.

#### Scenario: Lint violation aborts at the lint gate

- **WHEN** a deliberate lint violation is introduced into a `.go`
  file
- **AND** the previous mutating steps run successfully
- **THEN** `lint:gated` SHALL exit non-zero
- **AND** `vuln:check`, `test`, and `build` SHALL NOT run

#### Scenario: Vulnerability finding aborts at the vuln gate

- **WHEN** a deliberate vulnerable import is added to a Go source
  file
- **AND** all preceding verify steps pass
- **THEN** `vuln:check` SHALL exit non-zero
- **AND** `test` and `build` SHALL NOT run

#### Scenario: Format drift is repaired and the chain completes

- **WHEN** a file is intentionally mis-formatted
- **THEN** `format` SHALL rewrite it
- **AND** the chain SHALL continue without manual intervention
- **AND** the final exit code SHALL be 0 iff every step succeeded

### Requirement: `lint:gated` task combines gopls and golangci-lint

gopls SHALL be a first-class linter alongside golangci-lint. The
`mise run lint:gated` task SHALL run both tools in sequence:

1. `gopls check **/*.go` — type and semantic diagnostics across the
   module
2. `golangci-lint run` — the configured linter/formatter set

Failure of either tool SHALL fail the gate. Each tool SHALL also be
invokable on its own:

- `mise run lint:check` — golangci-lint only
- `mise run gopls:check` — gopls diagnostics only
- `mise run gopls:stats` — gopls analysis state summary

The gopls binary SHALL be provisioned by mise (PATH shim); running
the task without `mise install` SHALL fail with `command not found`.

#### Scenario: gopls flags an issue, golangci-lint does not

- **WHEN** a type or semantic error exists that gopls catches but
  golangci-lint does not
- **THEN** `mise run lint:gated` SHALL exit non-zero
- **AND** the error message SHALL surface the gopls diagnostic

#### Scenario: Both linters clean

- **WHEN** no issues exist
- **THEN** `mise run lint:gated` SHALL exit 0
- **AND** no file under the module SHALL be modified (read-only)

#### Scenario: Individual entry points

- **WHEN** a developer runs only `mise run gopls:check`
- **THEN** gopls diagnostics SHALL run and exit 0 on a clean module
- **AND** golangci-lint SHALL NOT be invoked
- **AND** the equivalent holds for `mise run lint:check` in the
  inverse direction

### Requirement: `vuln:check` task uses mise-installed binary

The `mise run vuln:check` task SHALL invoke the bare `govulncheck`
command (resolved through the mise PATH shim) and SHALL NOT use
`go run ...@latest`. The pre-commit hook that runs govulncheck SHALL
invoke the same bare binary; no pre-commit hook SHALL ever call
`go run` to fetch a tool.

#### Scenario: Bare govulncheck under mise

- **WHEN** a developer runs `mise install` then `mise run vuln:check`
- **THEN** the task SHALL invoke the mise-pinned `govulncheck`
  binary
- **AND** no module-fetch from the network SHALL occur at task
  execution time

#### Scenario: Govulncheck hook resolves via PATH

- **WHEN** the pre-commit hook for govulncheck runs in a directory
  where `mise install` has provisioned the toolchain
- **THEN** the hook SHALL invoke the bare `govulncheck` command
- **AND** `language: system` SHALL be configured so the hook reads
  `$PATH`

#### Scenario: Missing tool is not silently fetched

- **WHEN** `mise run vuln:check` runs in an environment where
  `mise install` was not performed
- **THEN** the task SHALL fail with a `command not found` error
- **AND** no `@latest` fallback SHALL be attempted
