package main

import (
	"fmt"
	"os"

	"strconv"

	"time"

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
)

func main() {
	target := envMustHave(OpsManagerURLKey)
	username := envMustHave(OpsManagerUsernameKey)
	password := envMustHave(OpsManagerPasswordKey)
	envMustHave(OutputPathKey)

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
	_, err := collector.Collect()
	if err != nil {
		os.Exit(1)
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
