//go:build e2e
// +build e2e

package e2e

import (
	"os"
	"path/filepath"

	. "github.com/onsi/ginkgo/v2"
	. "github.com/onsi/gomega"

	"twiggit/test/e2e/fixtures"
	"twiggit/test/e2e/helpers"
	testhelpers "twiggit/test/helpers"
)

var _ = Describe("Infrastructure Verification", func() {
	var fixture *fixtures.E2ETestFixture

	BeforeEach(func() {
		fixture = fixtures.NewE2ETestFixture()
	})

	AfterEach(func() {
		fixture.Cleanup()
	})

	It("validates generated config", func() {
		configHelper := helpers.NewConfigHelper()
		configDir := configHelper.Build()

		Expect(configDir).NotTo(BeEmpty())
		Expect(configHelper.GetConfigPath()).To(BeARegularFile())
	})

	It("creates and removes worktree using git CLI", func() {
		gitHelper := testhelpers.NewWorktreeTestHelper()
		fixture.SetupSingleProject("test")
		projectPath := fixture.GetProjectPath("test")

		worktreePath := filepath.Join(fixture.GetTempDir(), "wt-test")
		err := gitHelper.CreateWorktree(projectPath, worktreePath, "feature")
		Expect(err).NotTo(HaveOccurred())
		Expect(worktreePath).To(BeADirectory())

		err = gitHelper.RemoveWorktree(worktreePath, false)
		Expect(err).NotTo(HaveOccurred())
		Expect(worktreePath).NotTo(BeADirectory())
	})

	It("creates worktrees via fixture and cleans them up", func() {
		fixture.CreateWorktreeSetup("test")

		createdWorktrees := fixture.GetCreatedWorktrees()
		Expect(createdWorktrees).To(HaveLen(2))

		for _, wt := range createdWorktrees {
			Expect(wt).To(BeADirectory())
		}

		fixture.Cleanup()

		for _, wt := range createdWorktrees {
			_, err := os.Stat(wt)
			Expect(os.IsNotExist(err)).To(BeTrue(), "Worktree %s should be removed", wt)
		}
	})

	It("runs CLI commands from different contexts", func() {
		configDir := fixture.Build()
		cli := helpers.NewTwiggitCLI().WithConfigDir(configDir)
		ctxHelper := fixtures.NewContextHelper(fixture, cli)

		session := ctxHelper.FromOutsideGit("help")
		cli.ShouldSucceed(session)

		fixture.SetupSingleProject("test")
		session = ctxHelper.FromProjectDir("test", "help")
		cli.ShouldSucceed(session)
	})

	It("cleanup is idempotent and safe", func() {
		fixture.CreateWorktreeSetup("test-idempotent")

		fixture.Cleanup()
		fixture.Cleanup()
		fixture.Cleanup()
	})

	It("handles empty worktrees list gracefully", func() {
		fixture.SetupSingleProject("test")

		fixture.Cleanup()
	})

	It("validates cleanup was successful", func() {
		fixture.CreateWorktreeSetup("test-validate-cleanup")

		fixture.Cleanup()

		err := fixture.ValidateCleanup()
		Expect(err).NotTo(HaveOccurred())
	})

	It("provides debugging information via Inspect()", func() {
		fixture.CreateWorktreeSetup("test-inspect")

		state := fixture.Inspect()
		Expect(state).To(ContainSubstring("E2ETestFixture State"))
		Expect(state).To(ContainSubstring("Projects"))
		Expect(state).To(ContainSubstring("Worktrees"))
		Expect(state).To(ContainSubstring("test-inspect"))

		fixture.Cleanup()
	})
})
