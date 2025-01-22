package cmd

import (
	"errors"
	"os"

	"github.com/spf13/cobra"
)

type checkCmd struct {
	baseHelmCmd
}

func newCheckCommand() *cobra.Command {
	check := checkCmd{
		baseHelmCmd: baseHelmCmd{
			namespace: os.Getenv("HELM_NAMESPACE"),
		},
	}

	cmd := &cobra.Command{
		Use:   "check",
		Short: "Check chart resource requirements with cluster quotas",
		Long:  rootCmdLongUsage,
		Args: func(cmd *cobra.Command, args []string) error {
			if len(args) == 0 {
				return errors.New("requires an argument: chart path or release name")
			}
			return nil
		},
		RunE: func(cmd *cobra.Command, args []string) error {
			check.chart = args[0]
			return check.run()
		},
	}

	return check.propogateCmdFlags(cmd)
}

func (c checkCmd) run() error {
	return nil
}
