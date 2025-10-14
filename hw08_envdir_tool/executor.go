package main

import (
	"errors"
	"fmt"
	"os"
	"os/exec"
	"strings"
)

// RunCmd runs a command + arguments (cmd) with environment variables from env.
func RunCmd(cmd []string, env Environment) (returnCode int) {
	if len(cmd) == 0 {
		fmt.Print("no command provided")
		return 1
	}

	stringEnvs := make([]string, 0, len(env))
	for _, e := range os.Environ() {
		kv := strings.SplitN(e, "=", 2)
		k := kv[0]
		if v, ok := env[k]; ok && v.NeedRemove {
			continue
		}
		stringEnvs = append(stringEnvs, e)
	}

	for k, v := range env {
		stringEnvs = append(stringEnvs, fmt.Sprintf("%s=%s", k, v.Value))
	}

	command := exec.Command(cmd[0], cmd[1:]...) //nolint
	command.Env = stringEnvs
	command.Stdin = os.Stdin
	command.Stdout = os.Stdout
	command.Stderr = os.Stderr

	if err := command.Run(); err != nil {
		var exitErr *exec.ExitError
		if errors.As(err, &exitErr) {
			return exitErr.ExitCode()
		}
		fmt.Printf("command error: %v\n", err)
		return 1
	}
	return 0
}
