package agents

import (
	"github.com/spf13/cobra"
)

var Cmd = &cobra.Command{
	Use:   "agents",
	Short: "Manage agents",
	Long: `Commands for listing and inspecting agents.

Examples:
  passflow agents list
  passflow agents get <agent-id>`,
}

func init() {
	Cmd.AddCommand(listCmd)
	Cmd.AddCommand(getCmd)
}
