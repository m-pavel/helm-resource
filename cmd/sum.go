package cmd

import (
	"errors"
	"fmt"
	"os"

	"github.com/spf13/cobra"
)

type sumCmd struct {
	namespace string

	chart      string
	values     []string
	valueFiles []string
}

func newSumCommand() *cobra.Command {
	sum := sumCmd{
		namespace: os.Getenv("HELM_NAMESPACE"),
	}

	cmd := &cobra.Command{
		Use:   "sum",
		Short: "Show resource summary",
		Long:  rootCmdLongUsage,
		Args: func(cmd *cobra.Command, args []string) error {
			if len(args) == 0 {
				return errors.New("requires a chart path parameter")
			}
			return nil
		},
		RunE: func(cmd *cobra.Command, args []string) error {
			sum.chart = args[0]
			return sum.run()
		},
	}

	f := cmd.Flags()
	f.StringArrayVar(&sum.values, "set", []string{}, "set values on the command line (can specify multiple or separate values with commas: key1=val1,key2=val2)")
	f.StringArrayVarP(&sum.valueFiles, "values", "f", []string{}, "specify values in a YAML file (can specify multiple)")

	return cmd
}

func (s sumCmd) run() error {
	manifest, err := getTemplate(s.chart, s.namespace, s.values, s.valueFiles)
	if err != nil {
		return err
	}
	req, err := Parse(manifest)
	if err != nil {
		return err
	}
	fmt.Printf("CPU Request %v\n", req.Requests.Cpu())
	fmt.Printf("Memory Request %v\n", req.Requests.Memory())
	fmt.Printf("CPU Limit %v\n", req.Limits.Cpu())
	fmt.Printf("Memory Limit %v\n", req.Limits.Memory())

	return nil
}
