package main

import (
	"bytes"
	"encoding/json"
	"fmt"
	"io/ioutil"
	"os"
	"strconv"
	"strings"
	"text/template"

	"github.com/AlecAivazis/survey/v2"
	lua "github.com/yuin/gopher-lua"

	"github.com/spf13/cobra"

	"github.com/chalkan3/sloth-runner/internal/luainterface"
	"github.com/chalkan3/sloth-runner/internal/taskrunner"
	"github.com/chalkan3/sloth-runner/internal/types"
	"gopkg.in/yaml.v2" // Added for YAML parsing
)

// TemplateData struct to hold data passed to the Go template
type TemplateData struct {
	Env          string
	IsProduction bool
	Shards       []int
}

var (
	configFilePath string
	env            string
	isProduction   bool
	shardsStr      string
	targetTasksStr string
	targetGroup    string
	valuesFilePath string // New: Path to a values.yaml file
	returnOutput   bool
)

// loadAndRenderLuaConfig reads, renders, and loads Lua task definitions.
func loadAndRenderLuaConfig(configFilePath, env, shardsStr string, isProduction bool, valuesFilePath string) (map[string]types.TaskGroup, error) {
	// Read the Lua template file
	templateContent, err := ioutil.ReadFile(configFilePath)
	if err != nil {
		return nil, fmt.Errorf("error reading Lua template file %s: %w", configFilePath, err)
	}

	// Parse the template
	tmpl, err := template.New("lua_config").Parse(string(templateContent))
	if err != nil {
		return nil, fmt.Errorf("error parsing Lua template: %w", err)
	}

	// Parse shards string into []int
	var shards []int
	if shardsStr != "" {
		shardStrings := strings.Split(shardsStr, ",")
		for _, s := range shardStrings {
			shard, err := strconv.Atoi(strings.TrimSpace(s))
			if err != nil {
				return nil, fmt.Errorf("invalid shard number '%s': %w", s, err)
			}
			shards = append(shards, shard)
		}
	}

	// Prepare data for the template
	data := TemplateData{
		Env:          env,
		IsProduction: isProduction,
		Shards:       shards,
	}

	// Execute the template into a buffer
	var renderedLua bytes.Buffer
	if err := tmpl.Execute(&renderedLua, data); err != nil {
		return nil, fmt.Errorf("error executing Lua template: %w", err)
	}

	// Create a new Lua state
	L := lua.NewState()
	defer L.Close()

	// Open the 'exec' library for shell command execution
	luainterface.OpenExec(L)

	// Open the 'fs' library for file system operations
	luainterface.OpenFs(L)

	// Open the 'net' library for network operations
	luainterface.OpenNet(L)

	// Open the 'data' library for data serialization/deserialization
	luainterface.OpenData(L)

	// Open the 'log' library for logging from Lua
	luainterface.OpenLog(L)

	// Open the 'salt' library for SaltStack operations
	luainterface.OpenSalt(L)

	// --- New: Load and expose values.yaml to Lua ---
	if valuesFilePath != "" {
		valuesContent, err := ioutil.ReadFile(valuesFilePath)
		if err != nil {
			return nil, fmt.Errorf("error reading values file %s: %w", valuesFilePath, err)
		}

		var goValues map[string]interface{} // Explicitly unmarshal into map[string]interface{}
		if err := yaml.Unmarshal(valuesContent, &goValues); err != nil {
			return nil, fmt.Errorf("error parsing values YAML from %s: %w", valuesFilePath, err)
		}

		luaValues := luainterface.GoValueToLua(L, goValues)
		L.SetGlobal("values", luaValues)
	}
	// --- End New ---

	// Load the rendered Lua script content and parse task definitions
	taskGroups, err := luainterface.LoadTaskDefinitions(L, renderedLua.String())
	if err != nil {
		return nil, fmt.Errorf("error loading task definitions: %w", err)
	}

	return taskGroups, nil
}

var rootCmd = &cobra.Command{
	Use:   "sloth-runner",
	Short: "A flexible sloth-runner with Lua scripting capabilities",
	Long: "sloth-runner is a command-line tool that allows you to define and execute " +
		"tasks using Lua scripts. It supports pipelines, workflows, dynamic task generation, " +
		"and output manipulation.",
	Run: func(cmd *cobra.Command, args []string) {
		cmd.Help()
	},
}

var runCmd = &cobra.Command{
	Use:   "run",
	Short: "Executes tasks defined in a Lua template file",
	Long: `The run command executes tasks defined in a Lua template file.

You can specify the file, environment variables, and target specific tasks or groups.`,
	RunE: func(cmd *cobra.Command, args []string) error {
		taskGroups, err := loadAndRenderLuaConfig(configFilePath, env, shardsStr, isProduction, valuesFilePath)
		if err != nil {
			return err
		}

		var targetTasks []string
		if targetTasksStr != "" {
			targetTasks = strings.Split(targetTasksStr, ",")
			for i, task := range targetTasks {
				targetTasks[i] = strings.TrimSpace(task)
			}
		} else {
			// Interactive task selection
			var allTasks []string
			if targetGroup != "" {
				if group, ok := taskGroups[targetGroup]; ok {
					for _, task := range group.Tasks {
						allTasks = append(allTasks, task.Name)
					}
				} else {
					return fmt.Errorf("task group '%s' not found", targetGroup)
				}
			} else {
				for _, group := range taskGroups {
					for _, task := range group.Tasks {
						allTasks = append(allTasks, task.Name)
					}
				}
			}

			if len(allTasks) == 0 {
				fmt.Println("No tasks found to run.")
				return nil
			}

			prompt := &survey.MultiSelect{
				Message: "Select tasks to run:",
				Options: allTasks,
			}
			survey.AskOne(prompt, &targetTasks)
		}

		if len(targetTasks) == 0 {
			fmt.Println("No tasks selected.")
			return nil
		}

		// Create a new Lua state for the TaskRunner
		L := lua.NewState()
		defer L.Close()

		luainterface.OpenExec(L)
		luainterface.OpenFs(L)
		luainterface.OpenNet(L)
		luainterface.OpenData(L)
		luainterface.OpenLog(L)
		luainterface.OpenSalt(L)

		tr := taskrunner.NewTaskRunner(L, taskGroups, targetGroup, targetTasks)

		if err := tr.Run(); err != nil {
			return fmt.Errorf("error running tasks: %w", err)
		}

		if returnOutput {
			finalOutputs := make(map[string]interface{})
			for _, taskName := range targetTasks {
				if output, ok := tr.Outputs[taskName]; ok {
					finalOutputs[taskName] = output
				}
			}

			var outputToMarshal interface{}
			if len(targetTasks) == 1 {
				if val, ok := finalOutputs[targetTasks[0]]; ok {
					outputToMarshal = val
				} else {
					outputToMarshal = make(map[string]interface{})
				}
			} else {
				outputToMarshal = finalOutputs
			}

			jsonOutput, err := json.Marshal(outputToMarshal)
			if err != nil {
				return fmt.Errorf("error marshaling final task output to JSON: %w", err)
			}
			fmt.Println(string(jsonOutput))
		}

		return nil
	},
}
var listCmd = &cobra.Command{
	Use:   "list",
	Short: "Lists all available task groups and tasks",
	Long:  `The list command displays all task groups and their respective tasks, along with their descriptions and dependencies.`,
	RunE: func(cmd *cobra.Command, args []string) error {
		taskGroups, err := loadAndRenderLuaConfig(configFilePath, env, shardsStr, isProduction, valuesFilePath)
		if err != nil {
			return err
		}

		if len(taskGroups) == 0 {
			fmt.Println("No task groups found.")
			return nil
		}

		fmt.Println("Available Task Groups and Tasks:")
		for groupName, group := range taskGroups {
			fmt.Printf("\n  Group: %s (Description: %s)\n", groupName, group.Description)
			if len(group.Tasks) == 0 {
				fmt.Println("    No tasks defined in this group.")
				continue
			}
			for _, task := range group.Tasks {
				fmt.Printf("    - Task: %s\n", task.Name)
				fmt.Printf("      Description: %s\n", task.Description)
				if len(task.DependsOn) > 0 {
					fmt.Printf("      Depends On: %s\n", strings.Join(task.DependsOn, ", "))
				}
				fmt.Printf("      Async: %t\n", task.Async)
			}
		}
		return nil
	},
}

var validateCmd = &cobra.Command{
	Use:   "validate",
	Short: "Validates the syntax and structure of a Lua task file",
	Long:  `The validate command checks a Lua task file for syntax errors and ensures that the TaskDefinitions table is correctly structured.`,
	RunE: func(cmd *cobra.Command, args []string) error {
		_, err := loadAndRenderLuaConfig(configFilePath, env, shardsStr, isProduction, valuesFilePath)
		if err != nil {
			return err
		}

		fmt.Println("✅ Configuration file is valid.")
		return nil
	},
}

func init() {
	rootCmd.AddCommand(runCmd)
	rootCmd.AddCommand(listCmd)
	rootCmd.AddCommand(validateCmd)

	// Define flags for the run command
	runCmd.Flags().StringVarP(&configFilePath, "file", "f", "examples/basic_pipeline.lua", "Path to the Lua task configuration template file")
	runCmd.Flags().StringVarP(&env, "env", "e", "Development", "Environment for the tasks (e.g., Development, Production)")
	runCmd.Flags().BoolVarP(&isProduction, "prod", "p", false, "Set to true for production environment")
	runCmd.Flags().StringVar(&shardsStr, "shards", "1,2,3", "Comma-separated list of shard numbers (e.g., 1,2,3)")
	runCmd.Flags().StringVarP(&targetTasksStr, "tasks", "t", "", "Comma-separated list of specific tasks to run (e.g., task1,task2)")
	runCmd.Flags().StringVarP(&targetGroup, "group", "g", "", "Run tasks only from a specific task group")
	runCmd.Flags().StringVarP(&valuesFilePath, "values", "v", "", "Path to a YAML file with values to be passed to Lua tasks") // New flag for runCmd
	runCmd.Flags().BoolVar(&returnOutput, "return", false, "Return the output of the target tasks as JSON")

	// Flags for list command (can reuse configFilePath, env, isProduction, shardsStr)
	listCmd.Flags().StringVarP(&configFilePath, "file", "f", "examples/basic_pipeline.lua", "Path to the Lua task configuration template file")
	listCmd.Flags().StringVarP(&env, "env", "e", "Development", "Environment for the tasks (e.g., Development, Production)")
	listCmd.Flags().BoolVarP(&isProduction, "prod", "p", false, "Set to true for production environment")
	listCmd.Flags().StringVar(&shardsStr, "shards", "1,2,3", "Comma-separated list of shard numbers (e.g., 1,2,3)")
	listCmd.Flags().StringVarP(&valuesFilePath, "values", "v", "", "Path to a YAML file with values to be passed to Lua tasks") // New flag for listCmd

	// Flags for validate command
	validateCmd.Flags().StringVarP(&configFilePath, "file", "f", "examples/basic_pipeline.lua", "Path to the Lua task configuration template file")
	validateCmd.Flags().StringVarP(&env, "env", "e", "Development", "Environment for the tasks (e.g., Development, Production)")
	validateCmd.Flags().BoolVarP(&isProduction, "prod", "p", false, "Set to true for production environment")
	validateCmd.Flags().StringVar(&shardsStr, "shards", "1,2,3", "Comma-separated list of shard numbers (e.g., 1,2,3)")
	validateCmd.Flags().StringVarP(&valuesFilePath, "values", "v", "", "Path to a YAML file with values to be passed to Lua tasks")
}

func main() {
	if err := rootCmd.Execute(); err != nil {
		fmt.Fprintf(os.Stderr, "Error: %v\n", err)
		os.Exit(1)
	}
}