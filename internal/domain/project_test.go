package domain

import (
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestProject_NewProject(t *testing.T) {
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
		t.Run(tc.name, func(t *testing.T) {
			project, err := NewProject(tc.projectName, tc.path)

			if tc.expectError {
				require.Error(t, err)
				assert.Contains(t, err.Error(), tc.errorMessage)
				assert.Nil(t, project)
			} else {
				require.NoError(t, err)
				assert.NotNil(t, project)
				assert.Equal(t, tc.projectName, project.Name())
				assert.Equal(t, tc.path, project.Path())
			}
		})
	}
}
