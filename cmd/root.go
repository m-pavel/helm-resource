package cmd

import (
	"github.com/spf13/cobra"
)

const rootCmdLongUsage = `resource`

// New creates a new cobra client
func New() *cobra.Command {
	sumCommand := newSumCommand()

	cmd := &cobra.Command{
		Use:   "resource",
		Short: "Show resource summary",
		Long:  rootCmdLongUsage,
		Args:  sumCommand.Args,
	}

	// add flagset from chartCommand
	cmd.Flags().AddFlagSet(sumCommand.Flags())
	cmd.AddCommand(versionCmd(), sumCommand)
	cmd.SetHelpCommand(&cobra.Command{}) // Disable the help command
	return cmd
}
