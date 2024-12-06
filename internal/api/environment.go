package api

import (
	"fmt"
	"io/fs"
	"log"
	"os"
	"path/filepath"
	"sync"

	"github.com/diceone/for-IT/internal/models"
	"github.com/fsnotify/fsnotify"
	"gopkg.in/yaml.v3"
)

type EnvironmentManager struct {
	environments map[string][]models.Playbook
	mutex        sync.RWMutex
	watcher      *fsnotify.Watcher
	baseDir      string
}

func NewEnvironmentManager(baseDir string) (*EnvironmentManager, error) {
	watcher, err := fsnotify.NewWatcher()
	if err != nil {
		return nil, fmt.Errorf("failed to create watcher: %v", err)
	}

	manager := &EnvironmentManager{
		environments: make(map[string][]models.Playbook),
		watcher:     watcher,
		baseDir:     baseDir,
	}

	if err := manager.loadEnvironments(); err != nil {
		return nil, fmt.Errorf("failed to load environments: %v", err)
	}

	go manager.watchEnvironments()

	return manager, nil
}

func (m *EnvironmentManager) loadEnvironments() error {
	if err := os.MkdirAll(m.baseDir, 0755); err != nil {
		return fmt.Errorf("failed to create base directory: %v", err)
	}

	if err := m.watcher.Add(m.baseDir); err != nil {
		return fmt.Errorf("failed to watch base directory: %v", err)
	}

	entries, err := os.ReadDir(m.baseDir)
	if err != nil {
		return fmt.Errorf("failed to read base directory: %v", err)
	}

	for _, entry := range entries {
		if entry.IsDir() {
			if err := m.loadEnvironment(entry.Name()); err != nil {
				log.Printf("Error loading environment %s: %v", entry.Name(), err)
			}
		}
	}

	return nil
}

func (m *EnvironmentManager) loadEnvironment(envName string) error {
	envDir := filepath.Join(m.baseDir, envName)

	if err := m.watcher.Add(envDir); err != nil {
		return fmt.Errorf("failed to watch environment directory: %v", err)
	}

	var playbooks []models.Playbook
	err := filepath.WalkDir(envDir, func(path string, d fs.DirEntry, err error) error {
		if err != nil {
			return err
		}

		if !d.IsDir() && filepath.Ext(path) == ".yml" {
			playbook, err := m.loadPlaybook(path)
			if err != nil {
				log.Printf("Error loading playbook %s: %v", path, err)
				return nil
			}
			playbooks = append(playbooks, playbook)
		}

		return nil
	})

	if err != nil {
		return fmt.Errorf("failed to walk environment directory: %v", err)
	}

	m.mutex.Lock()
	m.environments[envName] = playbooks
	m.mutex.Unlock()

	return nil
}

func (m *EnvironmentManager) loadPlaybook(path string) (models.Playbook, error) {
	data, err := os.ReadFile(path)
	if err != nil {
		return models.Playbook{}, fmt.Errorf("failed to read playbook file: %v", err)
	}

	var playbook models.Playbook
	if err := yaml.Unmarshal(data, &playbook); err != nil {
		return models.Playbook{}, fmt.Errorf("failed to unmarshal playbook: %v", err)
	}

	return playbook, nil
}

func (m *EnvironmentManager) watchEnvironments() {
	for {
		select {
		case event, ok := <-m.watcher.Events:
			if !ok {
				return
			}

			// Handle directory events
			if filepath.Ext(event.Name) == "" {
				switch event.Op {
				case fsnotify.Create:
					if err := m.loadEnvironment(filepath.Base(event.Name)); err != nil {
						log.Printf("Error loading new environment: %v", err)
					}
				case fsnotify.Remove:
					m.mutex.Lock()
					delete(m.environments, filepath.Base(event.Name))
					m.mutex.Unlock()
				}
				continue
			}

			// Handle playbook file events
			if filepath.Ext(event.Name) == ".yml" {
				envName := filepath.Base(filepath.Dir(event.Name))
				if err := m.loadEnvironment(envName); err != nil {
					log.Printf("Error reloading environment %s: %v", envName, err)
				}
			}

		case err, ok := <-m.watcher.Errors:
			if !ok {
				return
			}
			log.Printf("Watcher error: %v", err)
		}
	}
}

func (m *EnvironmentManager) GetPlaybooksForHost(hostname string) []models.Task {
	var tasks []models.Task
	m.mutex.RLock()
	defer m.mutex.RUnlock()

	for _, playbooks := range m.environments {
		for _, playbook := range playbooks {
			for _, task := range playbook.Tasks {
				if task.When == "" || task.When == hostname {
					tasks = append(tasks, task)
				}
			}
		}
	}

	return tasks
}

func (m *EnvironmentManager) GetEnvironments() map[string][]models.Playbook {
	m.mutex.RLock()
	defer m.mutex.RUnlock()

	environments := make(map[string][]models.Playbook, len(m.environments))
	for env, playbooks := range m.environments {
		environments[env] = playbooks
	}

	return environments
}

func (m *EnvironmentManager) Close() error {
	return m.watcher.Close()
}
