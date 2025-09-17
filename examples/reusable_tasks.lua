-- examples/reusable_tasks.lua

-- Import the reusable Docker tasks. The path is relative to this file.
local docker_tasks = import("shared/docker.lua")

TaskDefinitions = {
    app_deployment = {
        description = "A workflow that builds and deploys an application using a reusable Docker module.",
        tasks = {
            -- Use the 'build' task from the imported module by referencing it.
            -- Override its default parameters directly.
            build = {
                uses = docker_tasks.build,
                description = "Build the main application Docker image",
                params = {
                    image_name = "my-app",
                    tag = "v1.0.0",
                    context = "./app"
                }
            },
            
            -- Define a new task that depends on the 'build' task defined above.
            deploy = {
                name = "deploy",
                description = "Deploys the application",
                depends_on = "build",
                command = function(params)
                    -- In a real scenario, this would use kubectl, helm, etc.
                    local image_name = "my-app"
                    local tag = "v1.0.0"
                    log.info(string.format("Deploying image %s:%s", image_name, tag))
                    return true, string.format("echo 'Deployed %s:%s'", image_name, tag)
                end
            }
        }
    }
}