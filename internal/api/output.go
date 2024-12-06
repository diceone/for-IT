package api

import (
	"fmt"
	"strings"
	"time"
)

const (
	ColorReset  = "\033[0m"
	ColorGreen  = "\033[32m"
	ColorYellow = "\033[33m"
	ColorRed    = "\033[31m"
	ColorBlue   = "\033[34m"
	ColorCyan   = "\033[36m"
)

type TaskResult struct {
	Name       string
	Status     string
	Changed    bool
	Failed     bool
	SkipReason string
	Duration   time.Duration
	Output     string
	Error      string
}

// formatTaskOutput formats task output in Ansible-like style
func formatTaskOutput(taskName string, result TaskResult, dryRun bool) string {
	var status, color string
	indent := "        "

	if dryRun {
		if result.SkipReason != "" {
			status = "skipped"
			color = ColorBlue
		} else {
			status = "check mode"
			color = ColorYellow
		}
	} else {
		switch {
		case result.Failed:
			status = "failed"
			color = ColorRed
		case result.Changed:
			status = "changed"
			color = ColorYellow
		case result.SkipReason != "":
			status = "skipped"
			color = ColorBlue
		default:
			status = "ok"
			color = ColorGreen
		}
	}

	// Format task header
	header := fmt.Sprintf("TASK [%s] %s", taskName, strings.Repeat("*", 80-7-len(taskName)))
	
	// Format task result line
	resultLine := fmt.Sprintf("%s%s: [localhost] => %s%s%s", 
		indent, 
		status,
		color,
		formatResultDetails(result, dryRun),
		ColorReset,
	)

	// Format output if present
	var outputLines []string
	if result.Output != "" || result.Error != "" {
		outputLines = append(outputLines, fmt.Sprintf("%sOutput:", indent))
		if result.Output != "" {
			for _, line := range strings.Split(result.Output, "\n") {
				if line != "" {
					outputLines = append(outputLines, fmt.Sprintf("%s  %s", indent, line))
				}
			}
		}
		if result.Error != "" {
			outputLines = append(outputLines, fmt.Sprintf("%s  Error: %s%s%s", 
				indent, 
				ColorRed,
				result.Error,
				ColorReset,
			))
		}
	}

	// Combine all parts
	parts := []string{header, resultLine}
	if len(outputLines) > 0 {
		parts = append(parts, outputLines...)
	}
	return strings.Join(parts, "\n")
}

func formatResultDetails(result TaskResult, dryRun bool) string {
	var details []string

	if dryRun {
		details = append(details, "\"check_mode\": true")
	}
	
	if result.Changed {
		details = append(details, "\"changed\": true")
	}
	
	if result.Failed {
		details = append(details, "\"failed\": true")
	}
	
	if result.SkipReason != "" {
		details = append(details, fmt.Sprintf("\"skip_reason\": \"%s\"", result.SkipReason))
	}
	
	details = append(details, fmt.Sprintf("\"duration\": %.2fs", result.Duration.Seconds()))

	return fmt.Sprintf("{%s}", strings.Join(details, ", "))
}

// formatPlaybookSummary formats the final playbook summary in Ansible style
func formatPlaybookSummary(results []TaskResult, duration time.Duration, dryRun bool) string {
	var ok, changed, failed, skipped int
	
	for _, result := range results {
		switch {
		case result.Failed:
			failed++
		case result.Changed:
			changed++
		case result.SkipReason != "":
			skipped++
		default:
			ok++
		}
	}

	header := "\nPLAY RECAP *********************************************************************"
	recap := fmt.Sprintf("localhost                  : %sok=%d    changed=%d    failed=%d    skipped=%d%s",
		ColorGreen, ok,
		ColorYellow, changed,
		ColorRed, failed,
		ColorBlue, skipped,
		ColorReset,
	)
	
	timing := fmt.Sprintf("\nPlaybook finished in %.2f seconds", duration.Seconds())
	
	if dryRun {
		return fmt.Sprintf("%s\n%s%s\n*** Playbook run in check mode ***", header, recap, timing)
	}
	return fmt.Sprintf("%s\n%s%s", header, recap, timing)
}
