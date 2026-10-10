//go:build e2e

// Package e2e covers the `twiggit status` command against a built
// binary. Smoke tests for each output shape, the `wtg status | jq`
// stream-separation scenario, and the outside-git -> UsageError
// exit-2 path live here. Golden fixtures back the four output
// shapes (default, json, table, plain) so the e2e surface stays
// stable across drift.
package e2e

import (
	"encoding/json"
	"twiggit/test/e2e/fixtures"
	"twiggit/test/e2e/helpers"

	. "github.com/onsi/ginkgo/v2"
	. "github.com/onsi/gomega"
)

// statusRowJSON is the wire shape pinned by cli-status for
// `status -o json` payload: a bare JSON array of
// {branch,path,base,ahead,behind,merged,dirty,last_commit_date,stale,skipped,skip_reason}
// objects. The `project` key is set under --all. Used by the
// structural decode assertions below.
type statusRowJSON struct {
	Project        string `json:"project,omitempty"`
	Branch         string `json:"branch"`
	Path           string `json:"path"`
	Base           string `json:"base"`
	Ahead          int    `json:"ahead"`
	Behind         int    `json:"behind"`
	Merged         bool   `json:"merged"`
	Dirty          bool   `json:"dirty"`
	LastCommitDate string `json:"last_commit_date"`
	Stale          bool   `json:"stale"`
	Skipped        bool   `json:"skipped"`
	SkipReason     string `json:"skip_reason,omitempty"`
}

var _ = Describe("status command", func() {
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

	It("renders one row per worktree in default output", func() {
		fixture.CreateWorktreeSetup("test-project")

		session := ctxHelper.FromProjectDir("test-project", "status")
		cli.ShouldSucceed(session)

		if session.ExitCode() != 0 {
			GinkgoT().Log(fixture.Inspect())
		}
	})

	It("renders one row per worktree with --all", func() {
		fixture.SetupMultiProject()

		session := cli.Run("status", "--all")
		cli.ShouldSucceed(session)

		if session.ExitCode() != 0 {
			GinkgoT().Log(fixture.Inspect())
		}
	})

	It("emits the project field under --all json", func() {
		fixture.CreateWorktreeSetup("status-proj-1")
		fixture.CreateWorktreeSetup("status-proj-2")

		session := cli.Run("status", "--all", "--output", "json")
		cli.ShouldSucceed(session)

		var got []statusRowJSON
		stdout := cli.GetOutput(session)
		Expect(json.Unmarshal([]byte(stdout), &got)).To(Succeed(),
			"status -o json must decode as []statusRowJSON; got %q", stdout)
		Expect(got).ToNot(BeEmpty(),
			"multi-project fixture must produce at least one worktree")
		seen := map[string]bool{}
		for _, row := range got {
			seen[row.Project] = true
		}
		Expect(len(seen)).To(BeNumerically(">=", 2),
			"--all JSON rows must carry project names from at least two projects; got %v", seen)
	})

	It("omits the project field without --all", func() {
		fixture.CreateWorktreeSetup("test-project")

		session := ctxHelper.FromProjectDir("test-project", "status", "--output", "json")
		cli.ShouldSucceed(session)

		var got []statusRowJSON
		stdout := cli.GetOutput(session)
		Expect(json.Unmarshal([]byte(stdout), &got)).To(Succeed(),
			"status -o json must decode as []statusRowJSON; got %q", stdout)
		Expect(got).ToNot(BeEmpty())
		for i, row := range got {
			Expect(row.Project).To(BeEmpty(),
				"single-project JSON must omit the project field; row %d had %q", i, row.Project)
		}
	})

	It("emits a bare JSON array with --output json", func() {
		result := fixture.CreateWorktreeSetup("test")

		session := ctxHelper.FromProjectDir("test", "status", "--output", "json")
		cli.ShouldSucceed(session)

		out := cli.GetOutput(session)
		Expect(out).To(HavePrefix("["), "--output json must emit a bare array; got %q", out)
		Expect(out).To(ContainSubstring(`"branch":"`+result.Feature1Branch+`"`),
			"--output json must surface the fixture branch name; got %q", out)
	})

	It("decodes --output json as a bare array of the documented shape", func() {
		result := fixture.CreateWorktreeSetup("test")

		session := ctxHelper.FromProjectDir("test", "status", "--output", "json")
		cli.ShouldSucceed(session)

		var got []statusRowJSON
		Expect(json.Unmarshal([]byte(cli.GetOutput(session)), &got)).To(Succeed(),
			"status -o json must decode as []statusRowJSON; got %q", cli.GetOutput(session))

		Expect(got).ToNot(BeEmpty(),
			"fixture must produce at least one worktree for this assertion to mean anything")
		Expect(got[0].Branch).To(Equal(result.Feature1Branch),
			"first element must be addressable as '.[0].branch' and yield the fixture branch")
		Expect(got[0].Path).ToNot(BeEmpty(), "first element must carry a non-empty path")
	})

	It("accepts --output table", func() {
		fixture.CreateWorktreeSetup("test-project")

		session := ctxHelper.FromProjectDir("test-project", "status", "--output", "table")
		cli.ShouldSucceed(session)

		if session.ExitCode() != 0 {
			GinkgoT().Log(fixture.Inspect())
		}
	})

	It("accepts --output plain", func() {
		fixture.CreateWorktreeSetup("test-project")

		session := ctxHelper.FromProjectDir("test-project", "status", "--output", "plain")
		cli.ShouldSucceed(session)

		if session.ExitCode() != 0 {
			GinkgoT().Log(fixture.Inspect())
		}
	})

	It("pipes through jq without interleaved stderr bytes", func() {
		fixture.CreateWorktreeSetup("test")

		session := ctxHelper.FromProjectDir("test", "status", "--output", "json")
		cli.ShouldSucceed(session)

		// The captured stdout is the bare JSON array; the
		// captured stderr must be empty for the documented
		// stream-separation contract (cli-output-formats).
		stdout := string(session.Out.Contents())
		Expect(stdout).To(HavePrefix("["))

		var got []statusRowJSON
		Expect(json.Unmarshal([]byte(stdout), &got)).To(Succeed(),
			"jq pipeline must receive a valid bare array; got %q", stdout)
	})

	It("emits exit 2 with UsageError when run outside git without --all", func() {
		session := ctxHelper.FromOutsideGit("status")
		cli.ShouldFailWithExit(session, 2)
		errOut := string(session.Err.Contents())
		Expect(errOut).To(ContainSubstring("--all"))
		Expect(errOut).To(ContainSubstring("project"))
	})
})
