package integration

import (
	"encoding/json"
	"fmt"
	"io/ioutil"
	"os"
	"os/exec"
	"path/filepath"
	"time"

	. "github.com/onsi/ginkgo"
	. "github.com/onsi/ginkgo/extensions/table"
	. "github.com/onsi/gomega"
	"github.com/onsi/gomega/gbytes"
	"github.com/onsi/gomega/gexec"
	"github.com/pivotal-cf/aqueduct-courier/cmd"
	"github.com/pivotal-cf/aqueduct-courier/file"
	"github.com/pivotal-cf/aqueduct-courier/ops"
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
			cmd.SkipTlsVerifyKey:      "true",
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
			Eventually(session, 30*time.Second).Should(gexec.Exit(0))

			contentDir := validatedContentDir(outputDirPath)
			assertContainsJsonFile(contentDir, "ops_manager_vm_types")
			assertMetadataFileIsCorrect(contentDir, "development")
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
			Eventually(session, 30*time.Second).Should(gexec.Exit(0))
			contentDir := validatedContentDir(outputDirPath)
			assertContainsJsonFile(contentDir, "ops_manager_vm_types")
			assertMetadataFileIsCorrect(contentDir, "development")
		})
	})

	Context("client/secret authentication", func() {
		It("succeeds with env variable configuration", func() {
			//needs to change to client/secret not user/pass
			delete(defaultEnvVars, cmd.OpsManagerUsernameKey)
			delete(defaultEnvVars, cmd.OpsManagerPasswordKey)
			defaultEnvVars[cmd.OpsManagerClientIdKey] = os.Getenv("TEST_OPS_MANAGER_CLIENT_ID")
			defaultEnvVars[cmd.OpsManagerClientSecretKey] = os.Getenv("TEST_OPS_MANAGER_CLIENT_SECRET")
			command := buildDefaultCommand(defaultEnvVars)
			session, err := gexec.Start(command, GinkgoWriter, GinkgoWriter)
			Expect(err).NotTo(HaveOccurred())
			Eventually(session, 30*time.Second).Should(gexec.Exit(0))

			contentDir := validatedContentDir(outputDirPath)
			assertContainsJsonFile(contentDir, "ops_manager_vm_types")
			assertMetadataFileIsCorrect(contentDir, "development")
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
			Eventually(session, 30*time.Second).Should(gexec.Exit(0))
			contentDir := validatedContentDir(outputDirPath)
			assertContainsJsonFile(contentDir, "ops_manager_vm_types")
			assertMetadataFileIsCorrect(contentDir, "development")
		})

	})

	DescribeTable(
		"succeeds with valid env type configuration",
		func(envType string) {
			command := buildDefaultCommand(defaultEnvVars)
			command.Env = append(command.Env, fmt.Sprintf("%s=%s", cmd.EnvTypeKey, envType))
			session, err := gexec.Start(command, GinkgoWriter, GinkgoWriter)
			Expect(err).NotTo(HaveOccurred())
			Eventually(session, 30*time.Second).Should(gexec.Exit(0))
		},
		Entry(cmd.EnvTypeDevelopment, cmd.EnvTypeDevelopment),
		Entry(cmd.EnvTypeQA, cmd.EnvTypeQA),
		Entry(cmd.EnvTypePreProduction, cmd.EnvTypePreProduction),
		Entry(cmd.EnvTypeProduction, cmd.EnvTypeProduction),
	)

	DescribeTable(
		"fails when required variable is not set",
		func(missingKey, missingFlag string) {
			command := exec.Command(aqueductBinaryPath, "collect")
			for k, v := range defaultEnvVars {
				if k != missingKey {
					command.Env = append(command.Env, fmt.Sprintf("%s=%s", k, v))
				}
			}
			session, err := gexec.Start(command, GinkgoWriter, GinkgoWriter)
			Expect(err).ToNot(HaveOccurred())
			Eventually(session, 30*time.Second).Should(gexec.Exit(1))
			Eventually(session.Err).Should(gbytes.Say(fmt.Sprintf(cmd.RequiredConfigErrorFormat, missingFlag)))
		},
		Entry(cmd.OpsManagerURLKey, cmd.OpsManagerURLKey, cmd.OpsManagerURLFlag),
		Entry(cmd.EnvTypeKey, cmd.EnvTypeKey, cmd.EnvTypeFlag),
		Entry(cmd.OutputPathKey, cmd.OutputPathKey, cmd.OutputPathFlag),
	)

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
			Eventually(session, 30*time.Second).Should(gexec.Exit(1))
			Eventually(session.Err).Should(gbytes.Say(cmd.InvalidAuthConfigurationMessage))
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
		Eventually(session, 30*time.Second).Should(gexec.Exit(1))
		Expect(session.Err).To(gbytes.Say(fmt.Sprintf(cmd.InvalidEnvTypeFailureFormat, "invalid-type")))
	})

	It("fails if data collection from Operations Manager fails", func() {
		defaultEnvVars[cmd.OpsManagerUsernameKey] = "non-real-user"
		command := buildDefaultCommand(defaultEnvVars)
		session, err := gexec.Start(command, GinkgoWriter, GinkgoWriter)
		Expect(err).NotTo(HaveOccurred())
		Eventually(session, 30*time.Second).Should(gexec.Exit(1))
		Expect(session.Err).To(gbytes.Say(ops.CollectFailureMessage))
	})
})

func validatedContentDir(outputDirPath string) string {
	fileInfos, err := ioutil.ReadDir(outputDirPath)
	Expect(err).NotTo(HaveOccurred())
	Expect(len(fileInfos)).To(Equal(1), fmt.Sprintf("Expected output dir %s to include a single directory", outputDirPath))
	Expect(fileInfos[0].IsDir()).To(BeTrue(), fmt.Sprintf("Expected file %s found in %s to be a directory", fileInfos[0], outputDirPath))
	Expect(fileInfos[0].Name()).To(MatchRegexp(fmt.Sprintf(`%s%s$`, file.OutputDirPrefix, UnixTimestampRegexp)))
	return filepath.Join(outputDirPath, fileInfos[0].Name())
}

func assertContainsJsonFile(contentDir, filename string) {
	jsonFilePath := filepath.Join(contentDir, filename)
	Expect(jsonFilePath).To(BeAnExistingFile())
	content, err := ioutil.ReadFile(jsonFilePath)
	Expect(err).NotTo(HaveOccurred())
	Expect(json.Valid(content)).To(BeTrue(), fmt.Sprintf("Expected file %s to contain valid json", jsonFilePath))
}

func assertMetadataFileIsCorrect(contentDir, expectedEnvType string) {
	content, err := ioutil.ReadFile(filepath.Join(contentDir, file.MetadataFileName))
	Expect(err).NotTo(HaveOccurred(), "Expected metadata file to exist but did not")
	var metadata file.Metadata
	Expect(json.Unmarshal(content, &metadata)).To(Succeed())
	Expect(metadata.EnvType).To(Equal(expectedEnvType))
}

func buildDefaultCommand(envVars map[string]string) *exec.Cmd {
	command := exec.Command(aqueductBinaryPath, "collect")
	command.Env = os.Environ()
	for k, v := range envVars {
		command.Env = append(command.Env, fmt.Sprintf("%s=%s", k, v))
	}
	return command
}
