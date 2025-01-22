package cmd

import (
	"errors"
	"fmt"
	"os"

	"github.com/spf13/cobra"
	cv1 "k8s.io/api/core/v1"
)

type checkCmd struct {
	baseHelmCmd
}

func newCheckCommand() *cobra.Command {
	check := checkCmd{}

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
	q, err := GetQuota(c.namespace)
	if err != nil {
		return err
	}
	req, err := c.GetRequirements()
	if err != nil {
		return err
	}

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

	qCpuLim := q.Status.Hard[cv1.ResourceLimitsCPU]
	cpuLimOk := req.Limits.Cpu().Cmp(qCpuLim) < 0
	cpuSumLimOk := sumCpuLim.Cmp(qCpuLim) < 0

	qMemLim := q.Status.Hard[cv1.ResourceLimitsMemory]
	memLimOk := req.Limits.Memory().Cmp(qMemLim) < 0
	memSumLimOk := sumMemLim.Cmp(qMemLim) < 0

	qCpuReq := q.Status.Hard[cv1.ResourceRequestsCPU]
	cpuReqOk := req.Requests.Cpu().Cmp(qCpuReq) < 0
	cpuSumReqOk := sumCpuReq.Cmp(qCpuReq) < 0

	qMemReq := q.Status.Hard[cv1.ResourceRequestsMemory]
	memReqOk := req.Requests.Memory().Cmp(qMemReq) < 0
	memSumReqOk := sumMemReq.Cmp(qMemReq) < 0

	w := os.Stdout

	line := func() error {
		if _, err := fmt.Fprint(w, "+---------------+---------------+---------------+---------------+---------------+---------------+--------------+\n"); err != nil {
			return err
		}
		return nil
	}
	if err := line(); err != nil {
		return err
	}
	if _, err := fmt.Fprint(w, "|               | Static wrkld  | Jobs          | Sum           | Quota         | Status ststic |Status sum    |\n"); err != nil {
		return err
	}
	if err := line(); err != nil {
		return err
	}
	if _, err := fmt.Fprintf(w, "|CPU Limit      | %13v | %13v | %13v | %13v | %13t |%13t |\n", req.Limits.Cpu(), &jobCpuLim, &sumCpuLim, &qCpuLim, cpuLimOk, cpuSumLimOk); err != nil {
		return err
	}
	if _, err := fmt.Fprintf(w, "|Memory Limit   | %13v | %13v | %13v | %13v | %13t |%13t |\n", req.Limits.Memory(), &jobMemLim, &sumMemLim, &qMemLim, memLimOk, memSumLimOk); err != nil {
		return err
	}
	if _, err := fmt.Fprintf(w, "|CPU Request    | %13v | %13v | %13v | %13v | %13t |%13t |\n", req.Requests.Cpu(), &jobCpuReq, &sumCpuReq, &qCpuReq, cpuReqOk, cpuSumReqOk); err != nil {
		return err
	}
	if _, err := fmt.Fprintf(w, "|Memory Request | %13v | %13v | %13v | %13v | %13t |%13t |\n", req.Requests.Memory(), &jobMemReq, &sumMemReq, &qMemReq, memReqOk, memSumReqOk); err != nil {
		return err
	}
	if err := line(); err != nil {
		return err
	}
	return nil
}
