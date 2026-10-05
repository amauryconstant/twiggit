package config_test

import (
	"fmt"
	"os"
	"twiggit/internal/config"
)

// ExampleNewManager walks the canonical Load flow: NewManager()
// constructs the koanf-backed manager, Load() reads the config file
// from XDG_CONFIG_HOME (or falls back to defaults when no file is
// present), and returns an immutable *core.Config copy. The example
// points XDG_CONFIG_HOME at an empty temp directory so the reader
// sees the defaults in action — DefaultSourceBranch is "main" and
// Git.IsCacheEnabled is true regardless of whether a config.toml
// exists. The original environment is restored on exit so the
// process state stays clean.
func ExampleNewManager() {
	dir, err := os.MkdirTemp("", "twiggit-config-example-")
	if err != nil {
		fmt.Println("setup error:", err)
		return
	}
	defer os.RemoveAll(dir)

	prev, had := os.LookupEnv("XDG_CONFIG_HOME")
	os.Setenv("XDG_CONFIG_HOME", dir)
	defer func() {
		if had {
			os.Setenv("XDG_CONFIG_HOME", prev)
		} else {
			os.Unsetenv("XDG_CONFIG_HOME")
		}
	}()

	cfg, err := config.NewManager().Load()
	if err != nil {
		fmt.Println("load error:", err)
		return
	}

	fmt.Println("DefaultSourceBranch:", cfg.DefaultSourceBranch)
	fmt.Println("IsCacheEnabled:", cfg.Git.IsCacheEnabled)
	// Output:
	// DefaultSourceBranch: main
	// IsCacheEnabled: true
}
