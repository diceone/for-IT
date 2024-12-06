package main

import (
	"flag"
	"log"
	"os"
	"time"

	"github.com/diceone/for-IT/internal/api"
	"github.com/diceone/for-IT/internal/logging"
)

func main() {
	serverAddr := flag.String("server", "localhost:8080", "Server address")
	checkInterval := flag.Duration("interval", 30*time.Minute, "Check interval")
	dryRun := flag.Bool("dry-run", false, "Show what would be executed without making changes")
	runOnce := flag.Bool("run-once", false, "Run once and exit")
	customer := flag.String("customer", "", "Customer name (required)")
	environment := flag.String("environment", "", "Environment name (required)")
	debug := flag.Bool("debug", true, "Enable debug logging")
	flag.Parse()

	// Setup logging
	if err := logging.SetupLogging("client"); err != nil {
		log.Fatalf("Failed to setup logging: %v", err)
	}

	if *debug {
		log.SetFlags(log.Ltime | log.Lshortfile)
	}

	if *customer == "" || *environment == "" {
		log.Fatal("Customer and environment parameters are required")
	}

	log.Printf("Connecting to server at %s (customer: %s, environment: %s)", *serverAddr, *customer, *environment)

	client, err := api.NewClient(*serverAddr, *checkInterval, *customer, *environment)
	if err != nil {
		log.Fatalf("Failed to create client: %v", err)
	}

	// Set dry run mode if requested
	if *dryRun {
		client.SetDryRun(true)
	}

	// If run-once flag is set, execute once and exit
	if *runOnce {
		log.Printf("Running in one-shot mode")
		err := client.CheckAndExecute()
		if err != nil {
			log.Printf("Error during execution: %v", err)
			os.Exit(1)
		}
		log.Printf("One-shot execution complete")
		os.Exit(0)
	}

	// Otherwise run in continuous mode
	log.Printf("Starting client, connecting to server at %s (check interval: %s)", *serverAddr, *checkInterval)
	log.Fatal(client.Start())
}
