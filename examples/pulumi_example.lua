-- examples/pulumi_example.lua
--
-- Este arquivo de exemplo demonstra o uso do módulo 'pulumi' para orquestrar stacks do Pulumi.

command = function()
    log.info("Iniciando exemplo de orquestração Pulumi...")

    -- Exemplo 1: Deploy de uma stack base (e.g., VPC)
    log.info("Deploying the base infrastructure stack (VPC)...")
    local vpc_stack = pulumi.stack("my-org/vpc-network/prod", {
        workdir = "./pulumi/vpc" -- Assumindo que o diretório do projeto Pulumi está aqui
    })

    -- Executa 'pulumi up' de forma não-interativa
    local vpc_result = vpc_stack:up({ non_interactive = true })

    -- Verifica o resultado do deploy da VPC
    if not vpc_result.success then
        log.error("VPC stack deployment failed: " .. vpc_result.stderr)
        return false, "VPC deployment failed."
    end
    log.info("VPC stack deployed successfully. Stdout: " .. vpc_result.stdout)

    -- Obtém os outputs da stack da VPC
    local vpc_outputs, outputs_err = vpc_stack:outputs()
    if outputs_err then
        log.error("Failed to get VPC stack outputs: " .. outputs_err)
        return false, "Failed to get VPC outputs."
    end

    local vpc_id = vpc_outputs.vpcId -- Assumindo que a stack exporta 'vpcId'
    if not vpc_id then
        log.warn("VPC stack did not export 'vpcId'. Continuing without it.")
        vpc_id = "unknown-vpc-id"
    end
    log.info("Obtained VPC ID from outputs: " .. vpc_id)

    -- Exemplo 2: Deploy de uma stack de aplicação, usando outputs da stack anterior como config
    log.info("Deploying the application stack into VPC: " .. vpc_id)
    local app_stack = pulumi.stack("my-org/app-server/prod", {
        workdir = "./pulumi/app" -- Assumindo que o diretório do projeto Pulumi da app está aqui
    })

    -- Executa 'pulumi up' passando outputs da stack anterior como configuração
    local app_result = app_stack:up({
        non_interactive = true,
        config = {
            ["my-app:vpcId"] = vpc_id,
            ["aws:region"] = "us-east-1"
        }
    })

    -- Verifica o resultado do deploy da aplicação
    if not app_result.success then
        log.error("Application stack deployment failed: " .. app_result.stderr)
        return false, "Application deployment failed."
    end
    log.info("Application stack deployed successfully. Stdout: " .. app_result.stdout)

    log.info("Exemplo de orquestração Pulumi concluído com sucesso.")
    return true, "Pulumi orchestration example finished."
end

TaskDefinitions = {
    pulumi_orchestration_example = {
        description = "Demonstrates using the 'pulumi' module to orchestrate infrastructure stacks.",
        tasks = {
            {
                name = "run_pulumi_orchestration",
                command = command
            }
        }
    }
}
