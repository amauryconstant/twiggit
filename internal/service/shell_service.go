package service

import (
	"context"
	"errors"
	"fmt"

	"twiggit/internal/application"
	"twiggit/internal/domain"
)

var _ application.ShellService = (*shellService)(nil)

// shellService implements the ShellService interface
type shellService struct {
	integration application.ShellInfrastructure
	config      *domain.Config
}

// NewShellService creates a new ShellService instance
func NewShellService(
	integration application.ShellInfrastructure,
	config *domain.Config,
) application.ShellService {
	return &shellService{
		integration: integration,
		config:      config,
	}
}

// detectShellAndConfig auto-detects shell type and config file when both are empty
// or infers one from the other when only one is provided
func (s *shellService) detectShellAndConfig(shellType domain.ShellType, configFile string) (domain.ShellType, string, error) {
	// Auto-detect shell and config file when both are empty
	if shellType == "" && configFile == "" {
		var err error
		shellType, err = domain.DetectShellFromEnv()
		if err != nil {
			return "", "", fmt.Errorf("shell auto-detection failed: %w", err)
		}

		detectedConfig, err := s.integration.DetectConfigFile(shellType)
		if err != nil {
			return "", "", fmt.Errorf("config file detection failed: %w", err)
		}
		return shellType, detectedConfig, nil
	}

	// Infer shell type from config file if not specified
	if shellType == "" && configFile != "" {
		inferredType, err := domain.InferShellTypeFromPath(configFile)
		if err != nil {
			return "", "", fmt.Errorf("failed to infer shell type: %w", err)
		}
		shellType = inferredType
	}

	// Use provided config file or detect one
	if configFile == "" {
		var err error
		configFile, err = s.integration.DetectConfigFile(shellType)
		if err != nil {
			return "", "", fmt.Errorf("failed to detect config file: %w", err)
		}
	}

	return shellType, configFile, nil
}

// SetupShell sets up shell integration for the specified shell type
func (s *shellService) SetupShell(_ context.Context, req *domain.SetupShellRequest) (*domain.SetupShellResult, error) {
	shellType, configFile, err := s.detectShellAndConfig(req.ShellType, req.ConfigFile)
	if err != nil {
		return nil, err
	}

	// Validate shell type
	if !domain.IsValidShellType(shellType) {
		return nil, domain.NewShellError(domain.ErrInvalidShellType, string(shellType), "unsupported shell type")
	}

	// Check existing installation
	if !req.ForceOverwrite {
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

	// Install wrapper
	if err := s.integration.InstallWrapper(shellType, wrapper, configFile, req.ForceOverwrite); err != nil {
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
	shellType, configFile, err := s.detectShellAndConfig(req.ShellType, req.ConfigFile)
	if err != nil {
		return nil, err
	}

	// Validate shell type
	if !domain.IsValidShellType(shellType) {
		return nil, domain.NewShellError(domain.ErrInvalidShellType, string(shellType), "unsupported shell type")
	}

	// Validate installation
	err = s.integration.ValidateInstallation(shellType, configFile)
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
		wrapper = s.integration.ComposeWrapper(req.CustomTemplate, req.ShellType)
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
