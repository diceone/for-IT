package api

import (
	"fmt"
	"io/ioutil"
	"log"
	"path/filepath"
	"strings"
	"sync"
	"time"

	"github.com/fsnotify/fsnotify"
	"gopkg.in/yaml.v3"
)

type EnvironmentManager struct {
	baseDir    string
	watcher    *fsnotify.Watcher
	playbooks  map[string]map[string][]Playbook // customer -> environment -> playbooks
	mu         sync.RWMutex
}

func NewEnvironmentManager(baseDir string) (*EnvironmentManager, error) {
	watcher, err := fsnotify.NewWatcher()
	if err != nil {
		return nil, fmt.Errorf("failed to create watcher: %v", err)
	}

	em := &EnvironmentManager{
		baseDir:   baseDir,
		watcher:   watcher,
		playbooks: make(map[string]map[string][]Playbook),
	}

	// Load all environments
	if err := em.loadAllEnvironments(); err != nil {
		return nil, fmt.Errorf("failed to load environments: %v", err)
	}

	// Watch for changes
	go em.watchForChanges()

	return em, nil
}

func (em *EnvironmentManager) loadAllEnvironments() error {
	// Get customer directories
	customers, err := ioutil.ReadDir(em.baseDir)
	if err != nil {
		return fmt.Errorf("failed to read base directory: %v", err)
	}

	for _, customer := range customers {
		if !customer.IsDir() {
			continue
		}

		customerDir := filepath.Join(em.baseDir, customer.Name())
		files, err := ioutil.ReadDir(customerDir)
		if err != nil {
			return fmt.Errorf("failed to read customer directory %s: %v", customer.Name(), err)
		}

		for _, file := range files {
			if !file.IsDir() && strings.HasSuffix(file.Name(), ".yml") {
				env := strings.TrimSuffix(file.Name(), ".yml")
				if err := em.loadEnvironment(customer.Name(), env); err != nil {
					return fmt.Errorf("failed to load environment %s/%s: %v", customer.Name(), env, err)
				}
			}
		}

		// Watch customer directory and its subdirectories
		if err := em.watcher.Add(customerDir); err != nil {
			return fmt.Errorf("failed to watch customer directory %s: %v", customer.Name(), err)
		}
	}

	return nil
}

func (em *EnvironmentManager) loadEnvironment(customer, env string) error {
	filePath := filepath.Join(em.baseDir, customer, env+".yml")
	data, err := ioutil.ReadFile(filePath)
	if err != nil {
		return err
	}

	var config struct {
		Playbooks map[string]struct {
			Name        string   `yaml:"name"`
			Description string   `yaml:"description"`
			Hosts       []string `yaml:"hosts"`
			IncludeRoles []string `yaml:"include_roles"`
		} `yaml:"playbooks"`
	}

	if err := yaml.Unmarshal(data, &config); err != nil {
		return err
	}

	em.mu.Lock()
	defer em.mu.Unlock()

	if em.playbooks[customer] == nil {
		em.playbooks[customer] = make(map[string][]Playbook)
	}

	var playbooks []Playbook
	for name, pb := range config.Playbooks {
		playbook := Playbook{
			Name:        name,
			Description: pb.Description,
			Hosts:       pb.Hosts,
			Tasks:       []Task{},
		}

		// Load tasks from each included role
		for _, roleName := range pb.IncludeRoles {
			tasks, err := em.loadRole(customer, roleName)
			if err != nil {
				log.Printf("Warning: failed to load role %s: %v", roleName, err)
				continue
			}
			playbook.Tasks = append(playbook.Tasks, tasks...)
		}

		playbooks = append(playbooks, playbook)
	}
	em.playbooks[customer][env] = playbooks

	return nil
}

func (em *EnvironmentManager) loadRole(customer, roleName string) ([]Task, error) {
	// First try customer-specific role
	rolePath := filepath.Join(em.baseDir, customer, "roles", roleName, "tasks.yml")
	data, err := ioutil.ReadFile(rolePath)
	if err != nil {
		// If not found in customer directory, try common roles
		rolePath = filepath.Join(em.baseDir, "roles", roleName, "tasks.yml")
		data, err = ioutil.ReadFile(rolePath)
		if err != nil {
			return nil, fmt.Errorf("role %s not found in customer or common directories", roleName)
		}
	}

	var roleConfig struct {
		Tasks []Task `yaml:"tasks"`
	}

	if err := yaml.Unmarshal(data, &roleConfig); err != nil {
		return nil, fmt.Errorf("failed to parse role %s: %v", roleName, err)
	}

	return roleConfig.Tasks, nil
}

func (em *EnvironmentManager) GetPlaybooksForHost(customer, env, hostname string) []Playbook {
	em.mu.RLock()
	defer em.mu.RUnlock()

	if em.playbooks[customer] == nil || em.playbooks[customer][env] == nil {
		return nil
	}

	var matchingPlaybooks []Playbook
	for _, playbook := range em.playbooks[customer][env] {
		for _, pattern := range playbook.Hosts {
			if glob.MustCompile(pattern).Match(hostname) {
				matchingPlaybooks = append(matchingPlaybooks, playbook)
				break
			}
		}
	}

	return matchingPlaybooks
}

func (em *EnvironmentManager) watchForChanges() {
	// Use a timer to debounce rapid file changes
	var debounceTimer *time.Timer
	const debounceDelay = 2 * time.Second

	for {
		select {
		case event, ok := <-em.watcher.Events:
			if !ok {
				return
			}

			// Check if the file is a YAML file
			if !strings.HasSuffix(event.Name, ".yaml") && !strings.HasSuffix(event.Name, ".yml") {
				continue
			}

			// Reset or create the debounce timer
			if debounceTimer != nil {
				debounceTimer.Stop()
			}
			debounceTimer = time.AfterFunc(debounceDelay, func() {
				log.Printf("Detected change in %s, reloading environments", event.Name)
				if err := em.loadAllEnvironments(); err != nil {
					log.Printf("Error reloading environments: %v", err)
				}
			})

		case err, ok := <-em.watcher.Errors:
			if !ok {
				return
			}
			log.Printf("Error watching for file changes: %v", err)
		}
	}
}

func (em *EnvironmentManager) Close() error {
	return em.watcher.Close()
}
