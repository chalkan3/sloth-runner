package scheduler

import (
	"fmt"
	"io/ioutil"
	"os"
	"os/exec"
	"sync"

	"github.com/robfig/cron/v3"
	"gopkg.in/yaml.v2"
)

// ScheduledTask represents a single task to be scheduled
type ScheduledTask struct {
	Name      string `yaml:"name"`
	Schedule  string `yaml:"schedule"`
	TaskFile  string `yaml:"task_file"`
	TaskGroup string `yaml:"task_group"`
	TaskName  string `yaml:"task_name"`
}

// SchedulerConfig holds the configuration for the scheduler
type SchedulerConfig struct {
	ScheduledTasks []ScheduledTask `yaml:"scheduled_tasks"`
}

// Scheduler manages the cron jobs
type Scheduler struct {
	cron *cron.Cron
	configPath string
	config     *SchedulerConfig
	mu         sync.Mutex
}

// NewScheduler creates a new Scheduler instance
func NewScheduler(configPath string) *Scheduler {
	return &Scheduler{
		cron: cron.New(),
		configPath: configPath,
	}
}

// LoadConfig loads the scheduler configuration from the specified path
func (s *Scheduler) LoadConfig() error {
	s.mu.Lock()
	defer s.mu.Unlock()

	data, err := ioutil.ReadFile(s.configPath)
	if err != nil {
		return fmt.Errorf("failed to read scheduler config file: %w", err)
	}

	var config SchedulerConfig
	if err := yaml.Unmarshal(data, &config); err != nil {
		return fmt.Errorf("failed to unmarshal scheduler config: %w", err)
	}
	s.config = &config
	return nil
}

// Start initializes and starts the cron scheduler
func (s *Scheduler) Start() error {
	s.mu.Lock()
	defer s.mu.Unlock()

	if s.config == nil {
		return fmt.Errorf("scheduler configuration not loaded")
	}

	for _, task := range s.config.ScheduledTasks {
		task := task // capture loop variable
		_, err := s.cron.AddFunc(task.Schedule, func() {
			s.RunTask(task)
		})
		if err != nil {
			return fmt.Errorf("failed to add cron job for task %s: %w", task.Name, err)
		}
		fmt.Printf("Scheduled task '%s' with schedule '%s'\n", task.Name, task.Schedule)
	}

	s.cron.Start()
	fmt.Println("Scheduler started.")
	return nil
}

// Stop stops the cron scheduler
func (s *Scheduler) Stop() {
	s.mu.Lock()
	defer s.mu.Unlock()
	s.cron.Stop()
	fmt.Println("Scheduler stopped.")
}

// Mockable exec.Command for testing
var execCommand = exec.Command

// RunTask executes a sloth-runner task
func (s *Scheduler) RunTask(task ScheduledTask) {
	fmt.Printf("Executing scheduled task '%s' (file: %s, group: %s, task: %s)...\n", task.Name, task.TaskFile, task.TaskGroup, task.TaskName)

	// Assuming sloth-runner executable is in the same directory or in PATH
	cmd := execCommand("sloth-runner", "run", "-f", task.TaskFile, "-g", task.TaskGroup, "-t", task.TaskName)

	cmd.Stdout = os.Stdout
	cmd.Stderr = os.Stderr

	if err := cmd.Run(); err != nil {
		fmt.Printf("Error executing scheduled task '%s': %v\n", task.Name, err)
	} else {
		fmt.Printf("Scheduled task '%s' completed successfully.\n", task.Name)
	}
}

// Config returns the current scheduler configuration
func (s *Scheduler) Config() *SchedulerConfig {
	s.mu.Lock()
	defer s.mu.Unlock()
	return s.config
}

// SetConfig sets the scheduler configuration
func (s *Scheduler) SetConfig(config *SchedulerConfig) {
	s.mu.Lock()
	defer s.mu.Unlock()
	s.config = config
}

// SaveConfig saves the current scheduler configuration to the specified path
func (s *Scheduler) SaveConfig() error {
	s.mu.Lock()
	defer s.mu.Unlock()

	if s.config == nil {
		return fmt.Errorf("no configuration to save")
	}

	data, err := yaml.Marshal(s.config)
	if err != nil {
		return fmt.Errorf("failed to marshal scheduler config: %w", err)
	}

	if err := ioutil.WriteFile(s.configPath, data, 0644); err != nil {
		return fmt.Errorf("failed to write scheduler config file: %w", err)
	}
	return nil
}
