package cmd

import (
	"fmt"

	"github.com/jaak-ai/passflow-cli/internal/config"
	"github.com/spf13/cobra"
)

var logoutCmd = &cobra.Command{
	Use:   "logout",
	Short: "Remove stored credentials",
	Long:  `Remove the stored JWT token from the configuration file.`,
	RunE: func(cmd *cobra.Command, args []string) error {
		if err := config.ClearToken(); err != nil {
			return fmt.Errorf("failed to clear token: %w", err)
		}

		fmt.Println("Successfully logged out!")
		return nil
	},
}
