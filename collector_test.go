package main_test

import (
	. "github.com/onsi/ginkgo"
	. "github.com/onsi/ginkgo/extensions/table"
	. "github.com/onsi/gomega"
	"github.com/onsi/gomega/gexec"

	"os/exec"

	"fmt"

	"os"

	"time"

	"github.com/onsi/gomega/gbytes"
	. "github.com/pivotal-cf/aqueduct-courier"
)

var _ = Describe("Collector", func() {
	var (
		collector      string
		defaultEnvVars = map[string]string{
			OpsManagerURLKey:      os.Getenv("TEST_OPSMANAGER_URL"),
			OpsManagerUsernameKey: os.Getenv("TEST_OPSMANAGER_USERNAME"),
			OpsManagerPasswordKey: os.Getenv("TEST_OPSMANAGER_PASSWORD"),
			OutputPathKey:         "path/on/disk",
			SkipTlsVerifyKey:      "true",
		}
	)

	BeforeSuite(func() {
		var err error
		collector, err = gexec.Build("github.com/pivotal-cf/aqueduct-courier")
		Expect(err).NotTo(HaveOccurred())
	})

	It("succeeds", func() {
		cmd := exec.Command(collector)
		for k, v := range defaultEnvVars {
			cmd.Env = append(cmd.Env, fmt.Sprintf("%s=%s", k, v))
		}
		session, err := gexec.Start(cmd, GinkgoWriter, GinkgoWriter)
		Expect(err).ToNot(HaveOccurred())
		Eventually(session, 10*time.Second).Should(gexec.Exit(0))
	})

	DescribeTable(
		"fails when required environment variable is not set",
		func(missingKey string) {
			cmd := exec.Command(collector)
			for k, v := range defaultEnvVars {
				if k != missingKey {
					cmd.Env = append(cmd.Env, fmt.Sprintf("%s=%s", k, v))
				}
			}
			session, err := gexec.Start(cmd, GinkgoWriter, GinkgoWriter)
			Expect(err).ToNot(HaveOccurred())
			Eventually(session).Should(gexec.Exit(1))
			Eventually(session.Err).Should(gbytes.Say(fmt.Sprintf("Requires %s to be set", missingKey)))
		},
		Entry(OpsManagerURLKey, OpsManagerURLKey),
		Entry(OpsManagerUsernameKey, OpsManagerUsernameKey),
		Entry(OpsManagerPasswordKey, OpsManagerPasswordKey),
		Entry(OutputPathKey, OutputPathKey),
	)
})
