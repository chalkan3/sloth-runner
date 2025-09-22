package repl

import (
	"fmt"
	"io/ioutil"
	"os"
	"strings"

	"github.com/c-bata/go-prompt"
	"github.com/chalkan3/sloth-runner/internal/luainterface"
	lua "github.com/yuin/gopher-lua"
)

var L *lua.LState

// executor is the function that gets called every time the user presses Enter.
func executor(in string) {
	if in == "exit" || in == "quit" {
		fmt.Println("Bye!")
		L.Close()
		os.Exit(0)
	}

	err := L.DoString(in)
	if err != nil {
		fmt.Println("Error:", err)
		return
	}

	// Print the returned values
	n := L.GetTop()
	if n > 0 {
		var results []string
		for i := 1; i <= n; i++ {
			results = append(results, L.Get(i).String())
		}
		fmt.Println(strings.Join(results, "\t"))
		L.Pop(n)
	}
}

// completer provides autocompletion suggestions.
func completer(d prompt.Document) []prompt.Suggest {
	var suggestions []prompt.Suggest

	text := d.TextBeforeCursor()
	parts := strings.Split(text, ".")

	if len(parts) <= 1 { // Top-level globals
		L.G.Global.ForEach(func(key, value lua.LValue) {
			if strings.HasPrefix(key.String(), text) {
				description := ""
				if value.Type() == lua.LTTable {
					description = "module"
				} else if value.Type() == lua.LTFunction {
					description = "function"
				}
			suggestions = append(suggestions, prompt.Suggest{Text: key.String(), Description: description})
			}
		})
	} else { // Methods of a table
		tableName := parts[0]
		methodPrefix := parts[len(parts)-1]
		table := L.GetGlobal(tableName)

		if tbl, ok := table.(*lua.LTable); ok {
			tbl.ForEach(func(key, value lua.LValue) {
				if strings.HasPrefix(key.String(), methodPrefix) {
					description := ""
					if value.Type() == lua.LTFunction {
						description = "function"
					}
				suggestions = append(suggestions, prompt.Suggest{Text: key.String(), Description: description})
				}
			})
		}
	}

	return prompt.FilterHasPrefix(suggestions, d.GetWordBeforeCursor(), true)
}

// Start begins the REPL session.
func Start(filePath string) {
	L = lua.NewState()
	// Do not close L here, it's closed in the executor on "exit"

	// Load all sloth-runner modules
	luainterface.OpenAll(L)

	fmt.Println("Sloth-Runner Interactive REPL")
	fmt.Println("Type 'exit' or 'quit' to leave.")

	// If a file is provided, execute it first to load its context.
	if filePath != "" {
		content, err := ioutil.ReadFile(filePath)
		if err != nil {
			fmt.Printf("Error reading file %s: %v\n", filePath, err)
		} else {
			if err := L.DoString(string(content)); err != nil {
				fmt.Printf("Error executing file %s: %v\n", filePath, err)
			} else {
				fmt.Printf("Successfully loaded context from %s\n", filePath)
			}
		}
	}

	p := prompt.New(
		executor,
		completer,
		prompt.OptionPrefix("sloth> "),
		prompt.OptionTitle("sloth-runner-repl"),
	)
	p.Run()
}
