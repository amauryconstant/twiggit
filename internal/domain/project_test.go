package domain

import (
	"testing"

	"github.com/stretchr/testify/suite"
)

type ProjectTestSuite struct {
	suite.Suite
}

func TestProjectSuite(t *testing.T) {
	suite.Run(t, new(ProjectTestSuite))
}

func (s *ProjectTestSuite) TestNewProject() {
	testCases := []struct {
		name         string
		projectName  string
		path         string
		expectError  bool
		errorMessage string
	}{
		{
			name:        "valid project",
			projectName: "my-project",
			path:        "/path/to/project",
			expectError: false,
		},
		{
			name:         "empty project name",
			projectName:  "",
			path:         "/path/to/project",
			expectError:  true,
			errorMessage: "new project: name cannot be empty",
		},
		{
			name:         "empty project path",
			projectName:  "my-project",
			path:         "",
			expectError:  true,
			errorMessage: "new project: path cannot be empty",
		},
	}

	for _, tc := range testCases {
		s.Run(tc.name, func() {
			project, err := NewProject(tc.projectName, tc.path)

			if tc.expectError {
				s.Require().Error(err)
				s.Contains(err.Error(), tc.errorMessage)
				s.Nil(project)
			} else {
				s.Require().NoError(err)
				s.NotNil(project)
				s.Equal(tc.projectName, project.Name())
				s.Equal(tc.path, project.Path())
			}
		})
	}
}
