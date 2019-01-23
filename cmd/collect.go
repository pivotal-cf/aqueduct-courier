package cmd

import (
	"fmt"
	"os"
	"path/filepath"
	"strings"
	"time"

	ogCredhub "code.cloudfoundry.org/credhub-cli/credhub"
	"code.cloudfoundry.org/credhub-cli/credhub/auth"
	"github.com/pivotal-cf/aqueduct-courier/credhub"

	"github.com/pivotal-cf/aqueduct-courier/operations"
	"github.com/pivotal-cf/aqueduct-courier/opsmanager"
	"github.com/pivotal-cf/aqueduct-utils/file"
	"github.com/pivotal-cf/om/api"
	"github.com/pivotal-cf/om/network"
	"github.com/pkg/errors"
	"github.com/spf13/cobra"
	"github.com/spf13/viper"
)

const (
	OpsManagerURLKey           = "OPS_MANAGER_URL"
	OpsManagerUsernameKey      = "OPS_MANAGER_USERNAME"
	OpsManagerPasswordKey      = "OPS_MANAGER_PASSWORD"
	OpsManagerClientIdKey      = "OPS_MANAGER_CLIENT_ID"
	OpsManagerClientSecretKey  = "OPS_MANAGER_CLIENT_SECRET"
	OpsManagerTimeoutKey       = "OPS_MANAGER_TIMEOUT"
	EnvTypeKey                 = "ENV_TYPE"
	OutputPathKey              = "OUTPUT_DIR"
	WithCredhubInfoKey         = "WITH_CREDHUB_INFO"
	OpsManagerURLFlag          = "url"
	OpsManagerUsernameFlag     = "username"
	OpsManagerPasswordFlag     = "password"
	OpsManagerClientIdFlag     = "client-id"
	OpsManagerClientSecretFlag = "client-secret"
	OpsManagerTimeoutFlag      = "ops-manager-timeout"
	CollectFromCredhubFlag     = "with-credhub-info"
	EnvTypeFlag                = "env-type"
	OutputPathFlag             = "output-dir"
	SkipTlsVerifyFlag          = "insecure-skip-tls-verify"

	EnvTypeDevelopment   = "development"
	EnvTypeQA            = "qa"
	EnvTypePreProduction = "pre-production"
	EnvTypeProduction    = "production"

	OutputFilePrefix                = "FoundationDetails_"
	CredhubClientError              = "Failed creating credhub client"
	InvalidEnvTypeFailureFormat     = "Invalid env-type %s. See help for the list of valid types."
	InvalidAuthConfigurationMessage = "Invalid auth configuration. Requires username/password or client/secret to be set."
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

	authedClient, _ := network.NewOAuthClient(
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

	credhubCollector, err := makeCredhubCollector(omService, viper.GetBool(CollectFromCredhubFlag))
	if err != nil {
		return err
	}

	tarFilePath := filepath.Join(
		viper.GetString(OutputPathFlag),
		fmt.Sprintf("%s%d.tar", OutputFilePrefix, time.Now().UTC().Unix()),
	)
	tarWriter, err := file.NewTarWriter(tarFilePath)
	if err != nil {
		return err
	}

	ce := operations.NewCollector(omCollector, credhubCollector, tarWriter)

	fmt.Printf("Collecting data from Operations Manager at %s\n", viper.GetString(OpsManagerURLFlag))
	err = ce.Collect(envType, version)
	if err != nil {
		os.Remove(tarFilePath)
		return err
	}

	fmt.Printf("Wrote output to %s\n", tarFilePath)
	fmt.Println("Success!")
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
		return noOpCredhubCollector{}, nil
	}
}

type credhubDataCollector interface {
	Collect() (credhub.Data, error)
}

type noOpCredhubCollector struct{}

func (n noOpCredhubCollector) Collect() (credhub.Data, error) {
	return credhub.Data{}, nil
}
