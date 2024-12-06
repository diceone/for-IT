package api

import (
	"crypto/sha256"
	"encoding/hex"
	"encoding/json"
	"fmt"
	"log"
	"net/http"
	"path/filepath"
	"sync"
)

type Server struct {
	envManager *EnvironmentManager
	inventory  *InventoryManager
	mu         sync.RWMutex
	addr       string
}

type ClientInfo struct {
	Hostname string `json:"hostname"`
}

func NewServer(addr, dataDir string) (*Server, error) {
	envManager, err := NewEnvironmentManager(filepath.Join(dataDir, "environments"))
	if err != nil {
		return nil, fmt.Errorf("failed to create environment manager: %v", err)
	}

	inventory, err := NewInventoryManager(dataDir)
	if err != nil {
		return nil, fmt.Errorf("failed to create inventory manager: %v", err)
	}

	return &Server{
		envManager: envManager,
		inventory:  inventory,
		addr:       addr,
	}, nil
}

func (s *Server) Start() error {
	mux := http.NewServeMux()
	
	// Playbooks endpoint
	mux.HandleFunc("/playbooks", func(w http.ResponseWriter, r *http.Request) {
		hostname := r.URL.Query().Get("hostname")
		if hostname == "" {
			http.Error(w, "hostname parameter required", http.StatusBadRequest)
			return
		}

		// Update inventory with client connection
		if err := s.inventory.UpdateClient(hostname, r.RemoteAddr); err != nil {
			log.Printf("Failed to update inventory for %s: %v", hostname, err)
		}

		playbooks := s.envManager.GetPlaybooksForHost(hostname)
		
		// Generate ETag for the playbooks
		playbooksJSON, err := json.Marshal(playbooks)
		if err != nil {
			http.Error(w, "Failed to marshal playbooks", http.StatusInternalServerError)
			return
		}

		hash := sha256.Sum256(playbooksJSON)
		etag := hex.EncodeToString(hash[:])
		w.Header().Set("ETag", etag)

		// Check If-None-Match header
		if r.Header.Get("If-None-Match") == etag {
			w.WriteHeader(http.StatusNotModified)
			return
		}

		w.Header().Set("Content-Type", "application/json")
		json.NewEncoder(w).Encode(playbooks)
	})

	// Inventory endpoint
	mux.HandleFunc("/inventory", func(w http.ResponseWriter, r *http.Request) {
		if r.Method != http.MethodGet {
			http.Error(w, "method not allowed", http.StatusMethodNotAllowed)
			return
		}

		inventory := s.inventory.GetInventory()
		w.Header().Set("Content-Type", "application/json")
		json.NewEncoder(w).Encode(inventory)
	})

	// Results endpoint
	mux.HandleFunc("/results", func(w http.ResponseWriter, r *http.Request) {
		if r.Method != http.MethodPost {
			http.Error(w, "method not allowed", http.StatusMethodNotAllowed)
			return
		}

		var result struct {
			Hostname    string `json:"hostname"`
			PlaybookName string `json:"playbook_name"`
			TaskName    string `json:"task_name"`
			Output      string `json:"output"`
			Error       string `json:"error,omitempty"`
		}

		if err := json.NewDecoder(r.Body).Decode(&result); err != nil {
			http.Error(w, err.Error(), http.StatusBadRequest)
			return
		}

		// Update inventory when receiving results
		if err := s.inventory.UpdateClient(result.Hostname, r.RemoteAddr); err != nil {
			log.Printf("Failed to update inventory for %s: %v", result.Hostname, err)
		}

		log.Printf("Received result from %s for task %s", result.Hostname, result.TaskName)
		if result.Error != "" {
			log.Printf("Task error from %s: %s", result.Hostname, result.Error)
		}

		w.WriteHeader(http.StatusOK)
	})

	log.Printf("Starting server on %s", s.addr)
	return http.ListenAndServe(s.addr, mux)
}

func (s *Server) Stop() error {
	return s.envManager.Close()
}
