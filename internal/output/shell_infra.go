// Package output also exposes the free functions used by cmd/init
// to detect the user's shell config file, install the twiggit wrapper
// into it, and validate an existing installation. These previously
// lived behind the application.ShellInfrastructure interface; with
// the legacy layer removed they are top-level functions callable
// directly by the cmd/run<Init> bodies.
package output

import (
	"errors"
	"os"
	"path/filepath"
	"strings"

	"twiggit/internal/core"
)

// DetectConfigFile returns the best config-file path for the shell
// type, falling back to the preferred config-file name under $HOME
// when no existing file is found. HOME is read from the environment
// first so test isolation via t.Setenv("HOME", ...) works without
// touching the real home directory.
func DetectConfigFile(shellType core.ShellType) (string, error) {
	home := os.Getenv("HOME")
	if home == "" {
		var err error
		home, err = os.UserHomeDir()
		if err != nil {
			return "", core.NewShellConfigError("", "failed to get home directory", err)
		}
	}

	configFiles := configFilesFor(shellType)
	if len(configFiles) == 0 {
		return "", core.NewShellInvalidTypeError(string(shellType), "no config files available for shell type", nil)
	}

	absHome, err := filepath.Abs(home)
	if err != nil {
		return "", core.NewShellConfigError("", "failed to resolve home directory", nil)
	}
	for _, configFile := range configFiles {
		configPath := filepath.Join(absHome, configFile)
		absPath, err := filepath.Abs(configPath)
		if err != nil {
			continue
		}
		if !strings.HasPrefix(absPath, absHome+string(filepath.Separator)) {
			continue
		}
		if _, err := os.Stat(absPath); err == nil { // #nosec G703 -- path validated to stay under home
			return configPath, nil
		}
	}

	// If no existing file found, return the preferred one.
	return filepath.Join(home, configFiles[0]), nil
}

// InstallWrapper writes the wrapper into the shell config file at
// configFile. When the file does not yet exist it is created. When
// it exists and a wrapper block is already present, the install is
// skipped unless force is true; on force, the existing wrapper block
// is removed before appending the new one.
func InstallWrapper(shellType core.ShellType, wrapper, configFile string, force bool) error {
	if configFile == "" {
		return core.NewShellConfigError("", "config file path is empty", nil)
	}

	parentDir := filepath.Dir(configFile)
	if _, err := os.Stat(parentDir); errors.Is(err, os.ErrNotExist) {
		return core.NewShellConfigError("", "parent directory does not exist", err)
	}

	fileExists := true
	if _, err := os.Stat(configFile); errors.Is(err, os.ErrNotExist) {
		fileExists = false
	}

	if !fileExists {
		// Create the file if it doesn't exist.
		if err := os.WriteFile(configFile, []byte(wrapper), 0644); err != nil { // #nosec G306 -- standard perms for shell configs
			return core.NewShellWrapperError(string(shellType), "installation", "failed to create config file", err)
		}
		return nil
	}

	// Read existing content.
	content, err := os.ReadFile(configFile) // #nosec G304 -- configFile from DetectConfigFile which validates path
	if err != nil {
		return core.NewShellWrapperError(string(shellType), "installation", "failed to read config file", err)
	}

	contentStr := string(content)

	// Check if wrapper block exists.
	if hasWrapperBlock(contentStr) {
		if !force {
			return core.NewShellAlreadyInstalledError(string(shellType), "wrapper already installed", nil)
		}
		// Remove existing wrapper block.
		contentStr = removeWrapperBlock(contentStr)
	}

	// Append wrapper to config file.
	updatedContent := appendWrapper(contentStr, wrapper)
	if err := os.WriteFile(configFile, []byte(updatedContent), 0644); err != nil { // #nosec G306,G703 -- standard perms for shell configs, path from DetectConfigFile
		return core.NewShellWrapperError(string(shellType), "installation", "failed to write wrapper to config file", err)
	}

	return nil
}

// ValidateInstallation reports whether a wrapper block is present at
// configFile. A typed core.OperationError with Op="shell.not_installed"
// is returned when the file or wrapper block is missing so callers can
// distinguish the absent-installed case from other failures.
func ValidateInstallation(shellType core.ShellType, configFile string) error {
	if configFile == "" {
		return core.NewShellConfigError("", "config file path is empty", nil)
	}

	if _, err := os.Stat(configFile); errors.Is(err, os.ErrNotExist) {
		return core.NewShellNotInstalledError(string(shellType), "config file does not exist", nil)
	}

	content, err := os.ReadFile(configFile) // #nosec G304 -- configFile from DetectConfigFile which validates path
	if err != nil {
		return core.NewShellNotInstalledError(string(shellType), "failed to read config file", err)
	}

	if !hasWrapperBlock(string(content)) {
		return core.NewShellNotInstalledError(string(shellType), "wrapper block not found", nil)
	}

	return nil
}

// configFilesFor returns the list of candidate config files for the
// shell type, in preference order.
func configFilesFor(shellType core.ShellType) []string {
	switch shellType {
	case core.ShellBash:
		return []string{".bashrc", ".bash_profile", ".profile"}
	case core.ShellZsh:
		return []string{".zshrc", ".zprofile", ".profile"}
	case core.ShellFish:
		return []string{".config/fish/config.fish", "config.fish", ".fishrc"}
	default:
		return []string{}
	}
}

const (
	beginWrapperDelimiter    = "### BEGIN TWIGGIT WRAPPER"
	endWrapperDelimiter      = "### END TWIGGIT WRAPPER"
	beginCompletionDelimiter = "### BEGIN TWIGGIT COMPLETION"
	endCompletionDelimiter   = "### END TWIGGIT COMPLETION"
)

func hasWrapperBlock(content string) bool {
	return strings.Contains(content, beginWrapperDelimiter) && strings.Contains(content, endWrapperDelimiter)
}

func removeWrapperBlock(content string) string {
	content = removeBlock(content, beginWrapperDelimiter, endWrapperDelimiter)
	content = removeBlock(content, beginCompletionDelimiter, endCompletionDelimiter)
	return content
}

func removeBlock(content, beginDelimiter, endDelimiter string) string {
	before, _, ok := strings.Cut(content, beginDelimiter)
	if !ok {
		return content
	}

	endIdx := strings.Index(content, endDelimiter)
	if endIdx == -1 {
		return content
	}

	newlineBefore := strings.LastIndex(before, "\n")
	if newlineBefore == -1 {
		newlineBefore = 0
	} else {
		newlineBefore++
	}

	endAfter := endIdx + len(endDelimiter)
	contentAfter := ""
	if endAfter < len(content) {
		nextNewline := strings.Index(content[endAfter:], "\n")
		if nextNewline != -1 {
			contentAfter = content[endAfter+nextNewline+1:]
		}
	}

	return content[:newlineBefore] + contentAfter
}

func appendWrapper(content, wrapper string) string {
	if content == "" {
		return wrapper + "\n"
	}
	return content + "\n" + wrapper + "\n"
}
