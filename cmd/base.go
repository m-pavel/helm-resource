package cmd

import (
	"github.com/spf13/cobra"
	cv1 "k8s.io/api/core/v1"
)

type baseHelmCmd struct {
	namespace string

	chart      string
	values     []string
	valueFiles []string

	remote  bool
	require bool

	defaultCpuLimit string
	defaultMemLimit string
	defaultCpuReq   string
	defaultMemReq   string
}

func (b baseHelmCmd) getDefault(k cv1.ResourceName, role string) string {
	if role == "limit" {
		if k == cv1.ResourceCPU {
			return b.defaultCpuLimit
		}
		return b.defaultMemLimit
	} else {
		if k == cv1.ResourceCPU {
			return b.defaultCpuReq
		}
		return b.defaultMemReq
	}
}

func (b *baseHelmCmd) propogateCmdFlags(cmd *cobra.Command) *cobra.Command {
	f := cmd.Flags()
	f.StringArrayVar(&b.values, "set", []string{}, "set values on the command line (can specify multiple or separate values with commas: key1=val1,key2=val2)")
	f.StringArrayVarP(&b.valueFiles, "values", "f", []string{}, "specify values in a YAML file (can specify multiple)")

	f.BoolVar(&b.remote, "remote", false, "Calculate for remote release instand of local chart")
	f.BoolVar(&b.require, "require", false, "Require CPU and Memory values to be defined for each container.")

	f.StringVar(&b.defaultCpuLimit, "default-cpu-limit", "", "Default value for CPU limit")
	f.StringVar(&b.defaultMemLimit, "default-mem-limit", "", "Default value for Memory limit")
	f.StringVar(&b.defaultCpuReq, "default-cpu-req", "", "Default value for CPU request")
	f.StringVar(&b.defaultMemReq, "default-mem-req", "", "Default value for Memory request")
	return cmd
}
