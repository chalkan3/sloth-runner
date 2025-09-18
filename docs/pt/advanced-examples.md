# Exemplos Avançados

Esta seção apresenta exemplos mais complexos e cenários de uso que combinam múltiplos módulos do Sloth-Runner para automação de ponta a ponta.

## Exemplo Completo: Pipeline de CI/CD End-to-End

Este tutorial demonstra como construir um pipeline de CI/CD completo usando os módulos `git`, `pulumi` e `salt` para versionar código, provisionar infraestrutura e implantar uma aplicação.

### Cenário

Imagine que você tem um projeto de infraestrutura Pulumi e um projeto de aplicação. Você quer automatizar o seguinte fluxo:

1.  Clonar o repositório da infraestrutura.
2.  Atualizar um arquivo de versão dentro do repositório.
3.  Committar e empurrar essa alteração para o Git.
4.  Executar `pulumi up` para provisionar ou atualizar a infraestrutura (por exemplo, um ambiente de staging).
5.  Usar o Salt para configurar os servidores provisionados e implantar a aplicação.

### Script Lua (`examples/pulumi_git_combined_example.lua`)

```lua
-- examples/pulumi_git_combined_example.lua

command = function(params)
    log.info("Iniciando exemplo combinado Pulumi e Git...")

    local pulumi_repo_url = "https://github.com/my-org/my-pulumi-infra.git" -- Exemplo de repo Pulumi
    local pulumi_repo_path = "./pulumi-infra-checkout"
    local new_infra_version = params.infra_version or "v1.0.0-infra"
    local pulumi_project_workdir = pulumi_repo_path .. "/my-vpc-project" -- Subdiretório dentro do repo clonado
    local repo

    -- 1. Clonar ou abrir o repositório Pulumi
    log.info("Step 1: Cloning or opening Pulumi repository...")
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
    log.info("Step 2: Pulling latest changes from Pulumi repository...")
    repo:checkout("main"):pull("origin", "main")
    local pull_result = repo:result()
    if not pull_result.success then
        log.error("Failed to pull Pulumi repository: " .. pull_result.stderr)
        return false, "Git pull failed."
    end
    log.info("Pulumi repository updated. Stdout: " .. pull_result.stdout)

    -- 3. Simular uma alteração no código Pulumi (e.g., atualizar um arquivo de versão)
    log.info("Step 3: Simulating a change in Pulumi code (updating version file)...")
    local infra_version_file = pulumi_repo_path .. "/INFRA_VERSION"
    fs.write(infra_version_file, new_infra_version)
    log.info("Updated INFRA_VERSION file to: " .. new_infra_version)

    -- 4. Commitar e empurrar as mudanças
    log.info("Step 4: Committing and pushing infrastructure version change...")
    local commit_message = "ci: Bump infrastructure version to " .. new_infra_version
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
    log.info("Step 5: Running pulumi up for the infrastructure project...")
    local infra_stack = pulumi.stack("my-org/my-infra/dev", {
        workdir = pulumi_project_workdir -- Usar o subdiretório do projeto Pulumi
    })

    local pulumi_up_result = infra_stack:up({ non_interactive = true })

    if not pulumi_up_result.success then
        log.error("Pulumi up failed: " .. pulumi_up_result.stderr)
        return false, "Pulumi up failed."
    end
    log.info("Pulumi up completed successfully. Stdout: " .. pulumi_up_result.stdout)

    -- 6. Configurar e implantar a aplicação usando Salt (Exemplo)
    log.info("Step 6: Configuring and deploying application using Salt...")
    -- Assumindo que o Pulumi up forneceu o IP ou hostname do servidor
    -- Para este exemplo, vamos usar um IP fictício
    local server_ip = "192.168.1.100" -- Substitua pelo output real do Pulumi, se houver
    local salt_target = salt.target(server_ip)

    log.info("Running Salt test.ping on " .. server_ip .. "...")
    salt_target:ping()
    local ping_result = salt_target:result()
    if not ping_result.success then
        log.error("Salt ping failed for " .. server_ip .. ": " .. ping_result.stderr)
        return false, "Salt ping failed."
    end
    log.info("Salt ping successful. Stdout: " .. data.to_json(ping_result.stdout)) -- Assumindo que ping retorna JSON

    log.info("Applying Salt state 'app.install' on " .. server_ip .. "...")
    salt_target:cmd('state.apply', 'app.install')
    local salt_apply_result = salt_target:result()
    if not salt_apply_result.success then
        log.error("Salt state.apply failed for " .. server_ip .. ": " .. salt_apply_result.stderr)
        return false, "Salt state.apply failed."
    end
    log.info("Salt state.apply successful. Stdout: " .. data.to_json(salt_apply_result.stdout))

    log.info("Exemplo combinado Pulumi e Git concluído com sucesso.")
    return true, "Combined Pulumi and Git example finished."
end

TaskDefinitions = {
    pulumi_git_combined_example = {
        description = "Demonstrates combined usage of 'pulumi' and 'git' modules for CI/CD pipeline.",
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
```

---
**Idiomas Disponíveis:**
[English](../en/advanced-examples.md) | [Português](./advanced-examples.md) | [中文](../zh/advanced-examples.md)