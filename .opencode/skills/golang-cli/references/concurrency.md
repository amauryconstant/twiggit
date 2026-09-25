# Signals and Concurrency

Commands run sequentially until parallelism is measurably worth it: add concurrency when independent I/O calls add up to user-perceptible latency (>500 ms). Concurrency is an I/O orchestration concern, so it lives in `runX` or adapters; the core stays synchronous. General patterns → See `samber/cc-skills-golang@golang-concurrency`.

## Contents

- [Signals and shutdown](#signals-and-shutdown)
- [errgroup](#errgroup)
- [Streaming results](#streaming-results)
- [Choosing a pattern](#choosing-a-pattern)

## Signals and shutdown

`run()` in [main.go](../assets/examples/main.go) creates the root context:

```go
ctx, stop := signal.NotifyContext(context.Background(), os.Interrupt, syscall.SIGTERM)
defer stop()
err := cmd.NewRootCommand(f).ExecuteContext(ctx)
```

Every command passes `cmd.Context()` down to each blocking call. On Ctrl-C:

1. The context cancels; in-flight I/O returns `context.Canceled`.
2. Commands stop, flush partial output, report progress to stderr (`uploaded 12/40 files`), and return.
3. `run()` sees `ctx.Err() != nil` and returns 130 without printing the cancellation error.

A second Ctrl-C after `stop()` has run gets Go's default behavior and kills the process immediately — the escape hatch for a stuck shutdown.

## errgroup

`golang.org/x/sync/errgroup` is the one concurrency primitive: shared context, cancellation of siblings on the first error, and a built-in worker pool through `SetLimit`.

Independent calls in parallel:

```go
g, ctx := errgroup.WithContext(opts.Ctx)
var info *core.RepoInfo
var items []core.Worktree
g.Go(func() (err error) { info, err = client.RepoInfo(ctx); return })
g.Go(func() (err error) { items, err = client.ListWorktrees(ctx); return })
if err := g.Wait(); err != nil {
    return err
}
```

Bounded fan-out over N items:

```go
g, ctx := errgroup.WithContext(opts.Ctx)
g.SetLimit(8) // concurrent I/O calls; tune to the remote's rate limit
for _, item := range items {
    g.Go(func() error { return process(ctx, item) })
}
return g.Wait()
```

## Streaming results

When results must print as they arrive (downloads, per-file status), send them through a channel drained by the command, with producers in an errgroup:

```go
results := make(chan Result)
g, ctx := errgroup.WithContext(opts.Ctx)
g.SetLimit(8)
go func() {
    for _, item := range items {
        g.Go(func() error {
            r, err := process(ctx, item)
            if err == nil {
                select {
                case results <- r:
                case <-ctx.Done():
                    return ctx.Err()
                }
            }
            return err
        })
    }
    _ = g.Wait() // the error is returned by the final Wait below
    close(results)
}()
for r := range results {
    fmt.Fprintln(opts.IO.Out, r)
}
return g.Wait()
```

## Choosing a pattern

| Need | Use |
| --- | --- |
| A few independent calls | `errgroup.WithContext` |
| N items, bounded parallelism | `errgroup` + `SetLimit(n)` |
| Print results as they complete | errgroup producers + one results channel |
| Staged transformation (parse → transform → write) | Channel pipeline, each stage owning its output channel |
| Deduplicate identical concurrent calls | `golang.org/x/sync/singleflight` |

**Diagnose:** 1- `go test -race ./...` — expect no data races in commands that fan out 2- `go.uber.org/goleak` in `TestMain` — expect no goroutines left after a command returns
