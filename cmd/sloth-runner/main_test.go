package main

import (
	"fmt"
	"io"
	"io/ioutil"
	"os"
	"os/exec"
	"path/filepath"
	"strconv"
	"syscall"
	"testing"
	"time"

	"github.com/chalkan3/sloth-runner/internal/scheduler"
	"github.com/stretchr/testify/assert"
	"github.com/spf13/cobra"
)

// Helper function to execute cobra commands
func executeCommand(root *cobra.Command, args ...string) (output string, err error) {
	r, w, _ := os.Pipe()
	defer w.Close() // Close writer when function exits

	root.SetOut(w)
	root.SetErr(w)

	_, _, err = executeCommandC(root, args...)

	// Read all output from the pipe
	outputBytes, readErr := ioutil.ReadAll(r)
	if readErr != nil {
		return "", readErr
	}
	output = string(outputBytes)

	return output, err
}

func executeCommandC(root *cobra.Command, args ...string) (c *cobra.Command, output string, err error) {
	// Set command arguments
	root.SetArgs(args)

	// Execute the command
	c, err = root.ExecuteC()

	return c, "", err // output is now captured globally
}

// Mocking os.Exit to prevent test runner from exiting
func TestMain(m *testing.M) {
	os.Exit(m.Run())
}

// TestHelperProcess is a helper for mocking exec.Command
func TestHelperProcess(t *testing.T) {
	if os.Getenv("GO_WANT_HELPER_PROCESS") != "1" {
		return
	}
	defer os.Exit(0)

	var writer io.Writer = os.Stdout // Declare writer here

	// Find the "--" separator
	var commandArgs []string
	foundSeparator := false
	for _, arg := range os.Args {
		if arg == "--" {
			foundSeparator = true
			continue
		}
		if foundSeparator {
			commandArgs = append(commandArgs, arg)
		}
	}

	if len(commandArgs) == 0 {
		fmt.Fprintf(os.Stderr, "No command provided to helper process\\n")
		os.Exit(1)
	}

	cmd := commandArgs[0]
	switch cmd {
	case "sloth-runner":
		// Simulate successful execution for the background scheduler process
		var f *os.File // Explicitly declare f here
		if outputPath := os.Getenv("GO_HELPER_PROCESS_OUTPUT"); outputPath != "" {
			var err error
			f, err = os.OpenFile(outputPath, os.O_APPEND|os.O_CREATE|os.O_WRONLY, 0644)
			if err != nil {
				fmt.Fprintf(os.Stderr, "Failed to open helper process output file: %v\\n", err)
				os.Exit(1)
			}
			defer f.Close()
			writer = f
		}

		fmt.Fprintln(writer, "Starting sloth-runner scheduler in background...")
		currentPid := os.Getpid()
		fmt.Fprintf(writer, "Scheduler started with PID %d. Logs will be redirected to stdout/stderr of the background process.\n", currentPid)
		fmt.Fprintln(writer, "To stop the scheduler, run: sloth-runner scheduler disable")

		// Write a dummy PID file for the enable command to find
		pidEnv := os.Getenv("GO_WANT_HELPER_PROCESS_PID")
		fmt.Fprintf(os.Stderr, "TestHelperProcess: GO_WANT_HELPER_PROCESS_PID = %s\\n", pidEnv)
		if pidEnv != "" {
			pidFile := pidEnv
			fmt.Fprintf(os.Stderr, "TestHelperProcess: pidFile = %s\\n", pidFile)
			pidFileDir := filepath.Dir(pidFile)
			fmt.Fprintf(os.Stderr, "TestHelperProcess: pidFileDir = %s\\n", pidFileDir)
			if err := os.MkdirAll(pidFileDir, 0755); err != nil {
				fmt.Fprintf(os.Stderr, "TestHelperProcess: Failed to create PID file directory %s: %v\\n", pidFileDir, err)
				os.Exit(1)
			}
			if err := ioutil.WriteFile(pidFile, []byte(strconv.Itoa(currentPid)), 0644); err != nil {
				fmt.Fprintf(os.Stderr, "TestHelperProcess: Failed to write mock PID file %s: %v\\n", pidFile, err)
				os.Exit(1)
			}
		} else {
			fmt.Fprintf(os.Stderr, "TestHelperProcess: GO_WANT_HELPER_PROCESS_PID is empty, not writing PID file\\n")
		}
		// Keep the process alive for a short duration to simulate background execution
		time.Sleep(100 * time.Millisecond)
		os.Exit(0)
	default:
		fmt.Fprintf(writer, "Unknown command: %s\\n", cmd)
		os.Exit(1)
	}
}

func TestSchedulerEnable(t *testing.T) {
	// Create a temporary directory for test artifacts
	tmpDir, err := ioutil.TempDir("", "sloth-runner-test-")
	assert.NoError(t, err)
	defer os.RemoveAll(tmpDir)

	// Ensure no PID file exists before starting the scheduler
	pidFile := filepath.Join(tmpDir, schedulerPIDFile)
	os.Remove(pidFile)

	// Create a dummy scheduler.yaml
	schedulerConfigPath := filepath.Join(tmpDir, "scheduler.yaml")
	dummyConfig := `scheduled_tasks:
  - name: "test_task"
    schedule: "@every 1s"
    task_file: "test.lua"
    task_group: "test_group"
    task_name: "test_name"`
	err = ioutil.WriteFile(schedulerConfigPath, []byte(dummyConfig), 0644)
	assert.NoError(t, err)

	// Set up environment for mocking exec.Command
	oldArgs := os.Args
	oldExecCommand := execCommand
	defer func() {
		os.Args = oldArgs
		execCommand = oldExecCommand
	}()

	execCommand = func(name string, arg ...string) *exec.Cmd {
		cs := []string{"-test.run=TestHelperProcess", "--"}
		cs = append(cs, "sloth-runner")
		cs = append(cs, arg...)
		cmd := exec.Command(os.Args[0], cs...)

		cmd.Env = append(os.Environ(), "GO_WANT_HELPER_PROCESS=1", "GO_WANT_HELPER_PROCESS_PID=" + filepath.Join(tmpDir, schedulerPIDFile), "GO_HELPER_PROCESS_OUTPUT=" + filepath.Join(tmpDir, "helper_output.txt"))
		cmd.SysProcAttr = &syscall.SysProcAttr{Setpgid: true} // Detach the process
		return cmd
	}

	// Execute the enable command
	output, err = executeCommand(rootCmd, "scheduler", "enable", "-c", schedulerConfigPath)
	assert.NoError(t, err, output)

	// Read helper process output from the temporary file
		helperOutputPath := filepath.Join(tmpDir, "helper_output.txt")
		helperOutputBytes, err = ioutil.ReadFile(helperOutputPath)
		assert.NoError(t, err)
		helperOutput := string(helperOutputBytes)

	// Combine output from cobra command and helper process
	combinedOutput := output + helperOutput

	assert.Contains(t, combinedOutput, "Starting sloth-runner scheduler in background...")
	assert.Contains(t, combinedOutput, "Scheduler started with PID ") // PID is now dynamic


	// Verify PID file exists
	pidFile := filepath.Join(tmpDir, schedulerPIDFile)
	assert.FileExists(t, pidFile)

	// Read the PID from the file
	pidBytes, err := ioutil.ReadFile(pidFile)
	assert.NoError(t, err)
	pid, err := strconv.Atoi(string(pidBytes))
	assert.NoError(t, err)
	assert.NotEqual(t, 0, pid)

	// Clean up PID file for subsequent tests
	os.Remove(pidFile)

	// Store the PID for TestSchedulerDisable
	// This is a simplification; in a real scenario, you might pass this via a channel or a more robust mechanism
	// For now, we'll use a package-level variable for simplicity in this test file.
	lastHelperProcessPID = pid
}

var lastHelperProcessPID int // Package-level variable to store PID

func TestSchedulerDisable(t *testing.T) {
	// Create a temporary directory for test artifacts
	tmpDir, err := ioutil.TempDir("", "sloth-runner-test-")
	assert.NoError(t, err)
	defer os.RemoveAll(tmpDir)

	// Create a dummy PID file for a mock process
	pidFile := filepath.Join(tmpDir, schedulerPIDFile)
	err = ioutil.WriteFile(pidFile, []byte(strconv.Itoa(lastHelperProcessPID)), 0644)
	assert.NoError(t, err)

	// Mock os.FindProcess and process.Signal
	oldFindProcess := osFindProcess
	oldProcessSignal := processSignal
	defer func() {
		osFindProcess = oldFindProcess
		processSignal = oldProcessSignal
	}()

	osFindProcess = func(pid int) (*os.Process, error) {
		// Expect the fixed PID from TestSchedulerEnable
		assert.Equal(t, lastHelperProcessPID, pid)
		return &os.Process{Pid: pid}, nil
	}
	processSignal = func(p *os.Process, sig os.Signal) error {
		assert.Equal(t, syscall.SIGTERM, sig)
		// Simulate process termination by removing the PID file
		os.Remove(pidFile)
		return nil // Simulate successful signal
	}

	// Execute the disable command
	output, err := executeCommand(rootCmd, "scheduler", "disable")
	assert.NoError(t, err, output)
	assert.Contains(t, output, fmt.Sprintf("Scheduler with PID %d stopped successfully.", lastHelperProcessPID))

	// Verify PID file is removed (mocked by not returning an error from os.Remove)
	assert.NoFileExists(t, pidFile)
}

func TestSchedulerList(t *testing.T) {
	// Create a temporary directory for test artifacts
	tmpDir, err := ioutil.TempDir("", "sloth-runner-test-")
	assert.NoError(t, err)
	defer os.RemoveAll(tmpDir)

	// Create a dummy scheduler.yaml
	schedulerConfigPath := filepath.Join(tmpDir, "scheduler.yaml")
	dummyConfig := `scheduled_tasks:
  - name: "list_test_task"
    schedule: "@every 1h"
    task_file: "list.lua"
    task_group: "list_group"
    task_name: "list_name"`
	err = ioutil.WriteFile(schedulerConfigPath, []byte(dummyConfig), 0644)
	assert.NoError(t, err)

	// Execute the list command
	output, err := executeCommand(rootCmd, "scheduler", "list", "-c", schedulerConfigPath)
	assert.NoError(t, err, output)
	assert.Contains(t, output, "Configured Scheduled Tasks")
	assert.Contains(t, output, "list_test_task")
	assert.Contains(t, output, "@every 1h")
}

func TestSchedulerDelete(t *testing.T) {
	// Create a temporary directory for test artifacts
	tmpDir, err := ioutil.TempDir("", "sloth-runner-test-")
	assert.NoError(t, err)
	defer os.RemoveAll(tmpDir)

	// Create a dummy scheduler.yaml with two tasks
	schedulerConfigPath := filepath.Join(tmpDir, "scheduler.yaml")
	dummyConfig := `scheduled_tasks:
  - name: "task_to_delete"
    schedule: "@every 1s"
    task_file: "file1.lua"
    task_group: "group1"
    task_name: "name1"
  - name: "task_to_keep"
    schedule: "@every 2s"
    task_file: "file2.lua"
    task_group: "group2"
    task_name: "name2"`
	err = ioutil.WriteFile(schedulerConfigPath, []byte(dummyConfig), 0644)
	assert.NoError(t, err)

	// Execute the delete command
	output, err := executeCommand(rootCmd, "scheduler", "delete", "task_to_delete", "-c", schedulerConfigPath)
	assert.NoError(t, err, output)
	assert.Contains(t, output, "Deleting scheduled task 'task_to_delete'...")
	assert.Contains(t, output, "Scheduled task 'task_to_delete' deleted successfully.")

	// Verify the scheduler.yaml is updated
	sched := scheduler.NewScheduler(schedulerConfigPath)
	err = sched.LoadConfig()
	assert.NoError(t, err)
	assert.NotNil(t, sched.Config())
	assert.Len(t, sched.Config().ScheduledTasks, 1)
	assert.Equal(t, "task_to_keep", sched.Config().ScheduledTasks[0].Name)
}