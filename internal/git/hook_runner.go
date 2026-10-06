package git

import (
	"context"
	"errors"
	"fmt"
	"log/slog"
	"os"
	"path/filepath"
	"strings"
	"time"
	"twiggit/internal/core"

	"github.com/knadh/koanf/parsers/toml"
	"github.com/knadh/koanf/providers/file"
	"github.com/knadh/koanf/v2"
)

// *HookRunner satisfies the consumer-side HookRunner interface declared
// in cmdutil implicitly (Run signature matches). We don't add an
// explicit assertion here because it would create an import cycle:
// cmdutil already imports git for Factory.GitClient.

// HookRunner executes post-create hooks from .twiggit.toml configs.
type HookRunner struct {
	executor       CommandExecutor
	defaultTimeout time.Duration
}

// NewHookRunner creates a new HookRunner for executing post-create hooks
func NewHookRunner(executor CommandExecutor, hookTimeoutSeconds ...int) *HookRunner {
	defaultTimeout := 30 * time.Second
	if len(hookTimeoutSeconds) > 0 {
		defaultTimeout = time.Duration(hookTimeoutSeconds[0]) * time.Second
	}
	return &HookRunner{
		executor:       executor,
		defaultTimeout: defaultTimeout,
	}
}

// Run executes hooks of the specified type with the given request context.
// PreRebase failures are surfaced as a HardFailure result so the cmd
// layer can short-circuit the rebase; all other failure modes are
// recorded in the result without blocking the calling operation.
func (r *HookRunner) Run(ctx context.Context, req *core.HookRunRequest) (*core.HookResult, error) {
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
		slog.Default().Warn("hook config parse warning", "path", req.ConfigFilePath, "err", err)
		return noOpResult(req), nil
	}

	if config == nil {
		return noOpResult(req), nil
	}

	definitions, hardFailure := r.selectDefinitions(config, req.HookType)
	if len(definitions) == 0 {
		return noOpResult(req), nil
	}

	return r.executeDefinitions(ctx, req, definitions, hardFailure), nil
}

// selectDefinitions returns the hook definitions for the given hook
// type and whether the lifecycle stage treats a non-zero exit as a
// hard failure (PreRebase) versus a warning (PostCreate, PostRebase,
// PostSync).
func (r *HookRunner) selectDefinitions(config *core.HookConfig, hookType core.HookType) ([]core.HookDefinition, bool) {
	switch hookType {
	case core.HookTypePostCreate:
		return config.PostCreate, false
	case core.HookTypePreRebase:
		return config.PreRebase, true
	case core.HookTypePostRebase:
		return config.PostRebase, false
	case core.HookTypePostSync:
		return config.PostSync, false
	default:
		return nil, false
	}
}

func noOpResult(req *core.HookRunRequest) *core.HookResult {
	return &core.HookResult{
		HookType:     req.HookType,
		HasExecuted:  false,
		IsSuccessful: true,
		Failures:     nil,
	}
}

func (r *HookRunner) readHookConfig(path string) (*core.HookConfig, error) {
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

// executeDefinitions runs each definition in order and aggregates the
// per-definition results into a single HookResult. Per-definition
// failure keeps going — one bad command does not abort the chain so
// every hook has a chance to run, matching the previous behaviour for
// multi-command single-definition configs.
//
// When hardFailure is true (PreRebase lifecycle stage) the first
// non-zero exit stops iteration so the cmd layer can short-circuit
// the gated operation.
func (r *HookRunner) executeDefinitions(ctx context.Context, req *core.HookRunRequest, definitions []core.HookDefinition, hardFailure bool) *core.HookResult {
	result := &core.HookResult{
		HookType:     req.HookType,
		HasExecuted:  true,
		IsSuccessful: true,
		Failures:     nil,
	}

	envExports := r.buildEnvExports(req)

	for _, def := range definitions {
		command := strings.TrimSpace(def.Command)
		if command == "" {
			continue
		}

		timeout := r.defaultTimeout
		if def.TimeoutSeconds > 0 {
			timeout = time.Duration(def.TimeoutSeconds) * time.Second
		}
		workDir := req.WorktreePath
		if def.WorkingDirectory != "" {
			workDir = def.WorkingDirectory
		}

		fullCmd := envExports + command
		cmdResult, err := r.executor.ExecuteWithTimeout(ctx, workDir, CmdSh, timeout, "-c", fullCmd)

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
			timedOut := errors.Is(err, context.DeadlineExceeded)
			result.Failures = append(result.Failures, core.HookFailure{
				Command:  command,
				ExitCode: exitCode,
				Output:   output,
				Error:    err,
				TimedOut: timedOut,
			})
			if hardFailure {
				break
			}
		}
	}

	return result
}

func (r *HookRunner) buildEnvExports(req *core.HookRunRequest) string {
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
	if req.HookType == core.HookTypePreRebase || req.HookType == core.HookTypePostRebase {
		if req.RebaseBase != "" {
			exports.WriteString(fmt.Sprintf("export TWIGGIT_REBASE_BASE=%q; ", req.RebaseBase))
		}
		if req.RebaseOldTip != "" {
			exports.WriteString(fmt.Sprintf("export TWIGGIT_REBASE_OLD_TIP=%q; ", req.RebaseOldTip))
		}
		if req.RebaseNewTip != "" {
			exports.WriteString(fmt.Sprintf("export TWIGGIT_REBASE_NEW_TIP=%q; ", req.RebaseNewTip))
		}
		if req.RebaseResult != "" {
			exports.WriteString(fmt.Sprintf("export TWIGGIT_REBASE_RESULT=%q; ", req.RebaseResult))
		}
	}
	if req.HookType == core.HookTypePostSync {
		if req.SyncRemote != "" {
			exports.WriteString(fmt.Sprintf("export TWIGGIT_SYNC_REMOTE=%q; ", req.SyncRemote))
		}
		if req.SyncBranch != "" {
			exports.WriteString(fmt.Sprintf("export TWIGGIT_SYNC_BRANCH=%q; ", req.SyncBranch))
		}
	}
	return exports.String()
}
