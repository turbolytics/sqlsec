package rules

import (
	"github.com/spf13/cobra"
	"github.com/turbolytics/sqlsec/internal/cli/api/rules/destinations"
)

func NewCommand() *cobra.Command {
	cmd := &cobra.Command{
		Use:   "rules",
		Short: "Manage rules",
	}

	cmd.AddCommand(NewCreateCmd())
	cmd.AddCommand(NewUpdateCmd())
	cmd.AddCommand(NewGetCmd())
	cmd.AddCommand(NewTestCmd())
	cmd.AddCommand(NewInstallCmd())

	cmd.AddCommand(destinations.NewCommand())

	return cmd
}
