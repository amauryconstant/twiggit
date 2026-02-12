# Go Concurrency Patterns

Go-specific patterns for context usage and concurrency in codebase auditing.

## Context Usage

### Patterns

- **First parameter**: `context.Context` as first parameter in exported functions
- **Background context**: Use `context.Background()` or `context.TODO()` when no parent context
- **Pass through**: Pass context through call chain, don't create new at each level
- **Timeouts**: Use `context.WithTimeout()` or `context.WithDeadline()` for time-bound operations

### Common Patterns

```go
// Service method with context
func (s *worktreeService) CreateWorktree(
    ctx context.Context,
    req *domain.CreateWorktreeRequest,
) (*domain.WorktreeInfo, error) {
    // Use ctx in all operations
    branch, err := s.gitClient.BranchExists(ctx, repoPath, req.BranchName)
    if err != nil {
        return nil, fmt.Errorf("failed to check branch: %w", err)
    }
    // ...
}

// HTTP client with context timeout
client := &http.Client{Timeout: 30 * time.Second}
req, _ := http.NewRequestWithContext(ctx, "GET", url, nil)
resp, err := client.Do(req)
```

### Anti-Patterns

- **Ignoring context**: Passing `context.Background()` when parent context available
- **Discarding context**: Not passing context through call chain
- **Context leaks**: Creating contexts without cancellation or timeout

## Concurrency

### Patterns

- **Goroutine lifecycle**: Ensure goroutines can be cancelled via context
- **Mutex for state**: Protect shared mutable state with `sync.Mutex`
- **Channels for coordination**: Use channels for goroutine communication
- **WaitGroups**: Use `sync.WaitGroup` for coordinating multiple goroutines

### Common Patterns

```go
// Goroutine with context
go func() {
    for {
        select {
        case <-ctx.Done():
            return // Cancel on context cancellation
        case <-ch:
            // Process work
        }
    }
}()

// Protecting shared state with mutex
type Service struct {
    mu sync.Mutex
    cache map[string]string
}

func (s *Service) Get(key string) string {
    s.mu.Lock()
    defer s.mu.Unlock()
    return s.cache[key]
}
```

### Anti-Patterns

- **Data races**: Concurrent access to shared mutable state without synchronization
- **Goroutine leaks**: Goroutines created but never terminated
- **Mutex contention**: Holding locks for too long in critical sections

## Audit-Specific Patterns

### For Concurrency Audits

- Check for goroutine lifecycle management (context cancellation)
- Identify data races (unprotected shared state)
- Look for goroutine leaks (unterminated goroutines)
- Verify proper mutex usage (defer unlock, lock scope)
- Check for race conditions (atomic operations where needed)
