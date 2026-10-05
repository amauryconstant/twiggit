package iostreams_test

import (
	"bytes"
	"context"
	"fmt"
	"log/slog"
	"os"
	"twiggit/internal/iostreams"
)

// ExampleNewLogger demonstrates the TWIGGIT_DEBUG gate that selects
// the slog level at construction time. Each call uses a fresh
// *bytes.Buffer so the per-writer cache returns a fresh logger
// instead of the singleton built the first time this example ran;
// the global env is restored so the test leaves the process state
// unchanged.
func ExampleNewLogger() {
	prev, had := os.LookupEnv("TWIGGIT_DEBUG")
	os.Setenv("TWIGGIT_DEBUG", "1")
	defer func() {
		if had {
			os.Setenv("TWIGGIT_DEBUG", prev)
		} else {
			os.Unsetenv("TWIGGIT_DEBUG")
		}
	}()

	ctx := context.Background()
	debugLogger := iostreams.NewLogger(&bytes.Buffer{})
	warnLogger := iostreams.NewLogger(&bytes.Buffer{})
	os.Unsetenv("TWIGGIT_DEBUG")
	prodLogger := iostreams.NewLogger(&bytes.Buffer{})

	fmt.Println("debug+debug:", debugLogger.Enabled(ctx, slog.LevelDebug))
	fmt.Println("warn+debug:", warnLogger.Enabled(ctx, slog.LevelWarn))
	fmt.Println("debug-off:", prodLogger.Enabled(ctx, slog.LevelDebug))
	fmt.Println("warn-off:", prodLogger.Enabled(ctx, slog.LevelWarn))
	// Output:
	// debug+debug: true
	// warn+debug: true
	// debug-off: false
	// warn-off: true
}
