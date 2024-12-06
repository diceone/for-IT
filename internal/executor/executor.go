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

	// Set up basic environment
	if env == nil {
		env = make(map[string]string)
	}

	// Ensure full PATH is available
	env["PATH"] = "/usr/local/sbin:/usr/local/bin:/usr/sbin:/usr/bin:/sbin:/bin"

	// Handle different package managers
	switch {
	case strings.Contains(command, "apt-get") || strings.Contains(command, "apt"):
		// Debian/Ubuntu package management
		env["DEBIAN_FRONTEND"] = "noninteractive"
		env["DEBIAN_PRIORITY"] = "critical"

		// Add -y flag to apt-get if not present
		if strings.Contains(command, "apt-get") && !strings.Contains(command, " -y ") && !strings.HasSuffix(command, " -y") {
			command = strings.Replace(command, "apt-get ", "apt-get -y ", 1)
		}

		// Add -y flag to apt if not present
		if strings.Contains(command, "apt ") && !strings.Contains(command, " -y ") && !strings.HasSuffix(command, " -y") {
			command = strings.Replace(command, "apt ", "apt -y ", 1)
		}

		// For install commands, add additional options
		if strings.Contains(command, "install") {
			if !strings.Contains(command, "-o Dpkg::Options") {
				command = command + " -o Dpkg::Options::=\"--force-confdef\" -o Dpkg::Options::=\"--force-confold\""
			}
		}

	case strings.Contains(command, "yum"):
		// RHEL/CentOS package management (older versions)
		if !strings.Contains(command, " -y ") && !strings.HasSuffix(command, " -y") {
			if strings.Contains(command, "yum ") {
				command = strings.Replace(command, "yum ", "yum -y ", 1)
			}
		}

	case strings.Contains(command, "dnf"):
		// RHEL/CentOS package management (newer versions)
		if !strings.Contains(command, " -y ") && !strings.HasSuffix(command, " -y") {
			if strings.Contains(command, "dnf ") {
				command = strings.Replace(command, "dnf ", "dnf -y ", 1)
			}
		}
	}

	// Set up full environment
	cmd.Env = os.Environ() // Start with current environment
	for k, v := range env {
		cmd.Env = append(cmd.Env, fmt.Sprintf("%s=%s", k, v))
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
		return output, fmt.Errorf("command failed: %v\nOutput: %s", err, output)
	}

	return output, nil
}
