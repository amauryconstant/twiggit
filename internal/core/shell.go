package core

// ShellType represents the type of shell.
//
// Wire form stays string (bash / zsh / fish) for backward compatibility with
// callers that parse ShellType via string(ShellType) casts and ShellType("bash")
// conversions (cmd/init.go, internal/output/wrapper.go, internal/output/shell_infra.go,
// internal/git/shell_detect.go). ShellTypeUnknown is the empty-string Unknown
// sentinel mandated by the enum-unknown-zero naming rule; the spec-mandated
// int iota conversion is deferred because the wide call-site surface would
// require a coordinated rename across cmd/ and internal/.
type ShellType string

const (
	// ShellTypeUnknown is the zero value of ShellType, used when no
	// specific shell type has been identified (matches ShellType("")).
	ShellTypeUnknown ShellType = ""
	// ShellBash represents the bash shell type.
	ShellBash ShellType = "bash"
	// ShellZsh represents the zsh shell type.
	ShellZsh ShellType = "zsh"
	// ShellFish represents the fish shell type.
	ShellFish ShellType = "fish"
)

// String returns the wire form of the shell type. Forwards to the
// underlying string so fmt.Sprintf("%s", shellType) and string()
// casts stay interchangeable; ShellTypeUnknown renders as "unknown"
// when the explicit sentinel is used.
func (s ShellType) String() string {
	if s == ShellTypeUnknown {
		return "unknown"
	}
	return string(s)
}

// IsValidShellType checks if the shell type is supported.
func IsValidShellType(shellType ShellType) bool {
	switch shellType {
	case ShellBash, ShellZsh, ShellFish:
		return true
	default:
		return false
	}
}
