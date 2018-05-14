package cmd

import (
	"errors"
	"fmt"
	"os"

	"github.com/spf13/cobra"
	"github.com/spf13/viper"
)

const RequiredConfigErrorFormat = "Requires --%s to be set"

var (
	version = "dev"
	rootCmd = &cobra.Command{
		Use:   "aqueduct",
		Short: "Utility for collecting information about a PCF Foundation",
	}
)

func Execute() {
	rootCmd.Version = version
	if err := rootCmd.Execute(); err != nil {
		os.Exit(1)
	}
}

func verifyRequiredConfig(keys ...string) error {
	for _, k := range keys {
		if viper.GetString(k) == "" {
			return errors.New(fmt.Sprintf(RequiredConfigErrorFormat, k))
		}
	}
	return nil
}
