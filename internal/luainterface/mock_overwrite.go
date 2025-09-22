package luainterface

import (
	"strings"

	lua "github.com/yuin/gopher-lua"
)

// A list of all module functions that can be mocked.
var mockableFunctions = []string{
	"aws.exec", "aws.s3.sync", "aws.secretsmanager.get_secret",
	"azure.exec", "azure.rg.delete", "azure.vm.list",
	"digitalocean.exec", "digitalocean.droplets.list", "digitalocean.droplets.delete",
	"docker.exec", "docker.build", "docker.push", "docker.run",
	"gcp.exec",
	"git.clone", "git.repo", // Note: repo methods like add, commit are harder to mock this way
	"notifications.slack.send", "notifications.ntfy.send",
	"pulumi.stack", // Note: stack methods like up, destroy are harder to mock this way
	"python.venv",
	"salt.client",
	"terraform.init", "terraform.plan", "terraform.apply", "terraform.destroy", "terraform.output",
	"exec.run",
}

// OverwriteModulesWithMocks replaces real module functions with mock handlers.
// This is the core of the mocking framework.
func OverwriteModulesWithMocks(L *lua.LState, testState *TestState) {
	for _, fullName := range mockableFunctions {
		parts := strings.Split(fullName, ".")
		if len(parts) < 2 {
			continue
		}

		moduleName := parts[0]
		funcName := parts[1]

		module := L.GetGlobal(moduleName)
		if module.Type() == lua.LTNil {
			continue
		}

		originalTable := module.(*lua.LTable)
		targetTable := originalTable
		
		// Handle nested modules like aws.s3
		if len(parts) > 2 {
			nestedTableName := parts[1]
			funcName = parts[2]
			nestedTableVal := L.GetField(originalTable, nestedTableName)
			if nestedTableVal.Type() == lua.LTTable {
				targetTable = nestedTableVal.(*lua.LTable)
			} else {
				// Create nested table if it doesn't exist
				newNestedTable := L.NewTable()
				originalTable.RawSetString(nestedTableName, newNestedTable)
				targetTable = newNestedTable
			}
		}

		// Create a closure for the mock handler
		handler := func(l *lua.LState) int {
			mock, found := testState.Mocks[fullName]
			if !found {
				l.RaiseError("strict mock error: function '%s' was called but not mocked in the test.", fullName)
				return 0
			}

			// The mock table is expected to contain a list of return values
			returnValues := mock.RawGetString("returns").(*lua.LTable)
			if returnValues == nil {
				l.RaiseError("invalid mock for '%s': 'returns' key must be a table/list of values.", fullName)
				return 0
			}
			
			returnValues.ForEach(func(_, val lua.LValue) {
				l.Push(val)
			})

			return returnValues.Len()
		}

		targetTable.RawSetString(funcName, L.NewFunction(handler))
	}
}
