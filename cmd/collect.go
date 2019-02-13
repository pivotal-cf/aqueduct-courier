package cmd

import (
	"fmt"
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
)

var collectCmd = &cobra.Command{
	Use:   "collect",
	Short: "Collects information from Operations Manager",
	Long:  `Collects information from Operations Manager and outputs the content to the configured directory`,
	RunE:  collect,
}

func init() {
	collectCmd.Flags().String(OpsManagerURLFlag, "", fmt.Sprintf("URL of Operations Manager to collect from [$%s]", OpsManagerURLKey))
	viper.BindPFlag(OpsManagerURLFlag, collectCmd.Flag(OpsManagerURLFlag))
	viper.BindEnv(OpsManagerURLFlag, OpsManagerURLKey)

	collectCmd.Flags().String(OpsManagerUsernameFlag, "", fmt.Sprintf("Operations Manager username [$%s]\nNote: not required if using client/secret authentication", OpsManagerUsernameKey))
	viper.BindPFlag(OpsManagerUsernameFlag, collectCmd.Flag(OpsManagerUsernameFlag))
	viper.BindEnv(OpsManagerUsernameFlag, OpsManagerUsernameKey)

	collectCmd.Flags().String(OpsManagerPasswordFlag, "", fmt.Sprintf("Operations Manager password [$%s]\nNote: not required if using client/secret authentication", OpsManagerPasswordKey))
	viper.BindPFlag(OpsManagerPasswordFlag, collectCmd.Flag(OpsManagerPasswordFlag))
	viper.BindEnv(OpsManagerPasswordFlag, OpsManagerPasswordKey)

	collectCmd.Flags().String(OpsManagerClientIdFlag, "", fmt.Sprintf("Operations Manager client id [$%s]\nNote: not required if using username/password authentication", OpsManagerClientIdKey))
	viper.BindPFlag(OpsManagerClientIdFlag, collectCmd.Flag(OpsManagerClientIdFlag))
	viper.BindEnv(OpsManagerClientIdFlag, OpsManagerClientIdKey)

	collectCmd.Flags().String(OpsManagerClientSecretFlag, "", fmt.Sprintf("Operations Manager client secret [$%s]\nNote: not required if using username/password authentication", OpsManagerClientSecretKey))
	viper.BindPFlag(OpsManagerClientSecretFlag, collectCmd.Flag(OpsManagerClientSecretFlag))
	viper.BindEnv(OpsManagerClientSecretFlag, OpsManagerClientSecretKey)

	collectCmd.Flags().Int(OpsManagerTimeoutFlag, 30, fmt.Sprintf("Timeout (in seconds) for Operations Manager HTTP requests [$%s]", OpsManagerTimeoutKey))
	viper.BindPFlag(OpsManagerTimeoutFlag, collectCmd.Flag(OpsManagerTimeoutFlag))
	viper.BindEnv(OpsManagerTimeoutFlag, OpsManagerTimeoutKey)

	collectCmd.Flags().String(EnvTypeFlag, "", fmt.Sprintf("Describe the type of environment you're collecting from [$%s]\nValid options: %s, %s, %s, %s", EnvTypeKey, EnvTypeDevelopment, EnvTypeQA, EnvTypePreProduction, EnvTypeProduction))
	viper.BindPFlag(EnvTypeFlag, collectCmd.Flag(EnvTypeFlag))
	viper.BindEnv(EnvTypeFlag, EnvTypeKey)

	collectCmd.Flags().String(OutputPathFlag, "", fmt.Sprintf("Local directory to write data [$%s]", OutputPathKey))
	viper.BindPFlag(OutputPathFlag, collectCmd.Flag(OutputPathFlag))
	viper.BindEnv(OutputPathFlag, OutputPathKey)

	collectCmd.Flags().Bool(SkipTlsVerifyFlag, false, "Skip TLS validation on http requests to Operations Manager")
	viper.BindPFlag(SkipTlsVerifyFlag, collectCmd.Flag(SkipTlsVerifyFlag))

	collectCmd.Flags().Bool(CollectFromCredhubFlag, false, fmt.Sprintf("Collect certificate expiry info from CredHub [$%s]", WithCredhubInfoKey))
	viper.BindPFlag(CollectFromCredhubFlag, collectCmd.Flag(CollectFromCredhubFlag))
	viper.BindEnv(CollectFromCredhubFlag, WithCredhubInfoKey)

	collectCmd.Flags().String(CfApiURLFlag, "", fmt.Sprintf("URL of the CF API used for UAA authentication in order to access the Usage Service [$%s]", CfApiURLKey))
	viper.BindPFlag(CfApiURLFlag, collectCmd.Flag(CfApiURLFlag))
	viper.BindEnv(CfApiURLFlag, CfApiURLKey)

	collectCmd.Flags().String(UsageServiceURLFlag, "", fmt.Sprintf("URL of the Usage Service [$%s]", UsageServiceURLKey))
	viper.BindPFlag(UsageServiceURLFlag, collectCmd.Flag(UsageServiceURLFlag))
	viper.BindEnv(UsageServiceURLFlag, UsageServiceURLKey)

	collectCmd.Flags().String(UsageServiceClientIDFlag, "", fmt.Sprintf("Usage Service client id [$%s]", UsageServiceClientIDKey))
	viper.BindPFlag(UsageServiceClientIDFlag, collectCmd.Flag(UsageServiceClientIDFlag))
	viper.BindEnv(UsageServiceClientIDFlag, UsageServiceClientIDKey)

	collectCmd.Flags().String(UsageServiceClientSecretFlag, "", fmt.Sprintf("Usage Service client secret [$%s]", UsageServiceClientSecretKey))
	viper.BindPFlag(UsageServiceClientSecretFlag, collectCmd.Flag(UsageServiceClientSecretFlag))
	viper.BindEnv(UsageServiceClientSecretFlag, UsageServiceClientSecretKey)

	collectCmd.Flags().Bool(UsageServiceSkipTlsVerifyFlag, false, fmt.Sprintf("Skip TLS validation for Usage Service components [$%s]", UsageServiceSkipTlsVerifyKey))
	viper.BindPFlag(UsageServiceSkipTlsVerifyFlag, collectCmd.Flag(UsageServiceSkipTlsVerifyFlag))
	viper.BindEnv(UsageServiceSkipTlsVerifyFlag, UsageServiceSkipTlsVerifyKey)

	collectCmd.Flags().BoolP("help", "h", false, "Help for collect")
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

	fmt.Printf("Collecting data from Operations Manager at %s\n", viper.GetString(OpsManagerURLFlag))
	err = collectExecutor.Collect(envType, version)
	if err != nil {
		tarFile.Close()
		os.Remove(tarFilePath)
		return err
	}

	fmt.Printf("Wrote output to %s\n", tarFilePath)
	fmt.Println("Success!")
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
	validEnvTypes := []string{EnvTypeDevelopment, EnvTypeQA, EnvTypePreProduction, EnvTypeProduction}
	envType := strings.ToLower(viper.GetString(EnvTypeFlag))
	for _, validType := range validEnvTypes {
		if validType == envType {
			return envType, nil
		}
	}
	return "", errors.Errorf(InvalidEnvTypeFailureFormat, envType)
}

func makeConsumptionCollector() (consumptionDataCollector, error) {
	if anyUsageServiceConfigsProvided() {
		err := validateUsageServiceConfig()
		if err != nil {
			return nil, err
		}

		client := network.NewClient(viper.GetBool(UsageServiceSkipTlsVerifyFlag))
		cfApiClient := cf.NewClient(viper.GetString(CfApiURLFlag), client)
		consumptionCollector := consumption.NewCollector(
			cfApiClient,
			client,
			viper.GetString(UsageServiceURLFlag),
			viper.GetString(UsageServiceClientIDFlag),
			viper.GetString(UsageServiceClientSecretFlag),
		)

		fmt.Printf("Collecting data from Usage Service at %s\n", viper.GetString(UsageServiceURLFlag))
		return consumptionCollector, nil
	}
	return nil, nil
}

func makeCredhubCollector(omService *opsmanager.Service, credhubCollectionEnabled bool) (credhubDataCollector, error) {
	if credhubCollectionEnabled {
		chCreds, err := omService.BoshCredentials()
		if err != nil {
			return nil, err
		}
		requestor, err := ogCredhub.New(
			"https://"+chCreds.Host+":8844",
			ogCredhub.SkipTLSValidation(true),
			ogCredhub.Auth(auth.UaaClientCredentials(chCreds.ClientID, chCreds.ClientSecret)),
		)
		if err != nil {
			return nil, errors.Wrap(err, CredhubClientError)
		}
		credhubService := credhub.NewCredhubService(requestor)
		return credhub.NewDataCollector(credhubService), nil
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
		omService,
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

type credhubDataCollector interface {
	Collect() (credhub.Data, error)
}

type consumptionDataCollector interface {
	Collect() (consumption.Data, error)
}
