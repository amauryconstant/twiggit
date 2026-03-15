## Purpose
Race detector validation for concurrent worktree operations

## Build Tag

```go
//go:build concurrent
```

Run with: `mise run test:race` or `go test -race -tags=concurrent ./test/concurrent/`

## Test Patterns

### Concurrent Operations on Same Project

```go
// Multiple goroutines listing worktrees on same project
gomega.Eventually(func() bool {
    return concurrentListOperations(t, repoPath, numGoroutines)
}, 5*time.Second, 100*time.Millisecond).Should(gomega.BeTrue())
```

### Concurrent Create/Delete

```go
// Create and delete different worktrees concurrently
var wg sync.WaitGroup
for i := 0; i < 10; i++ {
    wg.Add(2)
    go func(idx int) {
        defer wg.Done()
        // Create worktree
    }(i)
    go func(idx int) {
        defer wg.Done()
        // Delete worktree
    }(i)
}
wg.Wait()
```

## Quality Requirements

- All tests must pass race detector
- Use deterministic test data
- Ensure proper cleanup in all paths
- Keep tests under 5 seconds

## What to Test

| Scenario | Description |
|----------|-------------|
| Concurrent list | Multiple goroutines listing same project |
| Concurrent create | Creating different worktrees simultaneously |
| Concurrent delete | Deleting different worktrees simultaneously |
| Mixed create/delete | Create and delete operations interleaved |
| Prune while list | Prune operation during list |
