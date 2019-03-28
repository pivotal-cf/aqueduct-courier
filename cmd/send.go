package cmd

import (
	"fmt"
	"os"

	"net/http"

	"github.com/pivotal-cf/aqueduct-courier/operations"
	"github.com/pkg/errors"
	"github.com/spf13/cobra"
	"github.com/spf13/viper"
)

const (
	DataTarFilePathFlag = "path"
	DataTarFilePathKey  = "DATA_TAR_FILE_PATH"
	ApiKeyFlag          = "api-key"
	ApiKeyKey           = "API_KEY"

	SendFailureMessage      = "Failed to send data"
	FileNotFoundErrorFormat = "File not found at: %s"
)

var dataLoaderURL string

var sendCmd = &cobra.Command{
	Use:   "send",
	Short: "Sends information to Pivotal",
	Long:  `Sends specified file to Pivotal's secure data store at ` + dataLoaderURL,
	RunE:  send,
}

func init() {
	bindFlagAndEnvVar(sendCmd, ApiKeyFlag, "", fmt.Sprintf("``Telemetry Collector API Key used to authenticate with Pivotal [$%s]", ApiKeyKey), ApiKeyKey)
	bindFlagAndEnvVar(sendCmd, DataTarFilePathFlag, "", fmt.Sprintf("``The path to the file with data from the 'collect' command [$%s]\n", DataTarFilePathKey), DataTarFilePathKey)

	sendCmd.Flags().BoolP("help", "h", false, "Help for the send command\n")
	sendCmd.Flags().SortFlags = false

	sendCmd.Example = `
      Send data to Pivotal:
      telemetry-collector send --api-key --path`

	customUsageTextTemplate := `
USAGE EXAMPLES
{{.Example}}

FLAGS

{{.LocalFlags.FlagUsages}}`

	customHelpTextTemplate := fmt.Sprintf(`
Sends specified file to Pivotal's secure data store at %s
%s`, dataLoaderURL, customUsageTextTemplate)

	sendCmd.SetHelpTemplate(customHelpTextTemplate)
	sendCmd.SetUsageTemplate(customUsageTextTemplate)
	rootCmd.AddCommand(sendCmd)
}

func send(c *cobra.Command, _ []string) error {
	err := verifyRequiredConfig(DataTarFilePathFlag, ApiKeyFlag)
	if err != nil {
		return err
	}
	c.SilenceUsage = true

	sender := operations.SendExecutor{}
	tarFile, err := os.Open(viper.GetString(DataTarFilePathFlag))
	if err != nil {
		return errors.New(fmt.Sprintf(FileNotFoundErrorFormat, viper.GetString(DataTarFilePathFlag)))
	}

	logger.Printf("Sending %s to Pivotal at %s\n", viper.GetString(DataTarFilePathFlag), dataLoaderURL)
	err = sender.Send(http.DefaultClient, tarFile.Name(), dataLoaderURL, viper.GetString(ApiKeyFlag), version)
	if err != nil {
		return errors.Wrap(err, SendFailureMessage)
	}

	logger.Println("Success!")

	return nil
}
