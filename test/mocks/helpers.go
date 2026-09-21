package mocks

// variadicArgs builds the positional argument list for testify
// mock.Called from a single first argument and a variadic slice of any
// element type. testify's mock.Called expects positional arguments, not
// a slice; this helper expands a typed slice into that positional shape.
//
// Spread the result back to m.Called:
//
//	args := m.Called(variadicArgs(first, opts)...)
func variadicArgs[T any](first any, rest []T) []any {
	out := make([]any, 0, 1+len(rest))
	out = append(out, first)
	for _, v := range rest {
		out = append(out, v)
	}
	return out
}
