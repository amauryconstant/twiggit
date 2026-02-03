//go:build e2e
// +build e2e

package e2e

import (
	. "github.com/onsi/ginkgo/v2"

	"twiggit/test/e2e/helpers"
)

var _ = Describe("help command", func() {
	It("shows main help", func() {
		configHelper := helpers.NewConfigHelper()
		cli := helpers.NewTwiggitCLI().WithConfigDir(configHelper.Build())

		session := cli.Run("help")
		cli.ShouldSucceed(session)
		cli.ShouldOutput(session, "twiggit")
	})
})
