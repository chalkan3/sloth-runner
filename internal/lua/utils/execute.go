// Package utils provides shared helper functions for interacting with the gopher-lua
// state. These utilities are kept in a separate, low-level package to avoid
// import cycles between higher-level packages like 'luainterface' and 'runner'.
package utils

import (
	"context"
	"fmt"

	lua "github.com/yuin/gopher-lua"
)

// ExecuteLuaFunction is a helper to safely call a Lua function with a standardized
// set of arguments and return values. It handles the boilerplate of pushing the
// function and arguments to the stack and parsing the multiple return values.
//
// Parameters:
//   L: The Lua state.
//   fn: The Lua function to call.
//   params: A map of string key-value pairs, passed as the first table argument to the function.
//   secondArg: An arbitrary Lua value passed as the second argument.
//   nRet: The expected number of return values.
//   ctx: A Go context to associate with the Lua state for the duration of the call.
//   args: Additional variadic Lua values to pass to the function.
//
// Returns:
//   - bool: The first return value, expected to be a boolean indicating success.
//   - string: The second return value, expected to be a string message.
//   - *lua.LTable: The third return value, expected to be a table containing output data.
//   - error: Any error that occurred during the pcall execution.
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

// CopyTable performs a deep copy of a Lua table from one Lua state to another.
// This is essential when passing data between isolated Lua environments, such as
// in the testing framework.
func CopyTable(src *lua.LTable, destL *lua.LState) *lua.LTable {
	destT := destL.NewTable()
	src.ForEach(func(key, value lua.LValue) {
		destKey := CopyValue(key, destL)
		destValue := CopyValue(value, destL)
		destL.SetTable(destT, destKey, destValue)
	})
	return destT
}

// CopyValue recursively copies a Lua value from one state to another. It handles
// basic types and tables. Functions and userdata are not copied and will result in nil.
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