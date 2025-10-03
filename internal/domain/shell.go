package domain

import "fmt"

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
	ConfigFiles() []string
	WrapperTemplate() string
}

type shell struct {
	shellType ShellType
	path      string
	version   string
}

// NewShell creates a new Shell instance with validation
func NewShell(shellType ShellType, path, version string) (Shell, error) {
	if !isValidShellType(shellType) {
		return nil, fmt.Errorf("unsupported shell type: %s", shellType)
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

// ConfigFiles returns the list of configuration files for this shell
func (s *shell) ConfigFiles() []string {
	switch s.shellType {
	case ShellBash:
		return []string{".bashrc", ".bash_profile", ".profile"}
	case ShellZsh:
		return []string{".zshrc", ".zprofile", ".profile"}
	case ShellFish:
		return []string{"config.fish", ".fishrc"}
	default:
		return []string{}
	}
}

// WrapperTemplate returns the wrapper template for this shell
func (s *shell) WrapperTemplate() string {
	switch s.shellType {
	case ShellBash:
		return s.bashWrapperTemplate()
	case ShellZsh:
		return s.zshWrapperTemplate()
	case ShellFish:
		return s.fishWrapperTemplate()
	default:
		return ""
	}
}

// isValidShellType checks if the shell type is supported
func isValidShellType(shellType ShellType) bool {
	switch shellType {
	case ShellBash, ShellZsh, ShellFish:
		return true
	default:
		return false
	}
}

// bashWrapperTemplate returns the bash wrapper template
func (s *shell) bashWrapperTemplate() string {
	return `# Twiggit bash wrapper
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
func (s *shell) zshWrapperTemplate() string {
	return `# Twiggit zsh wrapper
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
func (s *shell) fishWrapperTemplate() string {
	return `# Twiggit fish wrapper
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
