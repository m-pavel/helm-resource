package cmd

import (
	"fmt"

	"github.com/spf13/cobra"
)

var Version = "HEAD"

func versionCmd() *cobra.Command {
	return &cobra.Command{
		Use:   "version",
		Short: "Show version of the helm resource plugin",
		Run: func(*cobra.Command, []string) {
			fmt.Println(Version)
		},
	}
}
