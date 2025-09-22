package luainterface

import (
	"bytes"
	"os/exec"

	"github.com/yuin/gopher-lua"
)

// AzureModule provides Azure functionalities to Lua scripts
type AzureModule struct{}

// NewAzureModule creates a new AzureModule
func NewAzureModule() *AzureModule {
	return &AzureModule{}
}

// Loader returns the Lua loader for the azure module
func (mod *AzureModule) Loader(L *lua.LState) int {
	azTable := L.NewTable()

	// Register azure.exec
	L.SetFuncs(azTable, map[string]lua.LGFunction{
		"exec": mod.azExec,
	})

	// Create and register the rg (resource group) submodule
	rgModule := L.NewTable()
	L.SetFuncs(rgModule, map[string]lua.LGFunction{
		"delete": mod.rgDelete,
	})
	azTable.RawSetString("rg", rgModule)

	// Create and register the vm submodule
	vmModule := L.NewTable()
	L.SetFuncs(vmModule, map[string]lua.LGFunction{
		"list": mod.vmList,
	})
	azTable.RawSetString("vm", vmModule)

	L.Push(azTable)
	return 1
}

// azExec is the generic executor for az commands.
// Lua usage: azure.exec({"group", "list"})
func (mod *AzureModule) azExec(L *lua.LState) int {
	argsTable := L.CheckTable(1)

	var args []string
	argsTable.ForEach(func(_, val lua.LValue) {
		args = append(args, val.String())
	})

	// Ensure JSON output for parsable results, unless already specified
	outputSpecified := false
	for _, arg := range args {
		if arg == "-o" || arg == "--output" {
			outputSpecified = true
			break
		}
	}
	if !outputSpecified {
		args = append(args, "--output", "json")
	}

	cmd := exec.Command("az", args...)
	var stdout, stderr bytes.Buffer
	cmd.Stdout = &stdout
	cmd.Stderr = &stderr

	err := cmd.Run()

	exitCode := 0
	if err != nil {
		if exitErr, ok := err.(*exec.ExitError); ok {
			exitCode = exitErr.ExitCode()
		} else {
			exitCode = -1
		}
	}

	result := L.NewTable()
	result.RawSetString("stdout", lua.LString(stdout.String()))
	result.RawSetString("stderr", lua.LString(stderr.String()))
	result.RawSetString("exit_code", lua.LNumber(exitCode))

	L.Push(result)
	return 1
}

// rgDelete provides a high-level interface for `az group delete`.
// Lua usage: azure.rg.delete({name="my-rg", yes=true})
func (mod *AzureModule) rgDelete(L *lua.LState) int {
	tbl := L.CheckTable(1)
	rgName := tbl.RawGetString("name").String()
	yes := lua.LVAsBool(tbl.RawGetString("yes"))

	if rgName == "" {
		L.Push(lua.LBool(false))
		L.Push(lua.LString("resource group 'name' is required"))
		return 2
	}

	args := []string{"group", "delete", "--name", rgName}
	if yes {
		args = append(args, "--yes")
	}

	// Call the generic exec function
	L.Push(L.NewFunction(mod.azExec))
	argsTable := L.NewTable()
	for _, arg := range args {
		argsTable.Append(lua.LString(arg))
	}
	L.Push(argsTable)
	L.Call(1, 1)

	result := L.CheckTable(L.GetTop())
	exitCode := int(result.RawGetString("exit_code").(lua.LNumber))

	if exitCode != 0 {
		L.Push(lua.LBool(false))
		L.Push(result.RawGetString("stderr"))
		return 2
	}

	L.Push(lua.LBool(true))
	return 1
}

// vmList provides a high-level interface for `az vm list`.
// Lua usage: azure.vm.list({resource_group="my-rg"})
func (mod *AzureModule) vmList(L *lua.LState) int {
	tbl := L.OptTable(1, L.NewTable())
	rgName := tbl.RawGetString("resource_group").String()

	args := []string{"vm", "list"}
	if rgName != "" {
		args = append(args, "--resource-group", rgName)
	}

	// Call the generic exec function
	L.Push(L.NewFunction(mod.azExec))
	argsTable := L.NewTable()
	for _, arg := range args {
		argsTable.Append(lua.LString(arg))
	}
	L.Push(argsTable)
	L.Call(1, 1)

	result := L.CheckTable(L.GetTop())
	exitCode := int(result.RawGetString("exit_code").(lua.LNumber))
	stdout := result.RawGetString("stdout").String()
	stderr := result.RawGetString("stderr").String()

	if exitCode != 0 {
		L.Push(lua.LNil)
		L.Push(lua.LString(stderr))
		return 2
	}

	// Parse the JSON output from az
	L.Push(L.GetGlobal("data").(*lua.LTable).RawGetString("parse_json"))
	L.Push(lua.LString(stdout))
	L.Call(1, 2)

	if L.Get(L.GetTop()).Type() == lua.LTNil {
		L.Push(lua.LNil)
		L.Push(L.Get(L.GetTop()))
		return 2
	}

	L.Push(L.Get(L.GetTop() - 1))
	return 1
}
