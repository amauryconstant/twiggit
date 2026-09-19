# 06 — Concurrency

Relevant when CLIs perform I/O-bound work: HTTP requests, file processing, parallel downloads, multi-target deployments.

---

## Signal Handling and Graceful Shutdown

This is the most important concurrency pattern for CLIs. Every CLI that does non-trivial work MUST handle signals:

```go
ctx, stop := signal.NotifyContext(context.Background(),
    os.Interrupt, syscall.SIGTERM)
defer stop()
```

Pass `ctx` through your entire call chain. All blocking operations should respect `ctx.Done()`. On signal: cancel in-flight work, flush buffers, print summary, exit with code 130 (standard for SIGINT).

---

## errgroup for Structured Concurrency

`golang.org/x/sync/errgroup` is the recommended default for bounded concurrent work in CLIs:

```go
g, ctx := errgroup.WithContext(ctx)
g.SetLimit(10) // max concurrent goroutines

for _, item := range items {
    g.Go(func() error {
        return process(ctx, item)
    })
}
if err := g.Wait(); err != nil {
    return err
}
```

`errgroup` provides: bounded concurrency via `SetLimit`, context cancellation on first error, and structured error collection. It's sufficient for the majority of CLI concurrency needs.

---

## Worker Pool

For more control over the job/result flow (e.g., you need to process results as they arrive), use a channel-based worker pool:

```go
func worker(ctx context.Context, jobs <-chan Task, results chan<- Result) {
    for j := range jobs {
        select {
        case <-ctx.Done():
            return
        case results <- process(j):
        }
    }
}
```

Launch N workers (typically `runtime.NumCPU()` for CPU-bound, a fixed limit for I/O-bound), send jobs into the channel, collect results.

---

## Pipeline

Stages connected by channels, each running concurrently. Natural fit for data transformation CLIs (parse → transform → output):

```go
func generate(ctx context.Context, items []string) <-chan string { ... }
func transform(ctx context.Context, in <-chan string) <-chan Result { ... }
func collect(ctx context.Context, in <-chan Result) ([]Result, error) { ... }

// Usage
results, err := collect(ctx, transform(ctx, generate(ctx, items)))
```

---

## Fan-Out / Fan-In

Distribute work across multiple goroutines (fan-out), then merge results into a single channel (fan-in). Useful for parallelizing HTTP calls, file hashing, or any embarrassingly parallel workload.

---

## Bounded Concurrency with Semaphore

For simple cases where `errgroup` is too heavy:

```go
sem := make(chan struct{}, runtime.NumCPU())
var wg sync.WaitGroup

for _, file := range files {
    sem <- struct{}{} // acquire
    wg.Add(1)
    go func(f string) {
        defer func() { <-sem; wg.Done() }()
        processFile(f)
    }(file)
}
wg.Wait()
```

---

## When to Use What

| Pattern | Use Case |
|---------|----------|
| `errgroup` | Default choice for bounded concurrent I/O |
| Worker pool | Need result streaming or complex job routing |
| Pipeline | Staged data transformation (parse → transform → output) |
| Fan-out/fan-in | Embarrassingly parallel tasks with merged output |
| Semaphore + WaitGroup | Simple bounded concurrency without error collection |

---

**See also:** `samber/cc-skills-golang@golang-concurrency` — full concurrency patterns: channels, `singleflight`, ownership rules, goroutine leak detection, and race condition prevention
