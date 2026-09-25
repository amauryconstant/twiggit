// Path: internal/config/config.go
package config

import (
	"errors"
	"fmt"
	"io/fs"
	"path/filepath"
	"strings"
	"time"

	"github.com/adrg/xdg"
	"github.com/knadh/koanf/parsers/toml/v2"
	"github.com/knadh/koanf/providers/env"
	"github.com/knadh/koanf/providers/file"
	"github.com/knadh/koanf/providers/structs"
	"github.com/knadh/koanf/v2"

	"github.com/you/myapp/internal/core"
)

type AppConfig struct {
	ProjectsDir   string        `koanf:"projects_dir"`
	DefaultBranch string        `koanf:"default_branch"`
	Timeout       time.Duration `koanf:"timeout"`
	Protected     []string      `koanf:"protected"`
}

func DefaultConfig() *AppConfig {
	return &AppConfig{
		ProjectsDir:   filepath.Join(xdg.Home, "Projects"),
		DefaultBranch: "main",
		Timeout:       30 * time.Second,
		Protected:     []string{"main", "master"},
	}
}

// Path is $XDG_CONFIG_HOME/myapp/config.toml.
func Path() string { return filepath.Join(xdg.ConfigHome, "myapp", "config.toml") }

// Load merges defaults → config file → MYAPP_* env, last wins. Flags win at the
// command, which reads them into its Options. A missing file is normal.
func Load() (*AppConfig, error) {
	k := koanf.New(".")
	if err := k.Load(structs.Provider(DefaultConfig(), "koanf"), nil); err != nil {
		return nil, err
	}
	if err := k.Load(file.Provider(Path()), toml.Parser()); err != nil && !errors.Is(err, fs.ErrNotExist) {
		return nil, fmt.Errorf("reading config %s: %w", Path(), err)
	}
	// MYAPP_DEFAULT_BRANCH → default_branch
	err := k.Load(env.Provider("MYAPP_", ".", func(s string) string {
		return strings.ToLower(strings.TrimPrefix(s, "MYAPP_"))
	}), nil)
	if err != nil {
		return nil, err
	}

	var cfg AppConfig
	if err := k.Unmarshal("", &cfg); err != nil {
		return nil, err
	}
	if err := rules.ValidateAll(&cfg); err != nil {
		return nil, fmt.Errorf("invalid config %s:\n%w", Path(), err)
	}
	return &cfg, nil
}

// rules run in collect-all mode so users fix every problem in one pass.
var rules = core.NewPipeline(
	func(c *AppConfig) error {
		if c.Timeout <= 0 {
			return &core.ValidationError{Field: "timeout", Value: c.Timeout.String(), Message: "must be positive"}
		}
		return nil
	},
	func(c *AppConfig) error {
		if c.DefaultBranch == "" {
			return &core.ValidationError{Field: "default_branch", Message: "must not be empty"}
		}
		return nil
	},
)
