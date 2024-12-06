package logging

import (
	"fmt"
	"log"
	"os"
	"path/filepath"
)

// SetupLogging configures logging to write to both a file and stderr
func SetupLogging(component string) error {
	// Try different log directories in order of preference
	logDirs := []string{
		"/var/log/for",           // Production directory
		"/tmp/for/log",           // Fallback for testing
		filepath.Join(os.TempDir(), "for", "log"), // Universal fallback
	}

	var logDir string
	var err error

	// Try each directory until we find one we can use
	for _, dir := range logDirs {
		err = os.MkdirAll(dir, 0755)
		if err == nil {
			logDir = dir
			break
		}
	}

	if logDir == "" {
		return fmt.Errorf("failed to create any log directory: %v", err)
	}

	logFile := filepath.Join(logDir, fmt.Sprintf("%s.log", component))
	errFile := filepath.Join(logDir, fmt.Sprintf("%s.error.log", component))

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

	// Log startup message
	log.Printf("%s service starting, logging to %s", component, logDir)
	errorLog.Printf("%s error logging initialized", component)

	return nil
}
