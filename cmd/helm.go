package cmd

import (
	"bufio"
	"bytes"
	"errors"
	"fmt"
	"os"
	"os/exec"

	appsv1 "k8s.io/api/apps/v1"
	bav1 "k8s.io/api/batch/v1"
	cv1 "k8s.io/api/core/v1"
	"k8s.io/apimachinery/pkg/api/resource"

	"sigs.k8s.io/yaml"
)

func getTemplate(template, namespace string, variables []string, values []string) ([]byte, error) {
	args := []string{"template", template}
	if namespace != "" {
		args = append(args, "--namespace", namespace)
	}
	for _, v := range variables {
		args = append(args, "--set", v)
	}
	for _, v := range values {
		args = append(args, "--values", v)
	}
	// fmt.Printf("Running: %s %v\n", os.Getenv("HELM_BIN"), args)
	cmd := exec.Command(os.Getenv("HELM_BIN"), args...)
	return outputWithRichError(cmd)
}

func outputWithRichError(cmd *exec.Cmd) ([]byte, error) {
	output, err := cmd.Output()
	var exitError *exec.ExitError
	if errors.As(err, &exitError) {
		return output, fmt.Errorf("%s: %s", exitError.Error(), string(exitError.Stderr))
	}
	return output, err
}

func Parse(manifest []byte) (*cv1.ResourceRequirements, error) {
	// Ensure we have a newline in front of the yaml separator
	scanner := bufio.NewScanner(bytes.NewReader(manifest))
	scanner.Split(scanYamlSpecs)
	// Allow for tokens (specs) up to 10MiB in size
	scanner.Buffer(make([]byte, bufio.MaxScanTokenSize), 10485760)

	// result := make(map[string]*MappingResult)

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
		if ok, err := parseDeployment(content, &cr); err == nil && ok {
			continue
		}
		if ok, err := parseStatefulset(content, &cr); err == nil && ok {
			continue
		}
		parseCronJob(content, &cr)
	}
	if err := scanner.Err(); err != nil {
		return nil, err
	}
	return &cr, nil
}

func procResourceRequirements(rr cv1.ResourceRequirements, tgt *cv1.ResourceRequirements, repl int32) {
	for k, v := range rr.Limits {
		if v.IsZero() {
			continue
		}
		if t, ok := tgt.Limits[k]; ok {
			v.Mul(int64(repl))
			t.Add(v)
			tgt.Limits[k] = t
		}
	}
	for k, v := range rr.Requests {
		if v.IsZero() {
			continue
		}
		if t, ok := tgt.Requests[k]; ok {
			v.Mul(int64(repl))
			t.Add(v)
			tgt.Requests[k] = t
		}
	}
}
func parseDeployment(content []byte, cr *cv1.ResourceRequirements) (bool, error) {
	depl := appsv1.Deployment{}

	err := yaml.Unmarshal(content, &depl)
	if err != nil {
		return false, err
	}
	if depl.Kind == "Deployment" {
		repl := int32(1)
		if depl.Spec.Replicas != nil {
			repl = *depl.Spec.Replicas
		}

		for _, c := range depl.Spec.Template.Spec.Containers {
			procResourceRequirements(c.Resources, cr, repl)
		}
		return true, nil
	}
	return false, nil
}

func parseStatefulset(content []byte, cr *cv1.ResourceRequirements) (bool, error) {
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
			procResourceRequirements(c.Resources, cr, repl)
		}
		return true, nil
	}
	return false, nil
}

func parseCronJob(content []byte, cr *cv1.ResourceRequirements) (bool, error) {
	depl := bav1.CronJob{}

	err := yaml.Unmarshal(content, &depl)
	if err != nil {
		fmt.Println(err)
		return false, err
	}
	if depl.Kind == "CronJob" {
		for _, c := range depl.Spec.JobTemplate.Spec.Template.Spec.Containers {
			procResourceRequirements(c.Resources, cr, 1)
		}
		return true, nil
	}
	return false, nil
}

var yamlSeparator = []byte("\n---\n")

func scanYamlSpecs(data []byte, atEOF bool) (int, []byte, error) {
	if atEOF && len(data) == 0 {
		return 0, nil, nil
	}
	if i := bytes.Index(data, yamlSeparator); i >= 0 {
		// We have a full newline-terminated line.
		return i + len(yamlSeparator), data[0:i], nil
	}
	// If we're at EOF, we have a final, non-terminated line. Return it.
	if atEOF {
		return len(data), data, nil
	}
	// Request more data.
	return 0, nil, nil
}
