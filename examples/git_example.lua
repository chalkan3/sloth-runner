-- examples/git_example.lua
--
-- Este arquivo de exemplo demonstra o uso do módulo 'git' para automação de repositórios.

command = function(params)
    log.info("Iniciando exemplo de automação Git...")

    local repo_url = "https://github.com/chalkan3/sloth-runner.git" -- Usando o próprio sloth-runner para exemplo
    local repo_path = "./sloth-runner-checkout"
    local new_version = params.version or "v1.0.0-test" -- Versão de exemplo
    local repo

    -- Clona o repositório se ele ainda não existir no disco
    if not fs.exists(repo_path) then
        log.info("Cloning repository: " .. repo_url .. " into " .. repo_path)
        local cloned_repo, clone_err = git.clone(repo_url, repo_path)
        if clone_err then
            log.error("Failed to clone repository: " .. clone_err)
            return false, "Git clone failed."
        end
        repo = cloned_repo
    else
        log.info("Repository already exists, opening local reference: " .. repo_path)
        local opened_repo, open_err = git.repo(repo_path) -- Apenas obtém o objeto para o repo local
        if open_err then
            log.error("Failed to open repository: " .. open_err)
            return false, "Git repo open failed."
        end
        repo = opened_repo
    end

    if not repo then
        return false, "Failed to clone or open repository."
    end

    log.info("Starting git operations on " .. repo_path .. "...")

    -- Executa uma sequência de comandos de forma fluente e encadeada
    -- Nota: Cada operação retorna o objeto 'repo' para encadeamento.
    -- Para verificar o sucesso de cada passo, você deve chamar :result() após cada um,
    -- ou no final da cadeia para o último comando.

    log.info("Checking out main branch and pulling latest changes...")
    repo:checkout("master"):pull("origin", "master")
    local pull_result = repo:result() -- Obtém o resultado do último comando (pull)
    if not pull_result.success then
        log.error("Failed to checkout or pull: " .. pull_result.stderr)
        return false, "Git checkout/pull failed."
    end
    log.info("Checkout and pull successful. Stdout: " .. pull_result.stdout)

    -- Simula uma alteração no repositório
    local version_file_path = repo_path .. "/VERSION_EXAMPLE" -- Usar um nome diferente para não conflitar
    fs.write(version_file_path, new_version)
    log.info("Updated VERSION_EXAMPLE file to: " .. new_version)

    -- Adiciona, commita, tagueia e empurra as mudanças de forma encadeada
    local commit_message = "ci: Example bump version to " .. new_version
    log.info("Adding, committing, tagging, and pushing changes...")

    -- Encadeamento: add -> commit -> tag -> push
    repo:add(version_file_path)
        :commit(commit_message)
        :tag(new_version, "Release " .. new_version)
        :push("origin", "master", { follow_tags = true })

    local final_push_result = repo:result() -- Obtém o resultado do último comando (push)

    -- Verifica o resultado da última operação na cadeia
    if not final_push_result.success then
        log.error("Failed to push changes: " .. final_push_result.stderr)
        return false, "Git push failed."
    end

    log.info("Successfully pushed version " .. new_version .. " to origin. Stdout: " .. final_push_result.stdout)
    log.info("Exemplo de automação Git concluído com sucesso.")
    return true, "Git automation example finished."
end

TaskDefinitions = {
    git_automation_example = {
        description = "Demonstrates using the 'git' module for repository automation.",
        tasks = {
            {
                name = "run_git_automation",
                command = command,
                params = {
                    version = "v1.0.0-test" -- Parâmetro de exemplo
                }
            }
        }
    }
}
