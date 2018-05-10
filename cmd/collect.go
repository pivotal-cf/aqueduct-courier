package cmd

import (
	"fmt"
	"os"
	"path/filepath"
	"strings"
	"time"

	"github.com/pivotal-cf/aqueduct-courier/file"
	"github.com/pivotal-cf/aqueduct-courier/ops"
	"github.com/pivotal-cf/aqueduct-courier/opsmanager"
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
	EnvTypeKey                 = "ENV_TYPE"
	OutputPathKey              = "OUTPUT_PATH"
	SkipTlsVerifyKey           = "INSECURE_SKIP_TLS_VERIFY"
	OpsManagerURLFlag          = "url"
	OpsManagerUsernameFlag     = "username"
	OpsManagerPasswordFlag     = "password"
	OpsManagerClientIdFlag     = "client-id"
	OpsManagerClientSecretFlag = "client-secret"
	EnvTypeFlag                = "env-type"
	OutputPathFlag             = "output"
	SkipTlsVerifyFlag          = "insecure-skip-tls-verify"

	EnvTypeDevelopment   = "development"
	EnvTypeQA            = "qa"
	EnvTypePreProduction = "pre-production"
	EnvTypeProduction    = "production"

	OutputFilePrefix                = "FoundationDetails_"
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

	collectCmd.Flags().String(OpsManagerUsernameFlag, "", fmt.Sprintf("Operations Manager username [$%s]", OpsManagerUsernameKey))
	viper.BindPFlag(OpsManagerUsernameFlag, collectCmd.Flag(OpsManagerUsernameFlag))
	viper.BindEnv(OpsManagerUsernameFlag, OpsManagerUsernameKey)

	collectCmd.Flags().String(OpsManagerPasswordFlag, "", fmt.Sprintf("Operations Manager password [$%s]", OpsManagerPasswordKey))
	viper.BindPFlag(OpsManagerPasswordFlag, collectCmd.Flag(OpsManagerPasswordFlag))
	viper.BindEnv(OpsManagerPasswordFlag, OpsManagerPasswordKey)

	collectCmd.Flags().String(OpsManagerClientIdFlag, "", fmt.Sprintf("Operations Manager client id [$%s]", OpsManagerClientIdKey))
	viper.BindPFlag(OpsManagerClientIdFlag, collectCmd.Flag(OpsManagerClientIdFlag))
	viper.BindEnv(OpsManagerClientIdFlag, OpsManagerClientIdKey)

	collectCmd.Flags().String(OpsManagerClientSecretFlag, "", fmt.Sprintf("Operations Manager client secret [$%s]", OpsManagerClientSecretKey))
	viper.BindPFlag(OpsManagerClientSecretFlag, collectCmd.Flag(OpsManagerClientSecretFlag))
	viper.BindEnv(OpsManagerClientSecretFlag, OpsManagerClientSecretKey)

	collectCmd.Flags().String(EnvTypeFlag, "", fmt.Sprintf("Environment type. Valid options are: %s, %s, %s, and %s [$%s]", EnvTypeDevelopment, EnvTypeQA, EnvTypePreProduction, EnvTypeProduction, EnvTypeKey))
	viper.BindPFlag(EnvTypeFlag, collectCmd.Flag(EnvTypeFlag))
	viper.BindEnv(EnvTypeFlag, EnvTypeKey)

	collectCmd.Flags().String(OutputPathFlag, "", fmt.Sprintf("Local file path to write data [$%s]", OutputPathKey))
	viper.BindPFlag(OutputPathFlag, collectCmd.Flag(OutputPathFlag))
	viper.BindEnv(OutputPathFlag, OutputPathKey)

	collectCmd.Flags().Bool(
		SkipTlsVerifyFlag,
		false,
		fmt.Sprintf("Skip TLS validation on http requests to Operations Manager. Defaults to false. [$%s]", SkipTlsVerifyKey),
	)
	viper.BindPFlag(SkipTlsVerifyFlag, collectCmd.Flag(SkipTlsVerifyFlag))
	viper.BindEnv(SkipTlsVerifyFlag, SkipTlsVerifyKey)

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
		30*time.Second,
	)

	omService := &opsmanager.Service{
		Requestor: api.NewRequestService(authedClient),
	}

	collector := opsmanager.NewDataCollector(
		omService,
		api.NewPendingChangesService(authedClient),
		api.NewDeployedProductsService(authedClient),
	)

	tarFilePath := filepath.Join(
		viper.GetString(OutputPathFlag),
		fmt.Sprintf("%s%d.tar", OutputFilePrefix, time.Now().UTC().Unix()),
	)
	tarWriter, err := file.NewTarWriter(tarFilePath)
	if err != nil {
		return err
	}

	ce := ops.NewCollector(collector, tarWriter)

	err = ce.Collect(envType)
	if err != nil {
		os.Remove(tarFilePath)
		return err
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
