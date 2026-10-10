package core

import (
	"encoding/json"
	"errors"
	"testing"
	"time"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestNewResult_NonSliceValuePassesThrough(t *testing.T) {
	t.Run("bool value passes through unchanged", func(t *testing.T) {
		is := assert.New(t)
		r := NewResult(true)
		is.True(r.Value)
		is.NoError(r.Error)
	})

	t.Run("int value passes through unchanged", func(t *testing.T) {
		is := assert.New(t)
		r := NewResult(42)
		is.Equal(42, r.Value)
	})

	t.Run("string value passes through unchanged", func(t *testing.T) {
		is := assert.New(t)
		r := NewResult("hello")
		is.Equal("hello", r.Value)
	})
}

func TestNewResult_SliceValueCloned(t *testing.T) {
	t.Run("mutating returned slice does not affect caller", func(t *testing.T) {
		is := assert.New(t)
		must := require.New(t)

		original := []string{"a", "b", "c"}
		r := NewResult(original)
		must.Equal(original, r.Value, "result value should match input")

		r.Value[0] = "MUTATED"

		is.Equal("a", original[0], "original must remain untouched after mutation through result")
	})

	t.Run("different slice types cloned independently", func(t *testing.T) {
		is := assert.New(t)
		must := require.New(t)

		ints := []int{1, 2, 3}
		rInts := NewResult(ints)
		rInts.Value[0] = 999
		is.Equal(1, ints[0])

		strs := []string{"x", "y"}
		rStrs := NewResult(strs)
		rStrs.Value[1] = "MUTATED"
		is.Equal("y", strs[1])
		must.NotNil(rInts.Value)
		must.NotNil(rStrs.Value)
	})

	t.Run("empty slice is cloned without panic", func(t *testing.T) {
		is := assert.New(t)
		original := []int{}
		r := NewResult(original)
		is.NotNil(r.Value)
		is.Empty(r.Value)
	})

	t.Run("nil slice passes through", func(t *testing.T) {
		is := assert.New(t)
		var nilSlice []int
		r := NewResult(nilSlice)
		is.Nil(r.Value)
	})
}

func TestNewErrResult_ZeroValue(t *testing.T) {
	t.Run("error result for bool", func(t *testing.T) {
		is := assert.New(t)
		sentinelErr := errors.New("test sentinel")
		r := NewErrResult[bool](sentinelErr)
		is.False(r.Value)
		is.ErrorIs(r.Error, sentinelErr)
	})

	t.Run("error result for int slice keeps nil zero value", func(t *testing.T) {
		is := assert.New(t)
		sentinelErr := errors.New("test sentinel")
		r := NewErrResult[[]int](sentinelErr)
		is.Nil(r.Value)
		is.ErrorIs(r.Error, sentinelErr)
	})
}

func TestWorktreeStatus_ZeroValueSafe(t *testing.T) {
	t.Run("zero value has safe field access without nil-deref", func(t *testing.T) {
		is := assert.New(t)
		var s WorktreeStatus
		is.False(s.IsClean)
		is.False(s.IsMerged)
		is.False(s.IsStale)
		is.False(s.IsSkipped)
		is.False(s.Dirty())
		is.Empty(s.SkipReason)
		is.Empty(s.Base)
		is.True(s.LastChecked.IsZero())
		is.True(s.LastCommitDate.IsZero())
	})

	t.Run("zero value marshals to {} without panic", func(t *testing.T) {
		var s WorktreeStatus
		data, err := json.Marshal(s)
		require.NoError(t, err, "zero-value marshal must not panic or error")
		require.NotNil(t, data)
		assert.Contains(t, string(data), "is_clean")
		assert.NotContains(t, string(data), "skip_reason",
			"empty skip_reason must be omitted via omitempty")
	})

	t.Run("zero value unmarshals into a fresh value without panic", func(t *testing.T) {
		is := assert.New(t)
		var s WorktreeStatus
		require.NoError(t, json.Unmarshal([]byte("{}"), &s))
		is.False(s.IsClean)
		is.False(s.IsMerged)
	})
}

func TestWorktreeStatus_JSONRoundTrip(t *testing.T) {
	now := time.Date(2026, 1, 2, 3, 4, 5, 0, time.UTC)
	commit := time.Date(2025, 12, 1, 0, 0, 0, 0, time.UTC)
	original := WorktreeStatus{
		Worktree: &Worktree{
			Path:   "/tmp/wt",
			Branch: "feat/x",
		},
		RepositoryStatus: &RepositoryStatus{
			IsClean: true,
			Branch:  "feat/x",
			Commit:  "abc123",
			Ahead:   2,
			Behind:  5,
		},
		LastChecked:           now,
		IsClean:               true,
		HasUncommittedChanges: false,
		Base:                  "main",
		IsMerged:              false,
		IsStale:               true,
		LastCommitDate:        commit,
		IsSkipped:             false,
	}

	data, err := json.Marshal(original)
	require.NoError(t, err)

	var got WorktreeStatus
	require.NoError(t, json.Unmarshal(data, &got))

	assert.Equal(t, original.Base, got.Base)
	assert.Equal(t, original.IsMerged, got.IsMerged)
	assert.Equal(t, original.IsStale, got.IsStale)
	assert.Equal(t, original.LastCommitDate, got.LastCommitDate)
	assert.Equal(t, original.IsSkipped, got.IsSkipped)
	assert.Equal(t, original.SkipReason, got.SkipReason)
	assert.Equal(t, original.IsClean, got.IsClean)
	require.NotNil(t, got.RepositoryStatus)
	assert.Equal(t, original.RepositoryStatus.Ahead, got.RepositoryStatus.Ahead)
	assert.Equal(t, original.RepositoryStatus.Behind, got.RepositoryStatus.Behind)
}

func TestWorktreeStatus_SkipReasonOmitEmpty(t *testing.T) {
	is := assert.New(t)
	row := WorktreeStatus{Base: "main", IsSkipped: true}
	data, err := json.Marshal(row)
	require.NoError(t, err)
	is.NotContains(string(data), "skip_reason",
		"empty skip_reason must be omitted via omitempty")
}

func TestWorktreeStatus_Dirty(t *testing.T) {
	cases := []struct {
		name string
		row  WorktreeStatus
		want bool
	}{
		{"clean tree is not dirty", WorktreeStatus{HasUncommittedChanges: false}, false},
		{"uncommitted changes are dirty", WorktreeStatus{HasUncommittedChanges: true}, true},
	}
	for _, tc := range cases {
		t.Run(tc.name, func(t *testing.T) {
			assert.Equal(t, tc.want, tc.row.Dirty())
		})
	}
}

func TestWorktreeStatus_ComputeIsStale(t *testing.T) {
	now := time.Now()
	cfg := &Config{
		Status: StatusConfig{StaleBehind: 20, StaleDays: 30},
	}

	cases := []struct {
		name string
		row  WorktreeStatus
		cfg  *Config
		want bool
	}{
		{
			name: "behind at threshold trips stale",
			row:  WorktreeStatus{RepositoryStatus: &RepositoryStatus{Behind: 20}},
			cfg:  cfg,
			want: true,
		},
		{
			name: "behind above threshold trips stale",
			row:  WorktreeStatus{RepositoryStatus: &RepositoryStatus{Behind: 25}},
			cfg:  cfg,
			want: true,
		},
		{
			name: "behind below threshold is fresh",
			row:  WorktreeStatus{RepositoryStatus: &RepositoryStatus{Behind: 5}},
			cfg:  cfg,
			want: false,
		},
		{
			name: "age above threshold trips stale",
			row:  WorktreeStatus{LastCommitDate: now.Add(-40 * 24 * time.Hour)},
			cfg:  cfg,
			want: true,
		},
		{
			name: "age below threshold is fresh",
			row:  WorktreeStatus{LastCommitDate: now.Add(-3 * 24 * time.Hour)},
			cfg:  cfg,
			want: false,
		},
		{
			name: "zero last commit date is not aged",
			row:  WorktreeStatus{LastCommitDate: time.Time{}},
			cfg:  cfg,
			want: false,
		},
		{
			name: "both trip in the same row",
			row: WorktreeStatus{
				RepositoryStatus: &RepositoryStatus{Behind: 30},
				LastCommitDate:   now.Add(-60 * 24 * time.Hour),
			},
			cfg:  cfg,
			want: true,
		},
		{
			name: "neither trip is fresh",
			row: WorktreeStatus{
				RepositoryStatus: &RepositoryStatus{Behind: 2},
				LastCommitDate:   now.Add(-1 * 24 * time.Hour),
			},
			cfg:  cfg,
			want: false,
		},
		{
			name: "both thresholds zero disables the heuristic",
			row: WorktreeStatus{
				RepositoryStatus: &RepositoryStatus{Behind: 999},
				LastCommitDate:   now.Add(-365 * 24 * time.Hour),
			},
			cfg:  &Config{Status: StatusConfig{}},
			want: false,
		},
		{
			name: "nil config returns false",
			row:  WorktreeStatus{RepositoryStatus: &RepositoryStatus{Behind: 999}},
			cfg:  nil,
			want: false,
		},
		{
			name: "nil repository status with behind threshold set",
			row:  WorktreeStatus{},
			cfg:  cfg,
			want: false,
		},
	}

	for _, tc := range cases {
		t.Run(tc.name, func(t *testing.T) {
			assert.Equal(t, tc.want, tc.row.ComputeIsStale(tc.cfg))
		})
	}
}
