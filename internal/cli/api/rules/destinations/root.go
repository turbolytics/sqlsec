package destinations

import "github.com/spf13/cobra"

func NewCommand() *cobra.Command {
	cmd := &cobra.Command{
		Use:   "destinations",
		Short: "Manage rule destinations",
	}

	cmd.AddCommand(NewAddCmd())

	return cmd
}
