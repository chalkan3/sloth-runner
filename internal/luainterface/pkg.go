package luainterface

import (
	"fmt"
	"os/exec"
	"strings"

	lua "github.com/yuin/gopher-lua"
)

// PkgModule provides functions for package management.
type PkgModule struct{}

// NewPkgModule creates a new PkgModule.
func NewPkgModule() *PkgModule {
	return &PkgModule{}
}

// Loader is the module loader function.
func (p *PkgModule) Loader(L *lua.LState) int {
	mod := L.SetFuncs(L.NewTable(), p.exports())
	L.Push(mod)
	return 1
}

func (p *PkgModule) exports() map[string]lua.LGFunction {
	return map[string]lua.LGFunction{
		"install": p.install,
	}
}

func (p *PkgModule) install(L *lua.LState) int {
	packages := L.ToString(1)

	// Detect package manager
	var cmd *exec.Cmd
	if _, err := exec.LookPath("apt-get"); err == nil {
		cmd = exec.Command("sudo", "apt-get", "install", "-y", packages)
	} else if _, err := exec.LookPath("yum"); err == nil {
		cmd = exec.Command("sudo", "yum", "install", "-y", packages)
	} else if _, err := exec.LookPath("brew"); err == nil {
		cmd = exec.Command("brew", "install", packages)
	} else {
		L.Push(lua.LFalse)
		L.Push(lua.LString("No supported package manager found (apt, yum, brew)"))
		return 2
	}

	output, err := cmd.CombinedOutput()
	if err != nil {
		L.Push(lua.LFalse)
		L.Push(lua.LString(fmt.Sprintf("Failed to install packages: %s\n%s", err, string(output))))
		return 2
	}

	L.Push(lua.LTrue)
	L.Push(lua.LString(string(output)))
	return 2
}
