package integration

import (
	"encoding/json"
	"fmt"
	"io/ioutil"
	"os"
	"os/exec"
	"path/filepath"
	"strings"

	"github.com/mholt/archiver"
	. "github.com/onsi/ginkgo"
	. "github.com/onsi/ginkgo/extensions/table"
	. "github.com/onsi/gomega"
	"github.com/onsi/gomega/gbytes"
	"github.com/onsi/gomega/gexec"
	"github.com/pivotal-cf/aqueduct-courier/cmd"
	"github.com/pivotal-cf/aqueduct-courier/file"
	"github.com/pivotal-cf/aqueduct-courier/ops"
	"github.com/pivotal-cf/aqueduct-utils/data"
)

const (
	UnixTimestampRegexp = `\d{10}`
)

var _ = Describe("Collect", func() {
	var (
		outputDirPath  string
		defaultEnvVars map[string]string
	)

	BeforeEach(func() {
		var err error
		outputDirPath, err = ioutil.TempDir("", "")
		Expect(err).NotTo(HaveOccurred())
		defaultEnvVars = map[string]string{
			cmd.OpsManagerURLKey:      os.Getenv("TEST_OPS_MANAGER_URL"),
			cmd.OpsManagerUsernameKey: os.Getenv("TEST_OPS_MANAGER_USERNAME"),
			cmd.OpsManagerPasswordKey: os.Getenv("TEST_OPS_MANAGER_PASSWORD"),
			cmd.EnvTypeKey:            "Development",
			cmd.OutputPathKey:         outputDirPath,
		}
	})

	AfterEach(func() {
		Expect(os.RemoveAll(outputDirPath)).To(Succeed())
	})

	Context("user/password authentication", func() {
		It("succeeds with env variable configuration", func() {
			command := buildDefaultCommand(defaultEnvVars)
			session, err := gexec.Start(command, GinkgoWriter, GinkgoWriter)
			Expect(err).NotTo(HaveOccurred())
			Eventually(session).Should(gexec.Exit(0))

			tarFilePath := validatedTarFilePath(outputDirPath)
			assertValidOutput(tarFilePath, "ops_manager_vm_types", "development")
			assertLogging(session, tarFilePath, defaultEnvVars[cmd.OpsManagerURLKey])
		})

		It("succeeds with flag configuration", func() {
			flagValues := map[string]string{
				cmd.OpsManagerURLFlag:      os.Getenv("TEST_OPS_MANAGER_URL"),
				cmd.OpsManagerUsernameFlag: os.Getenv("TEST_OPS_MANAGER_USERNAME"),
				cmd.OpsManagerPasswordFlag: os.Getenv("TEST_OPS_MANAGER_PASSWORD"),
				cmd.EnvTypeFlag:            "Development",
				cmd.SkipTlsVerifyFlag:      "true",
				cmd.OutputPathFlag:         outputDirPath,
			}
			command := exec.Command(aqueductBinaryPath, "collect")
			for k, v := range flagValues {
				command.Args = append(command.Args, fmt.Sprintf("--%s=%s", k, v))
			}
			session, err := gexec.Start(command, GinkgoWriter, GinkgoWriter)
			Expect(err).NotTo(HaveOccurred())
			Eventually(session).Should(gexec.Exit(0))
			tarFilePath := validatedTarFilePath(outputDirPath)
			assertValidOutput(tarFilePath, "ops_manager_vm_types", "development")
			assertLogging(session, tarFilePath, flagValues[cmd.OpsManagerURLFlag])
		})
	})

	Context("client/secret authentication", func() {
		It("succeeds with env variable configuration", func() {
			delete(defaultEnvVars, cmd.OpsManagerUsernameKey)
			delete(defaultEnvVars, cmd.OpsManagerPasswordKey)
			defaultEnvVars[cmd.OpsManagerClientIdKey] = os.Getenv("TEST_OPS_MANAGER_CLIENT_ID")
			defaultEnvVars[cmd.OpsManagerClientSecretKey] = os.Getenv("TEST_OPS_MANAGER_CLIENT_SECRET")
			command := buildDefaultCommand(defaultEnvVars)
			session, err := gexec.Start(command, GinkgoWriter, GinkgoWriter)
			Expect(err).NotTo(HaveOccurred())
			Eventually(session).Should(gexec.Exit(0))

			tarFilePath := validatedTarFilePath(outputDirPath)
			assertValidOutput(tarFilePath, "ops_manager_vm_types", "development")
			assertLogging(session, tarFilePath, defaultEnvVars[cmd.OpsManagerURLKey])
		})

		It("succeeds with flag configuration", func() {
			flagValues := map[string]string{
				cmd.OpsManagerURLFlag:          os.Getenv("TEST_OPS_MANAGER_URL"),
				cmd.OpsManagerClientIdFlag:     os.Getenv("TEST_OPS_MANAGER_CLIENT_ID"),
				cmd.OpsManagerClientSecretFlag: os.Getenv("TEST_OPS_MANAGER_CLIENT_SECRET"),
				cmd.EnvTypeFlag:                "Development",
				cmd.SkipTlsVerifyFlag:          "true",
				cmd.OutputPathFlag:             outputDirPath,
			}
			command := exec.Command(aqueductBinaryPath, "collect")
			for k, v := range flagValues {
				command.Args = append(command.Args, fmt.Sprintf("--%s=%s", k, v))
			}
			session, err := gexec.Start(command, GinkgoWriter, GinkgoWriter)
			Expect(err).NotTo(HaveOccurred())
			Eventually(session).Should(gexec.Exit(0))
			tarFilePath := validatedTarFilePath(outputDirPath)
			assertValidOutput(tarFilePath, "ops_manager_vm_types", "development")
			assertLogging(session, tarFilePath, flagValues[cmd.OpsManagerURLFlag])
		})

	})

	DescribeTable(
		"succeeds with valid env type configuration",
		func(envType string) {
			command := buildDefaultCommand(defaultEnvVars)
			command.Env = append(command.Env, fmt.Sprintf("%s=%s", cmd.EnvTypeKey, envType))
			session, err := gexec.Start(command, GinkgoWriter, GinkgoWriter)
			Expect(err).NotTo(HaveOccurred())
			Eventually(session).Should(gexec.Exit(0))
		},
		Entry(cmd.EnvTypeDevelopment, cmd.EnvTypeDevelopment),
		Entry(cmd.EnvTypeQA, cmd.EnvTypeQA),
		Entry(cmd.EnvTypePreProduction, cmd.EnvTypePreProduction),
		Entry(cmd.EnvTypeProduction, cmd.EnvTypeProduction),
	)

	It("fails with the correct error message when required variables are not set", func() {
		command := exec.Command(aqueductBinaryPath, "collect")
		session, err := gexec.Start(command, GinkgoWriter, GinkgoWriter)
		Expect(err).ToNot(HaveOccurred())
		Eventually(session).Should(gexec.Exit(1))
		requiredFlags := []string{"--" + cmd.OpsManagerURLFlag, "--" + cmd.EnvTypeFlag, "--" + cmd.OutputPathFlag}
		Expect(session.Err).To(gbytes.Say(fmt.Sprintf(cmd.RequiredConfigErrorFormat, strings.Join(requiredFlags, ", "))))
		Expect(session.Err).To(gbytes.Say("Usage:"))
		assertOutputDirEmpty(outputDirPath)
	})

	DescribeTable(
		"fails when there is no pair of auth credentials",
		func(keysToRemove ...string) {
			command := exec.Command(aqueductBinaryPath, "collect")
			for _, key := range keysToRemove {
				delete(defaultEnvVars, key)
			}
			for k, v := range defaultEnvVars {
				command.Env = append(command.Env, fmt.Sprintf("%s=%s", k, v))
			}

			session, err := gexec.Start(command, GinkgoWriter, GinkgoWriter)
			Expect(err).ToNot(HaveOccurred())
			Eventually(session).Should(gexec.Exit(1))
			Eventually(session.Err).Should(gbytes.Say(cmd.InvalidAuthConfigurationMessage))
			Expect(session.Err).To(gbytes.Say("Usage:"))
			assertOutputDirEmpty(outputDirPath)
		},
		Entry("none provided", cmd.OpsManagerUsernameKey, cmd.OpsManagerPasswordKey, cmd.OpsManagerClientIdKey, cmd.OpsManagerClientSecretKey),
		Entry("missing username and client id", cmd.OpsManagerUsernameKey, cmd.OpsManagerClientIdKey),
		Entry("missing username and client secret", cmd.OpsManagerUsernameKey, cmd.OpsManagerClientSecretKey),
		Entry("missing password and client id", cmd.OpsManagerPasswordKey, cmd.OpsManagerClientIdKey),
		Entry("missing password and client secret", cmd.OpsManagerPasswordKey, cmd.OpsManagerClientSecretKey),
	)

	It("fails if the passed in env type is invalid", func() {
		command := buildDefaultCommand(defaultEnvVars)
		command.Env = append(command.Env, fmt.Sprintf("%s=%s", cmd.EnvTypeKey, "invalid-type"))
		session, err := gexec.Start(command, GinkgoWriter, GinkgoWriter)
		Expect(err).NotTo(HaveOccurred())
		Eventually(session).Should(gexec.Exit(1))
		Expect(session.Err).To(gbytes.Say(fmt.Sprintf(cmd.InvalidEnvTypeFailureFormat, "invalid-type")))
		Expect(session.Err).To(gbytes.Say("Usage:"))
		assertOutputDirEmpty(outputDirPath)
	})

	It("fails if data collection from Operations Manager fails", func() {
		defaultEnvVars[cmd.OpsManagerUsernameKey] = "non-real-user"
		command := buildDefaultCommand(defaultEnvVars)
		session, err := gexec.Start(command, GinkgoWriter, GinkgoWriter)
		Expect(err).NotTo(HaveOccurred())
		Eventually(session).Should(gexec.Exit(1))
		Expect(session.Err).To(gbytes.Say(ops.CollectFailureMessage))
		Expect(session.Err).NotTo(gbytes.Say("Usage:"))
		assertOutputDirEmpty(outputDirPath)
	})

	It("fails if the output directory does not exist", func() {
		defaultEnvVars[cmd.OutputPathKey] = "/not/a/real/path"
		command := buildDefaultCommand(defaultEnvVars)
		session, err := gexec.Start(command, GinkgoWriter, GinkgoWriter)
		Expect(err).NotTo(HaveOccurred())
		Eventually(session).Should(gexec.Exit(1))
		Expect(session.Err).To(gbytes.Say(fmt.Sprintf(file.CreateTarFileFailureFormat, "")))
		Expect(session.Err).NotTo(gbytes.Say("Usage:"))
	})
})

func validatedTarFilePath(outputDirPath string) string {
	fileInfos, err := ioutil.ReadDir(outputDirPath)
	Expect(err).NotTo(HaveOccurred())
	Expect(len(fileInfos)).To(Equal(1), fmt.Sprintf("Expected output dir %s to include a single file", outputDirPath))
	Expect(fileInfos[0].Name()).To(MatchRegexp(fmt.Sprintf(`%s%s.tar$`, cmd.OutputFilePrefix, UnixTimestampRegexp)))
	return filepath.Join(outputDirPath, fileInfos[0].Name())
}

func assertMetadataFileIsCorrect(contentDir, expectedEnvType string) {
	content, err := ioutil.ReadFile(filepath.Join(contentDir, ops.MetadataFileName))
	Expect(err).NotTo(HaveOccurred(), "Expected metadata file to exist but did not")
	var metadata data.Metadata
	Expect(json.Unmarshal(content, &metadata)).To(Succeed())
	Expect(metadata.EnvType).To(Equal(expectedEnvType))
}

func assertOutputDirEmpty(outputDirPath string) {
	fileInfos, err := ioutil.ReadDir(outputDirPath)
	Expect(err).NotTo(HaveOccurred())
	Expect(len(fileInfos)).To(Equal(0), fmt.Sprintf("Expected output dir %s to be empty", outputDirPath))
}

func assertValidOutput(tarFilePath, filename, envType string) {
	tmpDir, err := ioutil.TempDir("", "")
	Expect(err).NotTo(HaveOccurred())
	defer os.RemoveAll(tmpDir)

	err = archiver.Tar.Open(tarFilePath, tmpDir)
	Expect(err).NotTo(HaveOccurred())

	jsonFilePath := filepath.Join(tmpDir, filename)
	Expect(jsonFilePath).To(BeAnExistingFile())
	content, err := ioutil.ReadFile(jsonFilePath)
	Expect(err).NotTo(HaveOccurred())
	Expect(json.Valid(content)).To(BeTrue(), fmt.Sprintf("Expected file %s to contain valid json", jsonFilePath))
	assertMetadataFileIsCorrect(tmpDir, envType)
}

func assertLogging(session *gexec.Session, tarFilePath, url string) {
	Expect(session.Out).To(gbytes.Say(fmt.Sprintf("Collecting data from Operations Manager at %s\n", url)))
	Expect(session.Out).To(gbytes.Say(fmt.Sprintf("Wrote output to %s\n", escapeWindowsPathRegex(tarFilePath))))
	Expect(session.Out).To(gbytes.Say("Success!\n"))
}

func buildDefaultCommand(envVars map[string]string) *exec.Cmd {
	command := exec.Command(aqueductBinaryPath, "collect", "--"+cmd.SkipTlsVerifyFlag)
	command.Env = os.Environ()
	for k, v := range envVars {
		command.Env = append(command.Env, fmt.Sprintf("%s=%s", k, v))
	}
	return command
}
