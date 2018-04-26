package integration

import (
	. "github.com/onsi/ginkgo"
	. "github.com/onsi/gomega"

	"fmt"
	"os/exec"
	"time"

	"github.com/onsi/gomega/gbytes"
	"github.com/onsi/gomega/gexec"
	"github.com/pivotal-cf/aqueduct-courier/cmd"
)

var _ = Describe("Send", func() {

	It("allows you to call the send command", func() {
		command := exec.Command(aqueductBinaryPath, "send", "--path=/path/to/data", "--api-key=best-key")
		session, err := gexec.Start(command, GinkgoWriter, GinkgoWriter)
		Expect(err).NotTo(HaveOccurred())
		Eventually(session, 30*time.Second).Should(gexec.Exit(0))
	})

	It("fails if the path flag has not been set", func() {
		command := exec.Command(aqueductBinaryPath, "send", "--api-key=best-key")
		session, err := gexec.Start(command, GinkgoWriter, GinkgoWriter)
		Expect(err).NotTo(HaveOccurred())
		Eventually(session, 30*time.Second).Should(gexec.Exit(1))
		Expect(session.Err).To(gbytes.Say(fmt.Sprintf(cmd.RequiredConfigErrorFormat, cmd.DirectoryPathFlag)))
	})

	It("fails if the api-key flag has not been set", func() {
		command := exec.Command(aqueductBinaryPath, "send", "--path=/path/to/data")
		session, err := gexec.Start(command, GinkgoWriter, GinkgoWriter)
		Expect(err).NotTo(HaveOccurred())
		Eventually(session, 30*time.Second).Should(gexec.Exit(1))
		Expect(session.Err).To(gbytes.Say(fmt.Sprintf(cmd.RequiredConfigErrorFormat, cmd.ApiKeyFlag)))
	})
})
