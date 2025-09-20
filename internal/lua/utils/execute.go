package utils

import (
	"context"
	"fmt"

	lua "github.com/yuin/gopher-lua"
)

// ExecuteLuaFunction is a helper to safely call a Lua function with arguments.
func ExecuteLuaFunction(L *lua.LState, fn *lua.LFunction, params map[string]string, secondArg lua.LValue, nRet int, ctx context.Context, args ...lua.LValue) (bool, string, *lua.LTable, error) {
	if ctx != nil {
		L.SetContext(ctx)
	}
	L.Push(fn)
	luaParams := L.NewTable()
	for k, v := range params {
		luaParams.RawSetString(k, lua.LString(v))
	}
	L.Push(luaParams)
	numArgs := 1
	if secondArg != nil {
		L.Push(secondArg)
		numArgs = 2
	}
	// Push additional args
	for _, arg := range args {
		L.Push(arg)
		numArgs++
	}

	if err := L.PCall(numArgs, lua.MultRet, nil); err != nil {
		return false, "", nil, fmt.Errorf("error executing Lua function: %w", err)
	}
	top := L.GetTop()
	var success bool
	var message string
	var outputTable *lua.LTable
	if top >= 1 {
		if L.Get(1).Type() == lua.LTBool {
			success = lua.LVAsBool(L.Get(1))
		} else {
			success = false
			message = fmt.Sprintf("unexpected first return type from Lua: %s", L.Get(1).Type().String())
		}
	}
	if top >= 2 {
		if L.Get(2).Type() == lua.LTString {
			message = lua.LVAsString(L.Get(2))
		}
	}
	if top >= 3 {
		if L.Get(3).Type() == lua.LTTable {
			outputTable = L.Get(3).(*lua.LTable)
		}
	}
	L.SetTop(0)
	return success, message, outputTable, nil
}

// CopyTable performs a deep copy of a table from one Lua state to another.
func CopyTable(src *lua.LTable, destL *lua.LState) *lua.LTable {
	destT := destL.NewTable()
	src.ForEach(func(key, value lua.LValue) {
		destKey := CopyValue(key, destL)
		destValue := CopyValue(value, destL)
		destL.SetTable(destT, destKey, destValue)
	})
	return destT
}

// CopyValue copies a Lua value from one state to another.
func CopyValue(value lua.LValue, destL *lua.LState) lua.LValue {
	switch value.Type() {
	case lua.LTBool:
		return lua.LBool(lua.LVAsBool(value))
	case lua.LTNumber:
		return lua.LNumber(lua.LVAsNumber(value))
	case lua.LTString:
		return lua.LString(lua.LVAsString(value))
	case lua.LTTable:
		return CopyTable(value.(*lua.LTable), destL)
	case lua.LTUserData:
		srcUD := value.(*lua.LUserData)
		destUD := destL.NewUserData()
		destUD.Value = srcUD.Value
		return destUD
	default:
		return lua.LNil
	}
}
