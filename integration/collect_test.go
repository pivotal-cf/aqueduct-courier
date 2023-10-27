package integration

import (
	"crypto/tls"
	"encoding/base64"
	"encoding/json"
	"fmt"
	"log"
	"net"
	"net/http"
	"net/url"
	"os"
	"os/exec"
	"path/filepath"
	"strings"

	"github.com/pivotal-cf/om/api"

	"github.com/pivotal-cf/aqueduct-courier/cf"

	"github.com/elazarl/goproxy"
	"github.com/mholt/archiver/v3"
	. "github.com/onsi/ginkgo"
	. "github.com/onsi/ginkgo/extensions/table"
	. "github.com/onsi/gomega"
	"github.com/onsi/gomega/gbytes"
	"github.com/onsi/gomega/gexec"
	"github.com/onsi/gomega/ghttp"
	"github.com/pivotal-cf/aqueduct-courier/cmd"
	"github.com/pivotal-cf/aqueduct-courier/operations"
	"github.com/pivotal-cf/telemetry-utils/collector_tar"
)

const (
	UnixTimestampRegexp = `\d{10}`
)

var _ = Describe("Collect", func() {
	var (
		outputDirPath    string
		defaultEnvVars   map[string]string
		opsManagerServer *ghttp.Server
		configDirPath    string
	)

	BeforeEach(func() {
		var err error
		outputDirPath, err = os.MkdirTemp("", "")
		Expect(err).NotTo(HaveOccurred())

		opsManagerServer = setupOpsManagerServer()
		defaultEnvVars = map[string]string{
			cmd.OpsManagerURLKey:      opsManagerServer.URL(),
			cmd.OpsManagerUsernameKey: "some-username",
			cmd.OpsManagerPasswordKey: "some-password",
			cmd.EnvTypeKey:            "Development",
			cmd.OutputPathKey:         outputDirPath,
		}

		configDirPath, err = os.MkdirTemp("", "")
		Expect(err).NotTo(HaveOccurred())
	})

	AfterEach(func() {
		opsManagerServer.Close()
		Expect(os.RemoveAll(outputDirPath)).To(Succeed())

		Expect(os.RemoveAll(configDirPath)).To(Succeed())
	})

	Context("aliases for config file input", func() {
		It("doesn't print the aliased commands in the help", func() {
			command := exec.Command(aqueductBinaryPath, "collect", "--help")

			session, err := gexec.Start(command, GinkgoWriter, GinkgoWriter)
			Expect(err).NotTo(HaveOccurred())
			Eventually(session).Should(gexec.Exit(0))

			stdOut := string(session.Out.Contents())
			Expect(stdOut).ToNot(ContainSubstring("--target"))
			Expect(stdOut).ToNot(ContainSubstring("--skip-ssl-validation"))
		})

		It("accepts an aliased url in the config file configuration", func() {
			config := fmt.Sprintf(`{
				"target": "%s",
				"username": "some-username",
				"password": "some-password",
				"env-type": "Development",
				"insecure-skip-tls-verify": true,
				"output-dir": "%s"
			}`, opsManagerServer.URL(), escapeFilePathForWindows(outputDirPath))
			configFile := filepath.Join(configDirPath, "config.yml")
			err := os.WriteFile(configFile, []byte(config), 0755)
			Expect(err).ToNot(HaveOccurred())

			command := exec.Command(aqueductBinaryPath, "collect", "--config", configFile)

			session, err := gexec.Start(command, GinkgoWriter, GinkgoWriter)
			Expect(err).NotTo(HaveOccurred())
			Eventually(session).Should(gexec.Exit(0))
			tarFilePath := validatedTarFilePath(outputDirPath)
			assertValidOutput(tarFilePath, collector_tar.OpsManagerCollectorDataSetId, "ops_manager_vm_types", "development")
			assertLogging(session, tarFilePath, false, false)
		})

		It("does not accept an aliased url in config file configuration if url is defined as a flag", func() {
			config := fmt.Sprintf(`{
				"target": "invalid.url.example.com",
				"username": "some-username",
				"password": "some-password",
				"env-type": "Development",
				"insecure-skip-tls-verify": true,
				"output-dir": "%s"
			}`, escapeFilePathForWindows(outputDirPath))
			configFile := filepath.Join(configDirPath, "config.yml")
			err := os.WriteFile(configFile, []byte(config), 0755)
			Expect(err).ToNot(HaveOccurred())

			command := exec.Command(aqueductBinaryPath, "collect",
				"--config", configFile,
				"--"+cmd.OpsManagerURLFlag, opsManagerServer.URL(),
			)

			session, err := gexec.Start(command, GinkgoWriter, GinkgoWriter)
			Expect(err).NotTo(HaveOccurred())
			Eventually(session).Should(gexec.Exit(0))
			tarFilePath := validatedTarFilePath(outputDirPath)
			assertValidOutput(tarFilePath, collector_tar.OpsManagerCollectorDataSetId, "ops_manager_vm_types", "development")
			assertLogging(session, tarFilePath, false, false)
		})

		It("does not accept an aliased url in config file configuration if url is defined as an environment variable", func() {
			config := fmt.Sprintf(`{
				"target": "invalid.url.example.com",
				"username": "some-username",
				"password": "some-password",
				"env-type": "Development",
				"insecure-skip-tls-verify": true,
				"output-dir": "%s"
			}`, escapeFilePathForWindows(outputDirPath))
			configFile := filepath.Join(configDirPath, "config.yml")
			err := os.WriteFile(configFile, []byte(config), 0755)
			Expect(err).ToNot(HaveOccurred())

			command := exec.Command(aqueductBinaryPath, "collect", "--config", configFile)
			command.Env = os.Environ()
			command.Env = append(command.Env, fmt.Sprintf("%s=%s", cmd.OpsManagerURLKey, opsManagerServer.URL()))

			session, err := gexec.Start(command, GinkgoWriter, GinkgoWriter)
			Expect(err).NotTo(HaveOccurred())
			Eventually(session).Should(gexec.Exit(0))
			tarFilePath := validatedTarFilePath(outputDirPath)
			assertValidOutput(tarFilePath, collector_tar.OpsManagerCollectorDataSetId, "ops_manager_vm_types", "development")
			assertLogging(session, tarFilePath, false, false)
		})

		It("accepts an aliased skip tls validation in the config file configuration", func() {
			config := fmt.Sprintf(`{
				"url": "%s",
				"username": "some-username",
				"password": "some-password",
				"env-type": "Development",
				"skip-ssl-validation": true,
				"output-dir": "%s"
			}`, opsManagerServer.URL(), escapeFilePathForWindows(outputDirPath))
			configFile := filepath.Join(configDirPath, "config.yml")
			err := os.WriteFile(configFile, []byte(config), 0755)
			Expect(err).ToNot(HaveOccurred())

			command := exec.Command(aqueductBinaryPath, "collect", "--config", configFile)

			session, err := gexec.Start(command, GinkgoWriter, GinkgoWriter)
			Expect(err).NotTo(HaveOccurred())
			Eventually(session).Should(gexec.Exit(0))
			tarFilePath := validatedTarFilePath(outputDirPath)
			assertValidOutput(tarFilePath, collector_tar.OpsManagerCollectorDataSetId, "ops_manager_vm_types", "development")
			assertLogging(session, tarFilePath, false, false)
		})

		It("does not accept an aliased skip tls validation in config file configuration if skip tls validation is defined as a flag", func() {
			config := fmt.Sprintf(
				`{
				"url": "%s",
				"username": "some-username",
				"password": "some-password",
				"env-type": "Development",
				"skip-ssl-validation": false,
				"output-dir": "%s"
			}`, opsManagerServer.URL(), escapeFilePathForWindows(outputDirPath))
			configFile := filepath.Join(configDirPath, "config.yml")
			err := os.WriteFile(configFile, []byte(config), 0755)
			Expect(err).ToNot(HaveOccurred())

			command := exec.Command(aqueductBinaryPath, "collect",
				"--config", configFile,
				"--"+cmd.SkipTlsVerifyFlag, "true",
			)

			session, err := gexec.Start(command, GinkgoWriter, GinkgoWriter)
			Expect(err).NotTo(HaveOccurred())
			Eventually(session).Should(gexec.Exit(0))
			tarFilePath := validatedTarFilePath(outputDirPath)
			assertValidOutput(tarFilePath, collector_tar.OpsManagerCollectorDataSetId, "ops_manager_vm_types", "development")
			assertLogging(session, tarFilePath, false, false)
		})

		It("does not accept an aliased skip tls validation in config file configuration if skip tls validation is defined as an environment variable", func() {
			config := fmt.Sprintf(`{
				"url": "%s",
				"username": "some-username",
				"password": "some-password",
				"env-type": "Development",
				"skip-ssl-validation": "false",
				"output-dir": "%s"
			}`, opsManagerServer.URL(), escapeFilePathForWindows(outputDirPath))
			configFile := filepath.Join(configDirPath, "config.yml")
			err := os.WriteFile(configFile, []byte(config), 0755)
			Expect(err).ToNot(HaveOccurred())

			command := exec.Command(aqueductBinaryPath, "collect", "--config", configFile)
			command.Env = os.Environ()
			command.Env = append(command.Env, fmt.Sprintf("%s=%s", cmd.SkipTlsVerifyKey, "true"))

			session, err := gexec.Start(command, GinkgoWriter, GinkgoWriter)
			Expect(err).NotTo(HaveOccurred())
			Eventually(session).Should(gexec.Exit(0))
			tarFilePath := validatedTarFilePath(outputDirPath)
			assertValidOutput(tarFilePath, collector_tar.OpsManagerCollectorDataSetId, "ops_manager_vm_types", "development")
			assertLogging(session, tarFilePath, false, false)
		})

		It("accepts an aliased skip tls validation with flag configuration", func() {
			flagValues := map[string]string{
				cmd.OpsManagerURLFlag:      opsManagerServer.URL(),
				cmd.OpsManagerUsernameFlag: "whatever",
				cmd.OpsManagerPasswordFlag: "whatever",
				cmd.EnvTypeFlag:            "Development",
				cmd.OutputPathFlag:         outputDirPath,
			}
			command := exec.Command(aqueductBinaryPath, "collect")
			for k, v := range flagValues {
				command.Args = append(command.Args, fmt.Sprintf("--%s=%s", k, v))
			}
			command.Args = append(command.Args, "--skip-ssl-validation")

			session, err := gexec.Start(command, GinkgoWriter, GinkgoWriter)
			Expect(err).NotTo(HaveOccurred())
			Eventually(session).Should(gexec.Exit(0))
			tarFilePath := validatedTarFilePath(outputDirPath)
			assertValidOutput(tarFilePath, collector_tar.OpsManagerCollectorDataSetId, "ops_manager_vm_types", "development")
			assertLogging(session, tarFilePath, false, false)
		})
	})

	DescribeTable(
		"succeeds with valid env type configuration from an environment variable",
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
		Entry(cmd.EnvTypeSandbox, cmd.EnvTypeSandbox),
	)

	DescribeTable(
		"succeeds with valid env type configuration from a config file",
		func(envType string) {
			config := fmt.Sprintf(`{
				"env-type": "%s",
				"url": "%s",
				"username": "some-username",
				"password": "some-password",
				"insecure-skip-tls-verify": "true",
				"output-dir": "%s"
			}`, envType, opsManagerServer.URL(), escapeFilePathForWindows(outputDirPath))

			configFile := filepath.Join(configDirPath, "config.yml")
			err := os.WriteFile(configFile, []byte(config), 0755)
			Expect(err).ToNot(HaveOccurred())

			command := exec.Command(aqueductBinaryPath, "collect", "--config", configFile)

			session, err := gexec.Start(command, GinkgoWriter, GinkgoWriter)
			Expect(err).NotTo(HaveOccurred())
			Eventually(session).Should(gexec.Exit(0))
		},
		Entry(cmd.EnvTypeDevelopment, cmd.EnvTypeDevelopment),
		Entry(cmd.EnvTypeQA, cmd.EnvTypeQA),
		Entry(cmd.EnvTypePreProduction, cmd.EnvTypePreProduction),
		Entry(cmd.EnvTypeProduction, cmd.EnvTypeProduction),
		Entry(cmd.EnvTypeSandbox, cmd.EnvTypeSandbox),
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
			Eventually(session).Should(gexec.Exit(1))
			Eventually(session.Err).Should(gbytes.Say(cmd.InvalidAuthConfigurationMessage))
			Expect(session.Err).To(gbytes.Say("USAGE EXAMPLES"))
			assertOutputDirEmpty(outputDirPath)
		},
		Entry("none provided", cmd.OpsManagerUsernameKey, cmd.OpsManagerPasswordKey, cmd.OpsManagerClientIdKey, cmd.OpsManagerClientSecretKey),
		Entry("missing username and client id", cmd.OpsManagerUsernameKey, cmd.OpsManagerClientIdKey),
		Entry("missing username and client secret", cmd.OpsManagerUsernameKey, cmd.OpsManagerClientSecretKey),
		Entry("missing password and client id", cmd.OpsManagerPasswordKey, cmd.OpsManagerClientIdKey),
		Entry("missing password and client secret", cmd.OpsManagerPasswordKey, cmd.OpsManagerClientSecretKey),
	)

	Context("with ops manager user/password authentication", func() {
		It("succeeds with env variable configuration", func() {
			command := buildDefaultCommand(defaultEnvVars)
			session, err := gexec.Start(command, GinkgoWriter, GinkgoWriter)
			Expect(err).NotTo(HaveOccurred())
			Eventually(session).Should(gexec.Exit(0))

			tarFilePath := validatedTarFilePath(outputDirPath)
			assertValidOutput(tarFilePath, collector_tar.OpsManagerCollectorDataSetId, "ops_manager_vm_types", "development")
			assertLogging(session, tarFilePath, false, false)
		})

		It("succeeds with flag configuration", func() {
			flagValues := map[string]string{
				cmd.OpsManagerURLFlag:      opsManagerServer.URL(),
				cmd.OpsManagerUsernameFlag: "some-username",
				cmd.OpsManagerPasswordFlag: "some-password",
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
			assertValidOutput(tarFilePath, collector_tar.OpsManagerCollectorDataSetId, "ops_manager_vm_types", "development")
			assertLogging(session, tarFilePath, false, false)
		})

		It("succeeds with config file configuration", func() {
			config := fmt.Sprintf(`{
				"url": "%s",
				"username": "some-username",
				"password": "some-password",
				"env-type": "Development",
				"insecure-skip-tls-verify": "true",
				"output-dir": "%s"
			}`, opsManagerServer.URL(), escapeFilePathForWindows(outputDirPath))

			configFile := filepath.Join(configDirPath, "config.yml")
			err := os.WriteFile(configFile, []byte(config), 0755)
			Expect(err).ToNot(HaveOccurred())

			command := exec.Command(aqueductBinaryPath, "collect", "--config", configFile)

			session, err := gexec.Start(command, GinkgoWriter, GinkgoWriter)
			Expect(err).NotTo(HaveOccurred())
			Eventually(session).Should(gexec.Exit(0))
			tarFilePath := validatedTarFilePath(outputDirPath)
			assertValidOutput(tarFilePath, collector_tar.OpsManagerCollectorDataSetId, "ops_manager_vm_types", "development")
			assertLogging(session, tarFilePath, false, false)
		})
	})

	Context("with ops manager client/secret authentication", func() {
		It("succeeds with env variable configuration", func() {
			delete(defaultEnvVars, cmd.OpsManagerUsernameKey)
			delete(defaultEnvVars, cmd.OpsManagerPasswordKey)
			defaultEnvVars[cmd.OpsManagerClientIdKey] = "some-client-id"
			defaultEnvVars[cmd.OpsManagerClientSecretKey] = "some-client-secret"
			command := buildDefaultCommand(defaultEnvVars)
			session, err := gexec.Start(command, GinkgoWriter, GinkgoWriter)
			Expect(err).NotTo(HaveOccurred())
			Eventually(session).Should(gexec.Exit(0))

			tarFilePath := validatedTarFilePath(outputDirPath)
			assertValidOutput(tarFilePath, collector_tar.OpsManagerCollectorDataSetId, "ops_manager_vm_types", "development")
			assertLogging(session, tarFilePath, false, false)
		})

		It("succeeds with flag configuration", func() {
			flagValues := map[string]string{
				cmd.OpsManagerURLFlag:          opsManagerServer.URL(),
				cmd.OpsManagerClientIdFlag:     "some-client-id",
				cmd.OpsManagerClientSecretFlag: "some-client-secret",
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
			assertValidOutput(tarFilePath, collector_tar.OpsManagerCollectorDataSetId, "ops_manager_vm_types", "development")
			assertLogging(session, tarFilePath, false, false)
		})

		It("succeeds with config file configuration", func() {
			config := fmt.Sprintf(`{
				"url": "%s",
				"client-id": "some-client-id",
				"client-secret": "some-client-secret",
				"env-type": "Development",
				"insecure-skip-tls-verify": "true",
				"output-dir": "%s"
			}`, opsManagerServer.URL(), escapeFilePathForWindows(outputDirPath))

			configFile := filepath.Join(configDirPath, "config.yml")
			err := os.WriteFile(configFile, []byte(config), 0755)
			Expect(err).ToNot(HaveOccurred())

			command := exec.Command(aqueductBinaryPath, "collect", "--config", configFile)

			session, err := gexec.Start(command, GinkgoWriter, GinkgoWriter)
			Expect(err).NotTo(HaveOccurred())
			Eventually(session).Should(gexec.Exit(0))
			tarFilePath := validatedTarFilePath(outputDirPath)
			assertValidOutput(tarFilePath, collector_tar.OpsManagerCollectorDataSetId, "ops_manager_vm_types", "development")
			assertLogging(session, tarFilePath, false, false)
		})
	})

	Context("specifying foundation nickname", func() {
		It("succeeds with env variable configuration", func() {
			defaultEnvVars[cmd.FoundationNicknameKey] = "some-nickname"
			command := buildDefaultCommand(defaultEnvVars)
			session, err := gexec.Start(command, GinkgoWriter, GinkgoWriter)
			Expect(err).NotTo(HaveOccurred())
			Eventually(session).Should(gexec.Exit(0))

			tarFilePath := validatedTarFilePath(outputDirPath)
			assertValidNickname(tarFilePath, collector_tar.OpsManagerCollectorDataSetId, "ops_manager_vm_types", "some-nickname")
		})

		It("succeeds with flag configuration", func() {
			flagValues := map[string]string{
				cmd.OpsManagerURLFlag:      opsManagerServer.URL(),
				cmd.OpsManagerUsernameFlag: "some-username",
				cmd.OpsManagerPasswordFlag: "some-password",
				cmd.EnvTypeFlag:            "Development",
				cmd.FoundationNicknameFlag: "some-nickname",
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
			assertValidNickname(tarFilePath, collector_tar.OpsManagerCollectorDataSetId, "ops_manager_vm_types", "some-nickname")
		})

		It("succeeds with config file configuration", func() {
			config := fmt.Sprintf(`{
				"url": "%s",
				"username": "some-username",
				"password": "some-password",
				"env-type": "Development",
				"foundation-nickname": "some-nickname",
				"insecure-skip-tls-verify": "true",
				"output-dir": "%s"
			}`, opsManagerServer.URL(), escapeFilePathForWindows(outputDirPath))

			configFile := filepath.Join(configDirPath, "config.yml")
			err := os.WriteFile(configFile, []byte(config), 0755)
			Expect(err).ToNot(HaveOccurred())

			command := exec.Command(aqueductBinaryPath, "collect", "--config", configFile)

			session, err := gexec.Start(command, GinkgoWriter, GinkgoWriter)
			Expect(err).NotTo(HaveOccurred())
			Eventually(session).Should(gexec.Exit(0))
			tarFilePath := validatedTarFilePath(outputDirPath)
			assertValidNickname(tarFilePath, collector_tar.OpsManagerCollectorDataSetId, "ops_manager_vm_types", "some-nickname")
		})
	})

	Context("with usage service client/secret authentication", func() {
		var (
			usageService *ghttp.Server
			cfService    *ghttp.Server
			uaaService   *ghttp.Server
		)
		BeforeEach(func() {
			uaaService, cfService, usageService = setupUsageService("")
		})

		AfterEach(func() {
			usageService.Close()
			cfService.Close()
			uaaService.Close()
		})

		It("succeeds with env variable configuration", func() {
			defaultEnvVars[cmd.CfApiURLKey] = cfService.URL()
			defaultEnvVars[cmd.UsageServiceURLKey] = usageService.URL()
			defaultEnvVars[cmd.UsageServiceClientIDKey] = "best-usage-service-client-id"
			defaultEnvVars[cmd.UsageServiceClientSecretKey] = "best-usage-service-client-secret"
			defaultEnvVars[cmd.UsageServiceSkipTlsVerifyKey] = "true"
			command := buildDefaultCommand(defaultEnvVars)
			session, err := gexec.Start(command, GinkgoWriter, GinkgoWriter)
			Expect(err).NotTo(HaveOccurred())
			Eventually(session).Should(gexec.Exit(0))

			tarFilePath := validatedTarFilePath(outputDirPath)
			assertValidOutput(tarFilePath, collector_tar.OpsManagerCollectorDataSetId, "ops_manager_vm_types", "development")
			assertValidOutput(tarFilePath, collector_tar.UsageServiceCollectorDataSetId, "app_usage", "development")
			assertLogging(session, tarFilePath, false, true)
		})

		It("succeeds with flag configuration", func() {
			flagValues := map[string]string{
				cmd.OpsManagerTimeoutFlag:         "1",
				cmd.OpsManagerRequestTimeoutFlag:  "10",
				cmd.OpsManagerURLFlag:             opsManagerServer.URL(),
				cmd.OpsManagerClientIdFlag:        "whatever",
				cmd.OpsManagerClientSecretFlag:    "whatever",
				cmd.SkipTlsVerifyFlag:             "true",
				cmd.EnvTypeFlag:                   "Development",
				cmd.OutputPathFlag:                outputDirPath,
				cmd.CfApiURLFlag:                  cfService.URL(),
				cmd.UsageServiceURLFlag:           usageService.URL(),
				cmd.UsageServiceClientIDFlag:      "best-usage-service-client-id",
				cmd.UsageServiceClientSecretFlag:  "best-usage-service-client-secret",
				cmd.UsageServiceSkipTlsVerifyFlag: "true",
				cmd.UsageServiceTimeoutFlag:       "10",
			}
			command := exec.Command(aqueductBinaryPath, "collect")
			for k, v := range flagValues {
				command.Args = append(command.Args, fmt.Sprintf("--%s=%s", k, v))
			}
			session, err := gexec.Start(command, GinkgoWriter, GinkgoWriter)
			Expect(err).NotTo(HaveOccurred())
			Eventually(session).Should(gexec.Exit(0))

			tarFilePath := validatedTarFilePath(outputDirPath)
			assertValidOutput(tarFilePath, collector_tar.OpsManagerCollectorDataSetId, "ops_manager_vm_types", "development")
			assertValidOutput(tarFilePath, collector_tar.UsageServiceCollectorDataSetId, "app_usage", "development")
			assertLogging(session, tarFilePath, false, true)
		})

		It("uses the usage service client credentials specified by config file configuration", func() {
			config := fmt.Sprintf(`{
				"ops-manager-timeout": 1,
				"ops-manager-request-timeout": 10,
				"url": "%s",
				"client-id": "whatever",
				"client-secret": "whatever",
				"insecure-skip-tls-verify": true,
				"env-type": "Development",
				"output-dir": "%s",
				"cf-api-url": "%s",
				"usage-service-url": "%s",
				"usage-service-client-id": "best-usage-service-client-id",
				"usage-service-client-secret": "best-usage-service-client-secret",
				"usage-service-insecure-skip-tls-verify": true,
				"usage-service-timeout": 10
			}`, opsManagerServer.URL(), escapeFilePathForWindows(outputDirPath), cfService.URL(), usageService.URL())
			configFile := filepath.Join(configDirPath, "config.yml")
			err := os.WriteFile(configFile, []byte(config), 0755)
			Expect(err).ToNot(HaveOccurred())

			command := exec.Command(aqueductBinaryPath, "collect", "--config", configFile)

			session, err := gexec.Start(command, GinkgoWriter, GinkgoWriter)
			Expect(err).NotTo(HaveOccurred())
			Eventually(session).Should(gexec.Exit(0))

			tarFilePath := validatedTarFilePath(outputDirPath)
			assertValidOutput(tarFilePath, collector_tar.OpsManagerCollectorDataSetId, "ops_manager_vm_types", "development")
			assertValidOutput(tarFilePath, collector_tar.UsageServiceCollectorDataSetId, "app_usage", "development")
			assertLogging(session, tarFilePath, false, true)
		})

		It("fails if the usage service URL is invalid", func() {
			defaultEnvVars[cmd.UsageServiceURLKey] = "-a:bad-url"
			defaultEnvVars[cmd.CfApiURLKey] = cfService.URL()
			defaultEnvVars[cmd.UsageServiceClientIDKey] = "best-usage-service-client-id"
			defaultEnvVars[cmd.UsageServiceClientSecretKey] = "best-usage-service-client-secret"
			defaultEnvVars[cmd.UsageServiceSkipTlsVerifyKey] = "true"
			command := buildDefaultCommand(defaultEnvVars)
			session, err := gexec.Start(command, GinkgoWriter, GinkgoWriter)
			Expect(err).NotTo(HaveOccurred())
			Eventually(session).Should(gexec.Exit(1))
			Expect(session.Err).To(gbytes.Say(cmd.UsageServiceURLParsingError))
			Expect(session.Err).NotTo(gbytes.Say("USAGE EXAMPLES"))
		})

		It("fails if getting the UAA URL fails", func() {
			cfService.RouteToHandler(http.MethodGet, "/v2/info", func(w http.ResponseWriter, req *http.Request) {
				w.WriteHeader(500)
			})

			defaultEnvVars[cmd.UsageServiceURLKey] = usageService.URL()
			defaultEnvVars[cmd.CfApiURLKey] = cfService.URL()
			defaultEnvVars[cmd.UsageServiceClientIDKey] = "best-usage-service-client-id"
			defaultEnvVars[cmd.UsageServiceClientSecretKey] = "best-usage-service-client-secret"
			defaultEnvVars[cmd.UsageServiceSkipTlsVerifyKey] = "true"
			command := buildDefaultCommand(defaultEnvVars)
			session, err := gexec.Start(command, GinkgoWriter, GinkgoWriter)
			Expect(err).NotTo(HaveOccurred())
			Eventually(session).Should(gexec.Exit(1))
			Expect(session.Err).To(gbytes.Say(cmd.GetUAAURLError))
			Expect(session.Err).To(gbytes.Say(fmt.Sprintf(cf.CFApiUnexpectedResponseStatusErrorFormat, http.StatusInternalServerError)))
		})

		It("fails if the UAA server TLS version is less than 1.2", func() {
			uaaService = serverWithMaxTLSVersion(tls.VersionTLS11)
			cfService = ghttp.NewTLSServer()
			usageService = ghttp.NewTLSServer()
			setupUsageServerServiceHandlers(uaaService, cfService, usageService, "")

			defaultEnvVars[cmd.UsageServiceURLKey] = usageService.URL()
			defaultEnvVars[cmd.CfApiURLKey] = cfService.URL()
			defaultEnvVars[cmd.UsageServiceClientIDKey] = "best-usage-service-client-id"
			defaultEnvVars[cmd.UsageServiceClientSecretKey] = "best-usage-service-client-secret"
			defaultEnvVars[cmd.UsageServiceSkipTlsVerifyKey] = "true"
			command := buildDefaultCommand(defaultEnvVars)
			session, err := gexec.Start(command, GinkgoWriter, GinkgoWriter)
			Expect(err).NotTo(HaveOccurred())
			Eventually(session).Should(gexec.Exit(1))
			Expect(session.Err).To(gbytes.Say("protocol version not supported"))
		})

		It("fails if cf api server TLS version is less than 1.2", func() {
			uaaService = ghttp.NewTLSServer()
			cfService = serverWithMaxTLSVersion(tls.VersionTLS11)
			usageService = ghttp.NewTLSServer()
			setupUsageServerServiceHandlers(uaaService, cfService, usageService, "")

			defaultEnvVars[cmd.UsageServiceURLKey] = usageService.URL()
			defaultEnvVars[cmd.CfApiURLKey] = cfService.URL()
			defaultEnvVars[cmd.UsageServiceClientIDKey] = "best-usage-service-client-id"
			defaultEnvVars[cmd.UsageServiceClientSecretKey] = "best-usage-service-client-secret"
			defaultEnvVars[cmd.UsageServiceSkipTlsVerifyKey] = "true"
			command := buildDefaultCommand(defaultEnvVars)
			session, err := gexec.Start(command, GinkgoWriter, GinkgoWriter)
			Expect(err).NotTo(HaveOccurred())
			Eventually(session).Should(gexec.Exit(1))
			Expect(session.Err).To(gbytes.Say("protocol version not supported"))
		})

		It("fails if the usage service server TLS version is less than 1.2", func() {
			uaaService = ghttp.NewTLSServer()
			cfService = ghttp.NewTLSServer()
			usageService = serverWithMaxTLSVersion(tls.VersionTLS11)

			setupUsageServerServiceHandlers(uaaService, cfService, usageService, "")

			defaultEnvVars[cmd.UsageServiceURLKey] = usageService.URL()
			defaultEnvVars[cmd.CfApiURLKey] = cfService.URL()
			defaultEnvVars[cmd.UsageServiceClientIDKey] = "best-usage-service-client-id"
			defaultEnvVars[cmd.UsageServiceClientSecretKey] = "best-usage-service-client-secret"
			defaultEnvVars[cmd.UsageServiceSkipTlsVerifyKey] = "true"
			command := buildDefaultCommand(defaultEnvVars)
			session, err := gexec.Start(command, GinkgoWriter, GinkgoWriter)
			Expect(err).NotTo(HaveOccurred())
			Eventually(session).Should(gexec.Exit(1))
			Expect(session.Err).To(gbytes.Say("protocol version not supported"))
		})

		DescribeTable(
			"returns an error when one but not all Usage Service required configs are provided",
			func(configName, configValue string) {
				flagValues := map[string]string{
					cmd.OpsManagerTimeoutFlag:        "1",
					cmd.OpsManagerRequestTimeoutFlag: "10",
					cmd.OpsManagerURLFlag:            opsManagerServer.URL(),
					cmd.OpsManagerClientIdFlag:       "whatever",
					cmd.OpsManagerClientSecretFlag:   "whatever",
					cmd.EnvTypeFlag:                  "Development",
					cmd.OutputPathFlag:               outputDirPath,
				}
				flagValues[configName] = configValue
				command := exec.Command(aqueductBinaryPath, "collect")
				for k, v := range flagValues {
					command.Args = append(command.Args, fmt.Sprintf("--%s=%s", k, v))
				}
				session, err := gexec.Start(command, GinkgoWriter, GinkgoWriter)
				Expect(err).NotTo(HaveOccurred())
				Eventually(session).Should(gexec.Exit(1))
				Eventually(session.Err).Should(gbytes.Say(cmd.InvalidUsageConfigurationMessage))
				assertOutputDirEmpty(outputDirPath)
			},
			Entry(cmd.CfApiURLFlag, cmd.CfApiURLFlag, "http://doesnt.matter.com/"),
			Entry(cmd.UsageServiceURLFlag, cmd.UsageServiceURLFlag, "http://also.doesnt.matter.com/"),
			Entry(cmd.UsageServiceClientIDFlag, cmd.UsageServiceClientIDFlag, "best-usage-service-client-id"),
			Entry(cmd.UsageServiceClientSecretFlag, cmd.UsageServiceClientSecretFlag, "best-usage-service-client-secret"),
		)
	})

	Context("when credhub collection is enabled", func() {
		var credhubServer *ghttp.Server

		BeforeEach(func() {
			credhubServer = setupCredHubServer()

			boshCredentialsResponse := func(w http.ResponseWriter, req *http.Request) {
				w.Header().Set("Content-Type", "application/json")
				_, _ = w.Write([]byte(`{ "credential": "BOSH_CLIENT=best_client BOSH_CLIENT_SECRET=best_secret BOSH_CA_CERT=/cool/path BOSH_ENVIRONMENT=127.0.0.1 bosh "}`))
			}
			opsManagerServer.RouteToHandler(http.MethodGet, "/api/v0/deployed/director/credentials/bosh_commandline_credentials", boshCredentialsResponse)
		})

		AfterEach(func() {
			credhubServer.Close()
		})

		It("collects information from credhub as well as ops manager with flag configuration", func() {
			flagValues := map[string]string{
				cmd.OpsManagerURLFlag:          opsManagerServer.URL(),
				cmd.OpsManagerClientIdFlag:     "some-client-id",
				cmd.OpsManagerClientSecretFlag: "some-client-secret",
				cmd.EnvTypeFlag:                "Development",
				cmd.SkipTlsVerifyFlag:          "true",
				cmd.CollectFromCredhubFlag:     "true",
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
			assertValidOutput(tarFilePath, collector_tar.OpsManagerCollectorDataSetId, "p-bosh_certificates", "development")
			assertLogging(session, tarFilePath, true, false)
		})

		It("collects information from credhub as well as ops manager with env variable configuration", func() {
			defaultEnvVars[cmd.OpsManagerURLKey] = opsManagerServer.URL()
			defaultEnvVars[cmd.WithCredhubInfoKey] = "true"
			command := buildDefaultCommand(defaultEnvVars)
			session, err := gexec.Start(command, GinkgoWriter, GinkgoWriter)
			Expect(err).NotTo(HaveOccurred())
			Eventually(session).Should(gexec.Exit(0))
			tarFilePath := validatedTarFilePath(outputDirPath)
			assertValidOutput(tarFilePath, collector_tar.OpsManagerCollectorDataSetId, "p-bosh_certificates", "development")
			assertLogging(session, tarFilePath, true, false)
		})

		It("collects information from credhub as well as ops manager with config file configuration", func() {
			config := fmt.Sprintf(`{
				"url": "%s",
				"username": "some-username",
				"password": "some-password",
				"env-type": "Development",
				"insecure-skip-tls-verify": "true",
				"output-dir": "%s",
				"with-credhub-info": "true"
			}`, opsManagerServer.URL(), escapeFilePathForWindows(outputDirPath))
			configFile := filepath.Join(configDirPath, "config.yml")
			err := os.WriteFile(configFile, []byte(config), 0755)
			Expect(err).ToNot(HaveOccurred())

			command := exec.Command(aqueductBinaryPath, "collect", "--config", configFile)

			session, err := gexec.Start(command, GinkgoWriter, GinkgoWriter)
			Expect(err).NotTo(HaveOccurred())
			Eventually(session).Should(gexec.Exit(0))
			tarFilePath := validatedTarFilePath(outputDirPath)
			assertValidOutput(tarFilePath, collector_tar.OpsManagerCollectorDataSetId, "p-bosh_certificates", "development")
			assertLogging(session, tarFilePath, true, false)
		})

		It("errors if fetching credentials for credhub auth fails", func() {
			opsManagerServer.RouteToHandler(http.MethodGet, "/api/v0/deployed/director/credentials/bosh_commandline_credentials", func(w http.ResponseWriter, req *http.Request) {
				w.WriteHeader(500)
			})
			flagValues := map[string]string{
				cmd.OpsManagerURLFlag:          opsManagerServer.URL(),
				cmd.OpsManagerClientIdFlag:     "some-client-id",
				cmd.OpsManagerClientSecretFlag: "some-client-secret",
				cmd.EnvTypeFlag:                "Development",
				cmd.SkipTlsVerifyFlag:          "true",
				cmd.CollectFromCredhubFlag:     "true",
				cmd.OutputPathFlag:             outputDirPath,
			}
			command := exec.Command(aqueductBinaryPath, "collect")
			for k, v := range flagValues {
				command.Args = append(command.Args, fmt.Sprintf("--%s=%s", k, v))
			}
			session, err := gexec.Start(command, GinkgoWriter, GinkgoWriter)
			Expect(err).NotTo(HaveOccurred())
			Eventually(session).Should(gexec.Exit(1))
			Expect(session.Err).To(gbytes.Say("unexpected status"))
			Expect(session.Err).NotTo(gbytes.Say("USAGE EXAMPLES"))
			assertOutputDirEmpty(outputDirPath)
		})

		It("errors if creating the credhub client fails", func() {
			credhubServer.RouteToHandler(http.MethodGet, "/info", func(w http.ResponseWriter, req *http.Request) {
				w.WriteHeader(500)
			})
			flagValues := map[string]string{
				cmd.OpsManagerURLFlag:          opsManagerServer.URL(),
				cmd.OpsManagerClientIdFlag:     "some-client-id",
				cmd.OpsManagerClientSecretFlag: "some-client-secret",
				cmd.EnvTypeFlag:                "Development",
				cmd.SkipTlsVerifyFlag:          "true",
				cmd.CollectFromCredhubFlag:     "true",
				cmd.OutputPathFlag:             outputDirPath,
			}
			command := exec.Command(aqueductBinaryPath, "collect")
			for k, v := range flagValues {
				command.Args = append(command.Args, fmt.Sprintf("--%s=%s", k, v))
			}
			session, err := gexec.Start(command, GinkgoWriter, GinkgoWriter)
			Expect(err).NotTo(HaveOccurred())
			Eventually(session).Should(gexec.Exit(1))
			Expect(session.Err).To(gbytes.Say(cmd.CredhubClientError))
			Expect(session.Err).NotTo(gbytes.Say("USAGE EXAMPLES"))
			assertOutputDirEmpty(outputDirPath)
		})
	})

	Context("when an https_proxy is set", func() {
		var (
			usageService  *ghttp.Server
			cfService     *ghttp.Server
			uaaService    *ghttp.Server
			credhubServer *ghttp.Server
			proxyServer   *http.Server
			listenerPort  int
		)

		BeforeEach(func() {
			uaaService, cfService, usageService = setupUsageService("https://uaa.example.com")
			uaaService.RouteToHandler(http.MethodPost, "/oauth/token", func(w http.ResponseWriter, req *http.Request) {
				w.Header().Set("Content-Type", "application/json")
				credentialBytes := []byte("best-usage-service-client-id:best-usage-service-client-secret")

				base64credentials := base64.StdEncoding.EncodeToString(credentialBytes)
				Expect(req.Header.Get("authorization")).To(Equal("Basic " + base64credentials))

				_, _ = w.Write([]byte(`{
					"access_token": "some-uaa-token",
					"token_type": "bearer",
					"expires_in": 3600
				}`))
			})

			credhubServer = setupCredHubServer()
			boshCredentialsResponse := func(w http.ResponseWriter, req *http.Request) {
				w.Header().Set("Content-Type", "application/json")
				_, _ = w.Write([]byte(`{ "credential": "BOSH_CLIENT=best_client BOSH_CLIENT_SECRET=best_secret BOSH_CA_CERT=/cool/path BOSH_ENVIRONMENT=credhub.example.com bosh "}`))
			}
			opsManagerServer.RouteToHandler(http.MethodGet, "/api/v0/deployed/director/credentials/bosh_commandline_credentials", boshCredentialsResponse)

			proxy := goproxy.NewProxyHttpServer()
			proxy.OnRequest().HandleConnect(goproxy.AlwaysMitm)
			proxy.OnRequest().DoFunc(
				func(r *http.Request, ctx *goproxy.ProxyCtx) (*http.Request, *http.Response) {
					var (
						err         error
						testServUrl *url.URL
					)
					switch r.Host {
					case "uaa.example.com":
						testServUrl, err = url.Parse(uaaService.URL())
					case "opsman.example.com":
						testServUrl, err = url.Parse(opsManagerServer.URL())
					case "cfapi.example.com":
						testServUrl, err = url.Parse(cfService.URL())
					case "usageService.example.com":
						testServUrl, err = url.Parse(usageService.URL())
					case "credhub.example.com:8844":
						testServUrl, err = url.Parse(credhubServer.URL())
					default:
						log.Fatalf("Unexpected host: %s", r.Host)
					}
					Expect(err).NotTo(HaveOccurred())
					testServUrl.Path = r.URL.Path
					r.URL = testServUrl
					return r, nil
				})

			listener, err := net.Listen("tcp", ":0")
			Expect(err).ToNot(HaveOccurred())
			listenerPort = listener.Addr().(*net.TCPAddr).Port
			proxyServer = &http.Server{Handler: proxy}
			go func() {
				_ = proxyServer.Serve(listener)
			}()
		})

		AfterEach(func() {
			proxyServer.Close()
			usageService.Close()
			cfService.Close()
			uaaService.Close()
			credhubServer.Close()
		})

		It("proxies traffic for all remote http services (ops manager, credhub, usage service, uaa, etc)", func() {
			defaultEnvVars[cmd.OpsManagerURLKey] = "https://opsman.example.com"
			defaultEnvVars[cmd.CfApiURLKey] = "https://cfapi.example.com"
			defaultEnvVars[cmd.UsageServiceURLKey] = "https://usageService.example.com"
			defaultEnvVars[cmd.UsageServiceClientIDKey] = "best-usage-service-client-id"
			defaultEnvVars[cmd.UsageServiceClientSecretKey] = "best-usage-service-client-secret"
			defaultEnvVars[cmd.UsageServiceSkipTlsVerifyKey] = "true"
			defaultEnvVars[cmd.WithCredhubInfoKey] = "true"

			defaultEnvVars["HTTPS_PROXY"] = fmt.Sprintf("http://localhost:%d", listenerPort)

			command := buildDefaultCommand(defaultEnvVars)
			session, err := gexec.Start(command, GinkgoWriter, GinkgoWriter)
			Expect(err).NotTo(HaveOccurred())
			Eventually(session).Should(gexec.Exit(0))

			tarFilePath := validatedTarFilePath(outputDirPath)
			assertValidOutput(tarFilePath, collector_tar.OpsManagerCollectorDataSetId, "ops_manager_vm_types", "development")
			assertLogging(session, tarFilePath, false, false)
		})
	})

	It("fails if the required variables are not set", func() {
		command := exec.Command(aqueductBinaryPath, "collect")
		session, err := gexec.Start(command, GinkgoWriter, GinkgoWriter)
		Expect(err).ToNot(HaveOccurred())
		Eventually(session).Should(gexec.Exit(1))
		requiredFlags := []string{"--" + cmd.OpsManagerURLFlag, "--" + cmd.EnvTypeFlag, "--" + cmd.OutputPathFlag}
		Expect(session.Err).To(gbytes.Say(fmt.Sprintf(cmd.RequiredConfigErrorFormat, strings.Join(requiredFlags, ", "))))
		Expect(session.Err).To(gbytes.Say("USAGE EXAMPLES"))
		assertOutputDirEmpty(outputDirPath)
	})

	It("fails if the passed in env type is invalid", func() {
		command := buildDefaultCommand(defaultEnvVars)
		command.Env = append(command.Env, fmt.Sprintf("%s=%s", cmd.EnvTypeKey, "invalid-type"))
		session, err := gexec.Start(command, GinkgoWriter, GinkgoWriter)
		Expect(err).NotTo(HaveOccurred())
		Eventually(session).Should(gexec.Exit(1))
		Expect(session.Err).To(gbytes.Say(fmt.Sprintf(cmd.InvalidEnvTypeFailureFormat, "invalid-type")))
		Expect(session.Err).To(gbytes.Say("USAGE EXAMPLES"))
		assertOutputDirEmpty(outputDirPath)
	})

	It("fails if data collection from Operations Manager fails", func() {
		failingServer := ghttp.NewServer()
		failingServer.RouteToHandler(http.MethodPost, "/uaa/oauth/token", func(w http.ResponseWriter, req *http.Request) {
			w.WriteHeader(500)
		})
		defer failingServer.Close()

		defaultEnvVars[cmd.OpsManagerURLKey] = failingServer.URL()
		command := buildDefaultCommand(defaultEnvVars)
		session, err := gexec.Start(command, GinkgoWriter, GinkgoWriter)
		Expect(err).NotTo(HaveOccurred())
		Eventually(session).Should(gexec.Exit(1))
		Expect(session.Err).To(gbytes.Say(operations.OpsManagerCollectFailureMessage))
		Expect(session.Err).NotTo(gbytes.Say("USAGE EXAMPLES"))
		assertOutputDirEmpty(outputDirPath)
	})

	It("fails if the output directory does not exist", func() {
		defaultEnvVars[cmd.OutputPathKey] = "/not/a/real/path"
		command := buildDefaultCommand(defaultEnvVars)
		session, err := gexec.Start(command, GinkgoWriter, GinkgoWriter)
		Expect(err).NotTo(HaveOccurred())
		Eventually(session).Should(gexec.Exit(1))
		Expect(session.Err).To(gbytes.Say(fmt.Sprintf(cmd.CreateTarFileFailureFormat, "")))
		Expect(session.Err).NotTo(gbytes.Say("USAGE EXAMPLES"))
	})

	It("fails if the Ops Manager server TLS version is less than 1.2", func() {
		opsManagerServer := serverWithMaxTLSVersion(tls.VersionTLS11)
		opsManagerServer.RouteToHandler(http.MethodPost, "/uaa/oauth/token", func(w http.ResponseWriter, req *http.Request) {
			w.Header().Set("Content-Type", "application/json")
			_, _ = w.Write([]byte(`{}`))
		})
		defaultEnvVars[cmd.OpsManagerURLKey] = opsManagerServer.URL()
		command := buildDefaultCommand(defaultEnvVars)
		session, err := gexec.Start(command, GinkgoWriter, GinkgoWriter)
		Expect(err).NotTo(HaveOccurred())
		Eventually(session).Should(gexec.Exit(1))
		Expect(session.Err).To(gbytes.Say("protocol version not supported"))
	})

	It("exits with success status when there are ops manager pending changes", func() {
		// This test exists because we used to error with a separate error code for pending changes
		opsManagerServer.RouteToHandler(http.MethodGet, "/api/v0/staged/pending_changes", func(w http.ResponseWriter, r *http.Request) {
			resp := api.PendingChangesOutput{
				ChangeList: []api.ProductChange{
					{
						Action: "not-unchanged",
					},
				},
			}
			Expect(json.NewEncoder(w).Encode(&resp)).NotTo(HaveOccurred())
		})

		command := buildDefaultCommand(defaultEnvVars)
		session, err := gexec.Start(command, GinkgoWriter, GinkgoWriter)
		Expect(err).NotTo(HaveOccurred())
		Eventually(session).Should(gexec.Exit(0))
	})
})

func validatedTarFilePath(outputDirPath string) string {
	fileInfos, err := os.ReadDir(outputDirPath)
	Expect(err).NotTo(HaveOccurred())
	Expect(len(fileInfos)).To(Equal(1), fmt.Sprintf("Expected output dir %s to include a single file", outputDirPath))
	Expect(fileInfos[0].Name()).To(MatchRegexp(fmt.Sprintf(`%s%s.tar$`, cmd.OutputFilePrefix, UnixTimestampRegexp)))
	return filepath.Join(outputDirPath, fileInfos[0].Name())
}

func assertMetadataFileIsCorrect(contentDir, expectedEnvType, dataSetType string) {
	content, err := os.ReadFile(filepath.Join(contentDir, dataSetType, collector_tar.MetadataFileName))
	Expect(err).NotTo(HaveOccurred(), "Expected metadata file to exist but did not")
	var metadata collector_tar.Metadata
	Expect(json.Unmarshal(content, &metadata)).To(Succeed())
	Expect(metadata.EnvType).To(Equal(expectedEnvType))
	Expect(metadata.CollectorVersion).To(Equal(testVersion))
}

func assertOutputDirEmpty(outputDirPath string) {
	fileInfos, err := os.ReadDir(outputDirPath)
	Expect(err).NotTo(HaveOccurred())
	Expect(len(fileInfos)).To(Equal(0), fmt.Sprintf("Expected output dir %s to be empty", outputDirPath))
}

func assertValidOutput(tarFilePath, dataSetType, filename, envType string) {
	tmpDir, err := os.MkdirTemp("", "")
	Expect(err).NotTo(HaveOccurred())
	defer os.RemoveAll(tmpDir)

	tar := &archiver.Tar{}
	err = tar.Unarchive(tarFilePath, tmpDir)
	Expect(err).NotTo(HaveOccurred())

	jsonFilePath := filepath.Join(tmpDir, dataSetType, filename)

	Expect(jsonFilePath).To(BeAnExistingFile())
	content, err := os.ReadFile(jsonFilePath)
	Expect(err).NotTo(HaveOccurred())
	Expect(json.Valid(content)).To(BeTrue(), fmt.Sprintf("Expected file %s to contain valid json", jsonFilePath))
	assertMetadataFileIsCorrect(tmpDir, envType, dataSetType)
}

func assertValidNickname(tarFilePath, dataSetType, filename, foundationNickname string) {
	tmpDir, err := os.MkdirTemp("", "")
	Expect(err).NotTo(HaveOccurred())
	defer os.RemoveAll(tmpDir)

	tar := &archiver.Tar{}
	err = tar.Unarchive(tarFilePath, tmpDir)
	Expect(err).NotTo(HaveOccurred())

	jsonFilePath := filepath.Join(tmpDir, dataSetType, filename)

	Expect(jsonFilePath).To(BeAnExistingFile())
	content, err := os.ReadFile(jsonFilePath)
	Expect(err).NotTo(HaveOccurred())
	Expect(json.Valid(content)).To(BeTrue(), fmt.Sprintf("Expected file %s to contain valid json", jsonFilePath))

	content, err = os.ReadFile(filepath.Join(tmpDir, dataSetType, collector_tar.MetadataFileName))
	Expect(err).NotTo(HaveOccurred(), "Expected metadata file to exist but did not")
	var metadata collector_tar.Metadata
	Expect(json.Unmarshal(content, &metadata)).To(Succeed())
	Expect(metadata.FoundationNickname).To(Equal(foundationNickname))
}

func assertLogging(session *gexec.Session, tarFilePath string, credHubEnabled, usageServiceEnabled bool) {
	Expect(session.Out).To(gbytes.Say("Collecting data from Operations Manager"))
	if credHubEnabled {
		Expect(session.Out).To(gbytes.Say("Collecting data from CredHub"))
	}
	if usageServiceEnabled {
		Expect(session.Out).To(gbytes.Say("Collecting data from Usage Service"))
	}
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

func setupOpsManagerServer() *ghttp.Server {
	opsManagerServer := ghttp.NewTLSServer()
	opsManagerServer.RouteToHandler(http.MethodPost, "/uaa/oauth/token", func(w http.ResponseWriter, req *http.Request) {
		w.Header().Set("Content-Type", "application/json")

		_, _ = w.Write([]byte(`{
					"access_token": "some-opsman-token",
					"token_type": "bearer",
					"expires_in": 3600
					}`))
	})
	emptyObjectResponse := func(w http.ResponseWriter, req *http.Request) {
		w.Header().Set("Content-Type", "application/json")
		_, _ = w.Write([]byte(`{}`))
	}
	emptyArrayResponse := func(w http.ResponseWriter, req *http.Request) {
		w.Header().Set("Content-Type", "application/json")
		_, _ = w.Write([]byte(`[]`))
	}
	emptyCSVResponse := func(w http.ResponseWriter, req *http.Request) {
		w.Header().Set("Content-Type", "text/csv")
		_, _ = w.Write([]byte(``))
	}

	opsManagerServer.RouteToHandler(http.MethodGet, "/api/v0/staged/pending_changes", emptyObjectResponse)
	opsManagerServer.RouteToHandler(http.MethodGet, "/api/v0/deployed/products", emptyArrayResponse)
	opsManagerServer.RouteToHandler(http.MethodGet, "/api/v0/vm_types", emptyArrayResponse)
	opsManagerServer.RouteToHandler(http.MethodGet, "/api/v0/diagnostic_report", emptyObjectResponse)
	opsManagerServer.RouteToHandler(http.MethodGet, "/api/v0/installations", emptyObjectResponse)
	opsManagerServer.RouteToHandler(http.MethodGet, "/api/v0/deployed/certificates", emptyObjectResponse)
	opsManagerServer.RouteToHandler(http.MethodGet, "/api/v0/certificate_authorities", emptyObjectResponse)
	opsManagerServer.RouteToHandler(http.MethodGet, "/api/v0/download_core_consumption", emptyCSVResponse)

	return opsManagerServer
}

func setupCredHubServer() *ghttp.Server {
	credhubServer := ghttp.NewUnstartedServer()

	listener, err := net.Listen("tcp", "127.0.0.1:8844")
	Expect(err).NotTo(HaveOccurred())
	credhubServer.HTTPTestServer.Listener = listener
	credhubServer.HTTPTestServer.StartTLS()
	credhubServer.RouteToHandler(http.MethodGet, "/info", func(w http.ResponseWriter, req *http.Request) {
		w.Header().Set("Content-Type", "application/json")
		_, _ = w.Write([]byte(`{ "auth-server": {"url": "https://127.0.0.1:8844"}}`))
	})

	credhubServer.RouteToHandler(http.MethodPost, "/oauth/token", func(w http.ResponseWriter, req *http.Request) {
		w.Header().Set("Content-Type", "application/json")
		_, _ = w.Write([]byte(`{
					"access_token": "some-credhub-token",
					"token_type": "bearer",
					"expires_in": 3600
					}`))
	})
	credhubServer.RouteToHandler(http.MethodGet, "/api/v1/certificates", func(w http.ResponseWriter, req *http.Request) {
		w.Header().Set("Content-Type", "application/json")
		_, _ = w.Write([]byte(`{}`))
	})

	return credhubServer
}

func setupUsageService(uaaServiceURLOverride string) (uaaService, cfService, usageService *ghttp.Server) {
	uaaService = ghttp.NewTLSServer()
	cfService = ghttp.NewTLSServer()
	usageService = ghttp.NewTLSServer()

	setupUsageServerServiceHandlers(uaaService, cfService, usageService, uaaServiceURLOverride)
	return
}

func setupUsageServerServiceHandlers(uaaService, cfService, usageService *ghttp.Server, uaaServiceURLOverride string) {
	uaaService.RouteToHandler(http.MethodPost, "/oauth/token", func(w http.ResponseWriter, req *http.Request) {
		w.Header().Set("Content-Type", "application/json")
		credentialBytes := []byte("best-usage-service-client-id:best-usage-service-client-secret")

		base64credentials := base64.StdEncoding.EncodeToString(credentialBytes)
		Expect(req.Header.Get("authorization")).To(Equal("Basic " + base64credentials))

		_, _ = w.Write([]byte(`{
					"access_token": "some-uaa-token",
					"token_type": "bearer",
					"expires_in": 3600
				}`))
	})

	cfService.RouteToHandler(http.MethodGet, "/v2/info", func(w http.ResponseWriter, req *http.Request) {
		w.Header().Set("Content-Type", "application/json")
		if uaaServiceURLOverride != "" {
			_, _ = w.Write([]byte(`{ "token_endpoint": "` + uaaServiceURLOverride + `" }`))
		} else {
			_, _ = w.Write([]byte(`{ "token_endpoint": "` + uaaService.URL() + `" }`))
		}
	})

	usageService.RouteToHandler(http.MethodGet, "/system_report/app_usages", func(w http.ResponseWriter, req *http.Request) {
		w.Header().Set("Content-Type", "application/json")
		_, _ = w.Write([]byte(`{}`))
		Expect(req.Header.Get("Authorization")).To(Equal("Bearer some-uaa-token"))
	})
	usageService.RouteToHandler(http.MethodGet, "/system_report/service_usages", func(w http.ResponseWriter, req *http.Request) {
		w.Header().Set("Content-Type", "application/json")
		_, _ = w.Write([]byte(`{}`))
		Expect(req.Header.Get("Authorization")).To(Equal("Bearer some-uaa-token"))
	})
	usageService.RouteToHandler(http.MethodGet, "/system_report/task_usages", func(w http.ResponseWriter, req *http.Request) {
		w.Header().Set("Content-Type", "application/json")
		_, _ = w.Write([]byte(`{}`))
		Expect(req.Header.Get("Authorization")).To(Equal("Bearer some-uaa-token"))
	})
}

func serverWithMaxTLSVersion(version uint16) *ghttp.Server {
	server := ghttp.NewUnstartedServer()
	listener, err := listenOnFreePort()
	Expect(err).NotTo(HaveOccurred())
	server.HTTPTestServer.Listener = listener
	server.HTTPTestServer.TLS = &tls.Config{MaxVersion: version}
	server.HTTPTestServer.Config.ErrorLog = log.New(GinkgoWriter, "", 0)
	server.HTTPTestServer.StartTLS()
	return server
}

func listenOnFreePort() (net.Listener, error) {
	listener, err := net.Listen("tcp", "127.0.0.1:")
	if err != nil {
		return nil, err
	}
	return listener, nil
}

func escapeFilePathForWindows(path string) string {
	return strings.Replace(path, `\`, `\\`, -1)
}
