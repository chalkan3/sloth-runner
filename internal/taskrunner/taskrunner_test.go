package taskrunner

import (
	"sync"
	"testing"

	"github.com/chalkan3/sloth-runner/internal/types"
	"github.com/stretchr/testify/assert"
	lua "github.com/yuin/gopher-lua"
)

func TestNewTaskRunner(t *testing.T) {
	L := lua.NewState()
	defer L.Close()
	groups := make(map[string]types.TaskGroup)
	targetGroup := "test"
	targetTasks := []string{"task1"}

	tr := NewTaskRunner(L, groups, targetGroup, targetTasks)

	assert.Equal(t, L, tr.L)
	assert.Equal(t, groups, tr.TaskGroups)
	assert.Equal(t, targetGroup, tr.TargetGroup)
	assert.Equal(t, targetTasks, tr.TargetTasks)
}

func TestResolveTasksToRun(t *testing.T) {
	task1 := &types.Task{Name: "task1", DependsOn: []string{"task2"}}
	task2 := &types.Task{Name: "task2"}
	task3 := &types.Task{Name: "task3"}
	originalTaskMap := map[string]*types.Task{
		"task1": task1,
		"task2": task2,
		"task3": task3,
	}

	tr := &TaskRunner{}

	t.Run("no target tasks", func(t *testing.T) {
		tasks, err := tr.resolveTasksToRun(originalTaskMap, []string{})
		assert.NoError(t, err)
		assert.ElementsMatch(t, []*types.Task{task1, task2, task3}, tasks)
	})

	t.Run("single target task with dependency", func(t *testing.T) {
		tasks, err := tr.resolveTasksToRun(originalTaskMap, []string{"task1"})
		assert.NoError(t, err)
		assert.ElementsMatch(t, []*types.Task{task1, task2}, tasks)
	})

	t.Run("multiple target tasks", func(t *testing.T) {
		tasks, err := tr.resolveTasksToRun(originalTaskMap, []string{"task1", "task3"})
		assert.NoError(t, err)
		assert.ElementsMatch(t, []*types.Task{task1, task2, task3}, tasks)
	})

	t.Run("dependency not found", func(t *testing.T) {
		task4 := &types.Task{Name: "task4", DependsOn: []string{"task5"}}
		taskMap := map[string]*types.Task{"task4": task4}
		_, err := tr.resolveTasksToRun(taskMap, []string{"task4"})
		assert.Error(t, err)
	})
}

func TestRunTask(t *testing.T) {
	L := lua.NewState()
	defer L.Close()

	tr := NewTaskRunner(L, nil, "", nil)

	t.Run("successful execution", func(t *testing.T) {
		task := &types.Task{
			Name:       "task1",
			CommandStr: "echo 'success'",
		}
		err := tr.runTask(task, L.NewTable(), &sync.Mutex{}, make(map[string]bool), make(map[string]*lua.LTable), make(map[string]bool))
		assert.NoError(t, err)
	})

	t.Run("execution with hooks", func(t *testing.T) {
		L.DoString(`
			function pre_exec_hook(params, inputs)
				return true, "pre-exec success"
			end
			function post_exec_hook(params, output)
				return true, "post-exec success"
			end
		`)
		task := &types.Task{
			Name:       "task2",
			CommandStr: "echo 'success'",
			PreExec:    L.GetGlobal("pre_exec_hook").(*lua.LFunction),
			PostExec:   L.GetGlobal("post_exec_hook").(*lua.LFunction),
		}
		err := tr.runTask(task, L.NewTable(), &sync.Mutex{}, make(map[string]bool), make(map[string]*lua.LTable), make(map[string]bool))
		assert.NoError(t, err)
	})

	t.Run("pre-exec hook failure", func(t *testing.T) {
		L.DoString(`
			function pre_exec_fail(params, inputs)
				return false, "pre-exec failure"
			end
		`)
		task := &types.Task{
			Name:    "task3",
			PreExec: L.GetGlobal("pre_exec_fail").(*lua.LFunction),
		}
		err := tr.runTask(task, L.NewTable(), &sync.Mutex{}, make(map[string]bool), make(map[string]*lua.LTable), make(map[string]bool))
		assert.Error(t, err)
	})

	t.Run("post-exec hook failure", func(t *testing.T) {
		L.DoString(`
			function post_exec_fail(params, output)
				return false, "post-exec failure"
			end
		`)
		task := &types.Task{
			Name:       "task4",
			CommandStr: "echo 'success'",
			PostExec:   L.GetGlobal("post_exec_fail").(*lua.LFunction),
		}
		err := tr.runTask(task, L.NewTable(), &sync.Mutex{}, make(map[string]bool), make(map[string]*lua.LTable), make(map[string]bool))
		assert.Error(t, err)
	})

	t.Run("command function failure", func(t *testing.T) {
		L.DoString(`
			function command_fail(params, inputs)
				return false, "command failure", nil
			end
		`)
		task := &types.Task{
			Name:        "task5",
			CommandFunc: L.GetGlobal("command_fail").(*lua.LFunction),
		}
		err := tr.runTask(task, L.NewTable(), &sync.Mutex{}, make(map[string]bool), make(map[string]*lua.LTable), make(map[string]bool))
		assert.Error(t, err)
	})

	t.Run("no command defined", func(t *testing.T) {
		task := &types.Task{Name: "task6"}
		err := tr.runTask(task, L.NewTable(), &sync.Mutex{}, make(map[string]bool), make(map[string]*lua.LTable), make(map[string]bool))
		assert.Error(t, err)
	})
}

func TestRun(t *testing.T) {
	L := lua.NewState()
	defer L.Close()

	t.Run("successful run with dependencies", func(t *testing.T) {
		task1 := types.Task{Name: "task1", CommandStr: "echo 'task1'", DependsOn: []string{"task2"}}
		task2 := types.Task{Name: "task2", CommandStr: "echo 'task2'"}
		groups := map[string]types.TaskGroup{
			"group1": {Tasks: []types.Task{task1, task2}},
		}
		tr := NewTaskRunner(L, groups, "", nil)
		err := tr.Run()
		assert.NoError(t, err)
	})

	t.Run("run with async tasks", func(t *testing.T) {
		task1 := types.Task{Name: "task1", CommandStr: "echo 'task1'", Async: true}
		task2 := types.Task{Name: "task2", CommandStr: "echo 'task2'", Async: true}
		groups := map[string]types.TaskGroup{
			"group1": {Tasks: []types.Task{task1, task2}},
		}
		tr := NewTaskRunner(L, groups, "", nil)
		err := tr.Run()
		assert.NoError(t, err)
	})

	t.Run("run with circular dependency", func(t *testing.T) {
		task1 := types.Task{Name: "task1", CommandStr: "echo 'task1'", DependsOn: []string{"task2"}}
		task2 := types.Task{Name: "task2", CommandStr: "echo 'task2'", DependsOn: []string{"task1"}}
		groups := map[string]types.TaskGroup{
			"group1": {Tasks: []types.Task{task1, task2}},
		}
		tr := NewTaskRunner(L, groups, "", nil)
		err := tr.Run()
		assert.Error(t, err)
	})

	t.Run("run with target group", func(t *testing.T) {
		task1 := types.Task{Name: "task1", CommandStr: "echo 'task1'"}
		task2 := types.Task{Name: "task2", CommandStr: "echo 'task2'"}
		groups := map[string]types.TaskGroup{
			"group1": {Tasks: []types.Task{task1}},
			"group2": {Tasks: []types.Task{task2}},
		}
		tr := NewTaskRunner(L, groups, "group1", nil)
		err := tr.Run()
		assert.NoError(t, err)
	})

	t.Run("run with target tasks", func(t *testing.T) {
		task1 := types.Task{Name: "task1", CommandStr: "echo 'task1'", DependsOn: []string{"task2"}}
		task2 := types.Task{Name: "task2", CommandStr: "echo 'task2'"}
		task3 := types.Task{Name: "task3", CommandStr: "echo 'task3'"}
		groups := map[string]types.TaskGroup{
			"group1": {Tasks: []types.Task{task1, task2, task3}},
		}
		tr := NewTaskRunner(L, groups, "", []string{"task1"})
		err := tr.Run()
		assert.NoError(t, err)
	})

	t.Run("run with target group not found", func(t *testing.T) {
		groups := map[string]types.TaskGroup{
			"group1": {},
		}
		tr := NewTaskRunner(L, groups, "group2", nil)
		err := tr.Run()
		assert.Error(t, err)
	})
}
