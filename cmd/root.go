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
	toolName                  = "telemetry-collector"
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

func bindFlagAndEnvVar(cmd *cobra.Command, flagName string, defaultValue interface{}, usageText, flagKey string) {
	switch val := defaultValue.(type) {
	case string:
		cmd.Flags().String(flagName, val, usageText)
	case int:
		cmd.Flags().Int(flagName, val, usageText)
	case bool:
		cmd.Flags().Bool(flagName, val, usageText)
	}
	viper.BindPFlag(flagName, cmd.Flag(flagName))
	viper.BindEnv(flagName, flagKey)
}
