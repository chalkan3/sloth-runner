package luainterface

import (
	"bytes"
	"fmt"
	"os/exec"

	"github.com/yuin/gopher-lua"
)

// DockerModule provides Docker functionalities to Lua scripts
type DockerModule struct{}

// NewDockerModule creates a new DockerModule
func NewDockerModule() *DockerModule {
	return &DockerModule{}
}

// Loader returns the Lua loader for the docker module
func (mod *DockerModule) Loader(L *lua.LState) int {
	dockerTable := L.NewTable()
	L.SetFuncs(dockerTable, map[string]lua.LGFunction{
		"exec":  mod.luaDockerExec,
		"build": mod.dockerBuild,
		"push":  mod.dockerPush,
		"run":   mod.dockerRun,
	})
	L.Push(dockerTable)
	return 1
}

// luaDockerExec is the generic executor for Docker commands exposed to Lua.
// Lua usage: docker.exec({"ps", "-a"})
func (mod *DockerModule) luaDockerExec(L *lua.LState) int {
	argsTable := L.CheckTable(1)
	var args []string
	argsTable.ForEach(func(_, val lua.LValue) {
		args = append(args, val.String())
	})
	return mod.goDockerExec(L, args)
}

// goDockerExec is the internal Go helper to run docker commands.
func (mod *DockerModule) goDockerExec(L *lua.LState, args []string) int {
	cmd := ExecCommand("docker", args...)
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

// docker.build({tag="my-image:latest", path=".", dockerfile="Dockerfile", build_args={...}})
func (mod *DockerModule) dockerBuild(L *lua.LState) int {
	tbl := L.CheckTable(1)
	tag := tbl.RawGetString("tag").String()
	path := tbl.RawGetString("path").String()
	dockerfile := tbl.RawGetString("dockerfile").String()
	buildArgsTbl, _ := tbl.RawGetString("build_args").(*lua.LTable)

	if tag == "" || path == "" {
		L.Push(lua.LNil)
		L.Push(lua.LString("tag and path are required for docker.build"))
		return 2
	}

	args := []string{"build", "-t", tag}
	if dockerfile != "" {
		args = append(args, "-f", dockerfile)
	}
	if buildArgsTbl != nil {
		buildArgsTbl.ForEach(func(key, value lua.LValue) {
			arg := fmt.Sprintf("%s=%s", key.String(), value.String())
			args = append(args, "--build-arg", arg)
		})
	}
	args = append(args, path)

	return mod.goDockerExec(L, args)
}

// docker.push({tag="my-image:latest"})
func (mod *DockerModule) dockerPush(L *lua.LState) int {
	tbl := L.CheckTable(1)
	tag := tbl.RawGetString("tag").String()

	if tag == "" {
		L.Push(lua.LNil)
		L.Push(lua.LString("tag is required for docker.push"))
		return 2
	}
	args := []string{"push", tag}
	return mod.goDockerExec(L, args)
}

// docker.run({image="...", name="...", ports={...}, env={...}, detach=true})
func (mod *DockerModule) dockerRun(L *lua.LState) int {
	tbl := L.CheckTable(1)
	image := tbl.RawGetString("image").String()
	name := tbl.RawGetString("name").String()
	portsTbl, _ := tbl.RawGetString("ports").(*lua.LTable)
	envTbl, _ := tbl.RawGetString("env").(*lua.LTable)
	detach := lua.LVAsBool(tbl.RawGetString("detach"))

	if image == "" {
		L.Push(lua.LNil)
		L.Push(lua.LString("image is required for docker.run"))
		return 2
	}

	args := []string{"run"}
	if detach {
		args = append(args, "-d")
	}
	if name != "" {
		args = append(args, "--name", name)
	}
	if portsTbl != nil {
		portsTbl.ForEach(func(_, value lua.LValue) {
			args = append(args, "-p", value.String())
		})
	}
	if envTbl != nil {
		envTbl.ForEach(func(key, value lua.LValue) {
			envVar := fmt.Sprintf("%s=%s", key.String(), value.String())
			args = append(args, "-e", envVar)
		})
	}
	args = append(args, image)

	return mod.goDockerExec(L, args)
}
