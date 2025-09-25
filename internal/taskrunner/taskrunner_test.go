package taskrunner

import (
	"testing"

	"github.com/chalkan3/sloth-runner/internal/types"
	"github.com/stretchr/testify/assert"
	lua "github.com/yuin/gopher-lua"
)

// TestRun_Successful_DependencyResolution validates that a simple dependency graph is resolved correctly.
func TestRun_Successful_DependencyResolution(t *testing.T) {
	L := lua.NewState()
	defer L.Close()

	// Using "true" as a command that is guaranteed to exist and succeed.
	task1 := types.Task{Name: "task1", CommandStr: "true", DependsOn: []string{"task2"}}
	task2 := types.Task{Name: "task2", CommandStr: "true"}
	groups := map[string]types.TaskGroup{
		"test_group": {Tasks: []types.Task{task1, task2}},
	}
	tr := NewTaskRunner(L, groups, "test_group", nil, false, false)
	err := tr.Run()
	assert.NoError(t, err)
}

// TestRun_Failure_CircularDependency validates that the runner correctly identifies a circular dependency.
func TestRun_Failure_CircularDependency(t *testing.T) {
	L := lua.NewState()
	defer L.Close()

	task1 := types.Task{Name: "task1", CommandStr: "true", DependsOn: []string{"task2"}}
	task2 := types.Task{Name: "task2", CommandStr: "true", DependsOn: []string{"task1"}}
	groups := map[string]types.TaskGroup{
		"test_group": {Tasks: []types.Task{task1, task2}},
	}
	tr := NewTaskRunner(L, groups, "test_group", []string{}, false, false)
	err := tr.Run()
	assert.Error(t, err)
	assert.Contains(t, err.Error(), "circular dependency")
}
