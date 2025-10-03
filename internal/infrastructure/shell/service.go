package shell

import (
	"fmt"
	"os"
	"path/filepath"
	"strings"
	"time"

	"twiggit/internal/domain"
	"twiggit/internal/infrastructure"
)

type shellService struct{}

// NewShellService creates a new shell service instance
func NewShellService() infrastructure.ShellInfrastructure {
	return &shellService{}
}

// GenerateWrapper generates a shell wrapper for the specified shell type
func (s *shellService) GenerateWrapper(shellType domain.ShellType) (string, error) {
	template := s.getWrapperTemplate(shellType)
	if template == "" {
		return "", domain.NewShellError(domain.ErrInvalidShellType, string(shellType), "unsupported shell type")
	}

	// Pure function composition for wrapper generation
	return s.composeWrapper(template, shellType), nil
}

// DetectConfigFile detects the appropriate config file for the shell type
func (s *shellService) DetectConfigFile(shellType domain.ShellType) (string, error) {
	home, err := os.UserHomeDir()
	if err != nil {
		return "", domain.NewShellErrorWithCause(domain.ErrConfigFileNotFound, string(shellType), "failed to get home directory", err)
	}

	configFiles := s.getConfigFiles(shellType)
	if len(configFiles) == 0 {
		return "", domain.NewShellError(domain.ErrInvalidShellType, string(shellType), "no config files available for shell type")
	}

	// Check for existing config files in order of preference
	for _, configFile := range configFiles {
		configPath := filepath.Join(home, configFile)
		if _, err := os.Stat(configPath); err == nil {
			return configPath, nil
		}
	}

	// If no existing file found, return the preferred one
	return filepath.Join(home, configFiles[0]), nil
}

// InstallWrapper installs the wrapper to the shell config file
func (s *shellService) InstallWrapper(shellType domain.ShellType, wrapper string) error {
	configFile, err := s.DetectConfigFile(shellType)
	if err != nil {
		return fmt.Errorf("failed to detect config file: %w", err)
	}

	// Check if file exists and is writable
	if _, err := os.Stat(configFile); os.IsNotExist(err) {
		// Create the file if it doesn't exist
		if err := os.WriteFile(configFile, []byte(wrapper), 0644); err != nil {
			return domain.NewShellErrorWithCause(domain.ErrWrapperInstallation, string(shellType), "failed to create config file", err)
		}
		return nil
	}

	// Check if file is writable
	file, err := os.OpenFile(configFile, os.O_WRONLY|os.O_APPEND, 0600)
	if err != nil {
		return domain.NewShellErrorWithCause(domain.ErrConfigFileNotWritable, string(shellType), "config file is not writable", err)
	}
	defer file.Close()

	// Check if wrapper is already installed
	content, err := os.ReadFile(configFile)
	if err != nil {
		return domain.NewShellErrorWithCause(domain.ErrWrapperInstallation, string(shellType), "failed to read config file", err)
	}

	if strings.Contains(string(content), "twiggit()") || strings.Contains(string(content), "function twiggit") {
		return domain.NewShellError(domain.ErrShellAlreadyInstalled, string(shellType), "wrapper already installed")
	}

	// Append wrapper to config file
	wrapperWithNewline := wrapper + "\n"
	if _, err := file.WriteString(wrapperWithNewline); err != nil {
		return domain.NewShellErrorWithCause(domain.ErrWrapperInstallation, string(shellType), "failed to write wrapper to config file", err)
	}

	return nil
}

// ValidateInstallation validates whether the wrapper is installed
func (s *shellService) ValidateInstallation(shellType domain.ShellType) error {
	configFile, err := s.DetectConfigFile(shellType)
	if err != nil {
		return fmt.Errorf("failed to detect config file: %w", err)
	}

	// Check if config file exists
	if _, err := os.Stat(configFile); os.IsNotExist(err) {
		return domain.NewShellError(domain.ErrShellNotInstalled, string(shellType), "config file does not exist")
	}

	// Read config file and check for wrapper
	content, err := os.ReadFile(configFile)
	if err != nil {
		return domain.NewShellErrorWithCause(domain.ErrShellNotInstalled, string(shellType), "failed to read config file", err)
	}

	// Check for wrapper content
	contentStr := string(content)
	hasBashWrapper := strings.Contains(contentStr, "twiggit() {") && strings.Contains(contentStr, "# Twiggit bash wrapper")
	hasZshWrapper := strings.Contains(contentStr, "twiggit() {") && strings.Contains(contentStr, "# Twiggit zsh wrapper")
	hasFishWrapper := strings.Contains(contentStr, "function twiggit") && strings.Contains(contentStr, "# Twiggit fish wrapper")

	switch shellType {
	case domain.ShellBash:
		if !hasBashWrapper {
			return domain.NewShellError(domain.ErrShellNotInstalled, string(shellType), "bash wrapper not found")
		}
	case domain.ShellZsh:
		if !hasZshWrapper {
			return domain.NewShellError(domain.ErrShellNotInstalled, string(shellType), "zsh wrapper not found")
		}
	case domain.ShellFish:
		if !hasFishWrapper {
			return domain.NewShellError(domain.ErrShellNotInstalled, string(shellType), "fish wrapper not found")
		}
	default:
		return domain.NewShellError(domain.ErrInvalidShellType, string(shellType), "unsupported shell type")
	}

	return nil
}

// getWrapperTemplate returns the wrapper template for the specified shell type
func (s *shellService) getWrapperTemplate(shellType domain.ShellType) string {
	switch shellType {
	case domain.ShellBash:
		return s.bashWrapperTemplate()
	case domain.ShellZsh:
		return s.zshWrapperTemplate()
	case domain.ShellFish:
		return s.fishWrapperTemplate()
	default:
		return ""
	}
}

// composeWrapper composes the wrapper with template replacements (pure function)
func (s *shellService) composeWrapper(template string, shellType domain.ShellType) string {
	// Pure function: no side effects, deterministic output
	replacements := map[string]string{
		"{{SHELL_TYPE}}": string(shellType),
		"{{TIMESTAMP}}":  time.Now().Format("2006-01-02 15:04:05"),
	}

	result := template
	for key, value := range replacements {
		result = strings.ReplaceAll(result, key, value)
	}

	return result
}

// getConfigFiles returns the list of config files for the shell type
func (s *shellService) getConfigFiles(shellType domain.ShellType) []string {
	switch shellType {
	case domain.ShellBash:
		return []string{".bashrc", ".bash_profile", ".profile"}
	case domain.ShellZsh:
		return []string{".zshrc", ".zprofile", ".profile"}
	case domain.ShellFish:
		return []string{".config/fish/config.fish", "config.fish", ".fishrc"}
	default:
		return []string{}
	}
}

// bashWrapperTemplate returns the bash wrapper template
func (s *shellService) bashWrapperTemplate() string {
	return `# Twiggit bash wrapper - Generated on {{TIMESTAMP}}
twiggit() {
    if [ "$1" = "cd" ]; then
        # Handle cd command with directory change
        target_dir=$(command twiggit "$@")
        if [ $? -eq 0 ] && [ -n "$target_dir" ]; then
            builtin cd "$target_dir"
        fi
    else
        # Pass through all other commands
        command twiggit "$@"
    fi
}`
}

// zshWrapperTemplate returns the zsh wrapper template
func (s *shellService) zshWrapperTemplate() string {
	return `# Twiggit zsh wrapper - Generated on {{TIMESTAMP}}
twiggit() {
    if [ "$1" = "cd" ]; then
        # Handle cd command with directory change
        target_dir=$(command twiggit "$@")
        if [ $? -eq 0 ] && [ -n "$target_dir" ]; then
            builtin cd "$target_dir"
        fi
    else
        # Pass through all other commands
        command twiggit "$@"
    fi
}`
}

// fishWrapperTemplate returns the fish wrapper template
func (s *shellService) fishWrapperTemplate() string {
	return `# Twiggit fish wrapper - Generated on {{TIMESTAMP}}
function twiggit
    if test "$argv[1]" = "cd"
        # Handle cd command with directory change
        set target_dir (command twiggit $argv)
        if test $status -eq 0 -a -n "$target_dir"
            builtin cd "$target_dir"
        end
    else
        # Pass through all other commands
        command twiggit $argv
    end
end`
}
