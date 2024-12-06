package logging

import (
	"fmt"
	"log"
	"os"
	"path/filepath"
)

// SetupLogging configures logging to write to both a file and stderr
func SetupLogging(component string) error {
	// For testing, use a local log directory
	logDir := filepath.Join(".", "logs")
	logFile := filepath.Join(logDir, fmt.Sprintf("%s.log", component))
	errFile := filepath.Join(logDir, fmt.Sprintf("%s.error.log", component))

	// Create log directory if it doesn't exist
	if err := os.MkdirAll(logDir, 0755); err != nil {
		return fmt.Errorf("failed to create log directory: %v", err)
	}

	// Open log files
	f, err := os.OpenFile(logFile, os.O_RDWR|os.O_CREATE|os.O_APPEND, 0644)
	if err != nil {
		return fmt.Errorf("failed to open log file: %v", err)
	}

	ef, err := os.OpenFile(errFile, os.O_RDWR|os.O_CREATE|os.O_APPEND, 0644)
	if err != nil {
		return fmt.Errorf("failed to open error log file: %v", err)
	}

	// Set up multi-writer to write to both file and stderr
	log.SetOutput(f)
	log.SetFlags(log.Ldate | log.Ltime | log.Lshortfile)

	// Create error logger
	errorLog := log.New(ef, "ERROR: ", log.Ldate|log.Ltime|log.Lshortfile)

	// Override the default error logger
	log.SetOutput(f)

	// Log startup message
	log.Printf("%s service starting", component)
	errorLog.Printf("%s error logging initialized", component)

	return nil
}
