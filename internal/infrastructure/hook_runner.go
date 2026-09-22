package infrastructure

import (
	"context"
	"errors"
	"fmt"
	"os"
	"path/filepath"
	"strings"
	"time"

	"github.com/knadh/koanf/parsers/toml"
	"github.com/knadh/koanf/providers/file"
	"github.com/knadh/koanf/v2"

	"twiggit/internal/application"
	"twiggit/internal/core"
)

var _ application.HookRunner = (*hookRunner)(nil)

type hookRunner struct {
	executor       CommandExecutor
	defaultTimeout time.Duration
}

// NewHookRunner creates a new HookRunner for executing post-create hooks
func NewHookRunner(executor CommandExecutor, hookTimeoutSeconds ...int) application.HookRunner {
	defaultTimeout := 30 * time.Second
	if len(hookTimeoutSeconds) > 0 {
		defaultTimeout = time.Duration(hookTimeoutSeconds[0]) * time.Second
	}
	return &hookRunner{
		executor:       executor,
		defaultTimeout: defaultTimeout,
	}
}

func (r *hookRunner) Run(ctx context.Context, req *application.HookRunRequest) (*core.HookResult, error) {
	if req.ConfigFilePath == "" {
		return noOpResult(req), nil
	}

	parentDir := filepath.Dir(req.ConfigFilePath)
	root, rootErr := os.OpenRoot(parentDir)
	if rootErr != nil {
		return noOpResult(req), nil
	}
	defer root.Close() //nolint:errcheck // read-only filesystem stat cleanup, no actionable error
	if _, err := root.Stat(filepath.Base(req.ConfigFilePath)); errors.Is(err, os.ErrNotExist) {
		return noOpResult(req), nil
	}

	config, err := r.readHookConfig(req.ConfigFilePath)
	if err != nil {
		fmt.Fprintf(os.Stderr, "warning: failed to parse %s: %v\n", req.ConfigFilePath, err)
		return noOpResult(req), nil
	}

	if config == nil {
		return noOpResult(req), nil
	}

	var definition *core.HookDefinition
	switch req.HookType {
	case core.HookPostCreate:
		definition = config.PostCreate
	default:
		return noOpResult(req), nil
	}

	if definition == nil || len(definition.Commands) == 0 {
		return noOpResult(req), nil
	}

	return r.executeCommands(ctx, req, definition.Commands)
}

func noOpResult(req *application.HookRunRequest) *core.HookResult {
	return &core.HookResult{
		HookType:     req.HookType,
		HasExecuted:  false,
		IsSuccessful: true,
		Failures:     nil,
	}
}

func (r *hookRunner) readHookConfig(path string) (*core.HookConfig, error) {
	k := koanf.New(".")

	if err := k.Load(file.Provider(path), toml.Parser()); err != nil {
		return nil, fmt.Errorf("failed to parse TOML: %w", err)
	}

	var hookConfig struct {
		Hooks *core.HookConfig `koanf:"hooks"`
	}

	if err := k.Unmarshal("", &hookConfig); err != nil {
		return nil, fmt.Errorf("failed to unmarshal config: %w", err)
	}

	return hookConfig.Hooks, nil
}

func (r *hookRunner) executeCommands(ctx context.Context, req *application.HookRunRequest, commands []string) (*core.HookResult, error) {
	result := &core.HookResult{
		HookType:     req.HookType,
		HasExecuted:  true,
		IsSuccessful: true,
		Failures:     nil,
	}

	envExports := r.buildEnvExports(req)

	for _, cmd := range commands {
		if strings.TrimSpace(cmd) == "" {
			continue
		}

		fullCmd := envExports + cmd
		cmdResult, err := r.executor.ExecuteWithTimeout(ctx, req.WorktreePath, "sh", r.defaultTimeout, "-c", fullCmd)

		if err != nil || cmdResult == nil || cmdResult.ExitCode != 0 {
			result.IsSuccessful = false
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
			result.Failures = append(result.Failures, core.HookFailure{
				Command:  cmd,
				ExitCode: exitCode,
				Output:   output,
			})
		}
	}

	return result, nil
}

func (r *hookRunner) buildEnvExports(req *application.HookRunRequest) string {
	var exports strings.Builder
	if req.WorktreePath != "" {
		exports.WriteString(fmt.Sprintf("export TWIGGIT_WORKTREE_PATH=%q; ", req.WorktreePath))
	}
	if req.ProjectName != "" {
		exports.WriteString(fmt.Sprintf("export TWIGGIT_PROJECT_NAME=%q; ", req.ProjectName))
	}
	if req.BranchName != "" {
		exports.WriteString(fmt.Sprintf("export TWIGGIT_BRANCH_NAME=%q; ", req.BranchName))
	}
	if req.SourceBranch != "" {
		exports.WriteString(fmt.Sprintf("export TWIGGIT_SOURCE_BRANCH=%q; ", req.SourceBranch))
	}
	if req.MainRepoPath != "" {
		exports.WriteString(fmt.Sprintf("export TWIGGIT_MAIN_REPO_PATH=%q; ", req.MainRepoPath))
	}
	return exports.String()
}
