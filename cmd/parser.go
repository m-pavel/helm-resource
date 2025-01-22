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

var ZERO = resource.MustParse("0")
var UNO = resource.MustParse("1")

const (
	jobCpu    = "x-job-cpu"
	jobMemory = "x-job-memory"
)

type TypeParser func(content []byte, cr *cv1.ResourceRequirements) (bool, error)

func (b baseHelmCmd) GetRequirements() (*cv1.ResourceRequirements, error) {
	var manifest []byte
	var err error
	if b.remote {
		manifest, err = getRelease(b.chart, b.namespace)
	} else {
		manifest, err = getTemplate(b.chart, b.namespace, b.values, b.valueFiles)
	}

	if err != nil {
		return nil, err
	}
	return b.Parse(manifest)
}

func (b baseHelmCmd) Parse(manifest []byte) (*cv1.ResourceRequirements, error) {
	scanner := bufio.NewScanner(bytes.NewReader(manifest))
	scanner.Split(scanYamlSpecs)
	scanner.Buffer(make([]byte, bufio.MaxScanTokenSize), 10485760)

	cr := cv1.ResourceRequirements{
		Limits: cv1.ResourceList{
			cv1.ResourceCPU:     resource.MustParse("0"),
			cv1.ResourceMemory:  resource.MustParse("0"),
			cv1.ResourceStorage: resource.MustParse("0"),

			cv1.ResourceConfigMaps:             resource.MustParse("0"),
			cv1.ResourceSecrets:                resource.MustParse("0"),
			cv1.ResourcePersistentVolumeClaims: resource.MustParse("0"),
			cv1.ResourceServices:               resource.MustParse("0"),

			jobCpu:    resource.MustParse("0"),
			jobMemory: resource.MustParse("0"),
		},
		Requests: cv1.ResourceList{
			cv1.ResourceCPU:     resource.MustParse("0"),
			cv1.ResourceMemory:  resource.MustParse("0"),
			cv1.ResourceStorage: resource.MustParse("0"),

			jobCpu:    resource.MustParse("0"),
			jobMemory: resource.MustParse("0"),
		},
	}

	parsers := []TypeParser{
		b.parseCronJob,
		b.parseDeployment,
		b.parseStatefulset,

		b.parseConfigmap,
		b.parseSecret,
		b.parsePvc,
		b.parseService,
	}

	for scanner.Scan() {
		content := scanner.Bytes()

		for _, p := range parsers {
			if ok, err := p(content, &cr); err != nil {
				return nil, err
			} else {
				if ok {
					break
				}
			}
		}
	}
	if err := scanner.Err(); err != nil {
		return nil, err
	}
	return &cr, nil
}

func (b baseHelmCmd) procRequirementSrc(resourceSrc cv1.ResourceName, resourceTgt cv1.ResourceName, pathid string, rr cv1.ResourceList, tgt cv1.ResourceList, repl int32, role string) error {
	v := rr[resourceSrc]

	if v.IsZero() {
		if vp, err := b.defaultResource(pathid, resourceSrc, b.getDefault(resourceSrc, role), role); err != nil {
			return err
		} else {
			v = *vp
		}

	}

	if t, ok := tgt[resourceTgt]; ok {
		v.Mul(int64(repl))
		t.Add(v)
		tgt[resourceTgt] = t
	}
	return nil
}

func (b baseHelmCmd) procRequirement(resource cv1.ResourceName, pathid string, rr cv1.ResourceList, tgt cv1.ResourceList, repl int32, role string) error {
	return b.procRequirementSrc(resource, resource, pathid, rr, tgt, repl, role)
}

func (b baseHelmCmd) procResourceRequirements(pathid string, rr cv1.ResourceRequirements, tgt *cv1.ResourceRequirements, repl int32) error {
	if err := b.procRequirement(cv1.ResourceCPU, pathid, rr.Limits, tgt.Limits, repl, "limit"); err != nil {
		return err
	}
	if err := b.procRequirement(cv1.ResourceMemory, pathid, rr.Limits, tgt.Limits, repl, "limit"); err != nil {
		return err
	}
	if err := b.procRequirement(cv1.ResourceCPU, pathid, rr.Requests, tgt.Requests, repl, "request"); err != nil {
		return err
	}
	if err := b.procRequirement(cv1.ResourceMemory, pathid, rr.Requests, tgt.Requests, repl, "request"); err != nil {
		return err
	}
	return nil
}

func (b baseHelmCmd) defaultResource(pathid string, typ cv1.ResourceName, val string, role string) (*resource.Quantity, error) {
	if typ == "cpu" {
		if val != "" {
			v, err := resource.ParseQuantity(val)
			return &v, err
		} else {
			if b.require {
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
			if b.require {
				return nil, fmt.Errorf("Memory %s not defined in %s", role, pathid)
			} else {
				return &ZERO, nil
			}
		}
	}
}

func (b baseHelmCmd) parseService(content []byte, cr *cv1.ResourceRequirements) (bool, error) {
	depl := cv1.Service{}

	err := yaml.Unmarshal(content, &depl)
	if err != nil {
		return false, err
	}
	if depl.Kind == "Service" {
		t := cr.Limits[cv1.ResourceServices]
		t.Add(UNO)
		cr.Limits[cv1.ResourceServices] = t
		return true, nil
	}
	return false, nil
}

func (b baseHelmCmd) parseConfigmap(content []byte, cr *cv1.ResourceRequirements) (bool, error) {
	depl := cv1.ConfigMap{}

	err := yaml.Unmarshal(content, &depl)
	if err != nil {
		return false, err
	}
	if depl.Kind == "ConfigMap" {
		t := cr.Limits[cv1.ResourceConfigMaps]
		t.Add(UNO)
		cr.Limits[cv1.ResourceConfigMaps] = t
		return true, nil
	}
	return false, nil
}

func (b baseHelmCmd) parseSecret(content []byte, cr *cv1.ResourceRequirements) (bool, error) {
	depl := cv1.Secret{}

	err := yaml.Unmarshal(content, &depl)
	if err != nil {
		return false, err
	}
	if depl.Kind == "Secret" {
		t := cr.Limits[cv1.ResourceSecrets]
		t.Add(UNO)
		cr.Limits[cv1.ResourceSecrets] = t
		return true, nil
	}
	return false, nil
}

func (b baseHelmCmd) parsePvc(content []byte, cr *cv1.ResourceRequirements) (bool, error) {
	depl := cv1.PersistentVolumeClaim{}

	err := yaml.Unmarshal(content, &depl)
	if err != nil {
		return false, err
	}
	if depl.Kind == "PersistentVolumeClaim" {
		t := cr.Limits[cv1.ResourcePersistentVolumeClaims]
		t.Add(UNO)
		cr.Limits[cv1.ResourcePersistentVolumeClaims] = t

		return true, nil
	}
	return false, nil
}

func (b baseHelmCmd) parseDeployment(content []byte, cr *cv1.ResourceRequirements) (bool, error) {
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
			if err = b.procResourceRequirements(fmt.Sprintf("Deployment: %s, Container: %s", depl.Name, c.Name), c.Resources, cr, repl); err != nil {
				return false, err
			}
		}
		return true, nil
	}
	return false, nil
}

func (b baseHelmCmd) parseStatefulset(content []byte, cr *cv1.ResourceRequirements) (bool, error) {
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
			if err = b.procResourceRequirements(fmt.Sprintf("StatefulSet: %s, Container: %s", depl.Name, c.Name), c.Resources, cr, repl); err != nil {
				return false, err
			}
		}
		return true, nil
	}
	return false, nil
}

func (b baseHelmCmd) parseCronJob(content []byte, cr *cv1.ResourceRequirements) (bool, error) {
	depl := bav1.CronJob{}

	err := yaml.Unmarshal(content, &depl)
	if err != nil {
		fmt.Println(err)
		return false, err
	}
	if depl.Kind == "CronJob" {
		for _, c := range depl.Spec.JobTemplate.Spec.Template.Spec.Containers {
			pathid := fmt.Sprintf("CronJob: %s, Container: %s", depl.Name, c.Name)
			if err := b.procRequirementSrc(cv1.ResourceCPU, jobCpu, pathid, c.Resources.Limits, cr.Limits, 1, "limit"); err != nil {
				return false, err
			}
			if err := b.procRequirementSrc(cv1.ResourceMemory, jobMemory, pathid, c.Resources.Limits, cr.Limits, 1, "limit"); err != nil {
				return false, err
			}
			if err := b.procRequirementSrc(cv1.ResourceCPU, jobCpu, pathid, c.Resources.Requests, cr.Requests, 1, "request"); err != nil {
				return false, err
			}
			if err := b.procRequirementSrc(cv1.ResourceMemory, jobMemory, pathid, c.Resources.Requests, cr.Requests, 1, "request"); err != nil {
				return false, err
			}
		}
		return true, nil
	}
	return false, nil
}
