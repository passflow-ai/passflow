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
	exportOutput    string
	exportFormat    string
	exportWorkspace string
)

var exportCmd = &cobra.Command{
	Use:   "export <agent-id>",
	Short: "Export an agent configuration as PAL",
	Long: `Export an existing agent's configuration in PAL format.

By default outputs to stdout in YAML format.

Examples:
  passflow pal export agent-123
  passflow pal export agent-123 -o agent.yaml
  passflow pal export agent-123 --format json`,
	Args: cobra.ExactArgs(1),
	RunE: func(cmd *cobra.Command, args []string) error {
		agentID := args[0]

		workspace := exportWorkspace
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
		result, err := client.ExportPAL(workspace, agentID, exportFormat)
		if err != nil {
			return err
		}

		if exportOutput != "" {
			if err := os.WriteFile(exportOutput, []byte(result.Content), 0644); err != nil {
				return fmt.Errorf("failed to write file: %w", err)
			}
			output.PrintSuccess(fmt.Sprintf("Exported to %s", exportOutput))
			return nil
		}

		fmt.Print(result.Content)
		return nil
	},
}

func init() {
	exportCmd.Flags().StringVarP(&exportOutput, "output", "o", "", "output file (default: stdout)")
	exportCmd.Flags().StringVar(&exportFormat, "format", "yaml", "output format (yaml, json)")
	exportCmd.Flags().StringVarP(&exportWorkspace, "workspace", "w", "", "workspace ID")
}
