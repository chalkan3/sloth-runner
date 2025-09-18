-- examples/pulumi_git_combined_example.lua
--
-- Este arquivo de exemplo demonstra o uso combinado dos módulos 'pulumi' e 'git'.

command = function(params)
    log.info("Iniciando exemplo combinado Pulumi e Git...")

    local pulumi_repo_url = "https://github.com/chalkan3/sloth-runner.git" -- Exemplo de repo Pulumi
    local pulumi_repo_path = "./pulumi-infra-checkout"
    local new_infra_version = params.infra_version or "v1.0.0-infra"
    local pulumi_project_workdir = pulumi_repo_path -- O diretório raiz do repo clonado
    local repo

    -- 1. Clonar ou abrir o repositório Pulumi
    if not fs.exists(pulumi_repo_path) then
        log.info("Cloning Pulumi repository: " .. pulumi_repo_url)
        local cloned_repo, clone_err = git.clone(pulumi_repo_url, pulumi_repo_path)
        if clone_err then
            log.error("Failed to clone Pulumi repository: " .. clone_err)
            return false, "Git clone failed."
        end
        repo = cloned_repo
    else
        log.info("Pulumi repository already exists, opening local reference.")
        local opened_repo, open_err = git.repo(pulumi_repo_path)
        if open_err then
            log.error("Failed to open Pulumi repository: " .. open_err)
            return false, "Git repo open failed."
        end
        repo = opened_repo
    end

    if not repo then
        return false, "Failed to get Pulumi repository reference."
    end

    -- 2. Atualizar o repositório (pull)
    log.info("Pulling latest changes from Pulumi repository...")
    repo:checkout("main"):pull("origin", "main")
    local pull_result = repo:result()
    if not pull_result.success then
        log.error("Failed to pull Pulumi repository: " .. pull_result.stderr)
        return false, "Git pull failed."
    end
    log.info("Pulumi repository updated. Stdout: " .. pull_result.stdout)

    -- 3. Simular uma alteração no código Pulumi (e.g., atualizar um arquivo de versão)
    local infra_version_file = pulumi_repo_path .. "/INFRA_VERSION"
    fs.write(infra_version_file, new_infra_version)
    log.info("Updated INFRA_VERSION file to: " .. new_infra_version)

    -- 4. Commitar e empurrar as mudanças
    local commit_message = "ci: Bump infrastructure version to " .. new_infra_version
    log.info("Committing and pushing infrastructure version change...")
    repo:add(infra_version_file)
        :commit(commit_message)
        :push("origin", "main") -- Sem follow_tags aqui, apenas o commit

    local push_result = repo:result()
    if not push_result.success then
        log.error("Failed to push infrastructure changes: " .. push_result.stderr)
        return false, "Git push failed for infra changes."
    end
    log.info("Infrastructure version change pushed. Stdout: " .. push_result.stdout)

    -- 5. Executar 'pulumi up' para o projeto
    log.info("Running pulumi up for the infrastructure project...")
    local infra_stack = pulumi.stack("my-org/my-infra/dev", {
        workdir = pulumi_project_workdir -- Usar o subdiretório do projeto Pulumi
    })

    local pulumi_up_result = infra_stack:up({ non_interactive = true })

    if not pulumi_up_result.success then
        log.error("Pulumi up failed: " .. pulumi_up_result.stderr)
        return false, "Pulumi up failed."
    end
    log.info("Pulumi up completed successfully. Stdout: " .. pulumi_up_result.stdout)

    log.info("Exemplo combinado Pulumi e Git concluído com sucesso.")
    return true, "Combined Pulumi and Git example finished."
end

TaskDefinitions = {
    pulumi_git_combined_example = {
        description = "Demonstrates combined usage of 'pulumi' and 'git' modules.",
        tasks = {
            {
                name = "run_combined_example",
                command = command,
                params = {
                    infra_version = "v1.0.0-test-combined"
                }
            }
        }
    }
}
