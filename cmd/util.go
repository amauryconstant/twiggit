package cmd

import (
	"fmt"
	"os"

	"github.com/spf13/cobra"
)

func logv(cmd *cobra.Command, level int, format string, args ...interface{}) {
	verbosity, _ := cmd.Flags().GetCount("verbose")

	if verbosity < level {
		return
	}

	prefix := ""
	if level > 1 {
		prefix = "  "
	}

	msg := fmt.Sprintf(format, args...)
	fmt.Fprintf(os.Stderr, "%s%s\n", prefix, msg)
}
