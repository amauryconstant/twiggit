package cmd

import (
	"errors"
	"fmt"
	"strings"

	"github.com/amaury/twiggit/internal/domain"
)

// FormatDomainError formats domain errors with suggestions for better user experience.
// This function transforms domain errors into user-friendly messages with:
// - Emoji indicators for quick visual identification
// - Structured error information (path, type, suggestions)
// - Consistent formatting across all CLI commands
//
// This function is designed to work with Cobra's error handling system and
// provides a unified error presentation layer for the CLI.
//
// Example output:
//
//	❌ worktree already exists
//	   Path: /path/to/worktree
//	   Type: worktree
//
//	💡 Suggestions:
//	   • Use 'twiggit cd' to navigate to existing worktree
//	   • Use 'twiggit list' to see all available worktrees
func FormatDomainError(err error) error {
	if err == nil {
		return nil
	}

	// Check if it's a domain error
	var domainErr *domain.DomainError
	if errors.As(err, &domainErr) {
		return formatDomainErrorWithDetails(domainErr)
	}

	// For wrapped errors, check if the underlying error is a domain error
	if errors.As(err, &domainErr) {
		return formatDomainErrorWithDetails(domainErr)
	}

	// For non-domain errors, apply consistent formatting
	return fmt.Errorf("❌ %s", err.Error())
}

// formatDomainErrorWithDetails formats a domain error with all its details.
// This internal function builds the complete error message string including
// emoji, main message, path context, entity type, suggestions, and error codes.
// The formatted output provides comprehensive error information for users.
func formatDomainErrorWithDetails(domainErr *domain.DomainError) error {
	var msg strings.Builder

	// Main error message with emoji
	msg.WriteString(fmt.Sprintf("❌ %s\n", domainErr.Message))

	// Add path context if available
	if domainErr.Path != "" {
		msg.WriteString(fmt.Sprintf("   Path: %s\n", domainErr.Path))
	}

	// Add entity type context if available
	if domainErr.EntityType != "" {
		msg.WriteString(fmt.Sprintf("   Type: %s\n", domainErr.EntityType))
	}

	// Add suggestions if available
	if len(domainErr.Suggestions) > 0 {
		msg.WriteString("\n💡 Suggestions:\n")
		for _, suggestion := range domainErr.Suggestions {
			msg.WriteString(fmt.Sprintf("   • %s\n", suggestion))
		}
	}

	return errors.New(msg.String())
}
