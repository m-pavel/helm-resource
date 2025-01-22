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
	baseHelmCmd
	output string
}

func newSumCommand() *cobra.Command {
	sum := sumCmd{}

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
	sum.propogateCmdFlags(cmd)
	f := cmd.Flags()
	f.StringVar(&sum.output, "output", "", "Output format")
	return cmd
}

func (s sumCmd) run() error {
	if req, err := s.GetRequirements(); err != nil {
		return err
	} else {
		return s.FormatOutput(os.Stdout, req)
	}
}

func (s sumCmd) FormatOutput(w io.Writer, req *cv1.ResourceRequirements) error {
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

	switch s.output {
	case "table":
		line := func() error {
			if _, err := fmt.Fprint(w, "+---------------+---------------+---------------+---------------+\n"); err != nil {
				return err
			}
			return nil
		}
		if err := line(); err != nil {
			return err
		}
		if _, err := fmt.Fprint(w, "|               | Static wrkld  | Jobs          | Sum           |\n"); err != nil {
			return err
		}
		if err := line(); err != nil {
			return err
		}
		if _, err := fmt.Fprintf(w, "|CPU Limit      | %13v | %13v | %13v |\n", req.Limits.Cpu(), &jobCpuLim, &sumCpuLim); err != nil {
			return err
		}
		if _, err := fmt.Fprintf(w, "|Memory Limit   | %13v | %13v | %13v |\n", req.Limits.Memory(), &jobMemLim, &sumMemLim); err != nil {
			return err
		}
		if _, err := fmt.Fprintf(w, "|CPU Request    | %13v | %13v | %13v |\n", req.Requests.Cpu(), &jobCpuReq, &sumCpuReq); err != nil {
			return err
		}
		if _, err := fmt.Fprintf(w, "|Memory Request | %13v | %13v | %13v |\n", req.Requests.Memory(), &jobMemReq, &sumMemReq); err != nil {
			return err
		}
		if err := line(); err != nil {
			return err
		}
	default:
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

	}

	return nil
}
