# 07 — Testing Strategy

This reference details the CLI-specific testing **pyramid**. The underlying Go
testing idioms — named subtests, `//go:build integration` tags, `t.Parallel()`,
`goleak`, fuzzing, `testing/synctest` — are owned by
`samber/cc-skills-golang@golang-testing`; testify's `assert`/`require` API by
`samber/cc-skills-golang@golang-stretchr-testify`. Load them alongside.

---

## The Four Levels

Same pyramid as `SKILL.md`, bottom (most tests) to top (fewest):

```
 ╱ E2E (few)           — full binary, -cover, testscript
╱  Integration (some)   — real I/O, //go:build integration
╱   Command (many)       — runF injection, captured output, golden files
╱    Core (most)          — table-driven, no mocks, pure values
```

### Core (unit) — pure logic

Standard Go table-driven tests against the functional core — validation, transformation, computation. No I/O, no mocks needed. Every case gets a `name` passed to `t.Run`.

```go
tests := []struct {
    name    string
    target  string
    tag     string
    want    *core.DeployResult
    wantErr bool
}{
    {"valid", "staging", "v1", &core.DeployResult{...}, false},
    {"empty target", "", "v1", nil, true},
}

for _, tt := range tests {
    t.Run(tt.name, func(t *testing.T) {
        got, err := core.Deploy(tt.target, tt.tag, false)
        if (err != nil) != tt.wantErr {
            t.Errorf("Deploy() error = %v, wantErr %v", err, tt.wantErr)
        }
        // compare got vs tt.want
    })
}
```

This is where the Functional Core / Imperative Shell split pays off. Pure core functions are trivially testable.

### Command tests — runF injection

The `runF` parameter on each `NewCmdXxx` constructor lets a test receive the
fully-parsed options and bypass Cobra, or run the real `runXxx` with test doubles
for the Factory's dependencies. This is the CLI-specific heart of the pyramid. (The
`runF` mechanism itself is introduced in `SKILL.md` §The Command Pattern; here we test
through it.)

```go
func TestRunCreate(t *testing.T) {
    ios, _, stdout, _ := iostreams.Test()
    f := &cmdutil.Factory{IOStreams: ios}

    var gotOpts *CreateOptions
    cmd := NewCmdCreate(f, func(opts *CreateOptions) error {
        gotOpts = opts
        return nil
    })
    cmd.SetArgs([]string{"feature/test", "--source", "develop"})
    require.NoError(t, cmd.Execute())

    assert.Equal(t, "feature/test", gotOpts.Name)
    assert.Equal(t, "develop", gotOpts.Source)
    _ = stdout
}
```

Capturing `stdout`/`stderr` here and diffing against `testdata/*.golden` is the
**golden-file** technique (see below) — a technique within this tier, not a
separate tier.

### Integration Tests (Component Boundaries)

Test the wiring between components with real I/O, gated by `//go:build integration`. Use interface mocks for external dependencies.

**Key libraries:**
- `testify` — assertions, practically universal
- `go-cmp` (Google) — deep equality with diff output, more flexible than `reflect.DeepEqual`
- `jarcoal/httpmock` — HTTP request/response mocking
- `afero` — in-memory filesystem for file operations
- `testcontainers-go` — real Docker containers for database/service testing

## Golden Files (technique, used in Command + E2E tiers)

Capture the full stdout/stderr output of a command and compare against a stored reference file. This is a technique applied within the Command and E2E tiers, not a tier of its own.

**How it works:**

1. Run the CLI function with known args
2. Capture output (stdout, stderr, exit code)
3. Compare against `testdata/<test-name>.golden`
4. Update with `-update` flag: `go test ./... -update`

```go
func TestDeployCommand(t *testing.T) {
    tests := []struct {
        name    string
        args    []string
        fixture string
    }{
        {"no arguments", []string{}, "deploy-no-args.golden"},
        {"dry run", []string{"--dry-run", "staging"}, "deploy-dry-run.golden"},
        {"json output", []string{"-o", "json", "staging"}, "deploy-json.golden"},
    }
    for _, tt := range tests {
        t.Run(tt.name, func(t *testing.T) {
            ios, _, stdout, _ := iostreams.Test()
            f := &cmdutil.Factory{IOStreams: ios}

            cmd := NewCmdDeploy(f, nil) // nil runF → real runDeploy
            cmd.SetArgs(tt.args)
            _ = cmd.Execute()

            // compare stdout.String() against testdata/<tt.fixture>
        })
    }
}
```

**Libraries:**
- `gotest.tools/v3/golden` — golden file utilities with `-update` flag
- `sebdah/goldie` — supports Go templates in golden files for dynamic values

### E2E Tests (Full Binary)

Build and run the actual binary as a subprocess. This tests the real `main()` path including signal handling, exit codes, and environment variable resolution.

Luca Pette's pattern: build the binary with `-cover` (Go 1.20+), run as subprocess, assert output against golden files, get coverage data from the instrumented binary.

```go
func TestCliArgs(t *testing.T) {
    binary := buildTestBinary(t)
    
    tests := []struct {
        name     string
        args     []string
        wantCode int
        fixture  string
    }{
        {"version", []string{"version"}, 0, "version.golden"},
        {"unknown command", []string{"bogus"}, 2, "unknown-cmd.golden"},
    }
    for _, tt := range tests {
        t.Run(tt.name, func(t *testing.T) {
            cmd := exec.Command(binary, tt.args...)
            output, err := cmd.CombinedOutput()
            exitCode := cmd.ProcessState.ExitCode()
            // assert exitCode == tt.wantCode
            // compare output against golden file
        })
    }
}
```

---

## Testing IOStreams

Because IOStreams wraps `io.Writer`/`io.Reader`, testing is straightforward. Use
the canonical `iostreams.Test()` constructor (defined in `SKILL.md`) — it returns
the streams plus the in/out/err buffers, with a non-TTY, no-color configuration:

```go
ios, in, out, errOut := iostreams.Test()
// ios.IsStdoutTTY() == false, colors disabled — the non-interactive path
```

This naturally exercises the non-TTY code path.

---

## TTY vs. Non-TTY Testing

CLIs SHOULD behave differently when stdout is not a terminal. Test both paths:

- **Non-TTY (default in tests):** `iostreams.Test()` gives buffers with `isTTY`
  false — no colors, no spinners, machine-parseable output.
- **TTY (explicit):** to verify colored/formatted output, construct an IOStreams
  with the TTY flag set (add a test helper such as `iostreams.TestTTY()` that
  flips `isTTY`), since `isTTY` is unexported and set by the constructor.

`golang.org/x/term.IsTerminal(fd)` detects TTY in `iostreams.System()`. In tests, output goes to a buffer (non-TTY), so your non-interactive code path is naturally exercised.

---

## Testing Interactive / TUI Commands

Bubble Tea provides `tea.WithInput()` and `tea.WithOutput()` for programmatic key sequences. The `huh` library's accessible mode (standard prompts) is testable via stdin/stdout redirection.

For commands using flags-first/prompt-as-fallback, test both paths: with flags provided (no prompting) and with stdin simulating user input.

---

## What to Test at Each Architecture Tier

Rows are the architecture tiers from [01-architecture.md](./01-architecture.md); columns are the test levels above.

| Arch tier | Core (unit) | Command (runF) | Integration | E2E |
|-----------|-------------|----------------|-------------|-----|
| 1 | `appEnv.run()` unit tests | — | Golden files against `run()` | Optional |
| 2 | Core unit tests | runF + captured output (golden) | IOStreams + mocked deps | Optional |
| 3 | Core unit tests | runF per command (golden) | Component integration | Binary E2E with exit codes |

---

**See also:** `samber/cc-skills-golang@golang-testing` — broader testing guide: fuzzing, benchmarks, goroutine leak detection with `goleak`, snapshot testing, and CI integration · `samber/cc-skills-golang@golang-stretchr-testify` — full `assert`/`require`/`mock`/`suite` API depth
