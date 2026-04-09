package pal

import (
	"fmt"
	"os"

	"github.com/passflow-ai/passflow/cmd/passflow-cli/internal/api"
	"github.com/passflow-ai/passflow/cmd/passflow-cli/internal/output"
	"github.com/spf13/cobra"
)

var validateMode string

var validateCmd = &cobra.Command{
	Use:   "validate <file>",
	Short: "Validate a PAL file",
	Long: `Validate the syntax and semantics of a PAL file.

Modes:
  strict (default) - Fail on any error
  warn             - Report errors as warnings

Examples:
  passflow pal validate agent.yaml
  passflow pal validate agent.yaml --mode warn`,
	Args: cobra.ExactArgs(1),
	RunE: func(cmd *cobra.Command, args []string) error {
		filePath := args[0]

		content, err := os.ReadFile(filePath)
		if err != nil {
			return fmt.Errorf("failed to read file: %w", err)
		}

		client := api.NewClient()
		result, err := client.ValidatePAL(string(content), validateMode)
		if err != nil {
			return err
		}

		if len(result.Errors) > 0 {
			output.PrintError(fmt.Sprintf("Validation failed with %d error(s)", len(result.Errors)))
			fmt.Println()
			for _, e := range result.Errors {
				if e.Line > 0 {
					fmt.Printf("  Line %d: %s\n", e.Line, e.Message)
				} else {
					fmt.Printf("  %s\n", e.Message)
				}
			}
			if validateMode != "warn" {
				return fmt.Errorf("validation failed")
			}
		}

		if len(result.Warnings) > 0 {
			output.PrintWarning(fmt.Sprintf("%d warning(s)", len(result.Warnings)))
			for _, w := range result.Warnings {
				if w.Line > 0 {
					fmt.Printf("  Line %d: %s\n", w.Line, w.Message)
				} else {
					fmt.Printf("  %s\n", w.Message)
				}
			}
		}

		if result.Valid {
			output.PrintSuccess(fmt.Sprintf("%s is valid", filePath))
		}

		return nil
	},
}

func init() {
	validateCmd.Flags().StringVarP(&validateMode, "mode", "m", "strict", "validation mode (strict, warn)")
}
