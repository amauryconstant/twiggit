package cmd

import (
	"fmt"
	"path/filepath"
	"twiggit/internal/cmdutil"
	"twiggit/internal/core"
	"twiggit/internal/git"
)

// detectContext runs the 5-step context-detection setup that was
// duplicated across runCreate / runDelete / runList / runCd /
// runPrune: load config, build the git client, build the context
// detector, resolve the working directory, and detect the current
// git context. Returns the detected Context and the composite
// *git.Client so callers can narrow via type assertion
// (e.g., `var br core.BranchReader = client`).
//
// The two function-typed parameters are the same shape as the
// cmdutil.Factory fields Config and GitClient; each Options type
// already stores both (set during RunE from the Factory) so callers
// pass opts.Config / opts.GitClient directly.
//
// Callers that need *core.Config after this call invoke
// loadConfig() again — the Factory wires Config via sync.OnceValues
// so the second call returns the cached value.
func detectContext(
	loadConfig func() (*core.Config, error),
	loadGitClient func() (cmdutil.Client, error),
) (*core.Context, *git.Client, error) {
	cfg, err := loadConfig()
	if err != nil {
		return nil, nil, fmt.Errorf("config load failed: %w", err)
	}

	client, err := loadGitClient()
	if err != nil {
		return nil, nil, fmt.Errorf("git client init failed: %w", err)
	}
	gitClient, ok := client.(*git.Client)
	if !ok {
		return nil, nil, fmt.Errorf("cmdutil: GitClient returned %T, expected *git.Client", client)
	}

	detector, err := git.NewContextDetector(cfg)
	if err != nil {
		return nil, nil, fmt.Errorf("context detector init failed: %w", err)
	}

	wd, err := filepath.Abs(".")
	if err != nil {
		return nil, nil, fmt.Errorf("get working directory: %w", err)
	}

	ctx, err := detector.DetectContext(wd)
	if err != nil {
		return nil, nil, fmt.Errorf("context detection failed: %w", err)
	}

	return ctx, gitClient, nil
}
