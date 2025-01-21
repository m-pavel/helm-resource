package cmd

import (
	"os"
	"strconv"
	"strings"

	"github.com/gonvenience/bunt"
	"github.com/mgutz/ansi"
	"github.com/spf13/cobra"
	"golang.org/x/term"
)

const rootCmdLongUsage = `resource`

// New creates a new cobra client
func New() *cobra.Command {
	sumCommand := newSumCommand()

	cmd := &cobra.Command{
		Use:   "resource",
		Short: "Show resource summary",
		Long:  rootCmdLongUsage,
		//Alias root command to chart subcommand
		Args: sumCommand.Args,
		// parse the flags and check for actions like suppress-secrets, no-colors
		PersistentPreRun: func(cmd *cobra.Command, args []string) {
			var fc *bool

			if cmd.Flags().Changed("color") {
				v, _ := cmd.Flags().GetBool("color")
				fc = &v
			} else {
				v, err := strconv.ParseBool(os.Getenv("HELM_DIFF_COLOR"))
				if err == nil {
					fc = &v
				}
			}

			if !cmd.Flags().Changed("output") {
				v, set := os.LookupEnv("HELM_DIFF_OUTPUT")
				if set && strings.TrimSpace(v) != "" {
					_ = cmd.Flags().Set("output", v)
				}
			}

			// Dyff relies on bunt, default to color=on
			bunt.SetColorSettings(bunt.ON, bunt.ON)
			nc, _ := cmd.Flags().GetBool("no-color")

			if nc || (fc != nil && !*fc) {
				ansi.DisableColors(true)
				bunt.SetColorSettings(bunt.OFF, bunt.OFF)
			} else if !cmd.Flags().Changed("no-color") && fc == nil {
				term := term.IsTerminal(int(os.Stdout.Fd()))
				// https://github.com/databus23/helm-diff/issues/281
				dumb := os.Getenv("TERM") == "dumb"
				ansi.DisableColors(!term || dumb)
				bunt.SetColorSettings(bunt.OFF, bunt.OFF)
			}
		},
	}

	// add flagset from chartCommand
	cmd.Flags().AddFlagSet(sumCommand.Flags())
	cmd.AddCommand(versionCmd(), sumCommand)
	cmd.SetHelpCommand(&cobra.Command{}) // Disable the help command
	return cmd
}
