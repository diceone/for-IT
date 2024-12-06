package api

import (
	"encoding/json"
	"fmt"
	"io"
	"log"
	"net/http"
	"os"
	"path/filepath"
	"strings"
	"sync"

	"github.com/diceone/for-IT/internal/models"
	"github.com/fsnotify/fsnotify"
	"gopkg.in/yaml.v3"
)

type Server struct {
	playbookDir string
	playbooks   map[string]models.Playbook
	mutex       sync.RWMutex
	watcher     *fsnotify.Watcher
}

func NewServer(playbookDir string) (*Server, error) {
	watcher, err := fsnotify.NewWatcher()
	if err != nil {
		return nil, fmt.Errorf("failed to create watcher: %v", err)
	}

	s := &Server{
		playbookDir: playbookDir,
		playbooks:   make(map[string]models.Playbook),
		watcher:     watcher,
	}

	if err := s.loadPlaybooks(); err != nil {
		return nil, fmt.Errorf("failed to load playbooks: %v", err)
	}

	go s.watchPlaybooks()

	return s, nil
}

func (s *Server) loadPlaybooks() error {
	err := filepath.Walk(s.playbookDir, func(path string, info os.FileInfo, err error) error {
		if err != nil {
			return err
		}
		if !info.IsDir() && strings.HasSuffix(info.Name(), ".yml") {
			relPath, err := filepath.Rel(s.playbookDir, path)
			if err != nil {
				return fmt.Errorf("failed to get relative path: %v", err)
			}
			if err := s.loadPlaybook(relPath); err != nil {
				log.Printf("Error loading playbook %s: %v", relPath, err)
			}
		}
		return nil
	})

	if err != nil {
		return fmt.Errorf("failed to walk playbook directory: %v", err)
	}

	log.Printf("Loaded %d playbooks", len(s.playbooks))
	return nil
}

func (s *Server) loadPlaybook(filename string) error {
	path := filepath.Join(s.playbookDir, filename)
	log.Printf("Loading playbook: %s", path)

	data, err := os.ReadFile(path)
	if err != nil {
		return fmt.Errorf("failed to read playbook file: %v", err)
	}

	var playbook models.Playbook
	if err := yaml.Unmarshal(data, &playbook); err != nil {
		return fmt.Errorf("failed to unmarshal playbook: %v", err)
	}

	s.mutex.Lock()
	s.playbooks[filename] = playbook
	s.mutex.Unlock()

	log.Printf("Loaded playbook %s: customer=%s, environment=%s, tasks=%d", 
		filename, playbook.Customer, playbook.Environment, len(playbook.Tasks))
	return nil
}

func (s *Server) watchPlaybooks() {
	// Watch the main playbook directory and all subdirectories
	err := filepath.Walk(s.playbookDir, func(path string, info os.FileInfo, err error) error {
		if err != nil {
			return err
		}
		if info.IsDir() {
			if err := s.watcher.Add(path); err != nil {
				log.Printf("Error watching directory %s: %v", path, err)
			} else {
				log.Printf("Watching directory: %s", path)
			}
		}
		return nil
	})

	if err != nil {
		log.Printf("Error setting up directory watchers: %v", err)
		return
	}

	for {
		select {
		case event, ok := <-s.watcher.Events:
			if !ok {
				return
			}
			if !strings.HasSuffix(event.Name, ".yml") {
				continue
			}

			// Get relative path for the playbook
			relPath, err := filepath.Rel(s.playbookDir, event.Name)
			if err != nil {
				log.Printf("Error getting relative path for %s: %v", event.Name, err)
				continue
			}

			switch event.Op {
			case fsnotify.Write, fsnotify.Create:
				log.Printf("Playbook modified: %s", relPath)
				if err := s.loadPlaybook(relPath); err != nil {
					log.Printf("Error reloading playbook %s: %v", relPath, err)
				}
			case fsnotify.Remove, fsnotify.Rename:
				log.Printf("Playbook removed: %s", relPath)
				s.mutex.Lock()
				delete(s.playbooks, relPath)
				s.mutex.Unlock()
			}

		case err, ok := <-s.watcher.Errors:
			if !ok {
				return
			}
			log.Printf("Watcher error: %v", err)
		}
	}
}

func (s *Server) Start(addr string) error {
	http.HandleFunc("/tasks", s.handleTasks)
	http.HandleFunc("/results", s.handleResults)
	return http.ListenAndServe(addr, nil)
}

func (s *Server) handleTasks(w http.ResponseWriter, r *http.Request) {
	if r.Method != http.MethodGet {
		http.Error(w, "Method not allowed", http.StatusMethodNotAllowed)
		return
	}

	hostname := r.URL.Query().Get("hostname")
	if hostname == "" {
		http.Error(w, "Hostname is required", http.StatusBadRequest)
		return
	}

	customer := r.URL.Query().Get("customer")
	if customer == "" {
		http.Error(w, "Customer is required", http.StatusBadRequest)
		return
	}

	environment := r.URL.Query().Get("environment")
	if environment == "" {
		http.Error(w, "Environment is required", http.StatusBadRequest)
		return
	}

	var tasks []models.Task
	s.mutex.RLock()
	for _, playbook := range s.playbooks {
		if playbook.Customer == customer && playbook.Environment == environment {
			tasks = append(tasks, playbook.Tasks...)
		}
	}
	s.mutex.RUnlock()

	w.Header().Set("Content-Type", "application/json")
	if err := json.NewEncoder(w).Encode(tasks); err != nil {
		http.Error(w, fmt.Sprintf("Failed to encode tasks: %v", err), http.StatusInternalServerError)
		return
	}
}

func (s *Server) handleResults(w http.ResponseWriter, r *http.Request) {
	if r.Method != http.MethodPost {
		http.Error(w, "Method not allowed", http.StatusMethodNotAllowed)
		return
	}

	body, err := io.ReadAll(r.Body)
	if err != nil {
		http.Error(w, fmt.Sprintf("Failed to read request body: %v", err), http.StatusBadRequest)
		return
	}
	defer r.Body.Close()

	var results []models.TaskResult
	if err := json.Unmarshal(body, &results); err != nil {
		http.Error(w, fmt.Sprintf("Failed to unmarshal results: %v", err), http.StatusBadRequest)
		return
	}

	// Process results (e.g., log them, store them, etc.)
	for _, result := range results {
		log.Printf("Task: %s, Changed: %v, Failed: %v, Output: %s", 
			result.Name, result.Changed, result.Failed, result.Output)
	}

	w.WriteHeader(http.StatusOK)
}
