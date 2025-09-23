package domain

import (
	"fmt"
	"strings"
)

// Package domain contains core business entities and validation rules for the twiggit application.
//
// Validation Strategy:
// - Domain Layer: Validates business rules and data formats only
// - Service Layer: Handles infrastructure-specific validation (filesystem access, git operations)
// - Validation Results: Unified ValidationResult type with entity-specific constructors
//
// Domain Validation Rules:
// - Project: Validates name format and git repository path format
// - Workspace: Validates path format and project name uniqueness within workspace
// - Worktree: Validates branch naming conventions and path format rules
//
// Common Validation Patterns:
// - Use ValidateNotEmpty for required string fields
// - Use ValidateNotNil for required entity references
// - Use MergeValidationResults to combine multiple validation results
// - Use entity-specific constructors for domain language clarity

// Validation constants for domain entities

const (
	// MinStringLength is the minimum length for string validation
	MinStringLength = 1
	// MaxStringLength is the maximum length for string validation
	MaxStringLength = 4096

	// MinBranchNameLength is the minimum length for branch names
	MinBranchNameLength = 1
	// MaxBranchNameLength is the maximum length for branch names
	MaxBranchNameLength = 250

	// MaxPathLength is the maximum length for filesystem paths
	MaxPathLength = 255
)

// InvalidBranchChars contains characters not allowed in git branch names
// Based on git-check-ref-format rules
var InvalidBranchChars = []string{" ", "\t", "\n", "\r", "~", "^", ":", "?", "*", "[", "\\", "..", "//"}

// ValidationResult represents a unified validation result for all domain entities
// This replaces ValidationResult, ProjectValidationResult, and WorkspaceValidationResult
type ValidationResult struct {
	Valid    bool
	Errors   []*DomainError
	Warnings []string
}

// NewValidationResult creates a new validation result with default valid state
func NewValidationResult() *ValidationResult {
	return &ValidationResult{
		Valid:    true,
		Errors:   make([]*DomainError, 0),
		Warnings: make([]string, 0),
	}
}

// AddError adds an error to the validation result and marks it as invalid
func (vr *ValidationResult) AddError(err *DomainError) {
	vr.Errors = append(vr.Errors, err)
	vr.Valid = len(vr.Errors) == 0
}

// AddWarning adds a warning to the validation result
func (vr *ValidationResult) AddWarning(warning string) {
	vr.Warnings = append(vr.Warnings, warning)
}

// HasErrors returns true if the validation result has any errors
func (vr *ValidationResult) HasErrors() bool {
	return len(vr.Errors) > 0
}

// FirstError returns the first error in the validation result, or nil if no errors
func (vr *ValidationResult) FirstError() *DomainError {
	if len(vr.Errors) == 0 {
		return nil
	}
	return vr.Errors[0]
}

// ToError returns the first error as an error interface, or nil if no errors
func (vr *ValidationResult) ToError() error {
	if len(vr.Errors) == 0 {
		return nil
	}
	return vr.Errors[0]
}

// Merge combines this validation result with another, returning a new result
func (vr *ValidationResult) Merge(other *ValidationResult) *ValidationResult {
	merged := NewValidationResult()

	// Merge errors
	merged.Errors = append(vr.Errors, other.Errors...)

	// Merge warnings
	merged.Warnings = append(vr.Warnings, other.Warnings...)

	// Update validity
	merged.Valid = len(merged.Errors) == 0

	return merged
}

// Entity-specific constructors for better domain language

// NewProjectValidationResult creates a validation result for project operations
func NewProjectValidationResult() *ValidationResult {
	return NewValidationResult()
}

// NewWorktreeValidationResult creates a validation result for worktree operations
func NewWorktreeValidationResult() *ValidationResult {
	return NewValidationResult()
}

// NewWorkspaceValidationResult creates a validation result for workspace operations
func NewWorkspaceValidationResult() *ValidationResult {
	return NewValidationResult()
}

// Legacy compatibility methods for workspace validation result
// These maintain backward compatibility during migration

// IsValid returns true if the validation result is valid (workspace compatibility)
func (vr *ValidationResult) IsValid() bool {
	return vr.Valid
}

// GetErrorCount returns the number of errors (workspace compatibility)
func (vr *ValidationResult) GetErrorCount() int {
	return len(vr.Errors)
}

// GetFirstError returns the first error (workspace compatibility)
func (vr *ValidationResult) GetFirstError() *DomainError {
	return vr.FirstError()
}

// Common validation helpers for cross-cutting concerns

// ValidateNotEmpty validates that a string is not empty after trimming
func ValidateNotEmpty(value, fieldName string, errorType DomainErrorType, errorConstructor func(DomainErrorType, string, string) *DomainError) *ValidationResult {
	result := NewValidationResult()

	if strings.TrimSpace(value) == "" {
		result.AddError(errorConstructor(
			errorType,
			fieldName+" cannot be empty",
			"",
		).WithSuggestion("Provide a valid " + fieldName))
	}

	return result
}

// ValidateNotNil validates that an entity is not nil
func ValidateNotNil(entity interface{}, entityName string, errorType DomainErrorType, errorConstructor func(DomainErrorType, string) *DomainError) *ValidationResult {
	result := NewValidationResult()

	if entity == nil {
		result.AddError(errorConstructor(
			errorType,
			entityName+" cannot be nil",
		))
	}

	return result
}

// MergeValidationResults merges multiple validation results into one
func MergeValidationResults(results ...*ValidationResult) *ValidationResult {
	merged := NewValidationResult()

	for _, result := range results {
		merged.Errors = append(merged.Errors, result.Errors...)
		merged.Warnings = append(merged.Warnings, result.Warnings...)
	}

	merged.Valid = len(merged.Errors) == 0
	return merged
}

// ValidateStringLength validates string length within bounds
func ValidateStringLength(value, fieldName string, minLength, maxLength int, errorType DomainErrorType, errorConstructor func(DomainErrorType, string, string) *DomainError) *ValidationResult {
	result := NewValidationResult()

	if len(value) < minLength {
		result.AddError(errorConstructor(
			errorType,
			fmt.Sprintf("%s too short (minimum %d characters)", fieldName, minLength),
			"",
		).WithSuggestion("Use a longer " + fieldName))
	}

	if len(value) > maxLength {
		result.AddError(errorConstructor(
			errorType,
			fmt.Sprintf("%s too long (maximum %d characters)", fieldName, maxLength),
			"",
		).WithSuggestion("Use a shorter " + fieldName))
	}

	return result
}

// ValidateCharacters validates that a string doesn't contain invalid characters
func ValidateCharacters(value, fieldName string, invalidChars []string, errorType DomainErrorType, errorConstructor func(DomainErrorType, string, string) *DomainError) *ValidationResult {
	result := NewValidationResult()

	for _, invalidChar := range invalidChars {
		if strings.Contains(value, invalidChar) {
			result.AddError(errorConstructor(
				errorType,
				fmt.Sprintf("%s contains invalid character: '%s'", fieldName, invalidChar),
				"",
			).WithSuggestion("Remove invalid characters from " + fieldName))
		}
	}

	return result
}
