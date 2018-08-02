package integration

import (
	"fmt"
	"os/exec"

	. "github.com/onsi/ginkgo"
	. "github.com/onsi/gomega"
	"github.com/onsi/gomega/gbytes"
	"github.com/onsi/gomega/gexec"
)

var _ = Describe("Flags", func() {
	Describe("--version", func() {
		It("prints the version when compiled with a version value", func() {
			expectedVersion := "0.100.99"
			binaryWithVersion, err := gexec.Build(
				"github.com/pivotal-cf/aqueduct-courier",
				"-ldflags",
				fmt.Sprintf("-X github.com/pivotal-cf/aqueduct-courier/cmd.version=%s", expectedVersion),
			)
			Expect(err).NotTo(HaveOccurred())

			command := exec.Command(binaryWithVersion, "--version")
			session, err := gexec.Start(command, GinkgoWriter, GinkgoWriter)
			Expect(err).NotTo(HaveOccurred())
			Eventually(session).Should(gexec.Exit(0))
			Expect(session.Out).To(gbytes.Say("platform-insights-collector version " + expectedVersion))
		})

		It("prints 'dev' as the version when not compiled with a version value", func() {
			command := exec.Command(aqueductBinaryPath, "--version")
			session, err := gexec.Start(command, GinkgoWriter, GinkgoWriter)
			Expect(err).NotTo(HaveOccurred())
			Eventually(session).Should(gexec.Exit(0))
			Expect(session.Out).To(gbytes.Say("platform-insights-collector version dev"))
		})
	})
})
