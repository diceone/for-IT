package main

import (
	"flag"
	"log"
	"os"
	"os/signal"
	"syscall"

	"github.com/mvogeler/for/internal/api"
)

func main() {
	addr := flag.String("addr", ":8080", "Server address")
	dataDir := flag.String("data-dir", "/etc/for", "Data directory")
	flag.Parse()

	// Create server instance
	server, err := api.NewServer(*addr, *dataDir)
	if err != nil {
		log.Fatalf("Failed to create server: %v", err)
	}

	// Handle graceful shutdown
	sigChan := make(chan os.Signal, 1)
	signal.Notify(sigChan, syscall.SIGINT, syscall.SIGTERM)

	go func() {
		sig := <-sigChan
		log.Printf("Received signal %v, shutting down...", sig)
		if err := server.Stop(); err != nil {
			log.Printf("Error during shutdown: %v", err)
		}
		os.Exit(0)
	}()

	// Start server
	log.Fatal(server.Start())
}
