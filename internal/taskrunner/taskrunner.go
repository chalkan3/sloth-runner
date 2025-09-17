package taskrunner

import (
	"encoding/json" // Import for JSON marshaling
	"fmt"
	"log"
	"sync"
	"time" // Import time for timestamp

	lua "github.com/yuin/gopher-lua"

	"github.com/chalkan3/sloth-runner/internal/luainterface"
	"sloth-runner/internal/types"
)

// ANSI color codes
const (
	Reset  = "\033[0m"
	Red    = "\033[31m"
	Green  = "\033[32m"
	Yellow = "\033[33m"
	Blue   = "\033[34m"
	Purple = "\033[35m"
	Cyan   = "\033[36m"
	White  = "\033[37m"
)

type TaskError struct {
	TaskName string
	Err      error
	Type     TaskErrorType
}

type TaskErrorType string

const (
	ErrorTypePreExec    TaskErrorType = "PreExecError"
	ErrorTypeCommand    TaskErrorType = "CommandError"
	ErrorTypePostExec   TaskErrorType = "PostExecError"
	ErrorTypeDependency TaskErrorType = "DependencyError"
	ErrorTypeUnknown    TaskErrorType = "UnknownError"
)

func (te *TaskError) Error() string {
	return fmt.Sprintf("task '%s' failed with %s: %v", te.TaskName, te.Type, te.Err)
}

func NewTaskError(taskName string, err error, errType TaskErrorType) *TaskError {
	return &TaskError{
		TaskName: taskName,
		Err:      err,
		Type:     errType,
	}
}

type TaskRunner struct {
	L           *lua.LState
	TaskGroups  map[string]types.TaskGroup
	TargetGroup string
	TargetTasks []string
}

func NewTaskRunner(L *lua.LState, groups map[string]types.TaskGroup, targetGroup string, targetTasks []string) *TaskRunner {
	return &TaskRunner{
		L:           L,
		TaskGroups:  groups,
		TargetGroup: targetGroup,
		TargetTasks: targetTasks,
	}
}

// runTask encapsulates the logic for executing a single task.
// inputFromDependencies is a Lua table where keys are dependency names and values are their outputs.
func (tr *TaskRunner) runTask(t *types.Task, inputFromDependencies *lua.LTable, mu *sync.Mutex, completedTasks map[string]bool, taskOutputs map[string]*lua.LTable, runningTasks map[string]bool) (taskErr error) {
	startTime := time.Now()
	// taskStatus and taskComment are now handled by the defer function based on taskErr
	// Initialize t.Output to an empty Lua table to prevent nil pointer dereferences in Lua
	t.Output = tr.L.NewTable()

	defer func() {
		mu.Lock()
		taskOutputs[t.Name] = t.Output // Store output for dependencies
		completedTasks[t.Name] = true // Mark task as completed after execution
		delete(runningTasks, t.Name) // Remove from running tasks
		mu.Unlock()

		duration := time.Since(startTime)
		emoji := "❌"
		taskStatus := "Failed"
		statusColor := Red
		taskComment := ""

		if taskErr == nil { // Check taskErr here
			taskStatus = "Success"
			emoji = "✅"
			statusColor = Green
			taskComment = fmt.Sprintf("Task '%s' completed successfully.", t.Name)
		} else {
			taskComment = taskErr.Error() // Use the error message as the comment
		}

		fmt.Printf("----------\n")
		fmt.Printf("%s%s ID: %s%s\n", statusColor, emoji, t.Name, Reset)
		fmt.Printf("    Function: %s\n", t.Description)
		fmt.Printf("      Result: %s%s%s\n", statusColor, taskStatus, Reset)
		fmt.Printf("     Comment: %s\n", taskComment) // Use the dynamic taskComment
		if t.Output != nil {
			goMap := luainterface.LuaTableToGoMap(tr.L, t.Output)
			jsonOutput, err := json.MarshalIndent(goMap, "             ", "  ")
			if err != nil {
				fmt.Printf("     Changes: Error marshaling output: %v\n", err)
			} else {
				fmt.Printf("     Changes: Output:\n%s\n", string(jsonOutput))
			}
		}
		fmt.Printf("    Duration: %s\n", duration)
		fmt.Printf("----------\n")
	}()

	// Call pre_exec hook if it exists
	if t.PreExec != nil {
		success, msg, _, err := luainterface.ExecuteLuaFunction(tr.L, t.PreExec, t.Params, inputFromDependencies, 2) // Expect 2 return values
		if err != nil {
			taskErr = NewTaskError(t.Name, fmt.Errorf("error executing pre_exec hook: %w", err), ErrorTypePreExec)
			log.Printf(taskErr.Error())
		} else if !success {
			taskErr = NewTaskError(t.Name, fmt.Errorf("pre-execution hook failed: %s", msg), ErrorTypePreExec)
		}
	}

	// Only proceed with command and post_exec if pre_exec was successful or didn't exist
	if taskErr == nil {
		if t.CommandFunc != nil {
			// Call the Lua command function to get the command string and output
			success, msg, outputTable, err := luainterface.ExecuteLuaFunction(tr.L, t.CommandFunc, t.Params, inputFromDependencies, 3) // Expect 3 return values
			if err != nil {
				taskErr = NewTaskError(t.Name, fmt.Errorf("error executing command function: %w", err), ErrorTypeCommand)
			} else if !success {
				taskErr = NewTaskError(t.Name, fmt.Errorf("command function returned failure: %s", msg), ErrorTypeCommand)
			} else {
				if outputTable != nil {
					t.Output = outputTable
				}
			}
		} else {
			// If no dynamic command, check if a static command string is available
			if t.CommandStr == "" {
				taskErr = NewTaskError(t.Name, fmt.Errorf("task has no command defined"), ErrorTypeCommand)
			}
		}
	}

	// Call post_exec hook if it exists, only if no error occurred so far
	if taskErr == nil && t.PostExec != nil {
		// Pass task params and its own output to post_exec
		var postExecSecondArg lua.LValue
		if t.Output != nil {
			postExecSecondArg = t.Output
		} else {
			postExecSecondArg = tr.L.NewTable() // Ensure an empty table is passed if t.Output is nil
		}

		success, msg, _, err := luainterface.ExecuteLuaFunction(tr.L, t.PostExec, t.Params, postExecSecondArg, 2) // Expect 2 return values
		if err != nil {
			taskErr = NewTaskError(t.Name, fmt.Errorf("error executing post_exec hook: %w", err), ErrorTypePostExec)
			log.Printf(taskErr.Error())
		} else if !success {
			taskErr = NewTaskError(t.Name, fmt.Errorf("post-execution hook failed: %s", msg), ErrorTypePostExec)
		}
	}

	return taskErr
}

func (tr *TaskRunner) Run() error {
	if len(tr.TaskGroups) == 0 {
		fmt.Println("No task groups defined.")
		return nil
	}

	fmt.Printf("%s---\n", Blue)
	fmt.Printf("%sExecuting Tasks\n", Blue)
	fmt.Printf("%s---\n", Blue)
	var allGroupErrors []error

	// Filter groups if TargetGroup is specified
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
		fmt.Printf("\n%sGroup: %s (Description: %s)%s\n", Cyan, groupName, group.Description, Reset)

		// Create a map for quick task lookup by name for the original group
		originalTaskMap := make(map[string]*types.Task)
		for i := range group.Tasks {
			originalTaskMap[group.Tasks[i].Name] = &group.Tasks[i]
		}

		// Resolve tasks to run, including dependencies
		tasksToRun, err := tr.resolveTasksToRun(originalTaskMap, tr.TargetTasks)
		if err != nil {
			allGroupErrors = append(allGroupErrors, fmt.Errorf("error resolving tasks for group '%s': %w", groupName, err))
			continue
		}

		if len(tasksToRun) == 0 {
			fmt.Printf("No tasks to run in group '%s' after filtering.\n", groupName)
			continue
		}

		// Map to store outputs of tasks within this group
		taskOutputs := make(map[string]*lua.LTable)
		completedTasks := make(map[string]bool)
		runningTasks := make(map[string]bool)
		var wg sync.WaitGroup
		var mu sync.Mutex
		errChan := make(chan error, len(tasksToRun))

		// Create a map for quick task lookup by name for tasksToRun
		taskMap := make(map[string]*types.Task)
		for _, task := range tasksToRun {
			taskMap[task.Name] = task
		}

		// Loop until all tasks are completed
		for {
			mu.Lock()
			if len(completedTasks) == len(tasksToRun) {
				mu.Unlock()
				break // All tasks completed
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
				// Check all dependencies
				for _, depName := range task.DependsOn {
					// Only consider dependencies that are part of the tasksToRun set
					if _, exists := taskMap[depName]; exists {
						mu.Lock()
						if !completedTasks[depName] {
							dependencyMet = false
						}
						mu.Unlock()
						if !dependencyMet {
							break
						}
					}
				}

				if dependencyMet {
					tasksLaunchedThisIteration++
					// Resolve input from dependencies
					inputFromDependencies := tr.L.NewTable()
					for _, depName := range task.DependsOn {
						if _, exists := taskMap[depName]; exists { // Only add output if dependency is part of tasksToRun
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
							if err := tr.runTask(t, input, &mu, completedTasks, taskOutputs, runningTasks); err != nil {
								errChan <- err
							}
						}(task, inputFromDependencies)
					} else {
						mu.Lock()
						runningTasks[task.Name] = true
						mu.Unlock()
						if err := tr.runTask(task, inputFromDependencies, &mu, completedTasks, taskOutputs, runningTasks); err != nil {
							errChan <- err
						}
					}
				}
			}

			// If no tasks were launched in this iteration, and not all tasks are completed,
			// it means there's a circular dependency or unresolvable tasks.
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
					circularErr := NewTaskError("group "+groupName, fmt.Errorf("circular dependency or unresolvable tasks. Uncompleted tasks: %v", uncompletedTasks), ErrorTypeDependency)
					log.Printf(circularErr.Error())
					errChan <- circularErr
					break // Break the inner loop to stop processing this group
				}
				mu.Unlock()
			}
			wg.Wait() // Wait for all tasks launched in this iteration to finish
		}

		close(errChan) // Close the channel after all tasks are done for this group

		// Collect errors for the current group
		var groupErrors []error
		for err := range errChan {
			groupErrors = append(groupErrors, err)
		}

		if len(groupErrors) > 0 {
			allGroupErrors = append(allGroupErrors, fmt.Errorf("task group '%s' encountered errors: %v", groupName, groupErrors))
		}
	}

	if len(allGroupErrors) > 0 {
		return fmt.Errorf("one or more task groups failed: %v", allGroupErrors)
	}
	return nil
}

// resolveTasksToRun determines the final list of tasks to execute, including their dependencies.
func (tr *TaskRunner) resolveTasksToRun(originalTaskMap map[string]*types.Task, targetTasks []string) ([]*types.Task, error) {
	if len(targetTasks) == 0 {
		// If no specific tasks are targeted, run all tasks in the group
		var allTasks []*types.Task
		for _, task := range originalTaskMap {
			allTasks = append(allTasks, task)
		}
		return allTasks, nil
	}

	resolved := make(map[string]*types.Task)
	queue := make([]string, 0, len(targetTasks))
	visited := make(map[string]bool)

	// Add target tasks to the queue, only if they exist in the current group
	var actualTargetTasksInGroup []string
	for _, taskName := range targetTasks {
		if _, ok := originalTaskMap[taskName]; ok {
			actualTargetTasksInGroup = append(actualTargetTasksInGroup, taskName)
		} else {
			log.Printf("Warning: Targeted task '%s' not found in group. Skipping for this group.", taskName)
		}
	}

	if len(actualTargetTasksInGroup) == 0 {
		return []*types.Task{}, nil // No tasks to run in this group if none of the targeted tasks exist here
	}

	for _, taskName := range actualTargetTasksInGroup {
		queue = append(queue, taskName)
		visited[taskName] = true
	}

	// BFS to find all dependencies
	head := 0
	for head < len(queue) {
		currentTaskName := queue[head]
		head++

		currentTask := originalTaskMap[currentTaskName]
		resolved[currentTaskName] = currentTask

		for _, depName := range currentTask.DependsOn {
			if _, ok := originalTaskMap[depName]; !ok {
				return nil, fmt.Errorf("dependency '%s' for task '%s' not found in group", depName, currentTaskName)
			}
			if !visited[depName] {
				visited[depName] = true
				queue = append(queue, depName)
			}
		}
	}

	// Convert resolved map to a slice
	var result []*types.Task
	for _, task := range resolved {
		result = append(result, task)
	}
	return result, nil
}

