package models

import "time"

// Task represents a single task to be executed
type Task struct {
	Name        string            `json:"name" yaml:"name"`
	Command     string            `json:"command" yaml:"command"`
	When        string            `json:"when,omitempty" yaml:"when,omitempty"`
	Variables   map[string]string `json:"variables,omitempty" yaml:"variables,omitempty"`
	Description string            `json:"description,omitempty" yaml:"description,omitempty"`
}

// Playbook represents a collection of tasks
type Playbook struct {
	Name        string   `json:"name" yaml:"name"`
	Description string   `json:"description,omitempty" yaml:"description,omitempty"`
	Hosts       []string `json:"hosts" yaml:"hosts"`
	Tasks       []Task   `json:"tasks" yaml:"tasks"`
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
