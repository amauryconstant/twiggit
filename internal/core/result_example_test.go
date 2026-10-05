package core_test

import (
	"errors"
	"fmt"
	"twiggit/internal/core"
)

// ExampleNewResult shows the success path of the Result helper:
// the returned value is reachable through .Value, and the helper
// methods IsSuccess / IsError tell the caller which branch is live
// without forcing them to compare against nil.
func ExampleNewResult() {
	r := core.NewResult(42)
	fmt.Println(r.IsSuccess())
	fmt.Println(r.IsError())
	fmt.Println(r.Value)
	// Output:
	// true
	// false
	// 42
}

// ExampleNewResult_error shows the error path: a Result built with
// NewErrResult keeps the supplied error reachable through .Error
// and flips the IsError / IsSuccess predicates so the caller can
// branch without a nil check.
func ExampleNewResult_error() {
	r := core.NewErrResult[int](errors.New("boom"))
	fmt.Println(r.IsSuccess())
	fmt.Println(r.IsError())
	fmt.Println(r.Error)
	// Output:
	// false
	// true
	// boom
}
