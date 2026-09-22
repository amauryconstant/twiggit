package infrastructure

import (
	"errors"
	"os"
	"path/filepath"
	"strings"
	"time"

	"twiggit/internal/application"
	"twiggit/internal/core"
)

var _ application.ShellInfrastructure = (*shellInfrastructure)(nil)

type shellInfrastructure struct{}

// NewShellInfrastructure creates a new shell infrastructure instance
func NewShellInfrastructure() application.ShellInfrastructure {
	return &shellInfrastructure{}
}

// GenerateWrapper generates a shell wrapper for the specified shell type
func (s *shellInfrastructure) GenerateWrapper(shellType core.ShellType) (string, error) {
	template := s.getWrapperTemplate(shellType)
	if template == "" {
		return "", core.NewShellInvalidTypeError(string(shellType), "unsupported shell type", nil)
	}

	// Pure function composition for wrapper generation
	return s.ComposeWrapper(template, shellType), nil
}

// DetectConfigFile detects the appropriate config file for the shell type
func (s *shellInfrastructure) DetectConfigFile(shellType core.ShellType) (string, error) {
	// Check HOME env var first (for test isolation), fallback to system home
	home := os.Getenv("HOME")
	if home == "" {
		var err error
		home, err = os.UserHomeDir()
		if err != nil {
			return "", core.NewShellConfigError("", "failed to get home directory", err)
		}
	}

	configFiles := s.getConfigFiles(shellType)
	if len(configFiles) == 0 {
		return "", core.NewShellInvalidTypeError(string(shellType), "no config files available for shell type", nil)
	}

	// Check for existing config files in order of preference
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

	// If no existing file found, return the preferred one
	return filepath.Join(home, configFiles[0]), nil
}

// InstallWrapper installs the wrapper to the shell config file
func (s *shellInfrastructure) InstallWrapper(shellType core.ShellType, wrapper, configFile string, force bool) error {
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
		// Create the file if it doesn't exist
		if err := os.WriteFile(configFile, []byte(wrapper), 0644); err != nil { // #nosec G306 -- standard perms for shell configs
			return core.NewShellWrapperError(string(shellType), "installation", "failed to create config file", err)
		}
		return nil
	}

	// Read existing content
	content, err := os.ReadFile(configFile) // #nosec G304 -- configFile from DetectConfigFile which validates path
	if err != nil {
		return core.NewShellWrapperError(string(shellType), "installation", "failed to read config file", err)
	}

	contentStr := string(content)

	// Check if wrapper block exists
	if s.hasWrapperBlock(contentStr) {
		if !force {
			return core.NewShellAlreadyInstalledError(string(shellType), "wrapper already installed", nil)
		}
		// Remove existing wrapper block
		contentStr = s.removeWrapperBlock(contentStr)
	}

	// Append wrapper to config file
	updatedContent := s.appendWrapper(contentStr, wrapper)
	if err := os.WriteFile(configFile, []byte(updatedContent), 0644); err != nil { // #nosec G306,G703 -- standard perms for shell configs, path from DetectConfigFile
		return core.NewShellWrapperError(string(shellType), "installation", "failed to write wrapper to config file", err)
	}

	return nil
}

// ValidateInstallation validates whether the wrapper is installed
func (s *shellInfrastructure) ValidateInstallation(shellType core.ShellType, configFile string) error {
	if configFile == "" {
		return core.NewShellConfigError("", "config file path is empty", nil)
	}

	if _, err := os.Stat(configFile); errors.Is(err, os.ErrNotExist) {
		return core.NewShellNotInstalledError(string(shellType), "config file does not exist", nil)
	}

	// Read config file and check for wrapper
	content, err := os.ReadFile(configFile) // #nosec G304 -- configFile from DetectConfigFile which validates path
	if err != nil {
		return core.NewShellNotInstalledError(string(shellType), "failed to read config file", err)
	}

	// Check for wrapper block delimiters
	if !s.hasWrapperBlock(string(content)) {
		return core.NewShellNotInstalledError(string(shellType), "wrapper block not found", nil)
	}

	return nil
}

// getWrapperTemplate returns the wrapper template for the specified shell type
func (s *shellInfrastructure) getWrapperTemplate(shellType core.ShellType) string {
	switch shellType {
	case core.ShellBash:
		return s.bashWrapperTemplate()
	case core.ShellZsh:
		return s.zshWrapperTemplate()
	case core.ShellFish:
		return s.fishWrapperTemplate()
	default:
		return ""
	}
}

// ComposeWrapper composes the wrapper with template replacements (pure function)
func (s *shellInfrastructure) ComposeWrapper(template string, shellType core.ShellType) string {
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
func (s *shellInfrastructure) getConfigFiles(shellType core.ShellType) []string {
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

// shellTemplateConfig defines shell-specific syntax for wrapper template
type shellTemplateConfig struct {
	caseBegin     string
	caseEnd       string
	elif          string
	caseSeparator string
	argsVar       string
	ifSyntax      string
	thenSyntax    string
	elseSyntax    string
	fiSyntax      string
	andOperator   string
	funcDef       string
	funcEnd       string
}

// getShellTemplateConfig returns shell-specific template configuration
func getShellTemplateConfig(shellType core.ShellType) shellTemplateConfig {
	switch shellType {
	case core.ShellBash, core.ShellZsh:
		return shellTemplateConfig{
			caseBegin:     `case "$1" in`,
			caseEnd:       "esac",
			elif:          ";;",
			caseSeparator: "    ",
			argsVar:       `"$@"`,
			ifSyntax:      "if [[",
			thenSyntax:    "]]; then",
			elseSyntax:    "else",
			fiSyntax:      "fi",
			andOperator:   "]] || [[",
			funcDef:       "twiggit() {",
			funcEnd:       "}",
		}
	case core.ShellFish:
		return shellTemplateConfig{
			caseBegin:     `switch "$argv[1]"`,
			caseEnd:       "end",
			elif:          "",
			caseSeparator: "    ",
			argsVar:       "$argv",
			ifSyntax:      "if",
			thenSyntax:    "",
			elseSyntax:    "else",
			fiSyntax:      "end",
			andOperator:   "or",
			funcDef:       "function twiggit",
			funcEnd:       "end",
		}
	default:
		return shellTemplateConfig{
			caseBegin:     `case "$1" in`,
			caseEnd:       "esac",
			elif:          ";;",
			caseSeparator: "    ",
			argsVar:       `"$@"`,
			ifSyntax:      "if [[",
			thenSyntax:    "]]; then",
			elseSyntax:    "else",
			fiSyntax:      "fi",
			andOperator:   "]] || [[",
			funcDef:       "twiggit() {",
			funcEnd:       "}",
		}
	}
}

// wrapperTemplate returns the shell wrapper template with shell-specific syntax
func (s *shellInfrastructure) wrapperTemplate(shellType core.ShellType) string {
	config := getShellTemplateConfig(shellType)

	return `### BEGIN TWIGGIT WRAPPER
# Twiggit ` + string(shellType) + ` wrapper - Generated on {{TIMESTAMP}}
` + config.funcDef + `
` + config.caseBegin + `
    cd)
        # Handle cd command with directory change
        target_dir=$(command twiggit ` + config.argsVar + `)
        if [ $? -eq 0 ] && [ -n "$target_dir" ]; then
            builtin cd "$target_dir"
        fi
        ` + config.elif + `
    create)
        # Handle create command with -C flag
	` + config.ifSyntax + ` " ` + config.argsVar + ` " == *" -C "* ` + config.andOperator + ` " ` + config.argsVar + ` " == *" --cd "* ` + config.thenSyntax + `
			target_dir=$(command twiggit ` + config.argsVar + `)
			if [ $? -eq 0 ] && [ -n "$target_dir" ]; then
				builtin cd "$target_dir"
			fi
		` + config.elseSyntax + `
			command twiggit ` + config.argsVar + `
		` + config.fiSyntax + `
		` + config.elif + `
	delete)
		# Handle delete command with -C flag
		` + config.ifSyntax + ` " ` + config.argsVar + ` " == *" -C "* ` + config.andOperator + ` " ` + config.argsVar + ` " == *" --cd "* ` + config.thenSyntax + `
            target_dir=$(command twiggit ` + config.argsVar + `)
            if [ $? -eq 0 ] && [ -n "$target_dir" ]; then
                builtin cd "$target_dir"
            fi
        ` + config.elseSyntax + `
            command twiggit ` + config.argsVar + `
        ` + config.fiSyntax + `
        ` + config.elif + `
    *)
        # Pass through all other commands
        command twiggit ` + config.argsVar + `
        ` + config.elif + `
` + config.caseEnd + `
` + config.funcEnd + `
### END TWIGGIT WRAPPER`
}

// bashWrapperTemplate returns the bash wrapper template with Carapace completion
func (s *shellInfrastructure) bashWrapperTemplate() string {
	return s.wrapperTemplate(core.ShellBash) + `
### BEGIN TWIGGIT COMPLETION
source <(command twiggit _carapace bash)
### END TWIGGIT COMPLETION`
}

// zshWrapperTemplate returns the zsh wrapper template with Carapace completion
func (s *shellInfrastructure) zshWrapperTemplate() string {
	return s.wrapperTemplate(core.ShellZsh) + `
### BEGIN TWIGGIT COMPLETION
source <(command twiggit _carapace zsh)
### END TWIGGIT COMPLETION`
}

// fishWrapperTemplate returns the fish wrapper template with Carapace completion
func (s *shellInfrastructure) fishWrapperTemplate() string {
	return s.wrapperTemplate(core.ShellFish) + `
### BEGIN TWIGGIT COMPLETION
command twiggit _carapace fish | source
### END TWIGGIT COMPLETION`
}

const (
	beginWrapperDelimiter    = "### BEGIN TWIGGIT WRAPPER"
	endWrapperDelimiter      = "### END TWIGGIT WRAPPER"
	beginCompletionDelimiter = "### BEGIN TWIGGIT COMPLETION"
	endCompletionDelimiter   = "### END TWIGGIT COMPLETION"
)

func (s *shellInfrastructure) hasWrapperBlock(content string) bool {
	return strings.Contains(content, beginWrapperDelimiter) && strings.Contains(content, endWrapperDelimiter)
}

func (s *shellInfrastructure) removeWrapperBlock(content string) string {
	content = s.removeBlock(content, beginWrapperDelimiter, endWrapperDelimiter)
	content = s.removeBlock(content, beginCompletionDelimiter, endCompletionDelimiter)
	return content
}

func (s *shellInfrastructure) removeBlock(content, beginDelimiter, endDelimiter string) string {
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
