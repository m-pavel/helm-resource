package cmd

import (
	"github.com/spf13/cobra"
)

const rootCmdLongUsage = `resource`

func New() *cobra.Command {
	sumCommand := newSumCommand()
	checkCommand := newCheckCommand()

	cmd := &cobra.Command{
		Use:   "resource",
		Short: "Show resource summary",
		Long:  rootCmdLongUsage,
		Args:  sumCommand.Args,
	}

	// add flagset from chartCommand
	cmd.Flags().AddFlagSet(sumCommand.Flags())
	cmd.Flags().AddFlagSet(checkCommand.Flags())
	cmd.AddCommand(versionCmd(), sumCommand, checkCommand)
	cmd.SetHelpCommand(&cobra.Command{})
	return cmd
}
