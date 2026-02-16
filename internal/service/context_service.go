package service

import (
	"fmt"
	"os"

	"twiggit/internal/application"
	"twiggit/internal/domain"
)

// contextService provides context-aware operations
type contextService struct {
	detector domain.ContextDetector
	resolver domain.ContextResolver
	config   *domain.Config
}

// NewContextService creates a new context service
func NewContextService(detector domain.ContextDetector, resolver domain.ContextResolver, cfg *domain.Config) application.ContextService {
	return &contextService{
		detector: detector,
		resolver: resolver,
		config:   cfg,
	}
}

// GetCurrentContext detects context from current working directory
func (cs *contextService) GetCurrentContext() (*domain.Context, error) {
	wd, err := os.Getwd()
	if err != nil {
		return nil, fmt.Errorf("failed to get working directory: %w", err)
	}

	ctx, err := cs.detector.DetectContext(wd)
	if err != nil {
		return nil, fmt.Errorf("failed to detect context: %w", err)
	}
	return ctx, nil
}

// DetectContextFromPath detects context from specified path
func (cs *contextService) DetectContextFromPath(path string) (*domain.Context, error) {
	ctx, err := cs.detector.DetectContext(path)
	if err != nil {
		return nil, fmt.Errorf("failed to detect context from path %s: %w", path, err)
	}
	return ctx, nil
}

// ResolveIdentifier resolves identifier based on current context
func (cs *contextService) ResolveIdentifier(identifier string) (*domain.ResolutionResult, error) {
	ctx, err := cs.GetCurrentContext()
	if err != nil {
		return nil, fmt.Errorf("failed to get current context: %w", err)
	}

	result, err := cs.resolver.ResolveIdentifier(ctx, identifier)
	if err != nil {
		return nil, fmt.Errorf("failed to resolve identifier '%s': %w", identifier, err)
	}
	return result, nil
}

// ResolveIdentifierFromContext resolves identifier based on specified context
func (cs *contextService) ResolveIdentifierFromContext(ctx *domain.Context, identifier string) (*domain.ResolutionResult, error) {
	result, err := cs.resolver.ResolveIdentifier(ctx, identifier)
	if err != nil {
		return nil, fmt.Errorf("failed to resolve identifier '%s': %w", identifier, err)
	}
	return result, nil
}

// GetCompletionSuggestions provides completion suggestions based on current context
func (cs *contextService) GetCompletionSuggestions(partial string, opts ...domain.SuggestionOption) ([]*domain.ResolutionSuggestion, error) {
	ctx, err := cs.GetCurrentContext()
	if err != nil {
		return nil, fmt.Errorf("failed to get current context: %w", err)
	}

	suggestions, err := cs.resolver.GetResolutionSuggestions(ctx, partial, opts...)
	if err != nil {
		return nil, fmt.Errorf("failed to get completion suggestions: %w", err)
	}
	return suggestions, nil
}

// GetCompletionSuggestionsFromContext provides completion suggestions based on specified context
func (cs *contextService) GetCompletionSuggestionsFromContext(ctx *domain.Context, partial string, opts ...domain.SuggestionOption) ([]*domain.ResolutionSuggestion, error) {
	suggestions, err := cs.resolver.GetResolutionSuggestions(ctx, partial, opts...)
	if err != nil {
		return nil, fmt.Errorf("failed to get completion suggestions: %w", err)
	}
	return suggestions, nil
}
