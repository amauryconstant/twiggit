//go:build e2e
// +build e2e

// Package e2e provides end-to-end tests for twiggit create command.
// Tests validate worktree creation across project, worktree, and outside git contexts.
package e2e

import (
	"os"
	"path/filepath"

	. "github.com/onsi/ginkgo/v2"
	. "github.com/onsi/gomega"

	"twiggit/test/e2e/fixtures"
	"twiggit/test/e2e/helpers"
)

var _ = Describe("create command", func() {
	var fixture *fixtures.E2ETestFixture
	var cli *helpers.TwiggitCLI
	var ctxHelper *fixtures.ContextHelper

	BeforeEach(func() {
		fixture = fixtures.NewE2ETestFixture()
		cli = helpers.NewTwiggitCLI()
		cli = cli.WithConfigDir(fixture.Build())
		ctxHelper = fixtures.NewContextHelper(fixture, cli)
	})

	AfterEach(func() {
		fixture.Cleanup()
	})

	It("creates worktree from project context with default source", func() {
		fixture.SetupSingleProject("test-project")

		testID := fixture.GetTestID()
		branchName := testID.BranchName("feature-1")

		session := ctxHelper.FromProjectDir("test-project", "create", branchName)
		cli.ShouldSucceed(session)
		cli.ShouldOutput(session, branchName)

		if session.ExitCode() != 0 {
			GinkgoT().Log(fixture.Inspect())
		}

		worktreePath := filepath.Join(fixture.GetConfigHelper().GetWorktreesDir(), "test-project", branchName)
		Expect(worktreePath).To(BeADirectory())
	})

	It("creates worktree with --source flag", func() {
		fixture.SetupSingleProject("test-project")
		projectPath := fixture.GetProjectPath("test-project")
		testID := fixture.GetTestID()
		gitHelper := fixture.GetGitHelper()

		err := gitHelper.CreateBranch(projectPath, "develop")
		Expect(err).NotTo(HaveOccurred())

		branchName := testID.BranchName("feature-new")
		session := cli.Run("create", "test-project/"+branchName, "--source", "develop")
		cli.ShouldSucceed(session)
		cli.ShouldOutput(session, branchName)

		if session.ExitCode() != 0 {
			GinkgoT().Log(fixture.Inspect())
		}

		worktreePath := filepath.Join(fixture.GetConfigHelper().GetWorktreesDir(), "test-project", branchName)
		Expect(worktreePath).To(BeADirectory())
	})

	It("creates worktree from outside git with project/branch", func() {
		fixture.SetupSingleProject("external-project")

		session := ctxHelper.FromOutsideGit("create", "external-project/new-feature")
		cli.ShouldSucceed(session)
		cli.ShouldOutput(session, "new-feature")

		if session.ExitCode() != 0 {
			GinkgoT().Log(fixture.Inspect())
		}

		worktreePath := filepath.Join(fixture.GetConfigHelper().GetWorktreesDir(), "external-project", "new-feature")
		Expect(worktreePath).To(BeADirectory())
	})

	It("shows level 1 verbose output with -v flag", func() {
		fixture.SetupSingleProject("test-project")

		testID := fixture.GetTestID()
		branchName := testID.BranchName("verbose-test")

		session := ctxHelper.FromProjectDir("test-project", "create", branchName, "-v")
		cli.ShouldSucceed(session)
		cli.ShouldOutput(session, branchName)
		cli.ShouldVerboseOutput(session, "Creating worktree for test-project/"+branchName)

		if session.ExitCode() != 0 {
			GinkgoT().Log(fixture.Inspect())
		}

		worktreePath := filepath.Join(fixture.GetConfigHelper().GetWorktreesDir(), "test-project", branchName)
		Expect(worktreePath).To(BeADirectory())
	})

	It("shows level 2 verbose output with -vv flag", func() {
		fixture.SetupSingleProject("test-project")

		testID := fixture.GetTestID()
		branchName := testID.BranchName("verbose-vv-test")

		session := ctxHelper.FromProjectDir("test-project", "create", branchName, "-vv")
		cli.ShouldSucceed(session)
		cli.ShouldOutput(session, branchName)
		cli.ShouldVerboseOutput(session, "Creating worktree for test-project/"+branchName)
		cli.ShouldVerboseOutput(session, "  from branch: main")
		cli.ShouldVerboseOutput(session, "  to path: test-project/"+branchName)

		if session.ExitCode() != 0 {
			GinkgoT().Log(fixture.Inspect())
		}

		worktreePath := filepath.Join(fixture.GetConfigHelper().GetWorktreesDir(), "test-project", branchName)
		Expect(worktreePath).To(BeADirectory())
	})

	It("shows no verbose output by default", func() {
		fixture.SetupSingleProject("test-project")

		testID := fixture.GetTestID()
		branchName := testID.BranchName("no-verbose-test")

		session := ctxHelper.FromProjectDir("test-project", "create", branchName)
		cli.ShouldSucceed(session)
		cli.ShouldNotHaveVerboseOutput(session)

		if session.ExitCode() != 0 {
			GinkgoT().Log(fixture.Inspect())
		}
	})

	It("with -C flag outputs path to stdout only", func() {
		fixture.SetupSingleProject("test-project")

		testID := fixture.GetTestID()
		branchName := testID.BranchName("feature-path")

		session := ctxHelper.FromProjectDir("test-project", "create", "-C", branchName)
		cli.ShouldSucceed(session)

		if session.ExitCode() != 0 {
			GinkgoT().Log(fixture.Inspect())
		}

		output := string(session.Out.Contents())
		worktreePath := filepath.Join(fixture.GetConfigHelper().GetWorktreesDir(), "test-project", branchName)

		Expect(output).To(Equal(worktreePath+"\n"), "Should output path only")
		Expect(output).NotTo(ContainSubstring("Created worktree"), "Should not include success message")
		Expect(worktreePath).To(BeADirectory())
	})

	It("without -C flag outputs success message", func() {
		fixture.SetupSingleProject("test-project")

		testID := fixture.GetTestID()
		branchName := testID.BranchName("feature-message")

		session := ctxHelper.FromProjectDir("test-project", "create", branchName)
		cli.ShouldSucceed(session)

		if session.ExitCode() != 0 {
			GinkgoT().Log(fixture.Inspect())
		}

		output := string(session.Out.Contents())

		Expect(output).To(ContainSubstring("Created worktree"), "Should include success message")
		Expect(output).To(ContainSubstring(branchName), "Should include branch name")
	})

	It("executes post-create hooks when .twiggit.toml exists", func() {
		fixture.SetupSingleProject("test-project")
		projectPath := fixture.GetProjectPath("test-project")

		hookConfigContent := `
[hooks.post-create]
commands = ["echo hook-ran-successfully"]
`
		hookConfigPath := filepath.Join(projectPath, ".twiggit.toml")
		err := os.WriteFile(hookConfigPath, []byte(hookConfigContent), 0644)
		Expect(err).NotTo(HaveOccurred())

		testID := fixture.GetTestID()
		branchName := testID.BranchName("hook-test")

		session := ctxHelper.FromProjectDir("test-project", "create", branchName)
		cli.ShouldSucceed(session)

		if session.ExitCode() != 0 {
			GinkgoT().Log(fixture.Inspect())
		}

		output := string(session.Out.Contents())
		Expect(output).To(ContainSubstring("Created worktree"), "Should show success message")
	})

	It("displays warning when post-create hook fails", func() {
		fixture.SetupSingleProject("test-project")
		projectPath := fixture.GetProjectPath("test-project")

		hookConfigContent := `
[hooks.post-create]
commands = ["exit 1"]
`
		hookConfigPath := filepath.Join(projectPath, ".twiggit.toml")
		err := os.WriteFile(hookConfigPath, []byte(hookConfigContent), 0644)
		Expect(err).NotTo(HaveOccurred())

		testID := fixture.GetTestID()
		branchName := testID.BranchName("hook-fail-test")

		session := ctxHelper.FromProjectDir("test-project", "create", branchName)
		cli.ShouldSucceed(session)

		if session.ExitCode() != 0 {
			GinkgoT().Log(fixture.Inspect())
		}

		stdout := string(session.Out.Contents())
		stderr := string(session.Err.Contents())
		Expect(stdout).To(ContainSubstring("Created worktree"), "Worktree should still be created")
		Expect(stderr).To(ContainSubstring("Warning"), "Should show warning for hook failure on stderr")
		Expect(stderr).To(ContainSubstring("exit 1"), "Should show failed command")
	})

	It("continues normally when no .twiggit.toml exists", func() {
		fixture.SetupSingleProject("test-project")

		testID := fixture.GetTestID()
		branchName := testID.BranchName("no-hook-test")

		session := ctxHelper.FromProjectDir("test-project", "create", branchName)
		cli.ShouldSucceed(session)

		if session.ExitCode() != 0 {
			GinkgoT().Log(fixture.Inspect())
		}

		stdout := string(session.Out.Contents())
		stderr := string(session.Err.Contents())
		Expect(stdout).To(ContainSubstring("Created worktree"), "Should show success message")
		Expect(stderr).NotTo(ContainSubstring("Warning"), "Should not show warning when no hooks")
	})
})
