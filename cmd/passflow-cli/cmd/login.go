package cmd

import (
	"fmt"

	"github.com/passflow-ai/passflow/cmd/passflow-cli/internal/config"
	"github.com/spf13/cobra"
)

var loginToken string

var loginCmd = &cobra.Command{
	Use:   "login",
	Short: "Authenticate with Passflow API",
	Long: `Authenticate with the Passflow API using a JWT token.

Examples:
  passflow login --token eyJhbGciOiJIUzI1NiIs...
  passflow login -t <your-token>`,
	RunE: func(cmd *cobra.Command, args []string) error {
		if loginToken == "" {
			return fmt.Errorf("token is required. Use --token flag")
		}

		if err := config.SetToken(loginToken); err != nil {
			return fmt.Errorf("failed to save token: %w", err)
		}

		fmt.Println("Successfully logged in!")
		return nil
	},
}

func init() {
	loginCmd.Flags().StringVarP(&loginToken, "token", "t", "", "JWT token for authentication")
}
