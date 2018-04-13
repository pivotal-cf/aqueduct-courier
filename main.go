package main

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
)

const (
	OpsManagerURLKey      = "OPS_MANAGER_URL"
	OpsManagerUsernameKey = "OPS_MANAGER_USERNAME"
	OpsManagerPasswordKey = "OPS_MANAGER_PASSWORD"
	OutputPathKey         = "OUTPUT_PATH"
	SkipTlsVerifyKey      = "INSECURE_SKIP_TLS_VERIFY"
	OutputDirPrefix       = "FoundationDetails_"
)

func main() {
	target := envMustHave(OpsManagerURLKey)
	username := envMustHave(OpsManagerUsernameKey)
	password := envMustHave(OpsManagerPasswordKey)
	outputPath := envMustHave(OutputPathKey)

	skipVerify, _ := strconv.ParseBool(os.Getenv(SkipTlsVerifyKey))

	authedClient, _ := network.NewOAuthClient(
		target,
		username,
		password,
		"",
		"",
		skipVerify,
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
		fmt.Fprintf(os.Stderr, "Failed collecting from Ops Manager: %s", err.Error())
		os.Exit(1)
	}

	timeString := time.Now().UTC().Format(time.RFC3339)
	outputFolderPath := filepath.Join(outputPath, fmt.Sprintf("%s%s", OutputDirPrefix, timeString))
	err = os.Mkdir(outputFolderPath, 0755)
	if err != nil {
		fmt.Fprintf(os.Stderr, "Failed creating directory %s:", outputFolderPath)
		os.Exit(1)
	}

	writer := file.Writer{}
	for _, data := range omData {
		err = writer.Write(data, outputFolderPath)
		if err != nil {
			fmt.Fprintf(os.Stderr, "Failed writing data to disk: %s", err.Error())
			os.Exit(1)
		}
	}
}

func envMustHave(key string) string {
	v := os.Getenv(key)
	if v == "" {
		fmt.Fprintf(os.Stderr, "Requires %s to be set", key)
		os.Exit(1)
	}

	return v
}
