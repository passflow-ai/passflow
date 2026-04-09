package pal

import (
	"fmt"
	"os"

	"github.com/jaak-ai/passflow-cli/internal/api"
	"github.com/jaak-ai/passflow-cli/internal/config"
	"github.com/jaak-ai/passflow-cli/internal/output"
	"github.com/spf13/cobra"
)

var (
	applyDryRun    bool
	applyWorkspace string
)

var applyCmd = &cobra.Command{
	Use:   "apply <file>",
	Short: "Apply a PAL file to create or update an agent",
	Long: `Apply a PAL configuration to create a new agent or update an existing one.

Use --dry-run to preview changes without applying them.

Examples:
  passflow pal apply agent.yaml
  passflow pal apply agent.yaml --dry-run
  passflow pal apply agent.yaml -w ws-abc123`,
	Args: cobra.ExactArgs(1),
	RunE: func(cmd *cobra.Command, args []string) error {
		filePath := args[0]

		workspace := applyWorkspace
		if workspace == "" {
			workspace = config.GetWorkspace()
		}
		if workspace == "" {
			return fmt.Errorf("workspace not specified. Use -w flag or set default with: passflow config set workspace <id>")
		}

		if !config.IsAuthenticated() {
			return fmt.Errorf("not authenticated. Run: passflow login --token <token>")
		}

		content, err := os.ReadFile(filePath)
		if err != nil {
			return fmt.Errorf("failed to read file: %w", err)
		}

		client := api.NewClient()
		result, err := client.ApplyPAL(string(content), workspace, applyDryRun)
		if err != nil {
			return err
		}

		if applyDryRun {
			output.PrintInfo("Dry run - no changes applied")
			fmt.Println()
		}

		actionVerb := "created"
		if result.Action == "updated" {
			actionVerb = "updated"
		}

		if applyDryRun {
			actionVerb = "would be " + actionVerb
		}

		name := result.AgentName
		if name == "" {
			name = result.AgentID
		}
		output.PrintSuccess(fmt.Sprintf("Agent '%s' %s", name, actionVerb))

		if len(result.Changes) > 0 {
			fmt.Println("\nChanges:")
			for _, c := range result.Changes {
				if c.OldValue != "" {
					fmt.Printf("  %s: %s → %s\n", c.Field, c.OldValue, c.NewValue)
				} else {
					fmt.Printf("  %s: %s\n", c.Field, c.NewValue)
				}
			}
		}

		if !applyDryRun {
			fmt.Printf("\nAgent ID: %s\n", result.AgentID)
		}

		return nil
	},
}

func init() {
	applyCmd.Flags().BoolVar(&applyDryRun, "dry-run", false, "preview changes without applying")
	applyCmd.Flags().StringVarP(&applyWorkspace, "workspace", "w", "", "workspace ID")
}
