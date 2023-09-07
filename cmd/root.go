package cmd

import (
	"fmt"
	"log"
	"os"
	"strings"

	"github.com/pkg/errors"

	"github.com/spf13/cobra"
	"github.com/spf13/viper"
)

const (
	RequiredConfigErrorFormat    = "Missing required flags: %s"
	toolName                     = "telemetry-collector"
	PendingChangesExistsExitCode = 3
)

var (
	version = "dev"
	logger  *log.Logger
	rootCmd = &cobra.Command{
		Use:   toolName,
		Short: "Utility for collecting information about a PCF Foundation",
	}
)

func Execute() {
	rootCmd.Version = version
	logger = log.New(os.Stdout, "", 0)

	rootCmd.Flags().BoolP("help", "h", false, fmt.Sprintf("Help for %s", toolName))
	rootCmd.Flags().BoolP("version", "v", false, fmt.Sprintf("Version for %s", toolName))

	rootCmd.Example = `
  "telemetry-collector [command]" executes a command
  "telemetry-collector [command] --help" shows information about a command`

	customHelpTextTemplate := `
Utility for collecting information about a PCF foundation

USAGE EXAMPLES
{{.Example}}

COMMANDS

  collect     Collects information from a PCF foundation
  send        Sends information to VMware
  help        Shows help about any command

FLAGS

{{.LocalFlags.FlagUsages}}`
	rootCmd.SetHelpTemplate(customHelpTextTemplate)

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
	_ = viper.BindPFlag(flagName, cmd.Flag(flagName))
	_ = viper.BindEnv(flagName, flagKey)
}
