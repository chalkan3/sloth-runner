// Package luainterface provides the core bridge between Go and the Lua scripting environment.
// It defines the APIs that are exposed to Lua scripts as global modules, such as 'exec',
// 'fs', 'log', 'test', and 'assert'.
package luainterface

import (
	"fmt"
	"log/slog"

	"github.com/chalkan3/sloth-runner/internal/runner"
	"github.com/chalkan3/sloth-runner/internal/types"
	"github.com/pterm/pterm"
	lua "github.com/yuin/gopher-lua"
)

// TestState holds the collective results of a test file execution. It tracks the
// number of assertions and failures and accumulates the results for the final report.
type TestState struct {
	Assertions int
	Failed     int
	CurrentSuite string
	Results    []pterm.LeveledListItem
}

// --- assert module ---

// newAssertModule creates the 'assert' Lua module. This module provides functions
// for making assertions within test cases.
func newAssertModule(ts *TestState) lua.LGFunction {
	return func(L *lua.LState) int {
		mod := L.SetFuncs(L.NewTable(), map[string]lua.LGFunction{
			// assert.is_true(value, message)
			"is_true": func(L *lua.LState) int {
				ts.Assertions++
				val := L.ToBool(1)
				message := L.ToString(2)
				if !val {
					ts.Failed++
					ts.Results = append(ts.Results, pterm.LeveledListItem{
						Level: 1,
						Text:  pterm.Red(fmt.Sprintf("✗ FAIL: %s - expected true, got false", message)),
					})
				} else {
					ts.Results = append(ts.Results, pterm.LeveledListItem{
						Level: 1,
						Text:  pterm.Green(fmt.Sprintf("✓ PASS: %s", message)),
					})
				}
				return 0
			},
			// assert.equals(actual, expected, message)
			"equals": func(L *lua.LState) int {
				ts.Assertions++
				actual := L.Get(1)
				expected := L.Get(2)
				message := L.ToString(3)
				if actual.String() != expected.String() {
					ts.Failed++
					ts.Results = append(ts.Results, pterm.LeveledListItem{
						Level: 1,
						Text:  pterm.Red(fmt.Sprintf("✗ FAIL: %s - expected '%s', got '%s'", message, expected.String(), actual.String())),
					})
				} else {
					ts.Results = append(ts.Results, pterm.LeveledListItem{
						Level: 1,
						Text:  pterm.Green(fmt.Sprintf("✓ PASS: %s", message)),
					})
				}
				return 0
			},
		})
		L.Push(mod)
		return 1
	}
}

// --- test module ---

// newTestModule creates the 'test' Lua module. This module provides functions
// for structuring tests and running tasks.
func newTestModule(ts *TestState, taskGroups map[string]types.TaskGroup) lua.LGFunction {
	return func(L *lua.LState) int {
		mod := L.SetFuncs(L.NewTable(), map[string]lua.LGFunction{
			// test.describe(name, function)
			"describe": func(L *lua.LState) int {
				suiteName := L.CheckString(1)
				fn := L.CheckFunction(2)
				ts.CurrentSuite = suiteName
				ts.Results = append(ts.Results, pterm.LeveledListItem{Level: 0, Text: suiteName})
				L.Push(fn)
				if err := L.PCall(0, 0, nil); err != nil {
					slog.Error("error executing test suite", "suite", suiteName, "err", err)
				}
				return 0
			},
			// test.it(function)
			"it": func(L *lua.LState) int {
				fn := L.CheckFunction(1)
				L.Push(fn)
				if err := L.PCall(0, 0, nil); err != nil {
					slog.Error("error executing test case", "err", err)
				}
				return 0
			},
			// test.run_task(taskName) -> result_table
			"run_task": func(L *lua.LState) int {
				taskName := L.CheckString(1)
				
				var targetTask *types.Task
				for _, group := range taskGroups {
					for _, task := range group.Tasks {
						if task.Name == taskName {
							targetTask = &task
							break
						}
					}
					if targetTask != nil {
						break
					}
				}

				if targetTask == nil {
					L.Push(lua.LNil)
					L.Push(lua.LString("task not found"))
					return 2
				}

				success, msg, output, duration, err := runner.RunSingleTask(L, targetTask)

				resultTable := L.NewTable()
				resultTable.RawSetString("success", lua.LBool(success))
				resultTable.RawSetString("message", lua.LString(msg))
				resultTable.RawSetString("duration", lua.LString(duration.String()))
				if output != nil {
					resultTable.RawSetString("output", output)
				}
				if err != nil {
					resultTable.RawSetString("error", lua.LString(err.Error()))
				}

				L.Push(resultTable)
				return 1
			},
		})
		L.Push(mod)
		return 1
	}
}

// OpenTesting loads the 'test' and 'assert' modules into the Lua state for a test run.
func OpenTesting(L *lua.LState, ts *TestState, taskGroups map[string]types.TaskGroup) {
	L.PreloadModule("assert", newAssertModule(ts))
	if err := L.DoString(`assert = require("assert")`); err != nil {
		panic(err)
	}
	L.PreloadModule("test", newTestModule(ts, taskGroups))
	if err := L.DoString(`test = require("test")`); err != nil {
		panic(err)
	}
}
