package models

// Playbook represents a set of tasks to be executed on specific hosts
type Playbook struct {
	Name        string   `yaml:"name" json:"name"`
	Description string   `yaml:"description" json:"description"`
	Hosts       []string `yaml:"hosts" json:"hosts"`
	Tasks       []Task   `yaml:"tasks" json:"tasks"`
}

// Task represents a single command to be executed
type Task struct {
	Name    string            `yaml:"name" json:"name"`
	Command string            `yaml:"command" json:"command"`
	Env     map[string]string `yaml:"env,omitempty" json:"env,omitempty"`
	When    string           `yaml:"when,omitempty" json:"when,omitempty"`
}

// Environment represents a collection of playbooks and their configurations
type Environment struct {
	Name        string               `yaml:"name" json:"name"`
	Description string               `yaml:"description" json:"description"`
	Playbooks   map[string]Playbook `yaml:"playbooks" json:"playbooks"`
	Variables   map[string]string    `yaml:"variables,omitempty" json:"variables,omitempty"`
}
