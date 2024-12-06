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
	files, err := os.ReadDir(s.playbookDir)
	if err != nil {
		return fmt.Errorf("failed to read playbook directory: %v", err)
	}

	for _, file := range files {
		if !file.IsDir() && strings.HasSuffix(file.Name(), ".yml") {
			if err := s.loadPlaybook(file.Name()); err != nil {
				log.Printf("Error loading playbook %s: %v", file.Name(), err)
				continue
			}
		}
	}

	return nil
}

func (s *Server) loadPlaybook(filename string) error {
	path := filepath.Join(s.playbookDir, filename)

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

	return nil
}

func (s *Server) watchPlaybooks() {
	if err := s.watcher.Add(s.playbookDir); err != nil {
		log.Printf("Error watching playbook directory: %v", err)
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

			filename := filepath.Base(event.Name)
			switch event.Op {
			case fsnotify.Write, fsnotify.Create:
				if err := s.loadPlaybook(filename); err != nil {
					log.Printf("Error reloading playbook %s: %v", filename, err)
				}
			case fsnotify.Remove:
				s.mutex.Lock()
				delete(s.playbooks, filename)
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
