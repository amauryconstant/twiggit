package domain

import (
	"testing"

	"github.com/stretchr/testify/suite"
)

// ValidationTestSuite provides test setup for validation tests
type ValidationTestSuite struct {
	suite.Suite
}

func TestValidationSuite(t *testing.T) {
	suite.Run(t, new(ValidationTestSuite))
}

func (s *ValidationTestSuite) TestValidationResult_AddError() {
	result := NewValidationResult()
	s.True(result.Valid)

	err := NewWorktreeError(ErrInvalidBranchName, "test error", "")
	result.AddError(err)

	s.False(result.Valid)
	s.Len(result.Errors, 1)
	s.Equal(err, result.Errors[0])
}

func (s *ValidationTestSuite) TestValidationResult_AddWarning() {
	result := NewValidationResult()
	result.AddWarning("test warning")

	s.True(result.Valid) // Warnings don't affect validity
	s.Len(result.Warnings, 1)
	s.Contains(result.Warnings, "test warning")
}

func (s *ValidationTestSuite) TestValidationResult_HasErrors() {
	result := NewValidationResult()
	s.False(result.HasErrors())

	result.AddError(NewWorktreeError(ErrValidation, "test", ""))
	s.True(result.HasErrors())
}

func (s *ValidationTestSuite) TestValidationResult_FirstError() {
	result := NewValidationResult()
	s.Require().Nil(result.FirstError())

	err1 := NewWorktreeError(ErrValidation, "first error", "")
	err2 := NewWorktreeError(ErrValidation, "second error", "")

	result.AddError(err1)
	result.AddError(err2)

	s.Equal(err1, result.FirstError())
}
