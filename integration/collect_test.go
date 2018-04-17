package integration

import (
	. "github.com/onsi/ginkgo"
	. "github.com/onsi/ginkgo/extensions/table"
	. "github.com/onsi/gomega"
	"github.com/onsi/gomega/gbytes"
	"github.com/onsi/gomega/gexec"

	"encoding/json"
	"fmt"
	"io/ioutil"
	"os"
	"os/exec"
	"path/filepath"
	"time"

	"github.com/pivotal-cf/aqueduct-courier/cmd"
)

const (
	RFC3339DateTimeUTCPermissiveRegexp = `\d{4}-\d{2}-\d{2}[Tt]\d{2}:\d{2}:\d{2}[Zz]`
)

var _ = Describe("Collect", func() {
	var (
		collector      string
		outputDirPath  string
		defaultEnvVars = map[string]string{
			cmd.OpsManagerURLKey:      os.Getenv("TEST_OPSMANAGER_URL"),
			cmd.OpsManagerUsernameKey: os.Getenv("TEST_OPSMANAGER_USERNAME"),
			cmd.OpsManagerPasswordKey: os.Getenv("TEST_OPSMANAGER_PASSWORD"),
			cmd.SkipTlsVerifyKey:      "true",
		}
		additionalEnvVars map[string]string
	)

	BeforeSuite(func() {
		var err error
		collector, err = gexec.Build("github.com/pivotal-cf/aqueduct-courier")
		Expect(err).NotTo(HaveOccurred())
	})

	BeforeEach(func() {
		var err error
		outputDirPath, err = ioutil.TempDir("", "")
		Expect(err).NotTo(HaveOccurred())
		additionalEnvVars = map[string]string{
			cmd.OutputPathKey: outputDirPath,
		}
	})

	It("succeeds", func() {
		command := exec.Command(collector, "collect")
		for k, v := range merge(defaultEnvVars, additionalEnvVars) {
			command.Env = append(command.Env, fmt.Sprintf("%s=%s", k, v))
		}
		session, err := gexec.Start(command, GinkgoWriter, GinkgoWriter)
		Expect(err).NotTo(HaveOccurred())
		Eventually(session, 30*time.Second).Should(gexec.Exit(0))

		contentDir := validatedContentDir(outputDirPath)
		assertContainsJsonFile(contentDir, "ops_manager_vm_types.json")
	})

	It("fails if data collection from Ops Manager fails", func() {
		additionalEnvVars[cmd.OpsManagerUsernameKey] = "non-real-user"
		command := exec.Command(collector, "collect")
		for k, v := range merge(defaultEnvVars, additionalEnvVars) {
			command.Env = append(command.Env, fmt.Sprintf("%s=%s", k, v))
		}
		session, err := gexec.Start(command, GinkgoWriter, GinkgoWriter)
		Expect(err).NotTo(HaveOccurred())
		Eventually(session, 30*time.Second).Should(gexec.Exit(1))
		Expect(session.Err).To(gbytes.Say("Failed collecting from Ops Manager:"))
	})

	It("fails if creating the output directory fails", func() {
		badDir := "not/a/real/path/on/disk"
		additionalEnvVars[cmd.OutputPathKey] = badDir
		command := exec.Command(collector, "collect")
		for k, v := range merge(defaultEnvVars, additionalEnvVars) {
			command.Env = append(command.Env, fmt.Sprintf("%s=%s", k, v))
		}
		session, err := gexec.Start(command, GinkgoWriter, GinkgoWriter)
		Expect(err).NotTo(HaveOccurred())
		Eventually(session, 30*time.Second).Should(gexec.Exit(1))
		Expect(session.Err).To(gbytes.Say(
			fmt.Sprintf(`Failed creating directory %s/%s%s:`, badDir, cmd.OutputDirPrefix, RFC3339DateTimeUTCPermissiveRegexp),
		))
		Consistently(session.Err).ShouldNot(
			gbytes.Say("Failed writing data to disk"),
		)
	})

	DescribeTable(
		"fails when required environment variable is not set",
		func(missingKey string) {
			command := exec.Command(collector, "collect")
			for k, v := range merge(defaultEnvVars, additionalEnvVars) {
				if k != missingKey {
					command.Env = append(command.Env, fmt.Sprintf("%s=%s", k, v))
				}
			}
			session, err := gexec.Start(command, GinkgoWriter, GinkgoWriter)
			Expect(err).ToNot(HaveOccurred())
			Eventually(session).Should(gexec.Exit(1))
			Eventually(session.Err).Should(gbytes.Say(fmt.Sprintf("Requires %s to be set", missingKey)))
		},
		Entry(cmd.OpsManagerURLKey, cmd.OpsManagerURLKey),
		Entry(cmd.OpsManagerUsernameKey, cmd.OpsManagerUsernameKey),
		Entry(cmd.OpsManagerPasswordKey, cmd.OpsManagerPasswordKey),
		Entry(cmd.OutputPathKey, cmd.OutputPathKey),
	)
})

func validatedContentDir(outputDirPath string) string {
	fileInfos, err := ioutil.ReadDir(outputDirPath)
	Expect(err).NotTo(HaveOccurred())
	Expect(len(fileInfos)).To(Equal(1), fmt.Sprintf("Expected output dir %s to include a single directory", outputDirPath))
	Expect(fileInfos[0].IsDir()).To(BeTrue(), fmt.Sprintf("Expected file %s found in %s to be a directory", fileInfos[0], outputDirPath))
	Expect(fileInfos[0].Name()).To(MatchRegexp(fmt.Sprintf(`%s%s$`, cmd.OutputDirPrefix, RFC3339DateTimeUTCPermissiveRegexp)))
	return filepath.Join(outputDirPath, fileInfos[0].Name())
}

func assertContainsJsonFile(contentDir, filename string) {
	contentFileInfos, err := ioutil.ReadDir(contentDir)
	Expect(err).NotTo(HaveOccurred())
	Expect(len(contentFileInfos)).To(BeNumerically(">", 0), fmt.Sprintf("Expected %s to contain at least 1 file", contentDir))
	expectedFileExists := false
	for _, i := range contentFileInfos {
		if i.Name() == filename {
			expectedFileExists = true
			content, err := ioutil.ReadFile(filepath.Join(contentDir, filename))
			Expect(err).NotTo(HaveOccurred())
			Expect(json.Valid(content)).To(BeTrue(), fmt.Sprintf("Expected file %s to contain valid json", filename))
		}
	}
	Expect(expectedFileExists).To(BeTrue(), fmt.Sprintf("Expected to find file with name %s, but did not", filename))
}

func merge(maps ...map[string]string) map[string]string {
	response := map[string]string{}
	for _, m := range maps {
		for k, v := range m {
			response[k] = v
		}
	}

	return response
}
