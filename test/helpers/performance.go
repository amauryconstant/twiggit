package helpers

import (
	"runtime"
	"testing"
	"time"
)

// BenchmarkResult represents the result of a benchmark operation
type BenchmarkResult struct {
	Iterations    int           `json:"iterations"`
	TotalDuration time.Duration `json:"total_duration"`
	AvgDuration   time.Duration `json:"avg_duration"`
	MinDuration   time.Duration `json:"min_duration"`
	MaxDuration   time.Duration `json:"max_duration"`
	LastResult    interface{}   `json:"last_result,omitempty"`
}

// PerformanceTestHelper provides functional performance testing utilities
type PerformanceTestHelper struct {
	t          *testing.T
	iterations int
	warmup     bool
}

// NewPerformanceTestHelper creates a new PerformanceTestHelper instance
func NewPerformanceTestHelper(t *testing.T) *PerformanceTestHelper {
	t.Helper()
	return &PerformanceTestHelper{
		t:          t,
		iterations: 1,
		warmup:     false,
	}
}

// WithIterations sets the number of iterations for benchmarking
func (h *PerformanceTestHelper) WithIterations(iterations int) *PerformanceTestHelper {
	h.iterations = iterations
	return h
}

// WithWarmup enables or disables warmup iterations
func (h *PerformanceTestHelper) WithWarmup(warmup bool) *PerformanceTestHelper {
	h.warmup = warmup
	return h
}

// MeasureFunction measures the execution time of a function
func (h *PerformanceTestHelper) MeasureFunction(fn func()) (time.Duration, error) {
	start := time.Now()
	fn()
	duration := time.Since(start)

	return duration, nil
}

// BenchmarkFunction benchmarks a function over multiple iterations
func (h *PerformanceTestHelper) BenchmarkFunction(iterations int, fn func() interface{}) (*BenchmarkResult, error) {
	if iterations <= 0 {
		iterations = h.iterations
	}

	result := &BenchmarkResult{
		Iterations:  iterations,
		MinDuration: time.Hour, // Initialize with a large value
	}

	// Warmup iteration if enabled
	if h.warmup {
		fn()
	}

	// Run benchmark iterations
	var totalDuration time.Duration
	for i := 0; i < iterations; i++ {
		start := time.Now()
		fnResult := fn()
		duration := time.Since(start)

		totalDuration += duration
		result.LastResult = fnResult

		if duration < result.MinDuration {
			result.MinDuration = duration
		}
		if duration > result.MaxDuration {
			result.MaxDuration = duration
		}
	}

	result.TotalDuration = totalDuration
	result.AvgDuration = totalDuration / time.Duration(iterations)

	return result, nil
}

// MeasureMemoryUsage measures memory usage before and after function execution
func (h *PerformanceTestHelper) MeasureMemoryUsage(fn func()) (before, after uint64, err error) {
	// Force garbage collection before measurement
	runtime.GC()

	var m1, m2 runtime.MemStats
	runtime.ReadMemStats(&m1)
	before = m1.Alloc

	// Execute the function
	fn()

	// Force garbage collection after function execution
	runtime.GC()
	runtime.ReadMemStats(&m2)
	after = m2.Alloc

	return before, after, nil
}

// MeasureFunctionWithMemory measures both execution time and memory usage
func (h *PerformanceTestHelper) MeasureFunctionWithMemory(fn func()) (duration time.Duration, before, after uint64, err error) {
	// Measure memory before
	runtime.GC()
	var m1 runtime.MemStats
	runtime.ReadMemStats(&m1)
	before = m1.Alloc

	// Measure execution time
	start := time.Now()
	fn()
	duration = time.Since(start)

	// Measure memory after
	runtime.GC()
	var m2 runtime.MemStats
	runtime.ReadMemStats(&m2)
	after = m2.Alloc

	return duration, before, after, nil
}

// AssertDuration asserts that a function completes within the specified duration
func (h *PerformanceTestHelper) AssertDuration(maxDuration time.Duration, fn func()) {
	duration, err := h.MeasureFunction(fn)
	if err != nil {
		h.t.Fatalf("Failed to measure function duration: %v", err)
	}

	if duration > maxDuration {
		h.t.Fatalf("Function took %v, expected <= %v", duration, maxDuration)
	}
}

// AssertMemoryIncrease asserts that memory usage increases by at most the specified amount
func (h *PerformanceTestHelper) AssertMemoryIncrease(maxIncrease uint64, fn func()) {
	before, after, err := h.MeasureMemoryUsage(fn)
	if err != nil {
		h.t.Fatalf("Failed to measure memory usage: %v", err)
	}

	increase := after - before
	if increase > maxIncrease {
		h.t.Fatalf("Memory increased by %d bytes, expected <= %d bytes", increase, maxIncrease)
	}
}
