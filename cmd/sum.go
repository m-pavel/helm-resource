package cmd

import (
	"errors"
	"fmt"
	"io"
	"os"

	"github.com/spf13/cobra"
	cv1 "k8s.io/api/core/v1"
)

type sumCmd struct {
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

func (s sumCmd) getDefaultLimit(k cv1.ResourceName) string {
	if k == "cpu" {
		return s.defaultCpuLimit
	}
	return s.defaultMemLimit
}

func (s sumCmd) getDefaultRequest(k cv1.ResourceName) string {
	if k == "cpu" {
		return s.defaultCpuReq
	}
	return s.defaultMemReq
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
				return errors.New("requires an argument: chart path or release name")
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

	f.BoolVar(&sum.remote, "remote", false, "Calculate for remote release instand of local chart")
	f.BoolVar(&sum.require, "require", false, "Require CPU and Memory values to be defined for each container.")

	f.StringVar(&sum.defaultCpuLimit, "default-cpu-limit", "", "Default value for CPU limit")
	f.StringVar(&sum.defaultMemLimit, "default-mem-limit", "", "Default value for Memory limit")
	f.StringVar(&sum.defaultCpuReq, "default-cpu-req", "", "Default value for CPU request")
	f.StringVar(&sum.defaultMemReq, "default-mem-req", "", "Default value for Memory request")

	return cmd
}

func (s sumCmd) run() error {
	var manifest []byte
	var err error
	if s.remote {
		manifest, err = getRelease(s.chart, s.namespace)
	} else {
		manifest, err = getTemplate(s.chart, s.namespace, s.values, s.valueFiles)
	}

	if err != nil {
		return err
	}
	req, err := s.Parse(manifest)
	if err != nil {
		return err
	}
	FormatOutput(os.Stdout, req)
	return nil
}

func FormatOutput(w io.Writer, req *cv1.ResourceRequirements) error {
	jobCpuReq := req.Requests[jobCpu]
	jobCpuLim := req.Limits[jobCpu]
	jobMemReq := req.Requests[jobMemory]
	jobMemLim := req.Limits[jobMemory]

	sumCpuReq := req.Requests.Cpu().DeepCopy()
	sumCpuReq.Add(jobCpuReq)
	sumMemReq := req.Requests.Memory().DeepCopy()
	sumMemReq.Add(jobMemReq)
	sumCpuLim := req.Limits.Cpu().DeepCopy()
	sumCpuLim.Add(jobCpuLim)
	sumMemLim := req.Limits.Memory().DeepCopy()
	sumMemLim.Add(jobMemLim)

	if _, err := fmt.Fprintf(w, "CPU Limit %v + %v (Jobs) = %v\n", req.Limits.Cpu(), &jobCpuLim, &sumCpuLim); err != nil {
		return err
	}
	if _, err := fmt.Fprintf(w, "Memory Limit %v + %v (Jobs) = %v\n", req.Limits.Memory(), &jobMemLim, &sumMemLim); err != nil {
		return err
	}
	if _, err := fmt.Fprintf(w, "CPU Request %v + %v (Jobs) = %v\n", req.Requests.Cpu(), &jobCpuReq, &sumCpuReq); err != nil {
		return err
	}
	if _, err := fmt.Fprintf(w, "Memory Request %v + %v (Jobs) = %v\n", req.Requests.Memory(), &jobMemReq, &sumMemReq); err != nil {
		return err
	}

	return nil
}
