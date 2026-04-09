package agents

import (
	"fmt"
	"strings"

	"github.com/jaak-ai/passflow-cli/internal/api"
	"github.com/jaak-ai/passflow-cli/internal/config"
	"github.com/jaak-ai/passflow-cli/internal/output"
	"github.com/spf13/cobra"
)

var (
	getWorkspace string
	getFormat    string
)

var getCmd = &cobra.Command{
	Use:   "get <agent-id>",
	Short: "Get agent details",
	Long: `Get detailed information about an agent.

Examples:
  passflow agents get agent-123
  passflow agents get agent-123 --format json
  passflow agents get agent-123 -w ws-abc123`,
	Args: cobra.ExactArgs(1),
	RunE: func(cmd *cobra.Command, args []string) error {
		agentID := args[0]

		workspace := getWorkspace
		if workspace == "" {
			workspace = config.GetWorkspace()
		}
		if workspace == "" {
			return fmt.Errorf("workspace not specified. Use -w flag or set default with: passflow config set workspace <id>")
		}

		if !config.IsAuthenticated() {
			return fmt.Errorf("not authenticated. Run: passflow login --token <token>")
		}

		client := api.NewClient()
		agent, err := client.GetAgent(workspace, agentID)
		if err != nil {
			return err
		}

		if getFormat == "json" || getFormat == "yaml" {
			return output.Print(getFormat, agent)
		}

		fmt.Printf("%s: %s\n", output.Bold("ID"), agent.ID)
		fmt.Printf("%s: %s\n", output.Bold("Name"), agent.Name)
		if agent.Description != "" {
			fmt.Printf("%s: %s\n", output.Bold("Description"), agent.Description)
		}
		fmt.Printf("%s: %s\n", output.Bold("Status"), agent.Status)
		fmt.Printf("%s: %s\n", output.Bold("Model"), agent.Model)
		if len(agent.Tools) > 0 {
			fmt.Printf("%s: %s\n", output.Bold("Tools"), strings.Join(agent.Tools, ", "))
		}
		fmt.Printf("%s: %s\n", output.Bold("Created"), agent.CreatedAt)
		fmt.Printf("%s: %s\n", output.Bold("Updated"), agent.UpdatedAt)

		return nil
	},
}

func init() {
	getCmd.Flags().StringVarP(&getWorkspace, "workspace", "w", "", "workspace ID")
	getCmd.Flags().StringVar(&getFormat, "format", "table", "output format (table, json, yaml)")
}
