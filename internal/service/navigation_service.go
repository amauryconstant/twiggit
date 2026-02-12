package service

import (
	"context"
	"errors"
	"fmt"
	"os"
	"path/filepath"

	"twiggit/internal/application"
	"twiggit/internal/domain"
)

// navigationService implements NavigationService interface
type navigationService struct {
	projectService application.ProjectService
	contextService application.ContextService
	config         *domain.Config
}

// NewNavigationService creates a new NavigationService instance
func NewNavigationService(
	projectService application.ProjectService,
	contextService application.ContextService,
	config *domain.Config,
) application.NavigationService {
	return &navigationService{
		projectService: projectService,
		contextService: contextService,
		config:         config,
	}
}

// ResolvePath resolves a target identifier to a concrete path
func (s *navigationService) ResolvePath(_ context.Context, req *domain.ResolvePathRequest) (*domain.ResolutionResult, error) {
	if req.Target == "" {
		return nil, errors.New("target cannot be empty")
	}

	// Delegate to ContextResolver for consistency
	result, err := s.contextService.ResolveIdentifierFromContext(req.Context, req.Target)
	if err != nil {
		return nil, fmt.Errorf("failed to resolve identifier: %w", err)
	}
	return result, nil
}

// ValidatePath validates that a path is accessible and valid
func (s *navigationService) ValidatePath(_ context.Context, path string) error {
	if path == "" {
		return errors.New("path cannot be empty")
	}

	// Check if path exists
	if _, err := os.Stat(path); os.IsNotExist(err) {
		return fmt.Errorf("path does not exist: %s", path)
	}

	// Check if path is accessible
	if !filepath.IsAbs(path) {
		absPath, err := filepath.Abs(path)
		if err != nil {
			return fmt.Errorf("failed to resolve absolute path: %w", err)
		}
		path = absPath
	}

	// Check if it's a directory
	info, err := os.Stat(path)
	if err != nil {
		return fmt.Errorf("failed to access path: %w", err)
	}

	if !info.IsDir() {
		return fmt.Errorf("path is not a directory: %s", path)
	}

	return nil
}

// GetNavigationSuggestions provides completion suggestions for navigation
func (s *navigationService) GetNavigationSuggestions(_ context.Context, context *domain.Context, partial string) ([]*domain.ResolutionSuggestion, error) {
	var suggestions []*domain.ResolutionSuggestion
	var err error

	// Get completion suggestions from context service
	// If a specific context is provided, use it; otherwise use current context
	if context != nil {
		suggestions, err = s.contextService.GetCompletionSuggestionsFromContext(context, partial)
	} else {
		suggestions, err = s.contextService.GetCompletionSuggestions(partial)
	}

	if err != nil {
		return nil, fmt.Errorf("failed to get completion suggestions: %w", err)
	}

	// Filter suggestions based on context if needed
	if context != nil && context.Type == domain.ContextProject {
		suggestions = s.filterSuggestionsForProject(suggestions, context.ProjectName)
	}

	// Apply limit from configuration
	maxSuggestions := s.config.Navigation.MaxSuggestions
	if maxSuggestions > 0 && len(suggestions) > maxSuggestions {
		suggestions = suggestions[:maxSuggestions]
	}

	return suggestions, nil
}

// filterSuggestionsForProject filters suggestions to be relevant to the current project
func (s *navigationService) filterSuggestionsForProject(suggestions []*domain.ResolutionSuggestion, projectName string) []*domain.ResolutionSuggestion {
	if projectName == "" {
		return suggestions
	}

	var filtered []*domain.ResolutionSuggestion
	for _, suggestion := range suggestions {
		// Include suggestions that match the current project or are general
		if suggestion.ProjectName == "" || suggestion.ProjectName == projectName {
			filtered = append(filtered, suggestion)
		}
	}

	return filtered
}
