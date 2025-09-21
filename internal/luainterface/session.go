package luainterface

import (
	"github.com/chalkan3/sloth-runner/internal/types"
	lua "github.com/yuin/gopher-lua"
)

// newExportFunction creates a new Lua function that allows scripts to export data.
// It uses the Exporter interface to avoid a direct dependency on the taskrunner package.
func newExportFunction(exporter types.Exporter) lua.LGFunction {
	return func(L *lua.LState) int {
		// Check if the first argument is a table
		tbl := L.CheckTable(1)

		// Convert the Lua table to a Go map
		exportedData := LuaTableToGoMap(L, tbl)

		// Use the interface to export the data
		exporter.Export(exportedData)

		return 0 // No return value
	}
}

// OpenSession registers session-specific functions like 'export' into the Lua state.
// It requires an Exporter to provide the export functionality.
func OpenSession(L *lua.LState, exporter types.Exporter) {
	L.SetGlobal("export", L.NewFunction(newExportFunction(exporter)))
}
