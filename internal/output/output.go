package output

import (
	"fmt"
	"time"

	"github.com/diceone/for-IT/internal/models"
)

func FormatTaskOutput(taskName string, result models.TaskResult, dryRun bool) string {
	output := fmt.Sprintf("TASK [%s] ****************************************************\n", taskName)
	if result.SkipReason != "" {
		output += fmt.Sprintf("skipping: [%s]\n", result.SkipReason)
	} else if result.Failed {
		output += fmt.Sprintf("failed: [%s]\n", result.Error)
	} else if result.Changed {
		output += fmt.Sprintf("changed: [%s]\n", result.Output)
	} else {
		output += "ok\n"
	}
	if dryRun {
		output += "(check mode)\n"
	}
	return output
}

func FormatPlaybookSummary(results []models.TaskResult, duration time.Duration, dryRun bool) string {
	output := fmt.Sprintf("PLAY RECAP *********************************************************************\n")
	for _, result := range results {
		if result.SkipReason != "" {
			output += fmt.Sprintf("%s                : skip=%s\n", result.Name, result.SkipReason)
		} else if result.Failed {
			output += fmt.Sprintf("%s                : failed=%s\n", result.Name, result.Error)
		} else if result.Changed {
			output += fmt.Sprintf("%s                : changed=%s\n", result.Name, result.Output)
		} else {
			output += fmt.Sprintf("%s                : ok\n", result.Name)
		}
	}
	output += fmt.Sprintf("Playbook run took %s\n", duration)
	if dryRun {
		output += "(check mode)\n"
	}
	return output
}
