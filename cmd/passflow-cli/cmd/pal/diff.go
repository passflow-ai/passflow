package pal

import (
	"fmt"
	"os"

	"github.com/jaak-ai/passflow-cli/internal/api"
	"github.com/jaak-ai/passflow-cli/internal/config"
	"github.com/jaak-ai/passflow-cli/internal/output"
	"github.com/spf13/cobra"
)

var diffWorkspace string

var diffCmd = &cobra.Command{
	Use:   "diff <agent-id> <file>",
	Short: "Show differences between an agent and a PAL file",
	Long: `Compare an existing agent's configuration with a local PAL file
and show the differences.

Examples:
  passflow pal diff agent-123 agent.yaml
  passflow pal diff agent-123 agent.yaml -w ws-abc123`,
	Args: cobra.ExactArgs(2),
	RunE: func(cmd *cobra.Command, args []string) error {
		agentID := args[0]
		filePath := args[1]

		workspace := diffWorkspace
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
		result, err := client.DiffPAL(workspace, agentID, string(content))
		if err != nil {
			return err
		}

		if !result.HasChanges {
			output.PrintSuccess("No changes detected")
			return nil
		}

		fmt.Printf("Comparing agent %s with %s\n\n", agentID, filePath)

		// Group changes by type
		var added, removed, modified []api.DiffChange
		for _, c := range result.Changes {
			switch c.Type {
			case "added":
				added = append(added, c)
			case "removed":
				removed = append(removed, c)
			case "modified":
				modified = append(modified, c)
			}
		}

		if len(added) > 0 {
			fmt.Println(output.Success("Added:"))
			for _, a := range added {
				fmt.Printf("  %s %s: %s\n", output.Success("+"), a.Path, a.NewValue)
			}
			fmt.Println()
		}

		if len(removed) > 0 {
			fmt.Println(output.Error("Removed:"))
			for _, r := range removed {
				fmt.Printf("  %s %s: %s\n", output.Error("-"), r.Path, r.OldValue)
			}
			fmt.Println()
		}

		if len(modified) > 0 {
			fmt.Println(output.Warning("Modified:"))
			for _, m := range modified {
				fmt.Printf("  %s:\n", m.Path)
				fmt.Printf("    %s %s\n", output.Error("-"), m.OldValue)
				fmt.Printf("    %s %s\n", output.Success("+"), m.NewValue)
			}
			fmt.Println()
		}

		output.PrintDiff(result.Summary.Added, result.Summary.Removed, result.Summary.Modified)

		return nil
	},
}

func init() {
	diffCmd.Flags().StringVarP(&diffWorkspace, "workspace", "w", "", "workspace ID")
}
