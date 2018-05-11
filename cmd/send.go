package cmd

import (
	"fmt"

	"github.com/pivotal-cf/aqueduct-courier/file"
	"github.com/pivotal-cf/aqueduct-courier/ops"
	"github.com/pkg/errors"
	"github.com/spf13/cobra"
	"github.com/spf13/viper"
)

const (
	DataTarFilePathFlag = "path"
	ApiKeyFlag          = "api-key"
	ApiKeyKey           = "API_KEY"

	SendFailureMessage = "Failed to send data"
)

var dataLoaderURL string

var sendCmd = &cobra.Command{
	Use:   "send",
	Short: "Sends information to Pivotal",
	Long:  `Sends information from your file system to Pivotal`,
	RunE:  send,
}

func init() {
	sendCmd.Flags().String(DataTarFilePathFlag, "", "Tar archive containing data from 'collect' command")
	viper.BindPFlag(DataTarFilePathFlag, sendCmd.Flag(DataTarFilePathFlag))

	sendCmd.Flags().String(ApiKeyFlag, "", fmt.Sprintf("API Key used to authenticate with Pivotal [$%s]", ApiKeyKey))
	viper.BindPFlag(ApiKeyFlag, sendCmd.Flag(ApiKeyFlag))
	viper.BindEnv(ApiKeyFlag, ApiKeyKey)

	rootCmd.AddCommand(sendCmd)
}

func send(c *cobra.Command, _ []string) error {
	err := verifyRequiredConfig(DataTarFilePathFlag, ApiKeyFlag)
	if err != nil {
		return err
	}
	c.SilenceUsage = true

	sender := ops.SendExecutor{}
	tarReader := file.NewTarReader(viper.GetString(DataTarFilePathFlag))
	err = sender.Send(tarReader, dataLoaderURL, viper.GetString(ApiKeyFlag))
	if err != nil {
		return errors.Wrap(err, SendFailureMessage)
	}

	return nil
}
