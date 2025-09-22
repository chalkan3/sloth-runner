package luainterface

import (
	"bytes"
	"os/exec"

	"github.com/yuin/gopher-lua"
)

// TerraformModule provides Terraform functionalities to Lua scripts
type TerraformModule struct{}

// NewTerraformModule creates a new TerraformModule
func NewTerraformModule() *TerraformModule {
	return &TerraformModule{}
}

// Loader returns the Lua loader for the terraform module
func (mod *TerraformModule) Loader(L *lua.LState) int {
	tfTable := L.NewTable()
	L.SetFuncs(tfTable, map[string]lua.LGFunction{
		"init":    mod.tfInit,
		"plan":    mod.tfPlan,
		"apply":   mod.tfApply,
		"destroy": mod.tfDestroy,
		"output":  mod.tfOutput,
	})
	L.Push(tfTable)
	return 1
}

// tfExec is the internal generic executor for Terraform commands.
func (mod *TerraformModule) tfExec(L *lua.LState, workdir string, args ...string) int {
	if workdir == "" {
		L.Push(lua.LNil)
		L.Push(lua.LString("workdir is required for all terraform commands"))
		return 2
	}

	cmd := exec.Command("terraform", args...)
	cmd.Dir = workdir
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
	result.RawSetString("success", lua.LBool(exitCode == 0))

	L.Push(result)
	return 1
}

// terraform.init({workdir="..."})
func (mod *TerraformModule) tfInit(L *lua.LState) int {
	tbl := L.CheckTable(1)
	workdir := tbl.RawGetString("workdir").String()
	return mod.tfExec(L, workdir, "init", "-input=false", "-no-color")
}

// terraform.plan({workdir="...", out="plan.out"})
func (mod *TerraformModule) tfPlan(L *lua.LState) int {
	tbl := L.CheckTable(1)
	workdir := tbl.RawGetString("workdir").String()
	outFile := tbl.RawGetString("out").String()

	args := []string{"plan", "-input=false", "-no-color"}
	if outFile != "" {
		args = append(args, "-out="+outFile)
	}

	return mod.tfExec(L, workdir, args...)
}

// terraform.apply({workdir="...", plan="plan.out", auto_approve=true})
func (mod *TerraformModule) tfApply(L *lua.LState) int {
	tbl := L.CheckTable(1)
	workdir := tbl.RawGetString("workdir").String()
	planFile := tbl.RawGetString("plan").String()
	autoApprove := lua.LVAsBool(tbl.RawGetString("auto_approve"))

	args := []string{"apply", "-input=false", "-no-color"}
	if autoApprove {
		args = append(args, "-auto-approve")
	}
	if planFile != "" {
		args = append(args, planFile)
	}

	return mod.tfExec(L, workdir, args...)
}

// terraform.destroy({workdir="...", auto_approve=true})
func (mod *TerraformModule) tfDestroy(L *lua.LState) int {
	tbl := L.CheckTable(1)
	workdir := tbl.RawGetString("workdir").String()
	autoApprove := lua.LVAsBool(tbl.RawGetString("auto_approve"))

	args := []string{"destroy", "-input=false", "-no-color"}
	if autoApprove {
		args = append(args, "-auto-approve")
	}

	return mod.tfExec(L, workdir, args...)
}

// terraform.output({workdir="...", name="output_name"})
func (mod *TerraformModule) tfOutput(L *lua.LState) int {
	tbl := L.CheckTable(1)
	workdir := tbl.RawGetString("workdir").String()
	outputName := tbl.RawGetString("name").String()

	args := []string{"output", "-json"}
	if outputName != "" {
		args = append(args, outputName)
	}

	// We need to call tfExec manually to process the output
	mod.tfExec(L, workdir, args...)
	result := L.CheckTable(L.GetTop())
	L.Pop(1) // remove result table

	if !lua.LVAsBool(result.RawGetString("success")) {
		L.Push(lua.LNil)
		L.Push(result.RawGetString("stderr"))
		return 2
	}

	stdout := result.RawGetString("stdout").String()
	L.Push(L.GetGlobal("data").(*lua.LTable).RawGetString("parse_json"))
	L.Push(lua.LString(stdout))
	L.Call(1, 2) // returns table, err

	if L.Get(L.GetTop()).Type() != lua.LTNil { // Check if parse_json returned an error
		L.Push(lua.LNil)
		L.Push(L.Get(L.GetTop())) // Forward the error from parse_json
		return 2
	}
	L.Pop(1) // remove nil error

	L.Push(L.Get(L.GetTop())) // Push the parsed table
	return 1
}
