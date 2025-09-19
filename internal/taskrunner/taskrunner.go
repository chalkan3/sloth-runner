package taskrunner

import (
	"context"
	"fmt"
	"io/ioutil"
	"log"
	"os"
	"os/exec"
	"path/filepath"
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

type TaskRunner struct {
	L           *lua.LState
	TaskGroups  map[string]types.TaskGroup
	TargetGroup string
	TargetTasks []string
	Results     []types.TaskResult
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

func (tr *TaskRunner) executeTaskWithRetries(t *types.Task, inputFromDependencies *lua.LTable, mu *sync.Mutex, completedTasks map[string]bool, taskOutputs map[string]*lua.LTable, runningTasks map[string]bool, session *types.SharedSession) error {
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
			tr.Results = append(tr.Results, types.TaskResult{
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
			tr.Results = append(tr.Results, types.TaskResult{
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
				spinner.Fail(fmt.Sprintf("Task '%s' has invalid timeout: %v", t.Name, err))
				return &TaskExecutionError{TaskName: t.Name, Err: fmt.Errorf("invalid timeout duration: %w", err)}
			}
			ctx, cancel = context.WithTimeout(context.Background(), timeout)
		} else {
			ctx, cancel = context.WithCancel(context.Background())
		}

		taskErr = tr.runTask(ctx, t, inputFromDependencies, mu, completedTasks, taskOutputs, runningTasks, session)
		cancel() // Cancel the context as soon as the task is done

		if taskErr == nil {
			spinner.Success(fmt.Sprintf("Task '%s' completed successfully", t.Name))
			return nil // Success
		}
	}

	spinner.Fail(fmt.Sprintf("Task '%s' failed after %d retries: %v", t.Name, t.Retries, taskErr))
	return taskErr // Final failure
}

func (tr *TaskRunner) runTask(ctx context.Context, t *types.Task, inputFromDependencies *lua.LTable, mu *sync.Mutex, completedTasks map[string]bool, taskOutputs map[string]*lua.LTable, runningTasks map[string]bool, session *types.SharedSession) (taskErr error) {
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
		tr.Results = append(tr.Results, types.TaskResult{
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
		// Prepara o ambiente Lua para garantir que os módulos globais (log, exec, etc.) estejam disponíveis.
		// Esta é a correção para o bug 'log is nil'.
		luainterface.OpenAll(tr.L)

		// Pass session to Lua context
		var sessionUD *lua.LUserData
		if session != nil {
			sessionUD = tr.L.NewUserData()
			sessionUD.Value = session
		}

		success, msg, outputTable, err := luainterface.ExecuteLuaFunction(tr.L, t.CommandFunc, t.Params, inputFromDependencies, 3, ctx, sessionUD)
		if err != nil {
			return &TaskExecutionError{TaskName: t.Name, Err: fmt.Errorf("error executing command function: %w", err)}
		} else if !success {
			return &TaskExecutionError{TaskName: t.Name, Err: fmt.Errorf("command function returned failure: %s", msg)}
		} else if outputTable != nil {
			t.Output = outputTable
		}
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

		// --- Início: Lógica de Gerenciamento do Workdir ---
		var workdir string
		var err error
		if group.CreateWorkdirBeforeRun {
			workdir = filepath.Join(os.TempDir(), groupName)
			if err := os.RemoveAll(workdir); err != nil {
				return fmt.Errorf("failed to clean fixed workdir %s: %w", workdir, err)
			}
		} else {
			// Cria um diretório temporário único
			workdir, err = ioutil.TempDir(os.TempDir(), groupName+"-*")
			if err != nil {
				return fmt.Errorf("failed to create ephemeral workdir: %w", err)
			}
		}

		if err := os.MkdirAll(workdir, 0755); err != nil {
			return fmt.Errorf("failed to create workdir %s: %w", workdir, err)
		}

		// Lógica de limpeza adiada
		defer func() {
			shouldClean := true // Padrão é limpar
			if group.CleanWorkdirAfterRunFunc != nil {
				// Pega o resultado da última tarefa executada no grupo
				var lastResult types.TaskResult
				if len(tr.Results) > 0 {
					lastResult = tr.Results[len(tr.Results)-1]
				}

				// Converte o resultado para uma tabela Lua para passar para a função
				resultTable := tr.L.NewTable()
				resultTable.RawSetString("success", lua.LBool(lastResult.Error == nil))
				// Adiciona mais campos se necessário no futuro

				success, _, _, err := luainterface.ExecuteLuaFunction(tr.L, group.CleanWorkdirAfterRunFunc, nil, resultTable, 1, context.Background())
				if err != nil {
					pterm.Error.Printf("Error executing clean_workdir_after_run for group '%s': %v\n", groupName, err)
				} else {
					shouldClean = success
				}
			}

			if shouldClean {
				pterm.Info.Printf("Cleaning up workdir for group '%s': %s\n", groupName, workdir)
				os.RemoveAll(workdir)
			} else {
				pterm.Warning.Printf("Workdir for group '%s' preserved at: %s\n", groupName, workdir)
			}
		}()
		// --- Fim: Lógica de Gerenciamento do Workdir ---

		// SHARED SESSION LOGIC
		var session *types.SharedSession
		if group.ExecutionMode == "shared_session" {
			var err error
			session, err = NewSharedSession()
			if err != nil {
				return fmt.Errorf("failed to create shared session for group '%s': %w", groupName, err)
			}
			defer session.Close()
		}

		originalTaskMap := make(map[string]*types.Task)
		for i := range group.Tasks {
			task := &group.Tasks[i]
			// Injeta o workdir nos parâmetros da tarefa
			if task.Params == nil {
				task.Params = make(map[string]string)
			}
			task.Params["workdir"] = workdir
			originalTaskMap[task.Name] = task
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

		// --- Task Execution Loop ---
		// This loop is simplified for clarity. The original logic for dependency resolution remains.
		// We will execute tasks sequentially for shared_session mode.
		if session != nil {
			// Sequential execution for shared session
			for _, taskName := range getExecutionOrder(taskMap) { // Assumes getExecutionOrder provides a topologically sorted list of task names
				task := taskMap[taskName]
				// Simplified input calculation for this example
				inputFromDependencies := tr.L.NewTable()
				// ... (dependency input logic would go here)

				// Pass the session to the task execution
				if err := tr.executeTaskWithRetries(task, inputFromDependencies, &mu, completedTasks, taskOutputs, runningTasks, session); err != nil {
					errChan <- err
					// In sequential execution, we might want to break on failure
					break
				}
			}
		} else {
			// Original parallel execution logic
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
					// Check 'depends_on' dependencies
					for _, depName := range task.DependsOn {
						if _, exists := taskMap[depName]; exists {
							mu.Lock()
							completed := completedTasks[depName]
							if !completed {
								dependencyMet = false
							} else {
								var depSuccess bool
								for _, res := range tr.Results {
									if res.Name == depName && res.Error == nil {
										depSuccess = true
										break
									}
								}
								if !depSuccess {
									dependencyMet = false
									tr.Results = append(tr.Results, types.TaskResult{
										Name:   task.Name,
										Status: "Skipped",
										Error:  fmt.Errorf("dependency '%s' failed", depName),
									})
									completedTasks[task.Name] = true
								}
							}
							mu.Unlock()
							if !dependencyMet {
								break
							}
						}
					}

					// Check 'next_if_fail' dependencies only if 'depends_on' are met
					if dependencyMet && len(task.NextIfFail) > 0 {
						nextIfFailMet := false
						allNextIfFailCompleted := true
						for _, depName := range task.NextIfFail {
							if _, exists := taskMap[depName]; exists {
								mu.Lock()
								completed := completedTasks[depName]
								if !completed {
									allNextIfFailCompleted = false
								} else {
									var depFailed bool
									for _, res := range tr.Results {
										if res.Name == depName && res.Error != nil {
											depFailed = true
											break
										}
									}
									if depFailed {
										nextIfFailMet = true
									}
								}
								mu.Unlock()
							}
						}

						if !allNextIfFailCompleted {
							dependencyMet = false
						} else if !nextIfFailMet {
							dependencyMet = false
							mu.Lock()
							tr.Results = append(tr.Results, types.TaskResult{
								Name:   task.Name,
								Status: "Skipped",
								Error:  fmt.Errorf("none of the next_if_fail dependencies for '%s' failed", task.Name),
							})
							completedTasks[task.Name] = true
							mu.Unlock()
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
						for _, depName := range task.NextIfFail {
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
								if err := tr.executeTaskWithRetries(t, input, &mu, completedTasks, taskOutputs, runningTasks, nil); err != nil {
									errChan <- err
								}
							}(task, inputFromDependencies)
						} else {
							mu.Lock()
							runningTasks[task.Name] = true
							mu.Unlock()
							if err := tr.executeTaskWithRetries(task, inputFromDependencies, &mu, completedTasks, taskOutputs, runningTasks, nil); err != nil {
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

// Helper function to get a sequential, dependency-respecting order of tasks.
// NOTE: This is a placeholder for a proper topological sort implementation.
func getExecutionOrder(taskMap map[string]*types.Task) []string {
	// This should be a topological sort based on DependsOn.
	// For this example, we'll use a simplified, hardcoded order.
	// A real implementation is required here.
	names := make([]string, 0, len(taskMap))
	for name := range taskMap {
		names = append(names, name)
	}
	// This is NOT a correct sort, just for placeholder.
	return names
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

// RunTasksParallel executes a slice of tasks concurrently and waits for them to complete.
func (tr *TaskRunner) RunTasksParallel(tasks []*types.Task, input *lua.LTable) ([]types.TaskResult, error) {
	var wg sync.WaitGroup
	var mu sync.Mutex

	resultsChan := make(chan types.TaskResult, len(tasks))
	errChan := make(chan error, len(tasks))

	for _, task := range tasks {
		wg.Add(1)
		go func(t *types.Task) {
			defer wg.Done()

			// These maps are for the context of a single parallel execution, not the whole runner
			completed := make(map[string]bool)
			outputs := make(map[string]*lua.LTable)
			running := make(map[string]bool)

			// We use a new mutex for the retry logic within the task execution
			var taskMu sync.Mutex

			err := tr.executeTaskWithRetries(t, input, &taskMu, completed, outputs, running, nil)

			// Find the result for this specific task. executeTaskWithRetries appends to tr.Results.
			// This is a bit of a workaround due to the existing structure.
			// A better long-term solution would be for executeTaskWithRetries to return the TaskResult.
			mu.Lock()
			var result types.TaskResult
			// Search from the end of the results slice as it's the most likely place
			for i := len(tr.Results) - 1; i >= 0; i-- {
				if tr.Results[i].Name == t.Name {
					result = tr.Results[i]
					break
				}
			}
			mu.Unlock()

			if err != nil {
				errChan <- err
			}
			resultsChan <- result

		}(task)
	}

	wg.Wait()
	close(resultsChan)
	close(errChan)

	var allErrors []error
	for err := range errChan {
		allErrors = append(allErrors, err)
	}

	var results []types.TaskResult
	for result := range resultsChan {
		results = append(results, result)
	}

	if len(allErrors) > 0 {
		// Combine multiple errors into one
		return results, fmt.Errorf("encountered %d errors during parallel execution: %v", len(allErrors), allErrors)
	}

	return results, nil
}
