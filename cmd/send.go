package cmd

import (
	"fmt"
	"os"

	"github.com/pivotal-cf/aqueduct-courier/ops"
	"github.com/pivotal-cf/aqueduct-utils/data"
	"github.com/pivotal-cf/aqueduct-utils/file"
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
	Long:  `Sends specified file to Pivotal's secure data store at ` + dataLoaderURL,
	RunE:  send,
}

func init() {
	sendCmd.Flags().String(DataTarFilePathFlag, "", "Tar archive containing data from 'collect' command")
	viper.BindPFlag(DataTarFilePathFlag, sendCmd.Flag(DataTarFilePathFlag))

	sendCmd.Flags().String(ApiKeyFlag, "", fmt.Sprintf("Insights Collector API Key used to authenticate with Pivotal [$%s]", ApiKeyKey))
	viper.BindPFlag(ApiKeyFlag, sendCmd.Flag(ApiKeyFlag))
	viper.BindEnv(ApiKeyFlag, ApiKeyKey)

	sendCmd.Flags().BoolP("help", "h", false, "Help for send")
	rootCmd.AddCommand(sendCmd)
}

func send(c *cobra.Command, _ []string) error {
	err := verifyRequiredConfig(DataTarFilePathFlag, ApiKeyFlag)
	if err != nil {
		return err
	}
	c.SilenceUsage = true

	sender := ops.SendExecutor{}
	tarFile, err := os.Open(viper.GetString(DataTarFilePathFlag))
	if err != nil {
		panic(err)
	}

	tarReader := file.NewTarReader(tarFile)
	tValidator := data.NewFileValidator(tarReader)

	fmt.Printf("Sending %s to Pivotal at %s\n", viper.GetString(DataTarFilePathFlag), dataLoaderURL)
	err = sender.Send(tarReader, tValidator, tarFile.Name(), dataLoaderURL, viper.GetString(ApiKeyFlag), version)
	if err != nil {
		return errors.Wrap(err, SendFailureMessage)
	}

	fmt.Println("Success!")

	return nil
}
