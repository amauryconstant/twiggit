package output_test

import (
	"fmt"
	"twiggit/internal/output"
)

// ExampleNewFormatter walks the constructor's resolution table so the
// reader can see the four documented outcomes side by side:
//
//   - "" returns (nil, nil) so the caller can defer to its per-command
//     human default.
//   - "json" | "table" | "plain" return a concrete formatter with a nil
//     error.
//   - any other name returns (nil, *core.UsageError) carrying the
//     recognised vocabulary inline so the user sees the fix without
//     opening the docs.
func ExampleNewFormatter() {
	for _, name := range []string{"", "json", "table", "plain", "xml"} {
		f, err := output.NewFormatter(name)
		fmt.Printf("%-5s formatter=%v err=%v\n", name, f != nil, err)
	}
	// Output:
	//       formatter=false err=<nil>
	// json  formatter=true err=<nil>
	// table formatter=true err=<nil>
	// plain formatter=true err=<nil>
	// xml   formatter=false err=invalid output format 'xml': must be 'json', 'table', or 'plain'
}
