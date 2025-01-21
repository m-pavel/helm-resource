package cmd

import (
	"bufio"
	"bytes"
	"fmt"

	appsv1 "k8s.io/api/apps/v1"
	bav1 "k8s.io/api/batch/v1"
	cv1 "k8s.io/api/core/v1"
	"k8s.io/apimachinery/pkg/api/resource"

	"sigs.k8s.io/yaml"
)

func (s sumCmd) Parse(manifest []byte) (*cv1.ResourceRequirements, error) {
	scanner := bufio.NewScanner(bytes.NewReader(manifest))
	scanner.Split(scanYamlSpecs)
	scanner.Buffer(make([]byte, bufio.MaxScanTokenSize), 10485760)

	cr := cv1.ResourceRequirements{
		Limits: cv1.ResourceList{
			"cpu":    resource.MustParse("0"),
			"memory": resource.MustParse("0"),
		},
		Requests: cv1.ResourceList{
			"cpu":    resource.MustParse("0"),
			"memory": resource.MustParse("0"),
		},
	}

	for scanner.Scan() {
		content := scanner.Bytes()
		if ok, err := s.parseDeployment(content, &cr); err != nil {
			return nil, err
		} else {
			if ok {
				continue
			}
		}

		if ok, err := s.parseStatefulset(content, &cr); err != nil {
			return nil, err
		} else {
			if ok {
				continue
			}
		}

		if _, err := s.parseCronJob(content, &cr); err != nil {
			return nil, err
		}
	}
	if err := scanner.Err(); err != nil {
		return nil, err
	}
	return &cr, nil
}

func (s sumCmd) procResourceRequirements(pathid string, rr cv1.ResourceRequirements, tgt *cv1.ResourceRequirements, repl int32) error {

	{
		v := rr.Limits.Cpu()
		if v.IsZero() {
			var err error
			if v, err = s.defaultResource(pathid, cv1.ResourceCPU, s.getDefaultLimit(cv1.ResourceCPU), "limit"); err != nil {
				return err
			}
		}

		if t, ok := tgt.Limits[cv1.ResourceCPU]; ok {
			v.Mul(int64(repl))
			t.Add(*v)
			tgt.Limits[cv1.ResourceCPU] = t
		}
	}
	{
		v := rr.Limits.Memory()
		if v.IsZero() {
			var err error
			if v, err = s.defaultResource(pathid, cv1.ResourceMemory, s.getDefaultLimit(cv1.ResourceMemory), "limit"); err != nil {
				return err
			}
		}

		if t, ok := tgt.Limits[cv1.ResourceMemory]; ok {
			v.Mul(int64(repl))
			t.Add(*v)
			tgt.Limits[cv1.ResourceMemory] = t
		}
	}
	{
		v := rr.Requests.Cpu()
		if v.IsZero() {
			var err error
			if v, err = s.defaultResource(pathid, cv1.ResourceCPU, s.getDefaultRequest(cv1.ResourceCPU), "request"); err != nil {
				return err
			}
		}
		if t, ok := tgt.Requests[cv1.ResourceCPU]; ok {
			v.Mul(int64(repl))
			t.Add(*v)
			tgt.Requests[cv1.ResourceCPU] = t
		}
	}
	{
		v := rr.Requests.Memory()
		if v.IsZero() {
			var err error
			if v, err = s.defaultResource(pathid, cv1.ResourceMemory, s.getDefaultRequest(cv1.ResourceMemory), "request"); err != nil {
				return err
			}
		}
		if t, ok := tgt.Requests[cv1.ResourceMemory]; ok {
			v.Mul(int64(repl))
			t.Add(*v)
			tgt.Requests[cv1.ResourceMemory] = t
		}
	}
	return nil
}

var ZERO = resource.MustParse("0")

func (s sumCmd) defaultResource(pathid string, typ cv1.ResourceName, val string, role string) (*resource.Quantity, error) {
	if typ == "cpu" {
		if val != "" {
			v, err := resource.ParseQuantity(val)
			return &v, err
		} else {
			if s.require {
				return nil, fmt.Errorf("CPU %s not defined in %s", role, pathid)
			} else {
				return &ZERO, nil
			}
		}
	} else {
		if val != "" {
			v, err := resource.ParseQuantity(val)
			return &v, err
		} else {
			if s.require {
				return nil, fmt.Errorf("Memory %s  not defined in %s", role, pathid)
			} else {
				return &ZERO, nil
			}
		}
	}
}

func (s sumCmd) parseDeployment(content []byte, cr *cv1.ResourceRequirements) (bool, error) {
	depl := appsv1.Deployment{}

	err := yaml.Unmarshal(content, &depl)
	if err != nil {
		// assume yaml is valid and error caused type incompatibility
		fmt.Println(err)
		return false, nil
	}
	if depl.Kind == "Deployment" {
		repl := int32(1)
		if depl.Spec.Replicas != nil {
			repl = *depl.Spec.Replicas
		}

		for _, c := range depl.Spec.Template.Spec.Containers {
			if err = s.procResourceRequirements(fmt.Sprintf("Deployment: %s, Container: %s", depl.Name, c.Name), c.Resources, cr, repl); err != nil {
				return false, err
			}
		}
		return true, nil
	}
	return false, nil
}

func (s sumCmd) parseStatefulset(content []byte, cr *cv1.ResourceRequirements) (bool, error) {
	depl := appsv1.StatefulSet{}

	err := yaml.Unmarshal(content, &depl)
	if err != nil {
		fmt.Println(err)
		return false, err
	}
	if depl.Kind == "StatefulSet" {
		repl := int32(1)
		if depl.Spec.Replicas != nil {
			repl = *depl.Spec.Replicas
		}

		for _, c := range depl.Spec.Template.Spec.Containers {
			if err = s.procResourceRequirements(fmt.Sprintf("StatefulSet: %s, Container: %s", depl.Name, c.Name), c.Resources, cr, repl); err != nil {
				return false, err
			}
		}
		return true, nil
	}
	return false, nil
}

func (s sumCmd) parseCronJob(content []byte, cr *cv1.ResourceRequirements) (bool, error) {
	depl := bav1.CronJob{}

	err := yaml.Unmarshal(content, &depl)
	if err != nil {
		fmt.Println(err)
		return false, err
	}
	if depl.Kind == "CronJob" {
		for _, c := range depl.Spec.JobTemplate.Spec.Template.Spec.Containers {
			if err = s.procResourceRequirements(fmt.Sprintf("CronJob: %s, Container: %s", depl.Name, c.Name), c.Resources, cr, 1); err != nil {
				return false, err
			}
		}
		return true, nil
	}
	return false, nil
}
