package bd

import (
	"fmt"
	"os/exec"
	"strings"
)

func RunJSON(args ...string) ([]byte, error) {
	command := exec.Command("bd", args...)
	output, err := command.CombinedOutput()
	if err != nil {
		message := strings.TrimSpace(string(output))
		if message != "" {
			message = "\n" + message
		}
		return nil, fmt.Errorf("bd %s failed: %w%s", strings.Join(args, " "), err, message)
	}

	return output, nil
}

func RunJSONInDir(dir string, args ...string) ([]byte, error) {
	command := exec.Command("bd", args...)
	command.Dir = dir
	output, err := command.CombinedOutput()
	if err != nil {
		message := strings.TrimSpace(string(output))
		if message != "" {
			message = "\n" + message
		}
		return nil, fmt.Errorf("bd %s failed: %w%s", strings.Join(args, " "), err, message)
	}

	return output, nil
}
