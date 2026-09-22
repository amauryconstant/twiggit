package core

import (
	"errors"
	"testing"

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
		require := require.New(t)

		original := []string{"a", "b", "c"}
		r := NewResult(original)
		require.Equal(original, r.Value, "result value should match input")

		r.Value[0] = "MUTATED"

		is.Equal("a", original[0], "original must remain untouched after mutation through result")
	})

	t.Run("different slice types cloned independently", func(t *testing.T) {
		is := assert.New(t)
		require := require.New(t)

		ints := []int{1, 2, 3}
		rInts := NewResult(ints)
		rInts.Value[0] = 999
		is.Equal(1, ints[0])

		strs := []string{"x", "y"}
		rStrs := NewResult(strs)
		rStrs.Value[1] = "MUTATED"
		is.Equal("y", strs[1])
		require.NotNil(rInts.Value)
		require.NotNil(rStrs.Value)
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
