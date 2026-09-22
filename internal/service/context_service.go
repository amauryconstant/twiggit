package service

import (
	"fmt"
	"os"

	"twiggit/internal/application"
	"twiggit/internal/core"
)

var _ application.ContextService = (*contextService)(nil)

type contextService struct {
	detector application.ContextDetector
	resolver application.ContextResolver
}

func NewContextService(detector application.ContextDetector, resolver application.ContextResolver) application.ContextService {
	return &contextService{
		detector: detector,
		resolver: resolver,
	}
}

func (cs *contextService) GetCurrentContext() (*core.Context, error) {
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

func (cs *contextService) DetectContextFromPath(path string) (*core.Context, error) {
	ctx, err := cs.detector.DetectContext(path)
	if err != nil {
		return nil, fmt.Errorf("failed to detect context from path %s: %w", path, err)
	}
	return ctx, nil
}

func (cs *contextService) ResolveIdentifier(identifier string) (*core.ResolutionResult, error) {
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

func (cs *contextService) ResolveIdentifierFromContext(ctx *core.Context, identifier string) (*core.ResolutionResult, error) {
	result, err := cs.resolver.ResolveIdentifier(ctx, identifier)
	if err != nil {
		return nil, fmt.Errorf("failed to resolve identifier '%s': %w", identifier, err)
	}
	return result, nil
}

func (cs *contextService) GetCompletionSuggestions(partial string, opts ...core.SuggestionOption) ([]*core.ResolutionSuggestion, error) {
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

func (cs *contextService) GetCompletionSuggestionsFromContext(ctx *core.Context, partial string, opts ...core.SuggestionOption) ([]*core.ResolutionSuggestion, error) {
	suggestions, err := cs.resolver.GetResolutionSuggestions(ctx, partial, opts...)
	if err != nil {
		return nil, fmt.Errorf("failed to get completion suggestions: %w", err)
	}
	return suggestions, nil
}
