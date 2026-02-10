package services

import (
	"context"
	"errors"
	"fmt"

	"twiggit/internal/application"
	"twiggit/internal/domain"
	"twiggit/internal/infrastructure"
)

// shellService implements the ShellService interface
type shellService struct {
	integration infrastructure.ShellInfrastructure
	config      *domain.Config
}

// NewShellService creates a new ShellService instance
func NewShellService(
	integration infrastructure.ShellInfrastructure,
	config *domain.Config,
) application.ShellService {
	return &shellService{
		integration: integration,
		config:      config,
	}
}

// SetupShell sets up shell integration for the specified shell type
func (s *shellService) SetupShell(_ context.Context, req *domain.SetupShellRequest) (*domain.SetupShellResult, error) {
	// Infer shell type from config file if not specified
	shellType := req.ShellType
	if shellType == "" && req.ConfigFile != "" {
		inferredType, err := domain.InferShellTypeFromPath(req.ConfigFile)
		if err != nil {
			return nil, fmt.Errorf("failed to infer shell type: %w", err)
		}
		shellType = inferredType
	}

	// Validate shell type
	if !s.isValidShellType(shellType) {
		return nil, domain.NewShellError(domain.ErrInvalidShellType, string(shellType), "unsupported shell type")
	}

	// Use provided config file or detect one
	configFile := req.ConfigFile
	if configFile == "" {
		var err error
		configFile, err = s.integration.DetectConfigFile(shellType)
		if err != nil {
			return nil, fmt.Errorf("failed to detect config file: %w", err)
		}
	}

	// Check existing installation
	if !req.Force {
		if err := s.integration.ValidateInstallation(shellType, configFile); err == nil {
			return &domain.SetupShellResult{
				ShellType:  shellType,
				Installed:  true,
				Skipped:    true,
				ConfigFile: configFile,
				Message:    "Shell wrapper already installed",
			}, nil
		}
	}

	// Generate wrapper
	wrapper, err := s.integration.GenerateWrapper(shellType)
	if err != nil {
		return nil, fmt.Errorf("failed to generate wrapper: %w", err)
	}

	// Handle dry run
	if req.DryRun {
		return &domain.SetupShellResult{
			ShellType:      shellType,
			Installed:      false,
			DryRun:         true,
			WrapperContent: wrapper,
			ConfigFile:     configFile,
			Message:        "Dry run completed",
		}, nil
	}

	// Install wrapper
	if err := s.integration.InstallWrapper(shellType, wrapper, configFile, req.Force); err != nil {
		// Check if it's already installed error
		var shellErr *domain.ShellError
		if errors.As(err, &shellErr) && shellErr.Code == domain.ErrShellAlreadyInstalled {
			return &domain.SetupShellResult{
				ShellType:  shellType,
				Installed:  true,
				Skipped:    true,
				ConfigFile: configFile,
				Message:    "Shell wrapper already installed",
			}, nil
		}
		return nil, fmt.Errorf("failed to install wrapper: %w", err)
	}

	return &domain.SetupShellResult{
		ShellType:  shellType,
		Installed:  true,
		ConfigFile: configFile,
		Message:    "Shell wrapper installed successfully",
	}, nil
}

// ValidateInstallation validates whether shell integration is installed
func (s *shellService) ValidateInstallation(_ context.Context, req *domain.ValidateInstallationRequest) (*domain.ValidateInstallationResult, error) {
	// Infer shell type from config file if not specified
	shellType := req.ShellType
	if shellType == "" && req.ConfigFile != "" {
		inferredType, err := domain.InferShellTypeFromPath(req.ConfigFile)
		if err != nil {
			return nil, fmt.Errorf("failed to infer shell type: %w", err)
		}
		shellType = inferredType
	}

	// Validate shell type
	if !s.isValidShellType(shellType) {
		return nil, domain.NewShellError(domain.ErrInvalidShellType, string(shellType), "unsupported shell type")
	}

	// Use provided config file or detect one
	configFile := req.ConfigFile
	if configFile == "" {
		var err error
		configFile, err = s.integration.DetectConfigFile(shellType)
		if err != nil {
			return nil, fmt.Errorf("failed to detect config file: %w", err)
		}
	}

	// Validate installation
	err := s.integration.ValidateInstallation(shellType, configFile)
	if err != nil {
		var shellErr *domain.ShellError
		if errors.As(err, &shellErr) && shellErr.Code == domain.ErrShellNotInstalled {
			return &domain.ValidateInstallationResult{
				ShellType:  shellType,
				Installed:  false,
				ConfigFile: configFile,
				Message:    "Shell wrapper not installed",
			}, nil
		}
		return nil, fmt.Errorf("failed to validate installation: %w", err)
	}

	return &domain.ValidateInstallationResult{
		ShellType:  shellType,
		Installed:  true,
		ConfigFile: configFile,
		Message:    "Shell wrapper is installed",
	}, nil
}

// GenerateWrapper generates a shell wrapper for the specified shell type
func (s *shellService) GenerateWrapper(_ context.Context, req *domain.GenerateWrapperRequest) (*domain.GenerateWrapperResult, error) {
	// Pure function: validate request first
	if err := req.ValidateGenerateWrapperRequest(); err != nil {
		return nil, fmt.Errorf("validation failed: %w", err)
	}

	// Generate wrapper
	var wrapper string
	var err error

	if req.CustomTemplate != "" {
		// Use custom template with composition
		wrapper = s.composeWrapper(req.CustomTemplate, req.ShellType)
	} else {
		// Use standard template
		wrapper, err = s.integration.GenerateWrapper(req.ShellType)
		if err != nil {
			return nil, fmt.Errorf("failed to generate wrapper: %w", err)
		}
	}

	templateUsed := "standard"
	if req.CustomTemplate != "" {
		templateUsed = "custom"
	}

	return &domain.GenerateWrapperResult{
		ShellType:      req.ShellType,
		WrapperContent: wrapper,
		TemplateUsed:   templateUsed,
		Message:        "Wrapper generated successfully",
	}, nil
}

// composeWrapper composes the wrapper with template replacements (pure function)
func (s *shellService) composeWrapper(template string, shellType domain.ShellType) string {
	// Pure function: no side effects, deterministic output
	replacements := map[string]string{
		"{{SHELL_TYPE}}": string(shellType),
	}

	result := template
	for key, value := range replacements {
		result = fmt.Sprintf(result, key, value)
	}

	return result
}

// isValidShellType checks if the shell type is supported
func (s *shellService) isValidShellType(shellType domain.ShellType) bool {
	switch shellType {
	case domain.ShellBash, domain.ShellZsh, domain.ShellFish:
		return true
	default:
		return false
	}
}
