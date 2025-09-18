package luainterface

import (
	"bytes"
	"encoding/json"
	"fmt"
	"log"
	"strings"

	lua "github.com/yuin/gopher-lua"
)

// luaPulumiStackTypeName is the name of the Lua userdata type for PulumiStack.
const luaPulumiStackTypeName = "pulumi_stack"

// PulumiStack holds the state for a fluent Pulumi API call.
type PulumiStack struct {
	StackName string
	WorkDir   string
}

// OpenPulumi registers the 'pulumi' module with the Lua state.
func OpenPulumi(L *lua.LState) {
	// Create the metatable for the PulumiStack type.
	mt := L.NewTypeMetatable(luaPulumiStackTypeName)
	L.SetGlobal(luaPulumiStackTypeName, mt) // Optional: make metatable available globally.

	// Register methods for the PulumiStack type.
	methods := map[string]lua.LGFunction{
		"up":      pulumiStackUp,
		"preview": pulumiStackPreview,
		"refresh": pulumiStackRefresh,
		"destroy": pulumiStackDestroy,
		"outputs": pulumiStackOutputs,
	}
	L.SetField(mt, "__index", L.SetFuncs(L.NewTable(), methods))

	// Create the main 'pulumi' module table.
	pulumiModule := L.NewTable()

	// Register the entry point function 'stack'.
	pulumiFuncs := map[string]lua.LGFunction{
		"stack": newPulumiStack,
	}
	L.SetFuncs(pulumiModule, pulumiFuncs)

	// Make the 'pulumi' module available globally.
	L.SetGlobal("pulumi", pulumiModule)
}

// newPulumiStack is the entry point for the fluent API, exposed as pulumi.stack(name, options_table).
func newPulumiStack(L *lua.LState) int {
	stackName := L.CheckString(1)
	options := L.OptTable(2, L.NewTable()) // Optional options table

	workDir := options.RawGetString("workdir").String()
	if workDir == "" {
		L.Push(lua.LNil) // Return nil for the stack object
		L.Push(lua.LString("workdir option is required for pulumi.stack()")) // Return error message
		return 2
	}

	stack := &PulumiStack{
		StackName: stackName,
		WorkDir:   workDir,
	}

	ud := L.NewUserData()
	ud.Value = stack
	L.SetMetatable(ud, L.GetTypeMetatable(luaPulumiStackTypeName))
	L.Push(ud)
	L.Push(lua.LNil) // No error
	return 2
}

// checkPulumiStack retrieves the PulumiStack struct from a Lua userdata.
func checkPulumiStack(L *lua.LState) *PulumiStack {
	ud := L.CheckUserData(1)
	if v, ok := ud.Value.(*PulumiStack); ok {
		return v
	}
	L.ArgError(1, "pulumi_stack expected")
	return nil
}

// runPulumiCommand executes a Pulumi CLI command and returns its structured output to Lua.
func runPulumiCommand(L *lua.LState, stack *PulumiStack, command string, args []string, options *lua.LTable) int {
	var pulumiArgs []string
	pulumiArgs = append(pulumiArgs, command)
	pulumiArgs = append(pulumiArgs, "--stack", stack.StackName)
	pulumiArgs = append(pulumiArgs, "--cwd", stack.WorkDir)

	// Handle common options
	if options != nil {
		if nonInteractive := options.RawGetString("non_interactive"); nonInteractive.Type() == lua.LTBool && lua.LVAsBool(nonInteractive) {
			pulumiArgs = append(pulumiArgs, "--non-interactive", "--yes") // --yes is often needed with --non-interactive
		}
		// Handle --config
		if configTable := options.RawGetString("config"); configTable.Type() == lua.LTTable {
			configTable.(*lua.LTable).ForEach(func(key, val lua.LValue) {
				pulumiArgs = append(pulumiArgs, "--config", fmt.Sprintf("%s=%s", key.String(), val.String()))
			})
		}
		// Add any other specific args passed directly
		if extraArgs := options.RawGetString("args"); extraArgs.Type() == lua.LTTable {
			extraArgs.(*lua.LTable).ForEach(func(_, argVal lua.LValue) {
				pulumiArgs = append(pulumiArgs, argVal.String())
			})
		}
	}

	cmd := ExecCommand("pulumi", pulumiArgs...)
	var stdoutBuf, stderrBuf bytes.Buffer
	cmd.Stdout = &stdoutBuf
	cmd.Stderr = &stderrBuf

	log.Printf("Executing Pulumi command: pulumi %s", strings.Join(pulumiArgs, " "))

	err := cmd.Run()

	resultTable := L.NewTable()
	resultTable.RawSetString("stdout", lua.LString(stdoutBuf.String()))
	resultTable.RawSetString("stderr", lua.LString(stderrBuf.String()))

	if err != nil {
		resultTable.RawSetString("success", lua.LBool(false))
		resultTable.RawSetString("error", lua.LString(err.Error()))
	} else {
		resultTable.RawSetString("success", lua.LBool(true))
		resultTable.RawSetString("error", lua.LNil)
	}

	L.Push(resultTable)
	return 1
}

// pulumiStackUp implements the .up() method for PulumiStack.
func pulumiStackUp(L *lua.LState) int {
	stack := checkPulumiStack(L)
	if stack == nil {
		return 0
	}
	options := L.OptTable(2, L.NewTable()) // Optional options table
	return runPulumiCommand(L, stack, "up", nil, options)
}

// pulumiStackPreview implements the .preview() method for PulumiStack.
func pulumiStackPreview(L *lua.LState) int {
	stack := checkPulumiStack(L)
	if stack == nil {
		return 0
	}
	options := L.OptTable(2, L.NewTable()) // Optional options table
	return runPulumiCommand(L, stack, "preview", nil, options)
}

// pulumiStackRefresh implements the .refresh() method for PulumiStack.
func pulumiStackRefresh(L *lua.LState) int {
	stack := checkPulumiStack(L)
	if stack == nil {
		return 0
	}
	options := L.OptTable(2, L.NewTable()) // Optional options table
	return runPulumiCommand(L, stack, "refresh", nil, options)
}

// pulumiStackDestroy implements the .destroy() method for PulumiStack.
func pulumiStackDestroy(L *lua.LState) int {
	stack := checkPulumiStack(L)
	if stack == nil {
		return 0
	}
	options := L.OptTable(2, L.NewTable()) // Optional options table
	return runPulumiCommand(L, stack, "destroy", nil, options)
}

// pulumiStackOutputs implements the .outputs() method for PulumiStack.
// It returns a Lua table representing the stack outputs.
func pulumiStackOutputs(L *lua.LState) int {
	stack := checkPulumiStack(L)
	if stack == nil {
		return 0
	}

	// Execute 'pulumi stack output --json'
	var pulumiArgs []string
	pulumiArgs = append(pulumiArgs, "stack", "output", "--json")
	pulumiArgs = append(pulumiArgs, "--stack", stack.StackName)
	pulumiArgs = append(pulumiArgs, "--cwd", stack.WorkDir)

	cmd := ExecCommand("pulumi", pulumiArgs...)
	var stdoutBuf, stderrBuf bytes.Buffer
	cmd.Stdout = &stdoutBuf
	cmd.Stderr = &stderrBuf

	log.Printf("Executing Pulumi command for outputs: pulumi %s", strings.Join(pulumiArgs, " "))

	err := cmd.Run()

	if err != nil {
		L.Push(lua.LNil) // No outputs table
		L.Push(lua.LString(fmt.Sprintf("failed to get pulumi outputs: %s, stderr: %s", err.Error(), stderrBuf.String())))
		return 2
	}

	var goOutputs map[string]interface{}
	jsonErr := json.Unmarshal(stdoutBuf.Bytes(), &goOutputs)
	if jsonErr != nil {
		L.Push(lua.LNil) // No outputs table
		L.Push(lua.LString(fmt.Sprintf("failed to parse pulumi outputs JSON: %s, raw stdout: %s", jsonErr.Error(), stdoutBuf.String())))
		return 2
	}

	luaOutputs := GoValueToLua(L, goOutputs)
	L.Push(luaOutputs)
	L.Push(lua.LNil) // No error
	return 2
}
