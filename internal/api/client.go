package api

import (
	"bytes"
	"encoding/json"
	"fmt"
	"log"
	"net/http"
	"os"
	"time"

	"for/internal/executor"
	"for/internal/models"
	"github.com/gobwas/glob"
)

type Client struct {
	serverAddr     string
	executor       *executor.Executor
	client         *http.Client
	hostname       string
	checkInterval  time.Duration
	lastETag       string
	dryRun         bool
	customer       string
	environment    string
}

type TaskResult struct {
	Name        string
	Changed     bool
	Failed      bool
	SkipReason  string
	Output      string
	Duration    time.Duration
	Error       string
}

func NewClient(serverAddr string, checkInterval time.Duration, customer string, environment string) (*Client, error) {
	hostname, err := os.Hostname()
	if err != nil {
		return nil, fmt.Errorf("failed to get hostname: %v", err)
	}

	if customer == "" {
		return nil, fmt.Errorf("customer name is required")
	}

	if environment == "" {
		return nil, fmt.Errorf("environment name is required")
	}

	return &Client{
		serverAddr:    serverAddr,
		executor:      executor.NewExecutor(),
		client:        &http.Client{Timeout: 30 * time.Second},
		hostname:      hostname,
		checkInterval: checkInterval,
		customer:      customer,
		environment:   environment,
		dryRun:        false,
	}, nil
}

func (c *Client) SetDryRun(enabled bool) {
	c.dryRun = enabled
}

func (c *Client) Start() error {
	ticker := time.NewTicker(c.checkInterval)
	defer ticker.Stop()

	for {
		if err := c.CheckAndExecute(); err != nil {
			log.Printf("Error checking for tasks: %v", err)
		}
		<-ticker.C
	}
}

func (c *Client) CheckAndExecute() error {
	hostname, err := os.Hostname()
	if err != nil {
		return fmt.Errorf("failed to get hostname: %v", err)
	}

	startTime := time.Now()
	fmt.Printf("\nPLAY [%s] ******************************************************************\n", hostname)

	// Get tasks from server
	tasks, newETag, err := c.getTasks(hostname)
	if err != nil {
		return err
	}

	// If ETag matches, no changes needed
	if newETag == c.lastETag {
		if c.dryRun {
			fmt.Println("No changes needed (check mode)")
		} else {
			fmt.Println("No changes needed")
		}
		return nil
	}

	// Track results for summary
	var results []TaskResult

	// Execute each task
	for _, task := range tasks {
		var result TaskResult
		result.Name = task.Name
		
		if c.dryRun {
			result.Changed = true // Assume change in dry run
			results = append(results, result)
			fmt.Println(formatTaskOutput(task.Name, result, c.dryRun))
		} else {
			if err := c.executeTask(task, &result); err != nil {
				return fmt.Errorf("failed to execute task %s: %v", task.Name, err)
			}
			results = append(results, result)
		}
	}

	// Print summary
	duration := time.Since(startTime)
	fmt.Println(formatPlaybookSummary(results, duration, c.dryRun))

	c.lastETag = newETag
	return nil
}

func (c *Client) executeTask(task models.Task, result *TaskResult) error {
	startTime := time.Now()
	result.Name = task.Name

	// Check conditions if present
	if task.When != "" {
		g, err := glob.Compile(task.When)
		if err != nil {
			result.Failed = true
			result.Error = fmt.Sprintf("Invalid condition: %v", err)
			fmt.Println(formatTaskOutput(task.Name, *result, c.dryRun))
			return nil
		}
		if !g.Match(c.hostname) {
			result.SkipReason = fmt.Sprintf("Condition not met: %s", task.When)
			fmt.Println(formatTaskOutput(task.Name, *result, c.dryRun))
			return nil
		}
	}

	if c.dryRun {
		result.Changed = true // Assume change in dry run
		fmt.Println(formatTaskOutput(task.Name, *result, c.dryRun))
		return nil
	}

	// Execute the task
	output, err := c.executor.ExecuteWithEnv(task.Command, task.Env)
	result.Duration = time.Since(startTime)
	result.Output = output
	
	if err != nil {
		result.Failed = true
		result.Error = err.Error()
	} else {
		result.Changed = true // Assume command made changes
	}

	// Print formatted output
	fmt.Println(formatTaskOutput(task.Name, *result, c.dryRun))

	// Send result to server
	return c.sendResult(result)
}

func (c *Client) getTasks(hostname string) ([]models.Task, string, error) {
	url := fmt.Sprintf("http://%s/playbooks?hostname=%s&customer=%s&environment=%s",
		c.serverAddr, hostname, c.customer, c.environment)
	
	resp, err := c.client.Get(url)
	if err != nil {
		return nil, "", fmt.Errorf("failed to fetch playbooks: %v", err)
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusOK {
		return nil, "", fmt.Errorf("unexpected status code: %d", resp.StatusCode)
	}

	var playbooks []models.Playbook
	if err := json.NewDecoder(resp.Body).Decode(&playbooks); err != nil {
		return nil, "", fmt.Errorf("failed to decode playbooks: %v", err)
	}

	var tasks []models.Task
	for _, playbook := range playbooks {
		tasks = append(tasks, playbook.Tasks...)
	}

	return tasks, resp.Header.Get("ETag"), nil
}

func (c *Client) sendResult(result interface{}) error {
	data, err := json.Marshal(result)
	if err != nil {
		return err
	}

	resp, err := c.client.Post(
		fmt.Sprintf("http://%s/results", c.serverAddr),
		"application/json",
		bytes.NewBuffer(data),
	)
	if err != nil {
		return err
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusOK {
		return fmt.Errorf("unexpected status code: %d", resp.StatusCode)
	}

	return nil
}

func formatTaskOutput(taskName string, result TaskResult, dryRun bool) string {
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

func formatPlaybookSummary(results []TaskResult, duration time.Duration, dryRun bool) string {
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
