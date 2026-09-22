package core

// ShellType represents the type of shell.
type ShellType string

const (
	// ShellBash represents the bash shell type.
	ShellBash ShellType = "bash"
	// ShellZsh represents the zsh shell type.
	ShellZsh ShellType = "zsh"
	// ShellFish represents the fish shell type.
	ShellFish ShellType = "fish"
)

// IsValidShellType checks if the shell type is supported.
func IsValidShellType(shellType ShellType) bool {
	switch shellType {
	case ShellBash, ShellZsh, ShellFish:
		return true
	default:
		return false
	}
}
