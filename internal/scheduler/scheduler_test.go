package scheduler

import (
	"fmt"
	"io/ioutil"
	"os"
	"os/exec"
	"testing"
	"time"

	"github.com/stretchr/testify/assert"
)

const testSchedulerConfig = `
scheduled_tasks:
  - name: "test_task_1"
    schedule: "@every 1s"
    task_file: "test_file_1.lua"
    task_group: "test_group_1"
    task_name: "task_1"
  - name: "test_task_2"
    schedule: "@every 2s"
    task_file: "test_file_2.lua"
    task_group: "test_group_2"
    task_name: "task_2"
`

func TestLoadConfig(t *testing.T) {
	// Create a temporary config file
	tmpFile, err := ioutil.TempFile("", "scheduler_test_*.yaml")
	assert.NoError(t, err)
	defer os.Remove(tmpFile.Name())

	_, err = tmpFile.WriteString(testSchedulerConfig)
	assert.NoError(t, err)
	tmpFile.Close()

	sched := NewScheduler(tmpFile.Name())
	err = sched.LoadConfig()
	assert.NoError(t, err)
	assert.NotNil(t, sched.Config())
	assert.Len(t, sched.Config().ScheduledTasks, 2)
	assert.Equal(t, "test_task_1", sched.Config().ScheduledTasks[0].Name)
	assert.Equal(t, "@every 1s", sched.Config().ScheduledTasks[0].Schedule)
}

func TestSaveConfig(t *testing.T) {
	// Create a temporary config file
	tmpFile, err := ioutil.TempFile("", "scheduler_test_*.yaml")
	assert.NoError(t, err)
	defer os.Remove(tmpFile.Name())

	sched := NewScheduler(tmpFile.Name())
	sched.SetConfig(&SchedulerConfig{
		ScheduledTasks: []ScheduledTask{
			{Name: "new_task", Schedule: "@every 3s", TaskFile: "new.lua", TaskGroup: "new_group", TaskName: "new_name"},
		},
	})

	err = sched.SaveConfig()
	assert.NoError(t, err)

	// Read the file back and verify content
	data, err := ioutil.ReadFile(tmpFile.Name())
	assert.NoError(t, err)
	assert.Contains(t, string(data), "new_task")
	assert.Contains(t, string(data), "@every 3s")
}

func TestConfigGetSet(t *testing.T) {
	sched := NewScheduler("dummy.yaml")
	assert.Nil(t, sched.Config())

	cfg := &SchedulerConfig{ScheduledTasks: []ScheduledTask{{Name: "test"}}}
	sched.SetConfig(cfg)
	assert.Equal(t, cfg, sched.Config())
}

// Mock exec.Command for testing RunTask
// var runCommandFunc func(name string, arg ...string) *os.ProcessState // This line was causing issues

func TestRunTask(t *testing.T) {
	// Override exec.Command for testing
	oldExecCommand := execCommand
	defer func() { execCommand = oldExecCommand }()

	execCommand = func(name string, arg ...string) *exec.Cmd {
		cs := []string{"-test.run=TestHelperProcess", "--"}
		cs = append(cs, arg...)
		cmd := exec.Command(os.Args[0], cs...)
		cmd.Env = []string{"GO_WANT_HELPER_PROCESS=1"}
		return cmd
	}

	sched := NewScheduler("dummy.yaml")
	task := ScheduledTask{
		Name:      "mock_task",
		TaskFile:  "mock.lua",
		TaskGroup: "mock_group",
		TaskName:  "mock_name",
	}

	sched.RunTask(task)
	// In a real test, you'd check logs or mock stdout/stderr to verify output
}

// TestHelperProcess is a helper for TestRunTask to mock exec.Command
func TestHelperProcess(t *testing.T) {
	if os.Getenv("GO_WANT_HELPER_PROCESS") != "1" {
		return
	}
	defer os.Exit(0)

	args := os.Args
	for len(args) > 0 {
		if args[0] == "--" {
			args = args[1:]
			break
		}
		args = args[1:]
	}
	if len(args) == 0 {
		fmt.Fprintf(os.Stderr, "No command\n")
		os.Exit(1)
	}

	cmd := args[0]
	switch cmd {
	case "sloth-runner":
		// Simulate successful execution
		fmt.Println("Mock sloth-runner run successful")
		os.Exit(0)
	default:
		fmt.Fprintf(os.Stderr, "Unknown command: %s\n", cmd)
		os.Exit(1)
	}
}

// Mock exec.Command for testing
// var execCommand = exec.Command // This is already defined in scheduler.go

func TestStartStop(t *testing.T) {
	// Create a temporary config file
	tmpFile, err := ioutil.TempFile("", "scheduler_test_*.yaml")
	assert.NoError(t, err)
	defer os.Remove(tmpFile.Name())

	_, err = tmpFile.WriteString(testSchedulerConfig)
	assert.NoError(t, err)
	tmpFile.Close()

	sched := NewScheduler(tmpFile.Name())
	err = sched.LoadConfig()
	assert.NoError(t, err)

	// Start the scheduler
	err = sched.Start()
	assert.NoError(t, err)

	// Give it a moment to schedule jobs
	time.Sleep(100 * time.Millisecond)

	// Stop the scheduler
	sched.Stop()

	// Ensure no more jobs are running (hard to assert directly without more complex mocking)
	// For now, just ensure Start/Stop don't panic or return errors
}