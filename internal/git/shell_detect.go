// Package git shell-detection helpers.
//
// DetectShellFromEnv reads the SHELL environment variable and infers the
// shell type. InferShellTypeFromPath infers the shell type from a config
// file path's filename. ProbeShellConfig stats the canonical config path
// for the given shell type and reports whether it exists.
//
// These helpers live here (not in internal/core) because they touch the
// environment and the filesystem; core is the pure functional core and
// must not import "os".
package git

import (
	"errors"
	"os"
	"path/filepath"
	"strings"

	"twiggit/internal/core"
)

// DetectShellFromEnv detects the shell type from the SHELL environment
// variable. Returns *core.OperationError with Op="shell.detect" if SHELL
// is unset or points at an unsupported shell.
func DetectShellFromEnv() (core.ShellType, error) {
	shellPath := os.Getenv("SHELL")
	if shellPath == "" {
		return "", core.NewShellDetectionError("SHELL environment variable not set", nil)
	}

	shellName := filepath.Base(shellPath)
	lowerName := strings.ToLower(shellName)

	switch {
	case strings.Contains(lowerName, "bash"):
		return core.ShellBash, nil
	case strings.Contains(lowerName, "zsh"):
		return core.ShellZsh, nil
	case strings.Contains(lowerName, "fish"):
		return core.ShellFish, nil
	default:
		return "", core.NewShellDetectionError(
			"unsupported shell detected: "+shellName,
			&core.ValidationError{
				Field:       "shellType",
				Value:       shellName,
				Message:     "unsupported shell type",
				Suggestions: []string{"use --shell to specify shell type (bash, zsh, fish)"},
			},
		)
	}
}

// InferShellTypeFromPath infers the shell type from a config file path.
// Pure filename parsing — no filesystem access. Returns *core.OperationError
// with Op="shell.infer" if the path does not match a known shell config.
func InferShellTypeFromPath(configPath string) (core.ShellType, error) {
	filename := filepath.Base(configPath)
	lowerFilename := strings.ToLower(filename)
	lowerPath := strings.ToLower(configPath)

	switch {
	case strings.HasPrefix(lowerFilename, ".bash") || strings.HasPrefix(lowerFilename, "bash") ||
		strings.HasSuffix(lowerFilename, ".bash") || lowerFilename == ".bash_profile" ||
		lowerFilename == ".profile" || strings.Contains(lowerFilename, "-bash-"):
		return core.ShellBash, nil

	case strings.HasPrefix(lowerFilename, ".zsh") || strings.HasPrefix(lowerFilename, "zsh") ||
		strings.HasSuffix(lowerFilename, ".zsh") || lowerFilename == ".zprofile" ||
		strings.Contains(lowerFilename, "-zsh-"):
		return core.ShellZsh, nil

	case strings.Contains(lowerFilename, "fish") || lowerFilename == "config.fish" ||
		lowerFilename == ".fishrc" || strings.Contains(lowerPath, "fish"):
		return core.ShellFish, nil

	default:
		return "", core.NewShellInferenceError(
			"",
			"cannot infer shell type from path: "+configPath,
			&core.ValidationError{
				Field:       "shellType",
				Message:     "cannot infer shell type",
				Suggestions: []string{"use --shell to specify shell type (bash, zsh, fish)"},
			},
		)
	}
}

// ProbeShellConfig stats the canonical config path for the given shell
// type under homeDir and reports whether the file exists. Stat failures
// other than ErrNotExist are wrapped as *core.OperationError with
// Op="shell.probe"; ErrNotExist returns exists=false with no error.
func ProbeShellConfig(shellType core.ShellType, homeDir string) (path string, exists bool, err error) {
	var rel string
	switch shellType {
	case core.ShellBash:
		rel = ".bashrc"
	case core.ShellZsh:
		rel = ".zshrc"
	case core.ShellFish:
		rel = ".config/fish/config.fish"
	default:
		return "", false, core.NewShellInvalidTypeError(
			"",
			"unsupported shell type for probe: "+string(shellType),
			nil,
		)
	}
	full := filepath.Join(homeDir, rel)
	info, statErr := os.Stat(full)
	if statErr != nil {
		if errors.Is(statErr, os.ErrNotExist) {
			return full, false, nil
		}
		return "", false, &core.OperationError{
			Op:      "shell.probe",
			Message: "stat failed for " + full,
			Cause:   statErr,
		}
	}
	if info.IsDir() {
		return full, false, nil
	}
	return full, true, nil
}
