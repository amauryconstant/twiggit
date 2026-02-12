package domain

import (
	"os"
	"path/filepath"
	"strings"
)

// ShellType represents the type of shell
type ShellType string

const (
	// ShellBash represents the bash shell type
	ShellBash ShellType = "bash"
	// ShellZsh represents the zsh shell type
	ShellZsh ShellType = "zsh"
	// ShellFish represents the fish shell type
	ShellFish ShellType = "fish"
)

// Shell represents a shell with its configuration
type Shell interface {
	Type() ShellType
	Path() string
	Version() string
}

type shell struct {
	shellType ShellType
	path      string
	version   string
}

// NewShell creates a new Shell instance with validation
func NewShell(shellType ShellType, path, version string) (Shell, error) {
	if !IsValidShellType(shellType) {
		return nil, NewValidationError("NewShell", "shellType", string(shellType), "unsupported shell type").
			WithSuggestions([]string{"Supported shells: bash, zsh, fish"})
	}

	return &shell{
		shellType: shellType,
		path:      path,
		version:   version,
	}, nil
}

// Type returns the shell type
func (s *shell) Type() ShellType {
	return s.shellType
}

// Path returns the shell path
func (s *shell) Path() string {
	return s.path
}

// Version returns the shell version
func (s *shell) Version() string {
	return s.version
}

// IsValidShellType checks if the shell type is supported
func IsValidShellType(shellType ShellType) bool {
	switch shellType {
	case ShellBash, ShellZsh, ShellFish:
		return true
	default:
		return false
	}
}

// InferShellTypeFromPath infers shell type from config file path
func InferShellTypeFromPath(configPath string) (ShellType, error) {
	filename := filepath.Base(configPath)
	lowerFilename := strings.ToLower(filename)
	lowerPath := strings.ToLower(configPath)

	switch {
	case strings.HasPrefix(lowerFilename, ".bash") || strings.HasPrefix(lowerFilename, "bash") ||
		strings.HasSuffix(lowerFilename, ".bash") || lowerFilename == ".bash_profile" ||
		lowerFilename == ".profile" || strings.Contains(lowerFilename, "-bash-"):
		return ShellBash, nil

	case strings.HasPrefix(lowerFilename, ".zsh") || strings.HasPrefix(lowerFilename, "zsh") ||
		strings.HasSuffix(lowerFilename, ".zsh") || lowerFilename == ".zprofile" ||
		strings.Contains(lowerFilename, "-zsh-"):
		return ShellZsh, nil

	case strings.Contains(lowerFilename, "fish") || lowerFilename == "config.fish" ||
		lowerFilename == ".fishrc" || strings.Contains(lowerPath, "fish"):
		return ShellFish, nil

	default:
		return "", NewShellErrorWithCause(
			ErrInferenceFailed,
			"",
			"cannot infer shell type from path: "+configPath,
			NewValidationError("InferShellTypeFromPath", "shellType", "", "cannot infer shell type").
				WithSuggestions([]string{"use --shell to specify shell type (bash, zsh, fish)"}),
		)
	}
}

// DetectShellFromEnv detects shell type from SHELL environment variable
func DetectShellFromEnv() (ShellType, error) {
	shellPath := os.Getenv("SHELL")
	if shellPath == "" {
		return "", NewShellError(ErrShellDetectionFailed, "", "SHELL environment variable not set")
	}

	shellName := filepath.Base(shellPath)
	lowerName := strings.ToLower(shellName)

	switch {
	case strings.Contains(lowerName, "bash"):
		return ShellBash, nil
	case strings.Contains(lowerName, "zsh"):
		return ShellZsh, nil
	case strings.Contains(lowerName, "fish"):
		return ShellFish, nil
	default:
		return "", NewShellErrorWithCause(
			ErrShellDetectionFailed,
			"",
			"unsupported shell detected: "+shellName,
			NewValidationError("DetectShellFromEnv", "shellType", shellName, "unsupported shell type").
				WithSuggestions([]string{"use --shell to specify shell type (bash, zsh, fish)"}),
		)
	}
}
