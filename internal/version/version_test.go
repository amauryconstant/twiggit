package version

import (
	"os"
	"testing"

	"github.com/amaury/twiggit/test/mocks"
	"github.com/stretchr/testify/suite"
)

type VersionTestSuite struct {
	suite.Suite
	originalDir    string
	tempDir        string
	mockFileSystem *mocks.FileSystemMock
}

func (s *VersionTestSuite) SetupTest() {
	var err error
	s.originalDir, err = os.Getwd()
	s.Require().NoError(err)

	s.tempDir, err = os.MkdirTemp("", "version-test-*")
	s.Require().NoError(err)

	// Initialize mock filesystem
	s.mockFileSystem = mocks.NewFileSystemMock()

	// Change to temp directory for testing
	err = os.Chdir(s.tempDir)
	s.Require().NoError(err)
}

func (s *VersionTestSuite) TearDownTest() {
	// Change back to original directory
	err := os.Chdir(s.originalDir)
	s.Require().NoError(err)

	// Clean up temp directory
	if s.tempDir != "" {
		_ = os.RemoveAll(s.tempDir)
	}
}

func TestVersionSuite(t *testing.T) {
	suite.Run(t, new(VersionTestSuite))
}

func (s *VersionTestSuite) TestVersion_WithValidVersionComment() {
	// Create go.mod with version comment
	goModContent := `module github.com/test/twiggit

// Version: 1.2.3
go 1.21

require (
	github.com/stretchr/testify v1.11.1
)
`

	// Mock reading go.mod file
	s.mockFileSystem.On("ReadFile", "go.mod").Return([]byte(goModContent), nil)

	version := Version(WithFileSystem(s.mockFileSystem))
	s.Equal("1.2.3", version)

	// Verify mock expectations
	s.mockFileSystem.AssertExpectations(s.T())
}

func (s *VersionTestSuite) TestVersion_WithVersionCommentAndExtraSpaces() {
	// Create go.mod with version comment containing extra spaces
	goModContent := `module github.com/test/twiggit

// Version: 2.0.1   
go 1.21

require (
	github.com/stretchr/testify v1.11.1
)
`

	// Mock reading go.mod file
	s.mockFileSystem.On("ReadFile", "go.mod").Return([]byte(goModContent), nil)

	version := Version(WithFileSystem(s.mockFileSystem))
	// The regex captures the version without leading spaces, but trims the result
	s.Equal("2.0.1", version)

	// Verify mock expectations
	s.mockFileSystem.AssertExpectations(s.T())
}

func (s *VersionTestSuite) TestVersion_WithNoVersionComment() {
	// Create go.mod without version comment
	goModContent := `module github.com/test/twiggit

go 1.21

require (
	github.com/stretchr/testify v1.11.1
)
`

	// Mock reading go.mod file
	s.mockFileSystem.On("ReadFile", "go.mod").Return([]byte(goModContent), nil)

	version := Version(WithFileSystem(s.mockFileSystem))
	s.Equal("dev", version)

	// Verify mock expectations
	s.mockFileSystem.AssertExpectations(s.T())
}

func (s *VersionTestSuite) TestVersion_WithMalformedVersionComment() {
	// Create go.mod with malformed version comment
	goModContent := `module github.com/test/twiggit

// Version: 
go 1.21

require (
	github.com/stretchr/testify v1.11.1
)
`

	// Mock reading go.mod file
	s.mockFileSystem.On("ReadFile", "go.mod").Return([]byte(goModContent), nil)

	version := Version(WithFileSystem(s.mockFileSystem))
	s.Equal("dev", version)

	// Verify mock expectations
	s.mockFileSystem.AssertExpectations(s.T())
}

func (s *VersionTestSuite) TestVersion_WithDifferentCommentFormat() {
	// Create go.mod with different comment format
	goModContent := `module github.com/test/twiggit

// Some other comment
// Version: 3.4.5
// Another comment
go 1.21

require (
	github.com/stretchr/testify v1.11.1
)
`

	// Mock reading go.mod file
	s.mockFileSystem.On("ReadFile", "go.mod").Return([]byte(goModContent), nil)

	version := Version(WithFileSystem(s.mockFileSystem))
	s.Equal("3.4.5", version)

	// Verify mock expectations
	s.mockFileSystem.AssertExpectations(s.T())
}

func (s *VersionTestSuite) TestVersion_WithMissingGoModFile() {
	// Mock file not found error
	s.mockFileSystem.On("ReadFile", "go.mod").Return([]byte{}, &os.PathError{Op: "read", Path: "go.mod", Err: os.ErrNotExist})

	version := Version(WithFileSystem(s.mockFileSystem))
	s.Equal("dev", version)

	// Verify mock expectations
	s.mockFileSystem.AssertExpectations(s.T())
}

func (s *VersionTestSuite) TestVersion_WithUnreadableGoModFile() {
	// Mock permission error when reading go.mod file
	s.mockFileSystem.On("ReadFile", "go.mod").Return([]byte{}, &os.PathError{Op: "read", Path: "go.mod", Err: os.ErrPermission})

	version := Version(WithFileSystem(s.mockFileSystem))
	s.Equal("dev", version)

	// Verify mock expectations
	s.mockFileSystem.AssertExpectations(s.T())
}

func (s *VersionTestSuite) TestVersion_WithComplexVersion() {
	// Create go.mod with complex version string
	goModContent := `module github.com/test/twiggit

// Version: 1.0.0-alpha.1+build.123
go 1.21

require (
	github.com/stretchr/testify v1.11.1
)
`

	// Mock reading go.mod file
	s.mockFileSystem.On("ReadFile", "go.mod").Return([]byte(goModContent), nil)

	version := Version(WithFileSystem(s.mockFileSystem))
	s.Equal("1.0.0-alpha.1+build.123", version)

	// Verify mock expectations
	s.mockFileSystem.AssertExpectations(s.T())
}

func (s *VersionTestSuite) TestVersion_WithMultipleVersionComments() {
	// Create go.mod with multiple version comments (should match first)
	goModContent := `module github.com/test/twiggit

// Version: 5.0.0
// Some other comment
// Version: 6.0.0
go 1.21

require (
	github.com/stretchr/testify v1.11.1
)
`

	// Mock reading go.mod file
	s.mockFileSystem.On("ReadFile", "go.mod").Return([]byte(goModContent), nil)

	version := Version(WithFileSystem(s.mockFileSystem))
	s.Equal("5.0.0", version)

	// Verify mock expectations
	s.mockFileSystem.AssertExpectations(s.T())
}

func (s *VersionTestSuite) TestVersion_WithVersionInDifferentContext() {
	// Create go.mod with version-like text in different context
	goModContent := `module github.com/test/twiggit

// This is not a version: 7.8.9
// Version: 8.9.0
// Another version-like text: 9.0.1
go 1.21

require (
	github.com/stretchr/testify v1.11.1
)
`

	// Mock reading go.mod file
	s.mockFileSystem.On("ReadFile", "go.mod").Return([]byte(goModContent), nil)

	version := Version(WithFileSystem(s.mockFileSystem))
	s.Equal("8.9.0", version)

	// Verify mock expectations
	s.mockFileSystem.AssertExpectations(s.T())
}

func (s *VersionTestSuite) TestVersion_WithEmptyFile() {
	// Mock reading empty go.mod file
	s.mockFileSystem.On("ReadFile", "go.mod").Return([]byte(""), nil)

	version := Version(WithFileSystem(s.mockFileSystem))
	s.Equal("dev", version)

	// Verify mock expectations
	s.mockFileSystem.AssertExpectations(s.T())
}
