package cmd

import (
	"fmt"
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
	OpsManagerURLKey       = "OPS_MANAGER_URL"
	OpsManagerUsernameKey  = "OPS_MANAGER_USERNAME"
	OpsManagerPasswordKey  = "OPS_MANAGER_PASSWORD"
	OutputPathKey          = "OUTPUT_PATH"
	SkipTlsVerifyKey       = "INSECURE_SKIP_TLS_VERIFY"
	OpsManagerURLFlag      = "url"
	OpsManagerUsernameFlag = "username"
	OpsManagerPasswordFlag = "password"
	OutputPathFlag         = "output"
	SkipTlsVerifyFlag      = "insecure-skip-tls-verify"

	RequiredConfigErrorFormat = "Requires --%s to be set"
)

var collectCmd = &cobra.Command{
	Use:   "collect",
	Short: "Collects information from Ops Manager",
	Long:  `Collects information from Ops Manager and outputs the content to the configured directory`,
	RunE:  collect,
}

func init() {
	collectCmd.Flags().String(OpsManagerURLFlag, "", fmt.Sprintf("URL of Ops Manager to collect from [$%s]", OpsManagerURLKey))
	viper.BindPFlag(OpsManagerURLFlag, collectCmd.Flag(OpsManagerURLFlag))
	viper.BindEnv(OpsManagerURLFlag, OpsManagerURLKey)

	collectCmd.Flags().String(OpsManagerUsernameFlag, "", fmt.Sprintf("Operations Manager username [$%s]", OpsManagerUsernameKey))
	viper.BindPFlag(OpsManagerUsernameFlag, collectCmd.Flag(OpsManagerUsernameFlag))
	viper.BindEnv(OpsManagerUsernameFlag, OpsManagerUsernameKey)

	collectCmd.Flags().String(OpsManagerPasswordFlag, "", fmt.Sprintf("Operations Manager password [$%s]", OpsManagerPasswordKey))
	viper.BindPFlag(OpsManagerPasswordFlag, collectCmd.Flag(OpsManagerPasswordFlag))
	viper.BindEnv(OpsManagerPasswordFlag, OpsManagerPasswordKey)

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

func collect(_ *cobra.Command, _ []string) error {
	err := validateConfig()
	if err != nil {
		return err
	}

	authedClient, _ := network.NewOAuthClient(
		viper.GetString(OpsManagerURLFlag),
		viper.GetString(OpsManagerUsernameFlag),
		viper.GetString(OpsManagerPasswordFlag),
		"",
		"",
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
	writer := file.Writer{}
	ce := ops.NewCollector(collector, writer)

	err = ce.Collect(viper.GetString(OutputPathFlag))
	if err != nil {
		return err
	}
	return nil
}

func validateConfig() error {
	keys := []string{
		OpsManagerURLFlag,
		OpsManagerUsernameFlag,
		OpsManagerPasswordFlag,
		OutputPathFlag,
	}
	for _, k := range keys {
		if viper.GetString(k) == "" {
			return errors.New(fmt.Sprintf(RequiredConfigErrorFormat, k))
		}
	}
	return nil
}
