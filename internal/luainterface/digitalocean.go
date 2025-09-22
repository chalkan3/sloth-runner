package luainterface

import (
	"bytes"
	"os/exec"

	"github.com/yuin/gopher-lua"
)

// DigitalOceanModule provides DigitalOcean functionalities to Lua scripts
type DigitalOceanModule struct{}

// NewDigitalOceanModule creates a new DigitalOceanModule
func NewDigitalOceanModule() *DigitalOceanModule {
	return &DigitalOceanModule{}
}

// Loader returns the Lua loader for the digitalocean module
func (mod *DigitalOceanModule) Loader(L *lua.LState) int {
	doTable := L.NewTable()

	// Register digitalocean.exec
	L.SetFuncs(doTable, map[string]lua.LGFunction{
		"exec": mod.doExec,
	})

	// Create and register the droplets submodule
	dropletsModule := L.NewTable()
	L.SetFuncs(dropletsModule, map[string]lua.LGFunction{
		"list":   mod.dropletsList,
		"delete": mod.dropletsDelete,
	})
	doTable.RawSetString("droplets", dropletsModule)

	L.Push(doTable)
	return 1
}

// doExec is the generic executor for doctl commands.
// Lua usage: digitalocean.exec({"compute", "droplet", "list"})
func (mod *DigitalOceanModule) doExec(L *lua.LState) int {
	argsTable := L.CheckTable(1)

	var args []string
	argsTable.ForEach(func(_, val lua.LValue) {
		args = append(args, val.String())
	})

	// Ensure JSON output for parsable results
	args = append(args, "--output", "json")

	cmd := exec.Command("doctl", args...)
	var stdout, stderr bytes.Buffer
	cmd.Stdout = &stdout
	cmd.Stderr = &stderr

	err := cmd.Run()

	exitCode := 0
	if err != nil {
		if exitErr, ok := err.(*exec.ExitError); ok {
			exitCode = exitErr.ExitCode()
		} else {
			exitCode = -1 // Indicates an error other than a non-zero exit code
		}
	}

	result := L.NewTable()
	result.RawSetString("stdout", lua.LString(stdout.String()))
	result.RawSetString("stderr", lua.LString(stderr.String()))
	result.RawSetString("exit_code", lua.LNumber(exitCode))

	L.Push(result)
	return 1
}

// dropletsList provides a high-level interface for `doctl compute droplet list`.
// Lua usage: digitalocean.droplets.list()
func (mod *DigitalOceanModule) dropletsList(L *lua.LState) int {
	args := []string{"compute", "droplet", "list"}

	// Call the generic exec function
	L.Push(L.NewFunction(mod.doExec))
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

	// Parse the JSON output from doctl
	L.Push(L.GetGlobal("data").(*lua.LTable).RawGetString("parse_json"))
	L.Push(lua.LString(stdout))
	L.Call(1, 2)

	if L.Get(L.GetTop()).Type() == lua.LTNil { // Check if parse_json returned an error
		L.Push(lua.LNil)
		L.Push(L.Get(L.GetTop())) // Forward the error from parse_json
		return 2
	}

	L.Push(L.Get(L.GetTop() - 1)) // The parsed table
	return 1
}

// dropletsDelete provides a high-level interface for `doctl compute droplet delete`.
// Lua usage: digitalocean.droplets.delete({id="...", force=true})
func (mod *DigitalOceanModule) dropletsDelete(L *lua.LState) int {
	tbl := L.CheckTable(1)
	dropletID := tbl.RawGetString("id").String()
	force := lua.LVAsBool(tbl.RawGetString("force"))

	if dropletID == "" {
		L.Push(lua.LBool(false))
		L.Push(lua.LString("droplet 'id' is required"))
		return 2
	}

	args := []string{"compute", "droplet", "delete", dropletID}
	if force {
		args = append(args, "--force")
	}

	// Call the generic exec function
	L.Push(L.NewFunction(mod.doExec))
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
