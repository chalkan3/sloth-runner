package main

import (
	"bytes"
	"fmt"
	"io"
	"io/ioutil"
	"log/slog"
	"os"
	"path/filepath"
	"regexp"
	"strconv"
	"strings"
	"syscall"
	"testing"
	"time"

	"github.com/chalkan3/sloth-runner/internal/scheduler"
	"github.com/pterm/pterm"
	"github.com/stretchr/testify/assert"
	"github.com/spf13/cobra"
)

var ansiRegex = regexp.MustCompile(`\x1b\[[0-?]*[ -/]*[@-~]`)

func stripAnsi(str string) string {
	return ansiRegex.ReplaceAllString(str, "")
}

// Helper function to execute cobra commands
func executeCommand(root *cobra.Command, args ...string) (output string, err error) {
	buf := new(bytes.Buffer)
	SetTestOutputBuffer(buf)
	defer SetTestOutputBuffer(nil)

	pterm.DefaultLogger.Writer = buf
	slog.SetDefault(slog.New(pterm.NewSlogHandler(&pterm.DefaultLogger)))

	root.SetOut(buf)
	root.SetErr(buf)

	// Temporarily disable error/usage silencing for testing
	oldSilenceErrors := rootCmd.SilenceErrors
	oldSilenceUsage := rootCmd.SilenceUsage
	rootCmd.SilenceErrors = true
	rootCmd.SilenceUsage = true
	defer func() {
		rootCmd.SilenceErrors = oldSilenceErrors
		rootCmd.SilenceUsage = oldSilenceUsage
	}()

	// No pterm output redirection here, as it's handled by the command itself

	root.SetArgs(args)
	err = root.Execute()

	output = buf.String()

	// If Execute() returns nil but there's an error in the output, capture it
	if err == nil && strings.Contains(output, "Error: ") {
		// Extract the error message from the output
		errorLine := ""
		lines := strings.Split(output, "\n")
		for _, line := range lines {
			if strings.HasPrefix(line, "Error: ") {
				errorLine = line
				break
			}
		}
		if errorLine != "" {
			err = fmt.Errorf("%s", errorLine)
		}
	}

	return output, err
}

func executeCommandC(root *cobra.Command, args ...string) (c *cobra.Command, output string, err error) {
	// Set command arguments
	root.SetArgs(args)

	// Execute the command
	// We need to temporarily set rootCmd to the test root for Execute() to work correctly
	oldRootCmd := rootCmd
	rootCmd = root
	defer func() { rootCmd = oldRootCmd }()

	// Temporarily disable error/usage silencing for testing
	oldSilenceErrors := rootCmd.SilenceErrors
	oldSilenceUsage := rootCmd.SilenceUsage
	rootCmd.SilenceErrors = false
	rootCmd.SilenceUsage = false
	defer func() {
		rootCmd.SilenceErrors = oldSilenceErrors
		rootCmd.SilenceUsage = oldSilenceUsage
	}()

	// No pterm output redirection here, as it's handled by the command itself

	err = Execute() // Call the new Execute() function

	return root, "", err // output is now captured globally
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

		// Check for --run-as-scheduler flag
		runAsSchedulerFlag := false
		for _, arg := range commandArgs {
			if arg == "--run-as-scheduler" {
				runAsSchedulerFlag = true
				break
			}
		}

		if runAsSchedulerFlag {
			fmt.Fprintln(writer, "Starting sloth-runner scheduler in background...")
			currentPid := os.Getpid()
			fmt.Fprintf(writer, "Scheduler started with PID %d. Logs will be redirected to stdout/stderr of the background process.\\n", currentPid)
			fmt.Fprintln(writer, "To stop the scheduler, run: sloth-runner scheduler disable")

			// Write a dummy PID file for the enable command to find
			pidEnv := os.Getenv("GO_WANT_HELPER_PROCESS_PID")
			if pidEnv != "" {
				pidFile := pidEnv
				if err := os.MkdirAll(filepath.Dir(pidFile), 0755); err != nil {
					fmt.Fprintf(os.Stderr, "TestHelperProcess: Failed to create PID file directory %s: %v\\n", filepath.Dir(pidFile), err)
					os.Exit(1)
				}
				if err := ioutil.WriteFile(pidFile, []byte(strconv.Itoa(currentPid)), 0644); err != nil {
					fmt.Fprintf(os.Stderr, "TestHelperProcess: Failed to write mock PID file %s: %v\\n", pidFile, err)
					os.Exit(1)
				}
			}
			// Keep the process alive for a short duration to simulate background execution
			time.Sleep(1 * time.Second)
			os.Exit(0)
		} else {
			// This is a regular sloth-runner command, not the scheduler
			// We need to simulate the behavior of the actual sloth-runner commands (enable, disable, list, delete)
			// For now, let's just print a generic message
			fmt.Fprintf(writer, "Simulating sloth-runner command: %v\\n", commandArgs)
		}
	default:
		fmt.Fprintf(writer, "Unknown command: %s\\n", cmd)
		os.Exit(1)
	}
}

/*
// func TestSchedulerEnable(t *testing.T) {
	// Create a temporary directory for test artifacts
	tmpDir, err := ioutil.TempDir("", "sloth-runner-test-")
	assert.NoError(t, err)
	defer os.RemoveAll(tmpDir)

	pidFile := filepath.Join(tmpDir, "sloth-runner-scheduler.pid")
	os.Remove(pidFile)

	schedulerConfigPath := filepath.Join(tmpDir, "scheduler.yaml")
	dummyConfig := `scheduled_tasks:
  - name: "test_task"
    schedule: "@every 1s"
    task_file: "test.lua"
    task_group: "test_group"
    task_name: "test_name"`
	err = ioutil.WriteFile(schedulerConfigPath, []byte(dummyConfig), 0644)
	assert.NoError(t, err)

	// Mock exec.Command to use the helper process
	oldExecCommand := execCommand
	defer func() { execCommand = oldExecCommand }()
	execCommand = func(name string, arg ...string) *exec.Cmd {
		cs := []string{"-test.run=TestHelperProcess", "--", "sloth-runner"}
		cs = append(cs, arg...)
		cmd := exec.Command(os.Args[0], cs...)
		cmd.Env = []string{"GO_WANT_HELPER_PROCESS=1",
			"GO_WANT_HELPER_PROCESS_PID=" + pidFile}
		return cmd
	}

	// Execute the enable command
	output, err := executeCommand(rootCmd, "scheduler", "enable", "-c", schedulerConfigPath)
	assert.NoError(t, err, output)

	// Assert output messages
	assert.Contains(t, output, "Starting sloth-runner scheduler in background...")
	// PID is dynamic, so we can't assert the exact PID, but we can check for the message
	assert.Regexp(t, `Scheduler started with PID \d+\.`, output)
	assert.Contains(t, output, "To stop the scheduler, run: sloth-runner scheduler disable")

	// Verify PID file exists
	assert.FileExists(t, pidFile)
	pidBytes, err := ioutil.ReadFile(pidFile)
	assert.NoError(t, err)
	pid, err := strconv.Atoi(string(pidBytes))
	assert.NoError(t, err)
	lastHelperProcessPID = pid

	// Run TestSchedulerDisable as a subtest
	t.Run("TestSchedulerDisable", func(t *testing.T) {
		testSchedulerDisable(t, tmpDir)
	})
}
*/

var lastHelperProcessPID int // Package-level variable to store PID

func testSchedulerDisable(t *testing.T, tmpDir string) {
	// Create a dummy PID file for a mock process
	pidFile := filepath.Join(tmpDir, "sloth-runner-scheduler.pid")
	err := ioutil.WriteFile(pidFile, []byte(strconv.Itoa(lastHelperProcessPID)), 0644)
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
	var terminated bool
	processSignal = func(p *os.Process, sig os.Signal) error {
		if sig == syscall.SIGTERM {
			assert.Equal(t, syscall.SIGTERM, sig)
			// Simulate process termination by removing the PID file
			os.Remove(pidFile)
			terminated = true
			return nil // Simulate successful signal
		} else if sig == syscall.Signal(0) {
			if terminated {
				// Simulate process not found after termination
				return os.ErrProcessDone
			}
			return nil // Simulate process still running
		}
		return nil
	}

	// Execute the disable command
	output, err := executeCommand(rootCmd, "scheduler", "disable", "-c", schedulerConfigPath) // Pass schedulerConfigPath
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
	output = stripAnsi(output)
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

/*
// func TestInteractiveRunner(t *testing.T) {
	// Create a temporary directory for test artifacts
	tmpDir, err := ioutil.TempDir("", "sloth-runner-test-")
	assert.NoError(t, err)
	defer os.RemoveAll(tmpDir)

	// Create a dummy Lua task file
	taskFilePath := filepath.Join(tmpDir, "interactive_tasks.lua")
	dummyTasks := `
TaskDefinitions = {
  interactive_group = {
    tasks = {
      { name = "task1", command = "echo 'Task 1 executed'" },
      { name = "task2", command = "echo 'Task 2 executed'" },
      { name = "task3", command = "echo 'Task 3 executed'" }
    }
  }
}
`
	err = ioutil.WriteFile(taskFilePath, []byte(dummyTasks), 0644)
	assert.NoError(t, err)

	// Mock survey.AskOne
	actions := []string{"run", "skip", "abort"}
	actionIndex := 0
	taskrunner.SetSurveyAskOne(func(p survey.Prompt, response interface{}, opts ...survey.AskOpt) error {
		if actionIndex < len(actions) {
			*(response.(*string)) = actions[actionIndex]
			actionIndex++
		}
		return nil
	})
	defer taskrunner.SetSurveyAskOne(survey.AskOne) // Restore original function

	// Execute the run command with --interactive
	output, err := executeCommand(rootCmd, "run", "-f", taskFilePath, "--interactive")
	output = stripAnsi(output)
	assert.Error(t, err) // Expect an error because we abort
	assert.Contains(t, err.Error(), "execution aborted by user")

	// Assert that the output contains expected messages
	assert.Contains(t, output, "Task 1 executed")
	assert.Contains(t, output, "Skipping task 'task2' by user choice.")
	assert.NotContains(t, output, "Task 2 executed")
	assert.NotContains(t, output, "Task 3 executed")
	assert.Contains(t, output, "Aborting execution by user choice.")
}
*/

func TestEnhancedValuesTemplating(t *testing.T) {
	// Create a temporary directory for test artifacts
	tmpDir, err := ioutil.TempDir("", "sloth-runner-test-")
	assert.NoError(t, err)
	defer os.RemoveAll(tmpDir)

	// Create a dummy values.yaml file
	valuesFilePath := filepath.Join(tmpDir, "values.yaml")
	dummyValues := `my_value: "Hello from {{ .Env.MY_TEST_VARIABLE }}"`
	err = ioutil.WriteFile(valuesFilePath, []byte(dummyValues), 0644)
	assert.NoError(t, err)

	// Create a dummy Lua task file
	taskFilePath := filepath.Join(tmpDir, "templated_values_task.lua")
	dummyTask := `
TaskDefinitions = {
  templated_values_group = {
    tasks = {
      {
        name = "print_templated_value",
        command = function()
          log.info("Templated value: " .. values.my_value)
          return true
        end
      }
    }
  }
}
`
	err = ioutil.WriteFile(taskFilePath, []byte(dummyTask), 0644)
	assert.NoError(t, err)

	// Set environment variable
	os.Setenv("MY_TEST_VARIABLE", "TestValue123")
	defer os.Unsetenv("MY_TEST_VARIABLE")

	// Execute the run command
	output, err := executeCommand(rootCmd, "run", "-f", taskFilePath, "-v", valuesFilePath, "--yes")
	output = stripAnsi(output)
	assert.NoError(t, err)

	// Assert that the output contains the templated value
	assert.Contains(t, output, "Templated value: Hello from TestValue123")
}
