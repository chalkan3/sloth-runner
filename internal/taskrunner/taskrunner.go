package taskrunner

import (
	"fmt"
	"log"
	"sync"
	"time"

	"github.com/chalkan3/sloth-runner/internal/luainterface"
	"github.com/chalkan3/sloth-runner/internal/types"
	"github.com/pterm/pterm"
	lua "github.com/yuin/gopher-lua"
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
}

func NewTaskRunner(L *lua.LState, groups map[string]types.TaskGroup, targetGroup string, targetTasks []string) *TaskRunner {
	return &TaskRunner{
		L:           L,
		TaskGroups:  groups,
		TargetGroup: targetGroup,
		TargetTasks: targetTasks,
		Outputs:     make(map[string]interface{}),
	}
}

func (tr *TaskRunner) runTask(t *types.Task, inputFromDependencies *lua.LTable, mu *sync.Mutex, completedTasks map[string]bool, taskOutputs map[string]*lua.LTable, runningTasks map[string]bool) (taskErr error) {
	startTime := time.Now()
	t.Output = tr.L.NewTable()

	spinner, _ := pterm.DefaultSpinner.Start(fmt.Sprintf("Executing task: %s", t.Name))

	defer func() {
		duration := time.Since(startTime)
		status := "Success"
		if taskErr != nil {
			status = "Failed"
			spinner.Fail(fmt.Sprintf("Task '%s' failed: %v", t.Name, taskErr))
		} else {
			spinner.Success(fmt.Sprintf("Task '%s' completed successfully", t.Name))
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
		success, msg, _, err := luainterface.ExecuteLuaFunction(tr.L, t.PreExec, t.Params, inputFromDependencies, 2)
		if err != nil {
			taskErr = NewTaskError(t.Name, fmt.Errorf("error executing pre_exec hook: %w", err), ErrorTypePreExec)
		} else if !success {
			taskErr = NewTaskError(t.Name, fmt.Errorf("pre-execution hook failed: %s", msg), ErrorTypePreExec)
		}
	}

	if taskErr == nil {
		if t.CommandFunc != nil {
			success, msg, outputTable, err := luainterface.ExecuteLuaFunction(tr.L, t.CommandFunc, t.Params, inputFromDependencies, 3)
			if err != nil {
				taskErr = NewTaskError(t.Name, fmt.Errorf("error executing command function: %w", err), ErrorTypeCommand)
			} else if !success {
				taskErr = NewTaskError(t.Name, fmt.Errorf("command function returned failure: %s", msg), ErrorTypeCommand)
			} else if outputTable != nil {
				t.Output = outputTable
			}
		} else if t.CommandStr == "" {
			taskErr = NewTaskError(t.Name, fmt.Errorf("task has no command defined"), ErrorTypeCommand)
		}
	}

	if taskErr == nil && t.PostExec != nil {
		var postExecSecondArg lua.LValue = t.Output
		if t.Output == nil {
			postExecSecondArg = tr.L.NewTable()
		}
		success, msg, _, err := luainterface.ExecuteLuaFunction(tr.L, t.PostExec, t.Params, postExecSecondArg, 2)
		if err != nil {
			taskErr = NewTaskError(t.Name, fmt.Errorf("error executing post_exec hook: %w", err), ErrorTypePostExec)
		} else if !success {
			taskErr = NewTaskError(t.Name, fmt.Errorf("post-execution hook failed: %s", msg), ErrorTypePostExec)
		}
	}

	return taskErr
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
		}
		tableData = append(tableData, []string{result.Name, status, result.Duration.String(), errStr})
	}
	pterm.DefaultTable.WithHasHeader().WithData(tableData).Render()

	if len(allGroupErrors) > 0 {
		return fmt.Errorf("one or more task groups failed: %v", allGroupErrors)
	}
	return nil
}

// resolveTasksToRun determines the final list of tasks to execute, including their dependencies.
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

