package taskrunner

import (
	"bytes"
	"context"
	"fmt"
	"log"
	"os/exec"
	"sync"
	"time"

	"github.com/chalkan3/sloth-runner/internal/luainterface"
	"github.com/chalkan3/sloth-runner/internal/types"
	"github.com/pterm/pterm"
	lua "github.com/yuin/gopher-lua"
)

// executeShellCondition executes a shell command and returns true if it succeeds (exit code 0).
func executeShellCondition(command string) (bool, error) {
	if command == "" {
		return false, fmt.Errorf("command cannot be empty")
	}
	cmd := exec.Command("bash", "-c", command)
	if err := cmd.Run(); err != nil {
		if _, ok := err.(*exec.ExitError); ok {
			// Command executed but returned a non-zero exit code.
			return false, nil
		}
		// Other error (e.g., command not found).
		return false, err
	}
	// Command succeeded.
	return true, nil
}

// TaskExecutionError provides a more context-rich error for task failures.
type TaskExecutionError struct {
	TaskName string
	Err      error
}

func (e *TaskExecutionError) Error() string {
	return fmt.Sprintf("task '%s' failed: %v", e.TaskName, e.Err)
}

// TaskResult holds the outcome of a single task execution.
type TaskResult struct {
	Name     string
	Status   string
	Duration time.Duration
	Error    error
}

type TaskRunner struct {
	L           *lua.LState
	TaskGroups  map[string]types.TaskGroup
	TargetGroup string
	TargetTasks []string
	Results     []TaskResult
	Outputs     map[string]interface{}
	DryRun      bool
}


func NewTaskRunner(L *lua.LState, groups map[string]types.TaskGroup, targetGroup string, targetTasks []string, dryRun bool) *TaskRunner {
	return &TaskRunner{
		L:           L,
		TaskGroups:  groups,
		TargetGroup: targetGroup,
		TargetTasks: targetTasks,
		Outputs:     make(map[string]interface{}),
		DryRun:      dryRun,
	}
}

func (tr *TaskRunner) executeTaskWithRetries(t *types.Task, inputFromDependencies *lua.LTable, mu *sync.Mutex, completedTasks map[string]bool, taskOutputs map[string]*lua.LTable, runningTasks map[string]bool) error {
	// AbortIf check
	if t.AbortIfFunc != nil {
		shouldAbort, _, _, err := luainterface.ExecuteLuaFunction(tr.L, t.AbortIfFunc, t.Params, inputFromDependencies, 1, nil)
		if err != nil {
			return &TaskExecutionError{TaskName: t.Name, Err: fmt.Errorf("failed to execute abort_if function: %w", err)}
		}
		if shouldAbort {
			return &TaskExecutionError{TaskName: t.Name, Err: fmt.Errorf("execution aborted by abort_if function")}
		}
	} else if t.AbortIf != "" {
		shouldAbort, err := executeShellCondition(t.AbortIf)
		if err != nil {
			return &TaskExecutionError{TaskName: t.Name, Err: fmt.Errorf("failed to execute abort_if condition: %w", err)}
		}
		if shouldAbort {
			return &TaskExecutionError{TaskName: t.Name, Err: fmt.Errorf("execution aborted by abort_if condition")}
		}
	}

	// RunIf check
	if t.RunIfFunc != nil {
		shouldRun, _, _, err := luainterface.ExecuteLuaFunction(tr.L, t.RunIfFunc, t.Params, inputFromDependencies, 1, nil)
		if err != nil {
			return &TaskExecutionError{TaskName: t.Name, Err: fmt.Errorf("failed to execute run_if function: %w", err)}
		}
		if !shouldRun {
			pterm.Info.Printf("Skipping task '%s' due to run_if function condition.\n", t.Name)
			mu.Lock()
			tr.Results = append(tr.Results, TaskResult{
				Name:   t.Name,
				Status: "Skipped",
			})
			completedTasks[t.Name] = true
			delete(runningTasks, t.Name)
			mu.Unlock()
			return nil
		}
	} else if t.RunIf != "" {
		shouldRun, err := executeShellCondition(t.RunIf)
		if err != nil {
			return &TaskExecutionError{TaskName: t.Name, Err: fmt.Errorf("failed to execute run_if condition: %w", err)}
		}
		if !shouldRun {
			pterm.Info.Printf("Skipping task '%s' due to run_if condition.\n", t.Name)
			mu.Lock()
			tr.Results = append(tr.Results, TaskResult{
				Name:   t.Name,
				Status: "Skipped",
			})
			completedTasks[t.Name] = true
			delete(runningTasks, t.Name)
			mu.Unlock()
			return nil
		}
	}

	var taskErr error
	var spinner *pterm.SpinnerPrinter

	for i := 0; i <= t.Retries; i++ {
		if i > 0 {
			pterm.Warning.Printf("Task '%s' failed. Retrying in 1s (%d/%d)...\n", t.Name, i, t.Retries)
			time.Sleep(1 * time.Second)
		}

		spinner, _ = pterm.DefaultSpinner.Start(fmt.Sprintf("Executing task: %s", t.Name))

		var ctx context.Context
		var cancel context.CancelFunc

		if t.Timeout != "" {
			timeout, err := time.ParseDuration(t.Timeout)
			if err != nil {
				return &TaskExecutionError{TaskName: t.Name, Err: fmt.Errorf("invalid timeout duration: %w", err)}
			}
			ctx, cancel = context.WithTimeout(context.Background(), timeout)
		} else {
			ctx, cancel = context.WithCancel(context.Background())
		}
		defer cancel()

		taskErr = tr.runTask(ctx, t, inputFromDependencies, mu, completedTasks, taskOutputs, runningTasks)

		if taskErr == nil {
			spinner.Success(fmt.Sprintf("Task '%s' completed successfully", t.Name))
			return nil // Success
		}
	}

	spinner.Fail(fmt.Sprintf("Task '%s' failed after %d retries: %v", t.Name, t.Retries, taskErr))
	return taskErr // Final failure
}

func (tr *TaskRunner) runTask(ctx context.Context, t *types.Task, inputFromDependencies *lua.LTable, mu *sync.Mutex, completedTasks map[string]bool, taskOutputs map[string]*lua.LTable, runningTasks map[string]bool) (taskErr error) {
	startTime := time.Now()
	t.Output = tr.L.NewTable()

	defer func() {
		if r := recover(); r != nil {
			taskErr = &TaskExecutionError{TaskName: t.Name, Err: fmt.Errorf("panic: %v", r)}
		}

		duration := time.Since(startTime)
		status := "Success"
		if taskErr != nil {
			status = "Failed"
		}

		mu.Lock()
		tr.Results = append(tr.Results, TaskResult{
			Name:     t.Name,
			Status:   status,
			Duration: duration,
			Error:    taskErr,
		})
		taskOutputs[t.Name] = t.Output
		completedTasks[t.Name] = true
		delete(runningTasks, t.Name)
		mu.Unlock()
	}()

	if t.PreExec != nil {
		success, msg, _, err := luainterface.ExecuteLuaFunction(tr.L, t.PreExec, t.Params, inputFromDependencies, 2, ctx)
		if err != nil {
			return &TaskExecutionError{TaskName: t.Name, Err: fmt.Errorf("error executing pre_exec hook: %w", err)}
		} else if !success {
			return &TaskExecutionError{TaskName: t.Name, Err: fmt.Errorf("pre-execution hook failed: %s", msg)}
		}
	}

	if t.CommandFunc != nil {
		success, msg, outputTable, err := luainterface.ExecuteLuaFunction(tr.L, t.CommandFunc, t.Params, inputFromDependencies, 3, ctx)
		if err != nil {
			return &TaskExecutionError{TaskName: t.Name, Err: fmt.Errorf("error executing command function: %w", err)}
		} else if !success {
			return &TaskExecutionError{TaskName: t.Name, Err: fmt.Errorf("command function returned failure: %s", msg)}
		} else if outputTable != nil {
			t.Output = outputTable
		}
	} else if t.CommandStr != "" {
		cmd := exec.CommandContext(ctx, "bash", "-c", t.CommandStr)
		var out bytes.Buffer
		cmd.Stdout = &out
		cmd.Stderr = &out
		if err := cmd.Run(); err != nil {
			return &TaskExecutionError{TaskName: t.Name, Err: fmt.Errorf("command '%s' failed: %w, output: %s", t.CommandStr, err, out.String())}
		}
	} else {
		return &TaskExecutionError{TaskName: t.Name, Err: fmt.Errorf("task has no command defined")}
	}

	if t.PostExec != nil {
		var postExecSecondArg lua.LValue = t.Output
		if t.Output == nil {
			postExecSecondArg = tr.L.NewTable()
		}
		success, msg, _, err := luainterface.ExecuteLuaFunction(tr.L, t.PostExec, t.Params, postExecSecondArg, 2, ctx)
		if err != nil {
			return &TaskExecutionError{TaskName: t.Name, Err: fmt.Errorf("error executing post_exec hook: %w", err)}
		} else if !success {
			return &TaskExecutionError{TaskName: t.Name, Err: fmt.Errorf("post-execution hook failed: %s", msg)}
		}
	}

	return nil
}

func (tr *TaskRunner) Run() error {
	if len(tr.TaskGroups) == 0 {
		pterm.Warning.Println("No task groups defined.")
		return nil
	}

	pterm.DefaultHeader.Println("Executing Tasks")
	var allGroupErrors []error

	filteredGroups := make(map[string]types.TaskGroup)
	if tr.TargetGroup != "" {
		if group, ok := tr.TaskGroups[tr.TargetGroup]; ok {
			filteredGroups[tr.TargetGroup] = group
		} else {
			return fmt.Errorf("task group '%s' not found", tr.TargetGroup)
		}
	} else {
		filteredGroups = tr.TaskGroups
	}

	for groupName, group := range filteredGroups {
		pterm.DefaultSection.Printf("Group: %s (Description: %s)\n", groupName, group.Description)

		originalTaskMap := make(map[string]*types.Task)
		for i := range group.Tasks {
			originalTaskMap[group.Tasks[i].Name] = &group.Tasks[i]
		}

		tasksToRun, err := tr.resolveTasksToRun(originalTaskMap, tr.TargetTasks)
		if err != nil {
			allGroupErrors = append(allGroupErrors, fmt.Errorf("error resolving tasks for group '%s': %w", groupName, err))
			continue
		}

		if len(tasksToRun) == 0 {
			pterm.Info.Printf("No tasks to run in group '%s' after filtering.\n", groupName)
			continue
		}

		taskOutputs := make(map[string]*lua.LTable)
		completedTasks := make(map[string]bool)
		runningTasks := make(map[string]bool)
		var wg sync.WaitGroup
		var mu sync.Mutex
		errChan := make(chan error, len(tasksToRun))

		taskMap := make(map[string]*types.Task)
		for _, task := range tasksToRun {
			taskMap[task.Name] = task
		}

		for {
			mu.Lock()
			if len(completedTasks) == len(tasksToRun) {
				mu.Unlock()
				break
			}
			mu.Unlock()

			tasksLaunchedThisIteration := 0
			for _, task := range taskMap {
				mu.Lock()
				if completedTasks[task.Name] || runningTasks[task.Name] {
					mu.Unlock()
					continue
				}
				mu.Unlock()

				dependencyMet := true
				for _, depName := range task.DependsOn {
					if _, exists := taskMap[depName]; exists {
						mu.Lock()
						completed := completedTasks[depName]
						if !completed {
							dependencyMet = false
						} else {
							// Check if the completed dependency was successful
							var depSuccess bool
							for _, res := range tr.Results {
								if res.Name == depName && res.Error == nil {
									depSuccess = true
									break
								}
							}
							if !depSuccess {
								dependencyMet = false
								// Mark this task as skipped since its dependency failed
								tr.Results = append(tr.Results, TaskResult{
									Name:   task.Name,
									Status: "Skipped",
									Error:  fmt.Errorf("dependency '%s' failed", depName),
								})
								completedTasks[task.Name] = true // Mark as completed to unblock the loop
							}
						}
						mu.Unlock()
						if !dependencyMet {
							break
						}
					}
				}

				if dependencyMet {
					tasksLaunchedThisIteration++
					inputFromDependencies := tr.L.NewTable()
					for _, depName := range task.DependsOn {
						if _, exists := taskMap[depName]; exists {
							mu.Lock()
							depOutput := taskOutputs[depName]
							mu.Unlock()
							if depOutput != nil {
								inputFromDependencies.RawSetString(depName, depOutput)
							} else {
								inputFromDependencies.RawSetString(depName, tr.L.NewTable())
							}
						}
					}

					if task.Async {
						mu.Lock()
						runningTasks[task.Name] = true
						mu.Unlock()
						wg.Add(1)
						go func(t *types.Task, input *lua.LTable) {
							defer wg.Done()
							if err := tr.executeTaskWithRetries(t, input, &mu, completedTasks, taskOutputs, runningTasks); err != nil {
								errChan <- err
							}
						}(task, inputFromDependencies)
					} else {
						mu.Lock()
						runningTasks[task.Name] = true
						mu.Unlock()
						if err := tr.executeTaskWithRetries(task, inputFromDependencies, &mu, completedTasks, taskOutputs, runningTasks); err != nil {
							errChan <- err
						}
					}
				}
			}

			if tasksLaunchedThisIteration == 0 {
				mu.Lock()
				if len(completedTasks) < len(tasksToRun) {
					uncompletedTasks := []string{}
					for _, task := range tasksToRun {
						if !completedTasks[task.Name] {
							uncompletedTasks = append(uncompletedTasks, task.Name)
						}
					}
					mu.Unlock()
					circularErr := &TaskExecutionError{TaskName: "group " + groupName, Err: fmt.Errorf("circular dependency or unresolvable tasks. Uncompleted tasks: %v", uncompletedTasks)}
					log.Println(circularErr.Error())
					errChan <- circularErr
					break
				}
				mu.Unlock()
			}
			wg.Wait()
		}

		close(errChan)

		var groupErrors []error
		for err := range errChan {
			groupErrors = append(groupErrors, err)
		}

		if len(groupErrors) > 0 {
			allGroupErrors = append(allGroupErrors, fmt.Errorf("task group '%s' encountered errors: %v", groupName, groupErrors))
		}

		mu.Lock()
		for taskName, outputTable := range taskOutputs {
			tr.Outputs[taskName] = luainterface.LuaTableToGoMap(tr.L, outputTable)
		}
		mu.Unlock()
	}

	pterm.DefaultSection.Println("Execution Summary")
	tableData := pterm.TableData{{"Task", "Status", "Duration", "Error"}}
	for _, result := range tr.Results {
		status := pterm.Green(result.Status)
		errStr := ""
		if result.Error != nil {
			status = pterm.Red(result.Status)
			errStr = result.Error.Error()
		} else if result.Status == "Skipped" {
			status = pterm.Yellow(result.Status)
		} else if result.Status == "DryRun" {
			status = pterm.Cyan(result.Status)
		}
		tableData = append(tableData, []string{result.Name, status, result.Duration.String(), errStr})
	}
	pterm.DefaultTable.WithHasHeader().WithData(tableData).Render()

	if len(allGroupErrors) > 0 {
		return fmt.Errorf("one or more task groups failed: %v", allGroupErrors)
	}
	return nil
}

func (tr *TaskRunner) resolveTasksToRun(originalTaskMap map[string]*types.Task, targetTasks []string) ([]*types.Task, error) {
	if len(targetTasks) == 0 {
		var allTasks []*types.Task
		for _, task := range originalTaskMap {
			allTasks = append(allTasks, task)
		}
		return allTasks, nil
	}

	resolved := make(map[string]*types.Task)
	queue := make([]string, 0, len(targetTasks))
	visited := make(map[string]bool)

	for _, taskName := range targetTasks {
		if !visited[taskName] {
			queue = append(queue, taskName)
			visited[taskName] = true
		}
	}

	head := 0
	for head < len(queue) {
		currentTaskName := queue[head]
		head++

		currentTask, ok := originalTaskMap[currentTaskName]
		if !ok {
			return nil, fmt.Errorf("task '%s' not found in group", currentTaskName)
		}
		resolved[currentTaskName] = currentTask

		for _, depName := range currentTask.DependsOn {
			if !visited[depName] {
				visited[depName] = true
				queue = append(queue, depName)
			}
		}
	}

	var result []*types.Task
	for _, task := range resolved {
		result = append(result, task)
	}
	return result, nil
}