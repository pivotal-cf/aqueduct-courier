package cmd

import (
	"github.com/spf13/cobra"
	"github.com/spf13/viper"
)

const (
	DirectoryPathFlag = "path"
	ApiKeyFlag        = "api-key"
)

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

	rootCmd.AddCommand(sendCmd)
}

func send(_ *cobra.Command, _ []string) error {
	err := verifyRequiredConfig(DirectoryPathFlag, ApiKeyFlag)
	if err != nil {
		return err
	}

	return nil
}
