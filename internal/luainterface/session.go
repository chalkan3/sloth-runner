package luainterface

import (
	"github.com/chalkan3/sloth-runner/internal/types"
	lua "github.com/yuin/gopher-lua"
)

const SharedSessionTypeName = "session"

func RegisterSharedSessionType(L *lua.LState) {
	mt := L.NewTypeMetatable(SharedSessionTypeName)
	L.SetGlobal(SharedSessionTypeName, mt)
	L.SetField(mt, "__index", L.NewFunction(sharedSession__index))
}

func checkSharedSession(L *lua.LState) *types.SharedSession {
	ud := L.CheckUserData(1)
	if v, ok := ud.Value.(*types.SharedSession); ok {
		return v
	}
	L.ArgError(1, "session expected")
	return nil
}

func sharedSession__index(L *lua.LState) int {
	s := checkSharedSession(L)
	key := L.CheckString(2)

	switch key {
	case "workdir":
		L.Push(lua.LString(s.Workdir))
	default:
		L.Push(lua.LNil)
	}

	return 1
}
