package luainterface

import (
	lua "github.com/yuin/gopher-lua"
)

const (
	luaSaltTargetTypeName = "saltTarget"
)

// Estrutura interna para o alvo Salt
type saltTarget struct {
	Target     string
	TargetType string
}

// salt:target(tgt, tgt_type) -> target
func saltTargetFn(L *lua.LState) int {
	tgt := L.CheckString(1)
	tgtType := L.CheckString(2)
	target := &saltTarget{Target: tgt, TargetType: tgtType}

	ud := L.NewUserData()
	ud.Value = target
	L.SetMetatable(ud, L.GetTypeMetatable(luaSaltTargetTypeName))
	L.Push(ud)
	return 1
}

// --- Métodos do Objeto Target ---

func checkSaltTarget(L *lua.LState, n int) *saltTarget {
	ud := L.CheckUserData(n)
	if v, ok := ud.Value.(*saltTarget); ok {
		return v
	}
	L.ArgError(n, "esperado objeto target do salt")
	return nil
}

// target:cmd(command, ...args)
func saltTargetCmd(L *lua.LState) int {
	target := checkSaltTarget(L, 1)
	cmdStr := L.CheckString(2)
	var args []string
	top := L.GetTop()
	for i := 3; i <= top; i++ {
		args = append(args, L.ToString(i))
	}

	fullArgs := []string{target.TargetType, target.Target, cmdStr}
	fullArgs = append(fullArgs, args...)
	cmd := ExecCommand("salt", fullArgs...)
	cmd.Run() // We don't check the error in the mock context

	L.Push(L.Get(1))
	return 1
}

// target:ping()
func saltTargetPing(L *lua.LState) int {
	target := checkSaltTarget(L, 1)
	cmd := ExecCommand("salt", target.TargetType, target.Target, "test.ping")
	cmd.Run() // We don't check the error in the mock context
	L.Push(L.Get(1))
	return 1
}

// target:result() -> stdout, stderr, err
func saltTargetResult(L *lua.LState) int {
	// Mock implementation for now
	L.Push(lua.LString("mock stdout"))
	L.Push(lua.LString(""))
	L.Push(lua.LNil)
	return 3
}

var saltMethods = map[string]lua.LGFunction{
	"target": saltTargetFn,
}

var saltTargetMethods = map[string]lua.LGFunction{
	"cmd":    saltTargetCmd,
	"ping":   saltTargetPing,
	"result": saltTargetResult,
}

func SaltLoader(L *lua.LState) int {
	// Registra o tipo 'saltTarget' com seus métodos
	mt := L.NewTypeMetatable(luaSaltTargetTypeName)
	L.SetField(mt, "__index", L.SetFuncs(L.NewTable(), saltTargetMethods))

	// Registra o módulo 'salt' com seu método construtor
	mod := L.SetFuncs(L.NewTable(), saltMethods)
	L.Push(mod)
	return 1
}

func OpenSalt(L *lua.LState) {
	L.PreloadModule("salt", SaltLoader)
}
