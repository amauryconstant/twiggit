package infrastructure

import (
	"context"
	"fmt"
	"os"
	"strings"
	"time"

	"github.com/knadh/koanf/parsers/toml"
	"github.com/knadh/koanf/providers/file"
	"github.com/knadh/koanf/v2"

	"twiggit/internal/domain"
)

type hookRunner struct {
	executor       CommandExecutor
	defaultTimeout time.Duration
}

// NewHookRunner creates a new HookRunner for executing post-create hooks
func NewHookRunner(executor CommandExecutor) HookRunner {
	return &hookRunner{
		executor:       executor,
		defaultTimeout: 30 * time.Second,
	}
}

func (r *hookRunner) Run(ctx context.Context, req *HookRunRequest) (*domain.HookResult, error) {
	if req.ConfigFilePath == "" {
		return &domain.HookResult{
			HookType: req.HookType,
			Executed: false,
			Success:  true,
			Failures: nil,
		}, nil
	}

	if _, err := os.Stat(req.ConfigFilePath); os.IsNotExist(err) {
		return &domain.HookResult{
			HookType: req.HookType,
			Executed: false,
			Success:  true,
			Failures: nil,
		}, nil
	}

	config, err := r.readHookConfig(req.ConfigFilePath)
	if err != nil {
		fmt.Fprintf(os.Stderr, "warning: failed to parse %s: %v\n", req.ConfigFilePath, err)
		return &domain.HookResult{
			HookType: req.HookType,
			Executed: false,
			Success:  true,
			Failures: nil,
		}, nil
	}

	if config == nil {
		return &domain.HookResult{
			HookType: req.HookType,
			Executed: false,
			Success:  true,
			Failures: nil,
		}, nil
	}

	var definition *domain.HookDefinition
	switch req.HookType {
	case domain.HookPostCreate:
		definition = config.PostCreate
	default:
		return &domain.HookResult{
			HookType: req.HookType,
			Executed: false,
			Success:  true,
			Failures: nil,
		}, nil
	}

	if definition == nil || len(definition.Commands) == 0 {
		return &domain.HookResult{
			HookType: req.HookType,
			Executed: false,
			Success:  true,
			Failures: nil,
		}, nil
	}

	return r.executeCommands(ctx, req, definition.Commands)
}

func (r *hookRunner) readHookConfig(path string) (*domain.HookConfig, error) {
	k := koanf.New(".")

	if err := k.Load(file.Provider(path), toml.Parser()); err != nil {
		return nil, fmt.Errorf("failed to parse TOML: %w", err)
	}

	var hookConfig struct {
		Hooks *domain.HookConfig `koanf:"hooks"`
	}

	if err := k.Unmarshal("", &hookConfig); err != nil {
		return nil, fmt.Errorf("failed to unmarshal config: %w", err)
	}

	return hookConfig.Hooks, nil
}

func (r *hookRunner) executeCommands(ctx context.Context, req *HookRunRequest, commands []string) (*domain.HookResult, error) {
	result := &domain.HookResult{
		HookType: req.HookType,
		Executed: true,
		Success:  true,
		Failures: nil,
	}

	envExports := r.buildEnvExports(req)

	for _, cmd := range commands {
		if strings.TrimSpace(cmd) == "" {
			continue
		}

		fullCmd := envExports + cmd
		cmdResult, err := r.executor.ExecuteWithTimeout(ctx, req.WorktreePath, "sh", r.defaultTimeout, "-c", fullCmd)

		if err != nil || cmdResult.ExitCode != 0 {
			result.Success = false
			exitCode := -1
			output := ""
			if cmdResult != nil {
				exitCode = cmdResult.ExitCode
				output = strings.TrimSpace(cmdResult.Stdout)
				if cmdResult.Stderr != "" {
					if output != "" {
						output += "\n"
					}
					output += strings.TrimSpace(cmdResult.Stderr)
				}
			}
			result.Failures = append(result.Failures, domain.HookFailure{
				Command:  cmd,
				ExitCode: exitCode,
				Output:   output,
			})
		}
	}

	return result, nil
}

func (r *hookRunner) buildEnvExports(req *HookRunRequest) string {
	var exports strings.Builder
	if req.WorktreePath != "" {
		exports.WriteString(fmt.Sprintf("export TWIGGIT_WORKTREE_PATH=%q ", req.WorktreePath))
	}
	if req.ProjectName != "" {
		exports.WriteString(fmt.Sprintf("export TWIGGIT_PROJECT_NAME=%q ", req.ProjectName))
	}
	if req.BranchName != "" {
		exports.WriteString(fmt.Sprintf("export TWIGGIT_BRANCH_NAME=%q ", req.BranchName))
	}
	if req.SourceBranch != "" {
		exports.WriteString(fmt.Sprintf("export TWIGGIT_SOURCE_BRANCH=%q ", req.SourceBranch))
	}
	if req.MainRepoPath != "" {
		exports.WriteString(fmt.Sprintf("export TWIGGIT_MAIN_REPO_PATH=%q ", req.MainRepoPath))
	}
	return exports.String()
}
