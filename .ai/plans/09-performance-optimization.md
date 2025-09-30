# Performance Optimization Layer Implementation Plan

## Purpose

Implement performance optimization layer for twiggit with caching, monitoring, and optimization features to meet performance targets while maintaining existing functionality.

## Context

This layer builds upon all established functional layers and testing infrastructure. Performance targets are optimization goals rather than hard requirements:

> "Startup time SHALL be <100ms on modern hardware" (implementation.md:74)
> "List operations SHALL complete in <500ms for <100 worktrees" (implementation.md:80)
> "Memory usage SHALL remain <50MB during operations" (implementation.md:83)

## Architecture Overview

### Core Components

- **Cache Interface**: Generic caching system with TTL support
- **Performance Monitor**: Metrics collection and timing wrapper
- **Memory Manager**: Memory-efficient data structures
- **Concurrent Operations**: Thread-safe parallel processing
- **Benchmarking Suite**: Performance testing utilities

### Integration Strategy

Performance optimizations SHALL be implemented as non-breaking wrappers around existing services, ensuring backward compatibility and gradual adoption.

## Implementation Steps

### Phase 1: Caching Infrastructure

#### 1.1 Generic Cache Interface
**File**: `internal/infrastructure/cache/interface.go`

```go
type Cache interface {
    Get(key string) (interface{}, bool)
    Set(key string, value interface{}, ttl time.Duration) error
    Delete(key string) error
    Clear() error
    InvalidatePattern(pattern string) error
}

type CacheItem struct {
    Value      interface{}
    Expiration time.Time
}
```

#### 1.2 In-Memory Cache Implementation
**File**: `internal/infrastructure/cache/memory.go`

- Thread-safe using sync.RWMutex
- TTL-based expiration with cleanup goroutine
- Pattern-based invalidation for filesystem changes
- Memory usage monitoring and limits

#### 1.3 Cache Configuration
**File**: `internal/domain/cache_config.go`

```go
type CacheConfig struct {
    RepositoryMetadata time.Duration `toml:"repository_metadata_ttl"`
    WorktreeList       time.Duration `toml:"worktree_list_ttl"`
    BranchInfo         time.Duration `toml:"branch_info_ttl"`
    ContextDetection   time.Duration `toml:"context_detection_ttl"`
    MaxMemoryMB        int           `toml:"max_memory_mb"`
    Enabled            bool          `toml:"enabled"`
}
```

#### 1.4 Cache TTL Configuration (from implementation.md:525-626)

```toml
[cache]
enabled = true
max_memory_mb = 32

[cache.ttl]
repository_metadata = "5m"    # Repository metadata: 5 minutes TTL
worktree_list = "2m"           # Worktree lists: 2 minutes TTL  
branch_info = "1m"             # Branch information: 1 minute TTL
context_detection = "30s"      # Context detection: 30 seconds TTL
```

### Phase 2: Performance Monitoring

#### 2.1 Metrics Collection Interface
**File**: `internal/infrastructure/metrics/interface.go`

```go
type Metrics interface {
    RecordOperation(name string, duration time.Duration, err error)
    RecordMemoryUsage(operation string, bytes int64)
    RecordCacheHit(cache string, hit bool)
    GetOperationStats(name string) OperationStats
}

type OperationStats struct {
    Count        int64
    TotalTime    time.Duration
    AverageTime  time.Duration
    ErrorCount   int64
    LastExecuted time.Time
}
```

#### 2.2 Performance Monitor Implementation
**File**: `internal/infrastructure/metrics/monitor.go`

- In-memory metrics storage
- Operation timing with error tracking
- Memory usage monitoring
- Cache hit/miss statistics
- Performance summary reporting

#### 2.3 Operation Wrapper
**File**: `internal/infrastructure/metrics/wrapper.go`

```go
func WithMetrics[T any](metrics Metrics, operation string, fn func() (T, error)) (T, error) {
    start := time.Now()
    result, err := fn()
    duration := time.Since(start)
    metrics.RecordOperation(operation, duration, err)
    return result, err
}
```

### Phase 3: Service Layer Optimization

#### 3.1 Cached Git Service Wrapper
**File**: `internal/infrastructure/git/cached_service.go`

```go
type CachedGitService struct {
    service GitService
    cache   Cache
    metrics Metrics
    config  CacheConfig
}

func (c *CachedGitService) ListWorktrees(repoPath string) ([]Worktree, error) {
    cacheKey := fmt.Sprintf("worktrees:%s", repoPath)
    
    if cached, found := c.cache.Get(cacheKey); found {
        c.metrics.RecordCacheHit("worktree_list", true)
        return cached.([]Worktree), nil
    }
    
    result, err := WithMetrics(c.metrics, "list_worktrees", func() ([]Worktree, error) {
        return c.service.ListWorktrees(repoPath)
    })
    
    if err == nil {
        c.cache.Set(cacheKey, result, c.config.WorktreeList)
        c.metrics.RecordCacheHit("worktree_list", false)
    }
    
    return result, err
}
```

#### 3.2 Cached Context Detection
**File**: `internal/domain/cached_context.go`

```go
type CachedContextDetector struct {
    detector ContextDetector
    cache    Cache
    metrics  Metrics
    config   CacheConfig
}

func (c *CachedContextDetector) DetectContext(path string) (Context, error) {
    cacheKey := fmt.Sprintf("context:%s", path)
    
    if cached, found := c.cache.Get(cacheKey); found {
        c.metrics.RecordCacheHit("context_detection", true)
        return cached.(Context), nil
    }
    
    result, err := WithMetrics(c.metrics, "detect_context", func() (Context, error) {
        return c.detector.DetectContext(path)
    })
    
    if err == nil {
        c.cache.Set(cacheKey, result, c.config.ContextDetection)
        c.metrics.RecordCacheHit("context_detection", false)
    }
    
    return result, err
}
```

#### 3.3 Filesystem Watcher for Cache Invalidation
**File**: `internal/infrastructure/cache/watcher.go`

```go
type CacheWatcher struct {
    cache   Cache
    watcher *fsnotify.Watcher
}

func (w *CacheWatcher) WatchRepository(repoPath string) error {
    return w.watcher.Add(repoPath)
}

func (w *CacheWatcher) handleEvent(event fsnotify.Event) {
    if event.Op&fsnotify.Write == fsnotify.Write || 
       event.Op&fsnotify.Create == fsnotify.Create ||
       event.Op&fsnotify.Remove == fsnotify.Remove {
        // Invalidate related cache entries
        w.cache.InvalidatePattern(fmt.Sprintf("*%s*", event.Name))
    }
}
```

### Phase 4: Memory Optimization

#### 4.1 Memory-Efficient Worktree Lists
**File**: `internal/domain/worktree_pool.go`

```go
type WorktreePool struct {
    pool sync.Pool
}

func NewWorktreePool() *WorktreePool {
    return &WorktreePool{
        pool: sync.Pool{
            New: func() interface{} {
                return make([]Worktree, 0, 50) // Pre-allocate for common case
            },
        },
    }
}

func (p *WorktreePool) Get() []Worktree {
    return p.pool.Get().([]Worktree)[:0] // Reset length but keep capacity
}

func (p *WorktreePool) Put(worktrees []Worktree) {
    if cap(worktrees) <= 100 { // Don't keep overly large slices
        p.pool.Put(worktrees)
    }
}
```

#### 4.2 Streaming Operations for Large Repositories
**File**: `internal/infrastructure/git/streaming_service.go`

```go
type StreamingGitService struct {
    service GitService
}

func (s *StreamingGitService) ListWorktreesStream(repoPath string) <-chan WorktreeResult {
    results := make(chan WorktreeResult, 10)
    
    go func() {
        defer close(results)
        // Stream worktrees one at a time to reduce memory footprint
        // Implementation depends on go-git capabilities
    }()
    
    return results
}
```

### Phase 5: Concurrent Operations

#### 5.1 Thread-Safe Concurrent Operations
**File**: `internal/infrastructure/git/concurrent_service.go`

```go
type ConcurrentGitService struct {
    service GitService
    limiter *semaphore.Weighted
}

func NewConcurrentGitService(maxConcurrent int) *ConcurrentGitService {
    return &ConcurrentGitService{
        limiter: semaphore.NewWeighted(int64(maxConcurrent)),
    }
}

func (c *ConcurrentGitService) ListWorktreesMultiple(repoPaths []string) (map[string][]Worktree, error) {
    var wg sync.WaitGroup
    results := make(map[string][]Worktree)
    resultsMu := sync.Mutex{}
    errors := make([]error, 0)
    errorsMu := sync.Mutex{}
    
    for _, repoPath := range repoPaths {
        wg.Add(1)
        go func(path string) {
            defer wg.Done()
            
            c.limiter.Acquire(context.Background(), 1)
            defer c.limiter.Release(1)
            
            worktrees, err := c.service.ListWorktrees(path)
            if err != nil {
                errorsMu.Lock()
                errors = append(errors, fmt.Errorf("%s: %w", path, err))
                errorsMu.Unlock()
                return
            }
            
            resultsMu.Lock()
            results[path] = worktrees
            resultsMu.Unlock()
        }(repoPath)
    }
    
    wg.Wait()
    
    if len(errors) > 0 {
        return results, fmt.Errorf("errors in %d repositories: %v", len(errors), errors)
    }
    
    return results, nil
}
```

### Phase 6: Benchmarking and Profiling

#### 6.1 Benchmarking Utilities
**File**: `internal/testing/benchmark/benchmark.go`

```go
type BenchmarkSuite struct {
    metrics Metrics
}

func (b *BenchmarkSuite) BenchmarkListWorktrees(repoPath string, iterations int) BenchmarkResult {
    var durations []time.Duration
    
    for i := 0; i < iterations; i++ {
        start := time.Now()
        // Execute list worktrees operation
        duration := time.Since(start)
        durations = append(durations, duration)
    }
    
    return BenchmarkResult{
        Operation:    "list_worktrees",
        Iterations:   iterations,
        MinDuration:  min(durations),
        MaxDuration:  max(durations),
        AvgDuration:  average(durations),
        P95Duration:  percentile(durations, 0.95),
        P99Duration:  percentile(durations, 0.99),
    }
}
```

#### 6.2 Performance Profiling Integration
**File**: `internal/testing/profiler/profiler.go`

```go
type Profiler struct {
    enabled bool
    output  string
}

func (p *Profiler) StartCPUProfile() error {
    if !p.enabled {
        return nil
    }
    
    f, err := os.Create(p.output + ".cpu.prof")
    if err != nil {
        return err
    }
    
    return pprof.StartCPUProfile(f)
}

func (p *Profiler) StopCPUProfile() {
    if p.enabled {
        pprof.StopCPUProfile()
    }
}

func (p *Profiler) WriteHeapProfile() error {
    if !p.enabled {
        return nil
    }
    
    f, err := os.Create(p.output + ".mem.prof")
    if err != nil {
        return err
    }
    defer f.Close()
    
    return pprof.WriteHeapProfile(f)
}
```

### Phase 7: Integration and Configuration

#### 7.1 Performance Service Integration
**File**: `internal/infrastructure/performance/service.go`

```go
type PerformanceService struct {
    cache    Cache
    metrics  Metrics
    profiler Profiler
    config   PerformanceConfig
}

func NewPerformanceService(config PerformanceConfig) *PerformanceService {
    cache := NewMemoryCache(config.Cache)
    metrics := NewMonitor()
    profiler := NewProfiler(config.Profiling)
    
    return &PerformanceService{
        cache:    cache,
        metrics:  metrics,
        profiler: profiler,
        config:   config,
    }
}

func (p *PerformanceService) WrapGitService(service GitService) GitService {
    if p.config.Cache.Enabled {
        return NewCachedGitService(service, p.cache, p.metrics, p.config.Cache)
    }
    return service
}

func (p *PerformanceService) WrapContextDetector(detector ContextDetector) ContextDetector {
    if p.config.Cache.Enabled {
        return NewCachedContextDetector(detector, p.cache, p.metrics, p.config.Cache)
    }
    return detector
}
```

#### 7.2 Performance Configuration
**File**: `internal/domain/performance_config.go`

```go
type PerformanceConfig struct {
    Cache      CacheConfig      `toml:"cache"`
    Profiling  ProfilingConfig  `toml:"profiling"`
    Concurrency ConcurrencyConfig `toml:"concurrency"`
}

type ProfilingConfig struct {
    Enabled bool   `toml:"enabled"`
    Output  string `toml:"output"`
}

type ConcurrencyConfig struct {
    MaxConcurrent int `toml:"max_concurrent"`
    Enabled       bool `toml:"enabled"`
}
```

#### 7.3 Configuration Integration
**File**: `internal/infrastructure/config/performance.go`

Add performance configuration to existing config loading:

```toml
[performance]
[performance.cache]
enabled = true
max_memory_mb = 32

[performance.cache.ttl]
repository_metadata = "5m"
worktree_list = "2m"
branch_info = "1m"
context_detection = "30s"

[performance.profiling]
enabled = false
output = "/tmp/twiggit-profile"

[performance.concurrency]
enabled = true
max_concurrent = 4
```

## Testing Strategy

### Unit Tests

#### Cache Testing
- TTL expiration behavior
- Thread safety under concurrent access
- Memory usage limits
- Pattern-based invalidation

#### Metrics Testing
- Accurate timing measurements
- Error counting
- Memory usage tracking
- Cache hit/miss statistics

#### Performance Wrapper Testing
- Correct fallback on cache miss
- Proper error handling
- Metrics collection accuracy

### Integration Tests

#### Cache Integration
- End-to-end caching with real git operations
- Cache invalidation on filesystem changes
- Performance improvements measurement

#### Concurrency Testing
- Thread-safe operations under load
- Resource limit enforcement
- Performance under concurrent load

### Performance Benchmarks

#### Benchmark Suite
- Startup time measurement
- List operation performance
- Memory usage profiling
- Cache effectiveness measurement

#### Regression Testing
- Performance targets validation
- Memory leak detection
- Cache efficiency monitoring

## Implementation Files

### New Files
```
internal/infrastructure/cache/
├── interface.go
├── memory.go
├── watcher.go
└── cache_test.go

internal/infrastructure/metrics/
├── interface.go
├── monitor.go
├── wrapper.go
└── metrics_test.go

internal/infrastructure/performance/
├── service.go
└── service_test.go

internal/infrastructure/git/
├── cached_service.go
├── concurrent_service.go
├── streaming_service.go
└── performance_test.go

internal/domain/
├── cache_config.go
├── performance_config.go
├── cached_context.go
├── worktree_pool.go
└── performance_test.go

internal/testing/benchmark/
├── benchmark.go
└── benchmark_test.go

internal/testing/profiler/
├── profiler.go
└── profiler_test.go
```

### Modified Files
```
internal/infrastructure/config/
├── config.go (add performance config loading)

cmd/
├── root.go (add performance flags)
└── list.go (add performance monitoring)
```

## Performance Targets and Monitoring

### Target Metrics
- **Startup Time**: <100ms (optimization target)
- **List Operations**: <500ms for <50 worktrees (optimization target)
- **Memory Usage**: <50MB during operations (optimization target)
- **Cache Hit Rate**: >80% for repeated operations

### Monitoring Implementation
- Real-time metrics collection
- Performance summary reporting
- Cache efficiency tracking
- Memory usage monitoring

### Performance Flags
```bash
twiggit --performance-stats list    # Show performance metrics
twiggit --profile-cpu list          # Enable CPU profiling
twiggit --profile-memory list       # Enable memory profiling
twiggit --cache-disable list        # Disable caching
twiggit --concurrency-disable list  # Disable concurrent operations
```

## Migration Strategy

### Phase 1: Infrastructure (Week 1)
- Implement cache interface and memory cache
- Add metrics collection system
- Create performance configuration

### Phase 2: Service Integration (Week 2)
- Wrap existing services with caching layer
- Add performance monitoring wrappers
- Implement cache invalidation

### Phase 3: Optimization (Week 3)
- Add memory-efficient data structures
- Implement concurrent operations
- Add streaming for large repositories

### Phase 4: Testing and Validation (Week 4)
- Comprehensive performance testing
- Benchmark suite implementation
- Performance target validation

## Success Criteria

### Functional Requirements
- All existing functionality preserved
- No breaking changes to public APIs
- Backward compatibility maintained

### Performance Requirements
- Measurable performance improvements
- Cache hit rates >80% for repeated operations
- Memory usage within optimization targets
- Startup time improvements >20%

### Quality Requirements
- Comprehensive test coverage (>90% for performance code)
- Performance regression tests passing
- Memory leak detection negative
- Thread safety validation passing

## Summary

This implementation plan provides a comprehensive performance optimization layer that enhances twiggit's speed and efficiency while maintaining all existing functionality. The modular design allows for gradual adoption and easy maintenance, with extensive testing and monitoring to ensure reliability and performance improvements.