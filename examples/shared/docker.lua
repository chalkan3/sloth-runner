-- examples/shared/docker.lua
-- A reusable module for Docker tasks.

local TaskDefinitions = {
    build = {
        name = "build",
        description = "Builds a Docker image",
        params = {
            tag = "latest",
            dockerfile = "Dockerfile",
            context = "."
        },
        command = function(params)
            local image_name = params.image_name or "my-default-image"
            local tag = params.tag
            local dockerfile = params.dockerfile
            local context = params.context
            
            if not image_name then
                return false, "image_name parameter is required"
            end

            local cmd = string.format("docker build -t %s:%s -f %s %s", image_name, tag, dockerfile, context)
            log.info("Executing Docker build: " .. cmd)
            return true, cmd
        end
    },
    push = {
        name = "push",
        description = "Pushes a Docker image to a registry",
        params = {
            tag = "latest"
        },
        command = function(params)
            local image_name = params.image_name
            local tag = params.tag

            if not image_name then
                return false, "image_name parameter is required"
            end

            local cmd = string.format("docker push %s:%s", image_name, tag)
            log.info("Executing Docker push: " .. cmd)
            return true, cmd
        end
    }
}

return TaskDefinitions
