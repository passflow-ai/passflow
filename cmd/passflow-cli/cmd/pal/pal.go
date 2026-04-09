package pal

import (
	"github.com/spf13/cobra"
)

var Cmd = &cobra.Command{
	Use:   "pal",
	Short: "Manage agents via PAL (Passflow Agent Language)",
	Long: `PAL commands allow you to validate, apply, export, and diff
agent configurations written in PAL format.

Examples:
  passflow pal validate agent.yaml
  passflow pal apply agent.yaml --dry-run
  passflow pal export <agent-id>
  passflow pal diff <agent-id> agent.yaml`,
}

func init() {
	Cmd.AddCommand(validateCmd)
	Cmd.AddCommand(applyCmd)
	Cmd.AddCommand(exportCmd)
	Cmd.AddCommand(diffCmd)
}
