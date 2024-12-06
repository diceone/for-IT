package main

import (
	"flag"
	"log"
	"os"
	"os/signal"
	"path/filepath"
	"syscall"

	"github.com/diceone/for-IT/internal/api"
)

func main() {
	var (
		addr       = flag.String("addr", ":8080", "Server address")
		playbookDir = flag.String("playbook-dir", "playbooks", "Directory containing playbook files")
	)
	flag.Parse()

	// Ensure absolute path for playbook directory
	absPlaybookDir, err := filepath.Abs(*playbookDir)
	if err != nil {
		log.Fatalf("Failed to get absolute path for playbook directory: %v", err)
	}

	// Create and start server
	server, err := api.NewServer(absPlaybookDir)
	if err != nil {
		log.Fatalf("Failed to create server: %v", err)
	}

	// Handle shutdown gracefully
	sigChan := make(chan os.Signal, 1)
	signal.Notify(sigChan, syscall.SIGINT, syscall.SIGTERM)
	errChan := make(chan error, 1)

	// Start server in a goroutine
	go func() {
		log.Printf("Starting server on %s", *addr)
		if err := server.Start(*addr); err != nil {
			errChan <- err
		}
	}()

	// Wait for shutdown signal
	select {
	case <-sigChan:
		log.Println("Shutting down server...")
	case err := <-errChan:
		log.Printf("Server error: %v", err)
	}
}
