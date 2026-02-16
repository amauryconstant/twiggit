package domain

// ContextType represents the type of git context
type ContextType int

const (
	// ContextUnknown represents an unknown context type
	ContextUnknown ContextType = iota
	// ContextProject represents a project context
	ContextProject
	// ContextWorktree represents a worktree context
	ContextWorktree
	// ContextOutsideGit represents being outside any git repository
	ContextOutsideGit
)

// String returns the string representation of ContextType
func (c ContextType) String() string {
	switch c {
	case ContextProject:
		return "project"
	case ContextWorktree:
		return "worktree"
	case ContextOutsideGit:
		return "outside-git"
	default:
		return "unknown"
	}
}

// Context represents the detected git context
type Context struct {
	Type        ContextType
	ProjectName string
	BranchName  string // Only for ContextWorktree
	Path        string // Absolute path to context root
	Explanation string // Human-readable explanation of detection
}

// PathType represents the type of resolved path
type PathType int

const (
	// PathTypeProject represents a project path
	PathTypeProject PathType = iota
	// PathTypeWorktree represents a worktree path
	PathTypeWorktree
	// PathTypeInvalid represents an invalid path
	PathTypeInvalid
)

// String returns the string representation of PathType
func (p PathType) String() string {
	switch p {
	case PathTypeProject:
		return "project"
	case PathTypeWorktree:
		return "worktree"
	default:
		return "invalid"
	}
}

// ResolutionResult represents the result of identifier resolution
type ResolutionResult struct {
	ResolvedPath string
	Type         PathType
	ProjectName  string
	BranchName   string
	Explanation  string
}

// ResolutionSuggestion represents a completion suggestion
type ResolutionSuggestion struct {
	Text        string
	Description string
	Type        PathType
	ProjectName string
	BranchName  string
}

// SuggestionOption is a functional option for configuring resolution suggestions
type SuggestionOption func(interface{})

// ContextDetector detects the current git context
type ContextDetector interface {
	// DetectContext detects the context from the given directory
	DetectContext(dir string) (*Context, error)
}

// ContextResolver resolves target identifiers based on current context
type ContextResolver interface {
	// ResolveIdentifier resolves target identifier based on context
	ResolveIdentifier(ctx *Context, identifier string) (*ResolutionResult, error)

	// GetResolutionSuggestions provides completion suggestions
	GetResolutionSuggestions(ctx *Context, partial string, opts ...SuggestionOption) ([]*ResolutionSuggestion, error)
}
