package cmd

import (
	"fmt"
	"os"

	"github.com/passflow-ai/passflow/cmd/passflow-cli/cmd/agents"
	"github.com/passflow-ai/passflow/cmd/passflow-cli/cmd/pal"
	"github.com/passflow-ai/passflow/cmd/passflow-cli/internal/config"
	"github.com/spf13/cobra"
)

var (
	cfgFile   string
	outputFmt string
	version   = "dev"
)

func SetVersion(v string) {
	version = v
}

var rootCmd = &cobra.Command{
	Use:   "passflow",
	Short: "Passflow CLI - Manage AI agents via PAL",
	Long: `Passflow CLI is a command-line tool for managing AI agents
using PAL (Passflow Agent Language).

Configure your API connection:
  passflow config set api-url https://api.passflow.ai
  passflow login --token <your-jwt-token>

Manage agents with PAL:
  passflow pal validate agent.yaml
  passflow pal apply agent.yaml
  passflow pal export <agent-id>`,
	SilenceUsage: true,
}

func Execute() {
	if err := rootCmd.Execute(); err != nil {
		os.Exit(1)
	}
}

func init() {
	cobra.OnInitialize(initConfig)

	rootCmd.PersistentFlags().StringVar(&cfgFile, "config", "", "config file (default is $HOME/.passflow/config.yaml)")
	rootCmd.PersistentFlags().StringVarP(&outputFmt, "output", "o", "table", "output format (table, json, yaml)")

	rootCmd.AddCommand(loginCmd)
	rootCmd.AddCommand(logoutCmd)
	rootCmd.AddCommand(configCmd)
	rootCmd.AddCommand(pal.Cmd)
	rootCmd.AddCommand(agents.Cmd)
	rootCmd.AddCommand(versionCmd)
}

func initConfig() {
	if cfgFile != "" {
		config.SetConfigFile(cfgFile)
	}

	if err := config.Load(); err != nil {
		fmt.Fprintf(os.Stderr, "Warning: %v\n", err)
	}
}

var versionCmd = &cobra.Command{
	Use:   "version",
	Short: "Print the version number",
	Run: func(cmd *cobra.Command, args []string) {
		fmt.Printf("passflow version %s\n", version)
	},
}
