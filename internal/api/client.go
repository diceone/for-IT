package api

import (
	"bytes"
	"encoding/json"
	"fmt"
	"log"
	"net/http"
	"os"
	"time"

	"github.com/diceone/for-IT/internal/executor"
	"github.com/diceone/for-IT/internal/models"
	"github.com/diceone/for-IT/internal/output"
	"github.com/gobwas/glob"
)

type Client struct {
	serverAddr     string
	executor       *executor.Executor
	client         *http.Client
	hostname       string
	customer       string
	environment    string
	checkInterval  time.Duration
	dryRun         bool
}

func NewClient(serverAddr string, checkInterval time.Duration, customer string, environment string) (*Client, error) {
	hostname, err := os.Hostname()
	if err != nil {
		return nil, fmt.Errorf("failed to get hostname: %v", err)
	}

	return &Client{
		serverAddr:    serverAddr,
		executor:      executor.NewExecutor(),
		client:        &http.Client{},
		hostname:      hostname,
		customer:      customer,
		environment:   environment,
		checkInterval: checkInterval,
	}, nil
}

func (c *Client) SetDryRun(enabled bool) {
	c.dryRun = enabled
}

func (c *Client) Start() error {
	for {
		if err := c.CheckAndExecute(); err != nil {
			log.Printf("Error checking tasks: %v", err)
		}

		if c.checkInterval == 0 {
			break
		}
		time.Sleep(c.checkInterval)
	}
	return nil
}

func (c *Client) CheckAndExecute() error {
	tasks, _, err := c.getTasks(c.hostname)
	if err != nil {
		return fmt.Errorf("failed to get tasks: %v", err)
	}

	if len(tasks) == 0 {
		return nil
	}

	var results []models.TaskResult
	startTime := time.Now()

	for _, task := range tasks {
		result := &models.TaskResult{
			Name: task.Name,
		}
		taskStartTime := time.Now()

		if task.When != "" {
			pattern := glob.MustCompile(task.When)
			if !pattern.Match(c.hostname) {
				result.SkipReason = fmt.Sprintf("Condition '%s' not met", task.When)
				results = append(results, *result)
				fmt.Print(output.FormatTaskOutput(task.Name, *result, c.dryRun))
				continue
			}
		}

		if err := c.executeTask(task, result); err != nil {
			result.Failed = true
			result.Error = err.Error()
		}

		result.Duration = time.Since(taskStartTime)
		results = append(results, *result)

		fmt.Print(output.FormatTaskOutput(task.Name, *result, c.dryRun))
	}

	duration := time.Since(startTime)
	fmt.Print(output.FormatPlaybookSummary(results, duration, c.dryRun))

	if err := c.sendResult(results); err != nil {
		return fmt.Errorf("failed to send results: %v", err)
	}

	return nil
}

func (c *Client) executeTask(task models.Task, result *models.TaskResult) error {
	if c.dryRun {
		result.Output = fmt.Sprintf("Would execute: %s", task.Command)
		return nil
	}

	output, err := c.executor.ExecuteWithEnv(task.Command, task.Variables)
	if err != nil {
		return err
	}

	result.Output = output
	result.Changed = true
	return nil
}

func (c *Client) getTasks(hostname string) ([]models.Task, string, error) {
	url := fmt.Sprintf("http://%s/tasks?hostname=%s&customer=%s&environment=%s",
		c.serverAddr, hostname, c.customer, c.environment)
	
	resp, err := c.client.Get(url)
	if err != nil {
		return nil, "", err
	}
	defer resp.Body.Close()

	if resp.StatusCode == http.StatusNotModified {
		return nil, resp.Header.Get("ETag"), nil
	}

	if resp.StatusCode != http.StatusOK {
		return nil, "", fmt.Errorf("unexpected status code: %d", resp.StatusCode)
	}

	var tasks []models.Task
	if err := json.NewDecoder(resp.Body).Decode(&tasks); err != nil {
		return nil, "", err
	}

	return tasks, resp.Header.Get("ETag"), nil
}

func (c *Client) sendResult(result interface{}) error {
	data, err := json.Marshal(result)
	if err != nil {
		return err
	}

	url := fmt.Sprintf("http://%s/results?hostname=%s", c.serverAddr, c.hostname)
	resp, err := c.client.Post(url, "application/json", bytes.NewReader(data))
	if err != nil {
		return err
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusOK {
		return fmt.Errorf("unexpected status code: %d", resp.StatusCode)
	}

	return nil
}
