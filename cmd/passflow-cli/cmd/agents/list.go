package agents

import (
	"fmt"

	"github.com/jaak-ai/passflow-cli/internal/api"
	"github.com/jaak-ai/passflow-cli/internal/config"
	"github.com/jaak-ai/passflow-cli/internal/output"
	"github.com/spf13/cobra"
)

var listWorkspace string

var listCmd = &cobra.Command{
	Use:   "list",
	Short: "List all agents",
	Long: `List all agents in a workspace.

Examples:
  passflow agents list
  passflow agents list -w ws-abc123`,
	RunE: func(cmd *cobra.Command, args []string) error {
		workspace := listWorkspace
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
		agents, err := client.ListAgents(workspace)
		if err != nil {
			return err
		}

		if len(agents) == 0 {
			output.PrintInfo("No agents found in workspace")
			return nil
		}

		headers := []string{"ID", "Name", "Status", "Model", "Updated"}
		var data [][]string
		for _, a := range agents {
			data = append(data, []string{
				a.ID,
				a.Name,
				a.Status,
				a.Model,
				a.UpdatedAt,
			})
		}

		output.Table(headers, data)
		return nil
	},
}

func init() {
	listCmd.Flags().StringVarP(&listWorkspace, "workspace", "w", "", "workspace ID")
}
