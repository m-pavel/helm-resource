package cmd

import (
	"bytes"
	"errors"
	"fmt"
	"os"
	"os/exec"
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

func getRelease(release, namespace string) ([]byte, error) {
	args := []string{"get", "manifest", release}
	if namespace != "" {
		args = append(args, "--namespace", namespace)
	}
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
