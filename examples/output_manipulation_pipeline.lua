TaskDefinitions = {
    output_manipulation_pipeline = {
        description = "Demonstrates manipulating command output for subsequent tasks",
        tasks = {
            {
                name = "get_file_list",
                description = "Gets a list of files using 'find . -name \"*.go\"'",
                command = function(params)
                    print("Lua: Executing 'find . -name \"*.go\"' to get file list...")
                    local stdout, stderr, err, exit_code = exec.command("/usr/bin/find", ".", "-name", "*.go")

                    if err then
                        print("Command failed: " .. err)
                        print("Stderr: " .. stderr)
                        return false, "find . -name \"*.go\" command failed: " .. err, { stdout = stdout, stderr = stderr, exit_code = exit_code }
                    end

                    -- Manipulate stdout: split into a table of lines
                    local files = {}
                    local current_pos = 1
                    while true do
                        local newline_pos = string.find(stdout, "\n", current_pos)
                        local line
                        if newline_pos then
                            line = string.sub(stdout, current_pos, newline_pos - 1)
                        else
                            line = string.sub(stdout, current_pos)
                        end

                        if line ~= "" then
                            table.insert(files, line)
                        end

                        if not newline_pos then
                            break
                        end
                        current_pos = newline_pos + 1
                    end

                    print("Found " .. #files .. " files.")
                    return true, "File list retrieved", { file_list = files }
                end,
                post_exec = function(params, output)
                    print("Lua Hook: get_file_list completed. Number of files: " .. #(output.file_list or {}))
                    return true, "get_file_list post_exec successful"
                end,
            },
            {
                name = "count_go_files",
                description = "Counts .go files from the previous task's output",
                depends_on = "get_file_list",
                command = function(params, input_from_dependency)
                    print("Lua: Counting .go files...")
                    local file_list = input_from_dependency.get_file_list.file_list
                    local go_file_count = 0

                    if file_list then
                        for _, filename in ipairs(file_list) do
                            if string.match(filename, "%.go$") then
                                go_file_count = go_file_count + 1
                            end
                        end
                    end

                    print("Found " .. go_file_count .. " .go files.")
                    return true, "Go files counted", { go_files_found = go_file_count }
                end,
                pre_exec = function(params, input_from_dependency)
                    local file_list = input_from_dependency.get_file_list.file_list
                    print("Lua Hook: count_go_files preparing. Received " .. #(file_list or {}) .. " files.")
                    return true, "count_go_files pre_exec successful"
                end,
            },
        }
    }
}
