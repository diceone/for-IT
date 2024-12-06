package api

import (
	"encoding/json"
	"fmt"
	"io/ioutil"
	"net"
	"os"
	"path/filepath"
	"sync"
	"time"
)

// InventoryEntry represents a client in the inventory
type InventoryEntry struct {
	Hostname    string    `json:"hostname"`
	IP          string    `json:"ip"`
	LastSeen    time.Time `json:"last_seen"`
	FirstSeen   time.Time `json:"first_seen"`
	Environment string    `json:"environment,omitempty"`
}

// InventoryManager handles the server's inventory of clients
type InventoryManager struct {
	inventoryFile string
	entries       map[string]InventoryEntry // hostname -> entry
	mu           sync.RWMutex
}

// NewInventoryManager creates a new inventory manager
func NewInventoryManager(dataDir string) (*InventoryManager, error) {
	inventoryFile := filepath.Join(dataDir, "inventory.json")
	
	im := &InventoryManager{
		inventoryFile: inventoryFile,
		entries:      make(map[string]InventoryEntry),
	}

	// Create data directory if it doesn't exist
	if err := os.MkdirAll(dataDir, 0755); err != nil {
		return nil, fmt.Errorf("failed to create data directory: %v", err)
	}

	// Load existing inventory if it exists
	if err := im.load(); err != nil && !os.IsNotExist(err) {
		return nil, fmt.Errorf("failed to load inventory: %v", err)
	}

	return im, nil
}

// UpdateClient updates or adds a client in the inventory
func (im *InventoryManager) UpdateClient(hostname, remoteAddr string) error {
	im.mu.Lock()
	defer im.mu.Unlock()

	// Extract IP from remoteAddr (removes port)
	ip, _, err := net.SplitHostPort(remoteAddr)
	if err != nil {
		ip = remoteAddr // fallback to full address if parsing fails
	}

	now := time.Now()
	entry, exists := im.entries[hostname]
	
	if !exists {
		entry = InventoryEntry{
			Hostname:  hostname,
			IP:        ip,
			FirstSeen: now,
		}
	} else {
		entry.IP = ip // Update IP in case it changed
	}
	
	entry.LastSeen = now
	im.entries[hostname] = entry

	return im.save()
}

// GetInventory returns all inventory entries
func (im *InventoryManager) GetInventory() []InventoryEntry {
	im.mu.RLock()
	defer im.mu.RUnlock()

	entries := make([]InventoryEntry, 0, len(im.entries))
	for _, entry := range im.entries {
		entries = append(entries, entry)
	}
	return entries
}

// load reads the inventory from disk
func (im *InventoryManager) load() error {
	data, err := ioutil.ReadFile(im.inventoryFile)
	if err != nil {
		return err
	}

	im.mu.Lock()
	defer im.mu.Unlock()

	return json.Unmarshal(data, &im.entries)
}

// save writes the inventory to disk
func (im *InventoryManager) save() error {
	data, err := json.MarshalIndent(im.entries, "", "  ")
	if err != nil {
		return fmt.Errorf("failed to marshal inventory: %v", err)
	}

	return ioutil.WriteFile(im.inventoryFile, data, 0644)
}
