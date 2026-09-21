package mocks

// variadicArgs expands a variadic argument list for testify mock.Called,
// which expects positional arguments rather than a slice.
func variadicArgs(first any, opts ...any) []any {
	args := []any{first}
	return append(args, opts...)
}
