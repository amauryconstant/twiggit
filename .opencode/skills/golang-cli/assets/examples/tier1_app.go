// Path: main.go (Tier 1 — the whole tool in package main, stdlib flag only)
package main

import (
	"encoding/json"
	"errors"
	"flag"
	"fmt"
	"io"
	"os"
)

func main() {
	os.Exit(run(os.Args[1:], os.Stdout, os.Stderr))
}

// appEnv holds parsed input and the output streams: the Tier 1 Factory.
type appEnv struct {
	out, errOut io.Writer
	verbose     bool
	format      string
	target      string
	dryRun      bool
}

// run returns the exit code, so tests call it with buffers.
func run(args []string, out, errOut io.Writer) int {
	app := &appEnv{out: out, errOut: errOut}
	if err := app.fromArgs(args); err != nil {
		if errors.Is(err, flag.ErrHelp) {
			return 0
		}
		return 2 // the flag set already printed the error and usage
	}
	if err := app.run(); err != nil {
		fmt.Fprintf(errOut, "error: %v\n", err)
		return 1
	}
	return 0
}

// fromArgs is Parse.
func (app *appEnv) fromArgs(args []string) error {
	fl := flag.NewFlagSet("myapp", flag.ContinueOnError)
	fl.SetOutput(app.errOut)
	fl.BoolVar(&app.verbose, "verbose", false, "show progress on stderr")
	fl.StringVar(&app.format, "format", "plain", "output format: plain, json")
	fl.StringVar(&app.target, "target", "", "target to process (required)")
	fl.BoolVar(&app.dryRun, "dry-run", false, "show what would happen")
	if err := fl.Parse(args); err != nil {
		return err
	}
	if app.target == "" {
		fmt.Fprintln(app.errOut, "missing required -target")
		fl.Usage()
		return errors.New("missing -target")
	}
	return nil
}

type result struct {
	Target string `json:"target"`
	DryRun bool   `json:"dry_run"`
}

// process is the Functional Core: values in, values out.
func process(target string, dryRun bool) (result, error) {
	return result{Target: target, DryRun: dryRun}, nil
}

// run is Execute → Respond.
func (app *appEnv) run() error {
	if app.verbose {
		fmt.Fprintf(app.errOut, "processing %s\n", app.target)
	}
	res, err := process(app.target, app.dryRun)
	if err != nil {
		return err
	}
	if app.format == "json" {
		return json.NewEncoder(app.out).Encode(res)
	}
	fmt.Fprintf(app.out, "processed %s (dry run: %t)\n", res.Target, res.DryRun)
	return nil
}
