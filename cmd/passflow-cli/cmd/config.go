package cmd

import (
	"fmt"

	"github.com/passflow-ai/passflow/cmd/passflow-cli/internal/config"
	"github.com/passflow-ai/passflow/cmd/passflow-cli/internal/output"
	"github.com/spf13/cobra"
)

var configCmd = &cobra.Command{
	Use:   "config",
	Short: "Manage CLI configuration",
	Long: `Manage the Passflow CLI configuration.

Examples:
  passflow config set api-url https://api.passflow.ai
  passflow config set workspace ws-abc123
  passflow config get api-url
  passflow config list`,
}

var configSetCmd = &cobra.Command{
	Use:   "set <key> <value>",
	Short: "Set a configuration value",
	Long: `Set a configuration value.

Available keys:
  api-url    - Base URL of the Passflow API
  workspace  - Default workspace ID

Examples:
  passflow config set api-url https://api.passflow.ai
  passflow config set workspace ws-abc123`,
	Args: cobra.ExactArgs(2),
	RunE: func(cmd *cobra.Command, args []string) error {
		key, value := args[0], args[1]

		switch key {
		case "api-url":
			if err := config.SetAPIURL(value); err != nil {
				return err
			}
		case "workspace":
			if err := config.SetWorkspace(value); err != nil {
				return err
			}
		default:
			return fmt.Errorf("unknown config key: %s", key)
		}

		fmt.Printf("Set %s = %s\n", key, value)
		return nil
	},
}

var configGetCmd = &cobra.Command{
	Use:   "get <key>",
	Short: "Get a configuration value",
	Args:  cobra.ExactArgs(1),
	RunE: func(cmd *cobra.Command, args []string) error {
		key := args[0]

		var value string
		switch key {
		case "api-url":
			value = config.GetAPIURL()
		case "workspace":
			value = config.GetWorkspace()
		case "token":
			value = maskToken(config.GetToken())
		default:
			return fmt.Errorf("unknown config key: %s", key)
		}

		if value == "" {
			fmt.Printf("%s is not set\n", key)
		} else {
			fmt.Println(value)
		}
		return nil
	},
}

var configListCmd = &cobra.Command{
	Use:   "list",
	Short: "List all configuration values",
	RunE: func(cmd *cobra.Command, args []string) error {
		cfg := config.GetAll()

		data := [][]string{
			{"api-url", cfg.APIURL},
			{"workspace", cfg.DefaultWorkspace},
			{"token", maskToken(cfg.Token)},
		}

		output.Table([]string{"Key", "Value"}, data)
		return nil
	},
}

func maskToken(token string) string {
	if token == "" {
		return ""
	}
	if len(token) < 20 {
		return "***"
	}
	return token[:10] + "..." + token[len(token)-5:]
}

func init() {
	configCmd.AddCommand(configSetCmd)
	configCmd.AddCommand(configGetCmd)
	configCmd.AddCommand(configListCmd)
}
