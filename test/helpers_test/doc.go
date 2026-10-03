// Package helpers_test carries the GitTestHelper coverage tests
// originally in test/helpers. Move-whole per the testing-helpers spec;
// per-package dissolve (where tests live next to the package they
// test, e.g. test/git/git_test.go) is deferred to a follow-up
// change.
//
// The package needs at least one non-_test.go file to satisfy the
// Go build (external _test packages require a sibling non-test
// source file in the same directory). This file supplies only the
// package declaration; the actual test bodies live in
// helpers_test.go.
package helpers_test
