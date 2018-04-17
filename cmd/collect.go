package cmd

import (
	"fmt"
	"os"
	"path/filepath"
	"strconv"
	"time"

	"github.com/pivotal-cf/aqueduct-courier/file"
	"github.com/pivotal-cf/aqueduct-courier/opsmanager"
	"github.com/pivotal-cf/om/api"
	"github.com/pivotal-cf/om/network"
	"github.com/pkg/errors"
	"github.com/spf13/cobra"
)

const (
	OpsManagerURLKey      = "OPS_MANAGER_URL"
	OpsManagerUsernameKey = "OPS_MANAGER_USERNAME"
	OpsManagerPasswordKey = "OPS_MANAGER_PASSWORD"
	OutputPathKey         = "OUTPUT_PATH"
	SkipTlsVerifyKey      = "INSECURE_SKIP_TLS_VERIFY"
	OutputDirPrefix       = "FoundationDetails_"
)

var collectCmd = &cobra.Command{
	Use:   "collect",
	Short: "Collects information from the configured Ops Manager",
	Long:  `Collects information from the configured Ops Manager and outputs the content to the configured directory`,
	RunE:  collect,
}

func init() {
	rootCmd.AddCommand(collectCmd)
}

func collect(_ *cobra.Command, _ []string) error {
	conf, err := parseConfig()
	if err != nil {
		return err
	}

	authedClient, _ := network.NewOAuthClient(
		conf.target,
		conf.username,
		conf.password,
		"",
		"",
		conf.skipVerify,
		false,
		30*time.Second,
	)

	omService := &opsmanager.Service{
		Requestor: api.NewRequestService(authedClient),
	}

	builder := opsmanager.DataCollectorBuilder{
		OmService:             omService,
		PendingChangesService: api.NewPendingChangesService(authedClient),
		DeployProductsService: api.NewDeployedProductsService(authedClient),
	}

	collector := opsmanager.NewDataCollector(builder)
	omData, err := collector.Collect()
	if err != nil {
		return errors.Wrap(err, "Failed collecting from Ops Manager")
	}

	timeString := time.Now().UTC().Format(time.RFC3339)
	outputFolderPath := filepath.Join(conf.outputPath, fmt.Sprintf("%s%s", OutputDirPrefix, timeString))
	err = os.Mkdir(outputFolderPath, 0755)
	if err != nil {
		return errors.Wrap(err, fmt.Sprintf("Failed creating directory %s:", outputFolderPath))
	}

	writer := file.Writer{}
	for _, data := range omData {
		err = writer.Write(data, outputFolderPath)
		if err != nil {
			errors.Wrap(err, "Failed writing data to disk")
		}
	}
	return nil
}

type config struct {
	target     string
	username   string
	password   string
	outputPath string
	skipVerify bool
}

func parseConfig() (config, error) {
	conf := config{
		target:     os.Getenv(OpsManagerURLKey),
		username:   os.Getenv(OpsManagerUsernameKey),
		password:   os.Getenv(OpsManagerPasswordKey),
		outputPath: os.Getenv(OutputPathKey),
	}

	if conf.target == "" {
		return config{}, fmt.Errorf("Requires %s to be set", OpsManagerURLKey)
	}

	if conf.username == "" {
		return config{}, fmt.Errorf("Requires %s to be set", OpsManagerUsernameKey)
	}

	if conf.password == "" {
		return config{}, fmt.Errorf("Requires %s to be set", OpsManagerPasswordKey)
	}

	if conf.outputPath == "" {
		return config{}, fmt.Errorf("Requires %s to be set", OutputPathKey)
	}

	conf.skipVerify, _ = strconv.ParseBool(os.Getenv(SkipTlsVerifyKey))
	return conf, nil
}
