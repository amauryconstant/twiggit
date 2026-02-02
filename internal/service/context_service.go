package service

import (
	"fmt"
	"os"

	"twiggit/internal/domain"
)

// ContextService provides context-aware operations
type ContextService struct {
	detector domain.ContextDetector
	resolver domain.ContextResolver
	config   *domain.Config
}

// NewContextService creates a new context service
func NewContextService(detector domain.ContextDetector, resolver domain.ContextResolver, cfg *domain.Config) *ContextService {
	return &ContextService{
		detector: detector,
		resolver: resolver,
		config:   cfg,
	}
}

// GetCurrentContext detects context from current working directory
func (cs *ContextService) GetCurrentContext() (*domain.Context, error) {
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
func (cs *ContextService) DetectContextFromPath(path string) (*domain.Context, error) {
	ctx, err := cs.detector.DetectContext(path)
	if err != nil {
		return nil, fmt.Errorf("failed to detect context from path %s: %w", path, err)
	}
	return ctx, nil
}

// ResolveIdentifier resolves identifier based on current context
func (cs *ContextService) ResolveIdentifier(identifier string) (*domain.ResolutionResult, error) {
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
func (cs *ContextService) ResolveIdentifierFromContext(ctx *domain.Context, identifier string) (*domain.ResolutionResult, error) {
	result, err := cs.resolver.ResolveIdentifier(ctx, identifier)
	if err != nil {
		return nil, fmt.Errorf("failed to resolve identifier '%s': %w", identifier, err)
	}
	return result, nil
}

// GetCompletionSuggestions provides completion suggestions based on current context
func (cs *ContextService) GetCompletionSuggestions(partial string) ([]*domain.ResolutionSuggestion, error) {
	ctx, err := cs.GetCurrentContext()
	if err != nil {
		return nil, fmt.Errorf("failed to get current context: %w", err)
	}

	suggestions, err := cs.resolver.GetResolutionSuggestions(ctx, partial)
	if err != nil {
		return nil, fmt.Errorf("failed to get completion suggestions: %w", err)
	}
	return suggestions, nil
}

// GetCompletionSuggestionsFromContext provides completion suggestions based on specified context
func (cs *ContextService) GetCompletionSuggestionsFromContext(ctx *domain.Context, partial string) ([]*domain.ResolutionSuggestion, error) {
	suggestions, err := cs.resolver.GetResolutionSuggestions(ctx, partial)
	if err != nil {
		return nil, fmt.Errorf("failed to get completion suggestions: %w", err)
	}
	return suggestions, nil
}
