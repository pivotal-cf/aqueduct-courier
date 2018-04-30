package cmd

import (
	"github.com/pivotal-cf/aqueduct-courier/ops"
	"github.com/pkg/errors"
	"github.com/spf13/cobra"
	"github.com/spf13/viper"
)

const (
	DirectoryPathFlag = "path"
	ApiKeyFlag        = "api-key"
	ApiKeyKey         = "API_KEY"

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
	sendCmd.Flags().String(DirectoryPathFlag, "", "Directory containing files from 'collect' command")
	viper.BindPFlag(DirectoryPathFlag, sendCmd.Flag(DirectoryPathFlag))

	sendCmd.Flags().String(ApiKeyFlag, "", "API Key used to authenticate with Pivotal")
	viper.BindPFlag(ApiKeyFlag, sendCmd.Flag(ApiKeyFlag))
	viper.BindEnv(ApiKeyFlag, ApiKeyKey)

	rootCmd.AddCommand(sendCmd)
}

func send(_ *cobra.Command, _ []string) error {
	err := verifyRequiredConfig(DirectoryPathFlag, ApiKeyFlag)
	if err != nil {
		return err
	}

	sender := ops.SendExecutor{}
	err = sender.Send(viper.GetString(DirectoryPathFlag), dataLoaderURL, viper.GetString(ApiKeyFlag))
	if err != nil {
		return errors.Wrap(err, SendFailureMessage)
	}

	return nil
}
