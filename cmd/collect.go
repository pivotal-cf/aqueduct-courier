package cmd

import (
	"fmt"
	"net/url"
	"os"
	"path/filepath"
	"strings"
	"time"

	"github.com/pivotal-cf/om/api"

	"github.com/gofrs/uuid"

	"github.com/pivotal-cf/aqueduct-courier/network"

	"github.com/pivotal-cf/aqueduct-courier/consumption"

	ogCredhub "code.cloudfoundry.org/credhub-cli/credhub"
	"code.cloudfoundry.org/credhub-cli/credhub/auth"
	"github.com/pivotal-cf/aqueduct-courier/cf"
	"github.com/pivotal-cf/aqueduct-courier/credhub"

	"github.com/pivotal-cf/aqueduct-courier/operations"
	"github.com/pivotal-cf/aqueduct-courier/opsmanager"
	"github.com/pivotal-cf/aqueduct-utils/file"
	omNetwork "github.com/pivotal-cf/om/network"
	"github.com/pkg/errors"
	"github.com/spf13/cobra"
	"github.com/spf13/viper"
)

const (
	OpsManagerURLKey             = "OPS_MANAGER_URL"
	OpsManagerUsernameKey        = "OPS_MANAGER_USERNAME"
	OpsManagerPasswordKey        = "OPS_MANAGER_PASSWORD"
	OpsManagerClientIdKey        = "OPS_MANAGER_CLIENT_ID"
	OpsManagerClientSecretKey    = "OPS_MANAGER_CLIENT_SECRET"
	OpsManagerTimeoutKey         = "OPS_MANAGER_TIMEOUT"
	EnvTypeKey                   = "ENV_TYPE"
	OutputPathKey                = "OUTPUT_DIR"
	SkipTlsVerifyKey             = "INSECURE_SKIP_TLS_VERIFY"
	WithCredhubInfoKey           = "WITH_CREDHUB_INFO"
	UsageServiceURLKey           = "USAGE_SERVICE_URL"
	UsageServiceClientIDKey      = "USAGE_SERVICE_CLIENT_ID"
	UsageServiceClientSecretKey  = "USAGE_SERVICE_CLIENT_SECRET"
	CfApiURLKey                  = "CF_API_URL"
	UsageServiceSkipTlsVerifyKey = "USAGE_SERVICE_INSECURE_SKIP_TLS_VERIFY"

	OpsManagerURLFlag             = "url"
	OpsManagerUsernameFlag        = "username"
	OpsManagerPasswordFlag        = "password"
	OpsManagerClientIdFlag        = "client-id"
	OpsManagerClientSecretFlag    = "client-secret"
	OpsManagerTimeoutFlag         = "ops-manager-timeout"
	CollectFromCredhubFlag        = "with-credhub-info"
	EnvTypeFlag                   = "env-type"
	OutputPathFlag                = "output-dir"
	SkipTlsVerifyFlag             = "insecure-skip-tls-verify"
	UsageServiceURLFlag           = "usage-service-url"
	UsageServiceClientIDFlag      = "usage-service-client-id"
	UsageServiceClientSecretFlag  = "usage-service-client-secret"
	CfApiURLFlag                  = "cf-api-url"
	UsageServiceSkipTlsVerifyFlag = "usage-service-insecure-skip-tls-verify"

	EnvTypeSandbox       = "sandbox"
	EnvTypeDevelopment   = "development"
	EnvTypeQA            = "qa"
	EnvTypePreProduction = "pre-production"
	EnvTypeProduction    = "production"

	OutputFilePrefix                 = "FoundationDetails_"
	CredhubClientError               = "Failed creating credhub client"
	InvalidEnvTypeFailureFormat      = "Invalid env-type %s. See help for the list of valid types."
	InvalidAuthConfigurationMessage  = "Invalid auth configuration. Requires username/password or client/secret to be set."
	InvalidUsageConfigurationMessage = "Not all usage service configurations provided."
	CreateTarFileFailureFormat       = "Could not create tar file %s"
	UsageServiceURLParsingError      = "error parsing Usage Service URL"
	GetUAAURLError                   = "error getting UAA URL"
)

var collectCmd = &cobra.Command{
	Use:   "collect",
	Short: "Collects information from a PCF foundation",
	Long:  "Collects information from Operations Manager and outputs the content to the configured directory.\nOptionally collects information from Usage Service and/or Credhub.",
	RunE:  collect,
}

func init() {
	bindFlagAndEnvVar(collectCmd, OpsManagerURLFlag, "", fmt.Sprintf("``Ops Manager URL [$%s]", OpsManagerURLKey), OpsManagerURLKey)
	bindFlagAndEnvVar(collectCmd, OpsManagerUsernameFlag, "", fmt.Sprintf("``Ops Manager username [$%s]", OpsManagerUsernameKey), OpsManagerUsernameKey)
	bindFlagAndEnvVar(collectCmd, OpsManagerPasswordFlag, "", fmt.Sprintf("``Ops Manager password [$%s]", OpsManagerPasswordKey), OpsManagerPasswordKey)
	bindFlagAndEnvVar(collectCmd, OpsManagerClientIdFlag, "", fmt.Sprintf("``Ops Manager client id [$%s]", OpsManagerClientIdKey), OpsManagerClientIdKey)
	bindFlagAndEnvVar(collectCmd, OpsManagerClientSecretFlag, "", fmt.Sprintf("``Ops Manager client secret [$%s]", OpsManagerClientSecretKey), OpsManagerClientSecretKey)
	bindFlagAndEnvVar(collectCmd, EnvTypeFlag, "", fmt.Sprintf("``Specify environment type (sandbox, development, qa, pre-production, production) [$%s]", EnvTypeKey), EnvTypeKey)
	bindFlagAndEnvVar(collectCmd, OpsManagerTimeoutFlag, 30, fmt.Sprintf("Ops Manager http request timeout in seconds [$%s]", OpsManagerTimeoutKey), OpsManagerTimeoutKey)
	bindFlagAndEnvVar(collectCmd, SkipTlsVerifyFlag, false, fmt.Sprintf("``Skip TLS validation on http requests to Ops Manager [$%s]\n", SkipTlsVerifyKey), SkipTlsVerifyKey)

	bindFlagAndEnvVar(collectCmd, CfApiURLFlag, "", fmt.Sprintf("``CF API URL for UAA authentication to access Usage Service [$%s]", CfApiURLKey), CfApiURLKey)
	bindFlagAndEnvVar(collectCmd, UsageServiceURLFlag, "", fmt.Sprintf("``Usage Service URL [$%s]", UsageServiceURLKey), UsageServiceURLKey)
	bindFlagAndEnvVar(collectCmd, UsageServiceClientIDFlag, "", fmt.Sprintf("``Usage Service client id [$%s]", UsageServiceClientIDKey), UsageServiceClientIDKey)
	bindFlagAndEnvVar(collectCmd, UsageServiceClientSecretFlag, "", fmt.Sprintf("``Usage Service client secret [$%s]", UsageServiceClientSecretKey), UsageServiceClientSecretKey)
	bindFlagAndEnvVar(collectCmd, UsageServiceSkipTlsVerifyFlag, false, fmt.Sprintf("``Skip TLS validation for Usage Service components [$%s]\n", UsageServiceSkipTlsVerifyKey), UsageServiceSkipTlsVerifyKey)

	bindFlagAndEnvVar(collectCmd, CollectFromCredhubFlag, false, fmt.Sprintf("Include CredHub certificate expiry information [$%s]\n", WithCredhubInfoKey), WithCredhubInfoKey)
	bindFlagAndEnvVar(collectCmd, OutputPathFlag, "", fmt.Sprintf("``Local directory to write data [$%s]\n", OutputPathKey), OutputPathKey)

	collectCmd.Flags().BoolP("help", "h", false, "Help for the collect command\n")

	collectCmd.Flags().SortFlags = false

	collectCmd.Example = `
      Collect data from Ops Manager only:
      telemetry-collector collect --url --username --password [or --client-id and
      --client-secret] --env-type --output-dir

      Collect data from Ops Manager and Usage Service:
      telemetry-collector collect --url --username --password [or --client-id and
      --client-secret] --usage-service-url --usage-service-client-id
      --usage-service-client-secret --cf-api-url --env-type --output-dir`

	customHelpTextTemplate := `
Collects information from a single Ops Manager (and optionally from
Usage Service and/or Credhub) and outputs the content to the configured directory.

USAGE EXAMPLES
{{.Example}}

FLAGS

{{.LocalFlags.FlagUsages}}`
	collectCmd.SetHelpTemplate(customHelpTextTemplate)
	rootCmd.AddCommand(collectCmd)
}

func collect(c *cobra.Command, _ []string) error {
	if err := verifyRequiredConfig(OpsManagerURLFlag, EnvTypeFlag, OutputPathFlag); err != nil {
		return err
	}
	if err := validateCredConfig(); err != nil {
		return err
	}
	envType, err := validateAndNormalizeEnvType()
	if err != nil {
		return err
	}

	c.SilenceUsage = true

	tarFilePath := filepath.Join(
		viper.GetString(OutputPathFlag),
		fmt.Sprintf("%s%d.tar", OutputFilePrefix, time.Now().UTC().Unix()),
	)
	tarFile, err := os.Create(tarFilePath)
	if err != nil {
		return errors.Wrapf(err, CreateTarFileFailureFormat, tarFilePath)
	}
	defer tarFile.Close()

	tarWriter := file.NewTarWriter(tarFile)

	collectExecutor, err := makeCollector(tarWriter)
	if err != nil {
		tarFile.Close()
		os.Remove(tarFilePath)
		return err
	}

	err = collectExecutor.Collect(envType, version)
	if err != nil {
		tarFile.Close()
		os.Remove(tarFilePath)
		return err
	}

	logger.Printf("Wrote output to %s\n", tarFilePath)
	logger.Println("Success!")
	return nil
}

func anyUsageServiceConfigsProvided() bool {
	return viper.GetString(CfApiURLFlag) != "" ||
		viper.GetString(UsageServiceURLFlag) != "" ||
		viper.GetString(UsageServiceClientIDFlag) != "" ||
		viper.GetString(UsageServiceClientSecretFlag) != ""
}

func validateUsageServiceConfig() error {
	if viper.GetString(CfApiURLFlag) == "" ||
		viper.GetString(UsageServiceURLFlag) == "" ||
		viper.GetString(UsageServiceClientIDFlag) == "" ||
		viper.GetString(UsageServiceClientSecretFlag) == "" {

		return errors.New(InvalidUsageConfigurationMessage)
	}
	return nil
}

func validateCredConfig() error {
	noUsernamePasswordAuth := viper.GetString(OpsManagerUsernameFlag) == "" || viper.GetString(OpsManagerPasswordFlag) == ""
	noClientSecretAuth := viper.GetString(OpsManagerClientIdFlag) == "" || viper.GetString(OpsManagerClientSecretFlag) == ""
	if noUsernamePasswordAuth && noClientSecretAuth {
		return errors.New(InvalidAuthConfigurationMessage)
	}

	return nil
}

func validateAndNormalizeEnvType() (string, error) {
	validEnvTypes := []string{EnvTypeSandbox, EnvTypeDevelopment, EnvTypeQA, EnvTypePreProduction, EnvTypeProduction}
	envType := strings.ToLower(viper.GetString(EnvTypeFlag))
	for _, validType := range validEnvTypes {
		if validType == envType {
			return envType, nil
		}
	}
	return "", errors.Errorf(InvalidEnvTypeFailureFormat, envType)
}

type consumptionDataCollector interface {
	Collect() ([]consumption.Data, error)
}

func makeConsumptionCollector() (consumptionDataCollector, error) {
	if anyUsageServiceConfigsProvided() {
		err := validateUsageServiceConfig()
		if err != nil {
			return nil, err
		}

		client := network.NewClient(viper.GetBool(UsageServiceSkipTlsVerifyFlag))
		cfApiClient := cf.NewClient(viper.GetString(CfApiURLFlag), client)

		usageURL, err := url.Parse(viper.GetString(UsageServiceURLFlag))
		if err != nil {
			return nil, errors.New(UsageServiceURLParsingError)
		}

		uaaURL, err := cfApiClient.GetUAAURL()
		if err != nil {
			return nil, errors.New(GetUAAURLError)
		}

		authedClient := cf.NewOAuthClient(
			uaaURL,
			viper.GetString(UsageServiceClientIDFlag),
			viper.GetString(UsageServiceClientSecretFlag),
			5*time.Second,
			client,
		)

		consumptionService := &consumption.Service{
			BaseURL: usageURL,
			Client:  authedClient,
		}

		consumptionCollector := consumption.NewDataCollector(
			*logger,
			consumptionService,
			viper.GetString(UsageServiceURLFlag),
		)

		return consumptionCollector, nil
	}
	return nil, nil
}

type credhubDataCollector interface {
	Collect() (credhub.Data, error)
}

func makeCredhubCollector(omService *opsmanager.Service, credhubCollectionEnabled bool) (credhubDataCollector, error) {
	if credhubCollectionEnabled {
		chCreds, err := omService.BoshCredentials()
		if err != nil {
			return nil, err
		}
		credHubURL := "https://" + chCreds.Host + ":8844"
		requestor, err := ogCredhub.New(
			credHubURL,
			ogCredhub.SkipTLSValidation(true),
			ogCredhub.Auth(auth.UaaClientCredentials(chCreds.ClientID, chCreds.ClientSecret)),
		)
		if err != nil {
			return nil, errors.Wrap(err, CredhubClientError)
		}
		credhubService := credhub.NewCredhubService(requestor)
		return credhub.NewDataCollector(*logger, credhubService, credHubURL), nil
	} else {
		return nil, nil
	}
}

func makeCollector(tarWriter *file.TarWriter) (*operations.CollectExecutor, error) {
	authedClient, _ := omNetwork.NewOAuthClient(
		viper.GetString(OpsManagerURLFlag),
		viper.GetString(OpsManagerUsernameFlag),
		viper.GetString(OpsManagerPasswordFlag),
		viper.GetString(OpsManagerClientIdFlag),
		viper.GetString(OpsManagerClientSecretFlag),
		viper.GetBool(SkipTlsVerifyFlag),
		false,
		time.Duration(viper.GetInt(OpsManagerTimeoutFlag))*time.Second,
		5*time.Second,
	)

	apiService := api.New(api.ApiInput{Client: authedClient})
	omService := &opsmanager.Service{
		Requestor: apiService,
	}

	omCollector := opsmanager.NewDataCollector(
		*logger,
		omService,
		viper.GetString(OpsManagerURLFlag),
		apiService,
		apiService,
	)

	consumptionCollector, err := makeConsumptionCollector()
	if err != nil {
		return nil, err
	}

	credhubCollector, err := makeCredhubCollector(omService, viper.GetBool(CollectFromCredhubFlag))
	if err != nil {
		return nil, err
	}

	return operations.NewCollector(omCollector, credhubCollector, consumptionCollector, tarWriter, uuid.DefaultGenerator), nil
}
