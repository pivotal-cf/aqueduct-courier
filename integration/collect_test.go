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
		defaultEnvVars = map[string]string{
			cmd.OpsManagerURLKey:      os.Getenv("TEST_OPSMANAGER_URL"),
			cmd.OpsManagerUsernameKey: os.Getenv("TEST_OPSMANAGER_USERNAME"),
			cmd.OpsManagerPasswordKey: os.Getenv("TEST_OPSMANAGER_PASSWORD"),
			cmd.EnvTypeKey:            "Development",
			cmd.SkipTlsVerifyKey:      "true",
		}
		additionalEnvVars map[string]string
	)

	BeforeEach(func() {
		var err error
		outputDirPath, err = ioutil.TempDir("", "")
		Expect(err).NotTo(HaveOccurred())
		additionalEnvVars = map[string]string{
			cmd.OutputPathKey: outputDirPath,
		}
	})

	AfterEach(func() {
		Expect(os.RemoveAll(outputDirPath)).To(Succeed())
	})

	It("succeeds with env variables", func() {
		command := exec.Command(aqueductBinaryPath, "collect")
		command.Env = os.Environ()
		for k, v := range merge(defaultEnvVars, additionalEnvVars) {
			command.Env = append(command.Env, fmt.Sprintf("%s=%s", k, v))
		}
		session, err := gexec.Start(command, GinkgoWriter, GinkgoWriter)
		Expect(err).NotTo(HaveOccurred())
		Eventually(session, 30*time.Second).Should(gexec.Exit(0))

		contentDir := validatedContentDir(outputDirPath)
		assertContainsJsonFile(contentDir, "ops_manager_vm_types")
		assertMetadataFileIsCorrect(contentDir, "development")
	})

	It("succeeds with flag configuration", func() {
		flagValues := map[string]string{
			cmd.OpsManagerURLFlag:      os.Getenv("TEST_OPSMANAGER_URL"),
			cmd.OpsManagerUsernameFlag: os.Getenv("TEST_OPSMANAGER_USERNAME"),
			cmd.OpsManagerPasswordFlag: os.Getenv("TEST_OPSMANAGER_PASSWORD"),
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

	DescribeTable(
		"succeeds with valid env type configuration",
		func(envType string) {
			command := exec.Command(aqueductBinaryPath, "collect")
			command.Env = os.Environ()
			for k, v := range merge(defaultEnvVars, additionalEnvVars) {
				command.Env = append(command.Env, fmt.Sprintf("%s=%s", k, v))
			}
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

	It("fails if data collection from Operations Manager fails", func() {
		additionalEnvVars[cmd.OpsManagerUsernameKey] = "non-real-user"
		command := exec.Command(aqueductBinaryPath, "collect")
		for k, v := range merge(defaultEnvVars, additionalEnvVars) {
			command.Env = append(command.Env, fmt.Sprintf("%s=%s", k, v))
		}
		session, err := gexec.Start(command, GinkgoWriter, GinkgoWriter)
		Expect(err).NotTo(HaveOccurred())
		Eventually(session, 30*time.Second).Should(gexec.Exit(1))
		Expect(session.Err).To(gbytes.Say(ops.CollectFailureMessage))
	})

	DescribeTable(
		"fails when required variable is not set",
		func(missingKey, missingFlag string) {
			command := exec.Command(aqueductBinaryPath, "collect")
			for k, v := range merge(defaultEnvVars, additionalEnvVars) {
				if k != missingKey {
					command.Env = append(command.Env, fmt.Sprintf("%s=%s", k, v))
				}
			}
			session, err := gexec.Start(command, GinkgoWriter, GinkgoWriter)
			Expect(err).ToNot(HaveOccurred())
			Eventually(session).Should(gexec.Exit(1))
			Eventually(session.Err).Should(gbytes.Say(fmt.Sprintf(cmd.RequiredConfigErrorFormat, missingFlag)))
		},
		Entry(cmd.OpsManagerURLKey, cmd.OpsManagerURLKey, cmd.OpsManagerURLFlag),
		Entry(cmd.OpsManagerUsernameKey, cmd.OpsManagerUsernameKey, cmd.OpsManagerUsernameFlag),
		Entry(cmd.OpsManagerPasswordKey, cmd.OpsManagerPasswordKey, cmd.OpsManagerPasswordFlag),
		Entry(cmd.EnvTypeKey, cmd.EnvTypeKey, cmd.EnvTypeFlag),
		Entry(cmd.OutputPathKey, cmd.OutputPathKey, cmd.OutputPathFlag),
	)

	It("fails if the passed in env type is invalid", func() {
		command := exec.Command(aqueductBinaryPath, "collect")
		command.Env = os.Environ()
		for k, v := range merge(defaultEnvVars, additionalEnvVars) {
			command.Env = append(command.Env, fmt.Sprintf("%s=%s", k, v))
		}
		command.Env = append(command.Env, fmt.Sprintf("%s=%s", cmd.EnvTypeKey, "invalid-type"))
		session, err := gexec.Start(command, GinkgoWriter, GinkgoWriter)
		Expect(err).NotTo(HaveOccurred())
		Eventually(session, 30*time.Second).Should(gexec.Exit(1))
		Expect(session.Err).To(gbytes.Say(fmt.Sprintf(cmd.InvalidEnvTypeFailureFormat, "invalid-type")))
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

func assertMetadataFileIsCorrect(contentDir, expectedEnvType string) {
	content, err := ioutil.ReadFile(filepath.Join(contentDir, file.MetadataFileName))
	Expect(err).NotTo(HaveOccurred(), "Expected metadata file to exist but did not")
	var metadata file.Metadata
	Expect(json.Unmarshal(content, &metadata)).To(Succeed())
	Expect(metadata.EnvType).To(Equal(expectedEnvType))
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
