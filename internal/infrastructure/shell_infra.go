package infrastructure

import (
	"os"
	"path/filepath"
	"strings"
	"time"

	"twiggit/internal/domain"
)

type shellInfrastructure struct{}

// NewShellInfrastructure creates a new shell infrastructure instance
func NewShellInfrastructure() ShellInfrastructure {
	return &shellInfrastructure{}
}

// GenerateWrapper generates a shell wrapper for the specified shell type
func (s *shellInfrastructure) GenerateWrapper(shellType domain.ShellType) (string, error) {
	template := s.getWrapperTemplate(shellType)
	if template == "" {
		return "", domain.NewShellError(domain.ErrInvalidShellType, string(shellType), "unsupported shell type")
	}

	// Pure function composition for wrapper generation
	return s.composeWrapper(template, shellType), nil
}

// DetectConfigFile detects the appropriate config file for the shell type
func (s *shellInfrastructure) DetectConfigFile(shellType domain.ShellType) (string, error) {
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
func (s *shellInfrastructure) InstallWrapper(shellType domain.ShellType, wrapper, configFile string, force bool) error {
	if configFile == "" {
		return domain.NewShellErrorWithCause(domain.ErrConfigFileNotFound, string(shellType), "config file path is empty", nil)
	}

	// Check if parent directory exists
	parentDir := filepath.Dir(configFile)
	if _, err := os.Stat(parentDir); os.IsNotExist(err) {
		return domain.NewShellErrorWithCause(domain.ErrConfigFileNotFound, string(shellType), "parent directory does not exist", err)
	}

	// Check if file exists
	fileExists := true
	if _, err := os.Stat(configFile); os.IsNotExist(err) {
		fileExists = false
	}

	if !fileExists {
		// Create the file if it doesn't exist
		if err := os.WriteFile(configFile, []byte(wrapper), 0644); err != nil {
			return domain.NewShellErrorWithCause(domain.ErrWrapperInstallation, string(shellType), "failed to create config file", err)
		}
		return nil
	}

	// Read existing content
	content, err := os.ReadFile(configFile)
	if err != nil {
		return domain.NewShellErrorWithCause(domain.ErrWrapperInstallation, string(shellType), "failed to read config file", err)
	}

	contentStr := string(content)

	// Check if wrapper block exists
	if s.hasWrapperBlock(contentStr) {
		if !force {
			return domain.NewShellError(domain.ErrShellAlreadyInstalled, string(shellType), "wrapper already installed")
		}
		// Remove existing wrapper block
		contentStr = s.removeWrapperBlock(contentStr)
	}

	// Append wrapper to config file
	updatedContent := s.appendWrapper(contentStr, wrapper)
	if err := os.WriteFile(configFile, []byte(updatedContent), 0644); err != nil {
		return domain.NewShellErrorWithCause(domain.ErrWrapperInstallation, string(shellType), "failed to write wrapper to config file", err)
	}

	return nil
}

// ValidateInstallation validates whether the wrapper is installed
func (s *shellInfrastructure) ValidateInstallation(shellType domain.ShellType, configFile string) error {
	if configFile == "" {
		return domain.NewShellErrorWithCause(domain.ErrConfigFileNotFound, string(shellType), "config file path is empty", nil)
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

	// Check for wrapper block delimiters
	if !s.hasWrapperBlock(string(content)) {
		return domain.NewShellError(domain.ErrShellNotInstalled, string(shellType), "wrapper block not found")
	}

	return nil
}

// getWrapperTemplate returns the wrapper template for the specified shell type
func (s *shellInfrastructure) getWrapperTemplate(shellType domain.ShellType) string {
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
func (s *shellInfrastructure) composeWrapper(template string, shellType domain.ShellType) string {
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
func (s *shellInfrastructure) getConfigFiles(shellType domain.ShellType) []string {
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
func (s *shellInfrastructure) bashWrapperTemplate() string {
	return `### BEGIN TWIGGIT WRAPPER
# Twiggit bash wrapper - Generated on {{TIMESTAMP}}
twiggit() {
    case "$1" in
        cd)
            # Handle cd command with directory change
            target_dir=$(command twiggit "$@")
            if [ $? -eq 0 ] && [ -n "$target_dir" ]; then
                builtin cd "$target_dir"
            fi
            ;;
        create)
            # Handle create command with -C flag
            if [[ " $@ " == *" -C "* ]] || [[ " $@ " == *" --cd "* ]]; then
                target_dir=$(command twiggit "$@")
                if [ $? -eq 0 ] && [ -n "$target_dir" ]; then
                    builtin cd "$target_dir"
                fi
            else
                command twiggit "$@"
            fi
            ;;
        delete)
            # Handle delete command with -C flag
            if [[ " $@ " == *" -C "* ]] || [[ " $@ " == *" --cd "* ]]; then
                target_dir=$(command twiggit "$@")
                if [ $? -eq 0 ] && [ -n "$target_dir" ]; then
                    builtin cd "$target_dir"
                fi
            else
                command twiggit "$@"
            fi
            ;;
        *)
            # Pass through all other commands
            command twiggit "$@"
            ;;
    esac
}
### END TWIGGIT WRAPPER`
}

// zshWrapperTemplate returns the zsh wrapper template
func (s *shellInfrastructure) zshWrapperTemplate() string {
	return `### BEGIN TWIGGIT WRAPPER
# Twiggit zsh wrapper - Generated on {{TIMESTAMP}}
twiggit() {
    case "$1" in
        cd)
            # Handle cd command with directory change
            target_dir=$(command twiggit "$@")
            if [ $? -eq 0 ] && [ -n "$target_dir" ]; then
                builtin cd "$target_dir"
            fi
            ;;
        create)
            # Handle create command with -C flag
            if [[ " $@ " == *" -C "* ]] || [[ " $@ " == *" --cd "* ]]; then
                target_dir=$(command twiggit "$@")
                if [ $? -eq 0 ] && [ -n "$target_dir" ]; then
                    builtin cd "$target_dir"
                fi
            else
                command twiggit "$@"
            fi
            ;;
        delete)
            # Handle delete command with -C flag
            if [[ " $@ " == *" -C "* ]] || [[ " $@ " == *" --cd "* ]]; then
                target_dir=$(command twiggit "$@")
                if [ $? -eq 0 ] && [ -n "$target_dir" ]; then
                    builtin cd "$target_dir"
                fi
            else
                command twiggit "$@"
            fi
            ;;
        *)
            # Pass through all other commands
            command twiggit "$@"
            ;;
    esac
}
### END TWIGGIT WRAPPER`
}

// fishWrapperTemplate returns the fish wrapper template
func (s *shellInfrastructure) fishWrapperTemplate() string {
	return `### BEGIN TWIGGIT WRAPPER
# Twiggit fish wrapper - Generated on {{TIMESTAMP}}
function twiggit
    switch "$argv[1]"
        case cd
            # Handle cd command with directory change
            set target_dir (command twiggit $argv)
            if test $status -eq 0 -a -n "$target_dir"
                builtin cd "$target_dir"
            end
        case create
            # Handle create command with -C flag
            if string match -q "* -C *" " $argv "; or string match -q "* --cd *" " $argv "
                set target_dir (command twiggit $argv)
                if test $status -eq 0 -a -n "$target_dir"
                    builtin cd "$target_dir"
                end
            else
                command twiggit $argv
            end
        case delete
            # Handle delete command with -C flag
            if string match -q "* -C *" " $argv "; or string match -q "* --cd *" " $argv "
                set target_dir (command twiggit $argv)
                if test $status -eq 0 -a -n "$target_dir"
                    builtin cd "$target_dir"
                end
            else
                command twiggit $argv
            end
        case '*'
            # Pass through all other commands
            command twiggit $argv
    end
end
### END TWIGGIT WRAPPER`
}

const (
	beginDelimiter = "### BEGIN TWIGGIT WRAPPER"
	endDelimiter   = "### END TWIGGIT WRAPPER"
)

func (s *shellInfrastructure) hasWrapperBlock(content string) bool {
	return strings.Contains(content, beginDelimiter) && strings.Contains(content, endDelimiter)
}

func (s *shellInfrastructure) removeWrapperBlock(content string) string {
	beginIdx := strings.Index(content, beginDelimiter)
	if beginIdx == -1 {
		return content
	}

	endIdx := strings.Index(content, endDelimiter)
	if endIdx == -1 {
		return content
	}

	newlineBefore := strings.LastIndex(content[:beginIdx], "\n")
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

func (s *shellInfrastructure) appendWrapper(content, wrapper string) string {
	if content == "" {
		return wrapper + "\n"
	}
	return content + "\n" + wrapper + "\n"
}
