package executor

import (
	"bytes"
	"fmt"
	"os"
	"os/exec"
	"runtime"
	"strings"
)

type Executor struct {
	shell string
}

func NewExecutor() *Executor {
	shell := "/bin/sh"
	if runtime.GOOS == "windows" {
		shell = "cmd.exe"
	}
	
	return &Executor{
		shell: shell,
	}
}

func (e *Executor) Execute(command string) (string, error) {
	return e.ExecuteWithEnv(command, nil)
}

func (e *Executor) ExecuteWithEnv(command string, env map[string]string) (string, error) {
	var cmd *exec.Cmd
	
	if runtime.GOOS == "windows" {
		cmd = exec.Command(e.shell, "/C", command)
	} else {
		cmd = exec.Command(e.shell, "-c", command)
	}

	// Set up environment
	if env != nil {
		cmd.Env = os.Environ() // Start with current environment
		for k, v := range env {
			cmd.Env = append(cmd.Env, fmt.Sprintf("%s=%s", k, v))
		}
	}

	var stdout, stderr bytes.Buffer
	cmd.Stdout = &stdout
	cmd.Stderr = &stderr

	err := cmd.Run()
	output := strings.TrimSpace(stdout.String())
	if err != nil {
		errStr := strings.TrimSpace(stderr.String())
		if errStr != "" {
			output = errStr
		}
		return output, err
	}

	return output, nil
}
