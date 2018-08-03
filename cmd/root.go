package cmd

import (
	"errors"
	"fmt"
	"os"
	"strings"

	"github.com/spf13/cobra"
	"github.com/spf13/viper"
)

const (
	RequiredConfigErrorFormat = "Missing required flags: %s"
	toolName                  = "insights-collector"
)

var (
	version = "dev"
	rootCmd = &cobra.Command{
		Use:   toolName,
		Short: "Utility for collecting information about a PCF Foundation",
	}
)

func Execute() {
	rootCmd.Version = version
	rootCmd.Flags().BoolP("help", "h", false, fmt.Sprintf("Help for %s", toolName))
	rootCmd.Flags().BoolP("version", "v", false, fmt.Sprintf("Version for %s", toolName))
	if err := rootCmd.Execute(); err != nil {
		os.Exit(1)
	}
}

func verifyRequiredConfig(keys ...string) error {
	var missingFlags []string
	for _, k := range keys {
		if viper.GetString(k) == "" {
			missingFlags = append(missingFlags, "--"+k)

		}
	}

	if len(missingFlags) > 0 {
		return errors.New(fmt.Sprintf(RequiredConfigErrorFormat, strings.Join(missingFlags, ", ")))
	}

	return nil
}
