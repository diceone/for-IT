package models

import "time"

// Task represents a single task to be executed
type Task struct {
	Name        string            `json:"name" yaml:"name"`
	Command     string            `json:"command" yaml:"command"`
	When        string            `json:"when,omitempty" yaml:"when,omitempty"`
	Variables   map[string]string `json:"variables,omitempty" yaml:"variables,omitempty"`
	Description string            `json:"description,omitempty" yaml:"description,omitempty"`
	Env         map[string]string `json:"env,omitempty" yaml:"env,omitempty"`
}

// Playbook represents a collection of tasks
type Playbook struct {
	Name        string   `json:"name" yaml:"name"`
	Description string   `json:"description" yaml:"description"`
	Hosts       []string `json:"hosts" yaml:"hosts"`
	Customer    string   `json:"customer" yaml:"customer"`
	Environment string   `json:"environment" yaml:"environment"`
	Tasks       []Task   `json:"tasks" yaml:"tasks"`
}

// Environment represents a collection of playbooks and their configurations
type Environment struct {
	Name        string               `yaml:"name" json:"name"`
	Description string               `yaml:"description" json:"description"`
	Variables   map[string]string    `yaml:"variables,omitempty" json:"variables,omitempty"`
	Playbooks   map[string]Playbook  `yaml:"playbooks" json:"playbooks"`
}

// TaskResult represents the result of executing a task
type TaskResult struct {
	Name       string        `json:"name"`
	Changed    bool          `json:"changed"`
	Failed     bool          `json:"failed"`
	SkipReason string        `json:"skip_reason,omitempty"`
	Output     string        `json:"output"`
	Duration   time.Duration `json:"duration"`
	Error      string        `json:"error,omitempty"`
}
