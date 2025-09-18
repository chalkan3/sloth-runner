# Módulo Pulumi

O módulo `pulumi` do Sloth-Runner permite que você orquestre suas stacks do Pulumi diretamente de seus scripts Lua. Isso é ideal para fluxos de trabalho de Infraestrutura como Código (IaC) onde você precisa provisionar, atualizar ou destruir recursos de nuvem como parte de uma pipeline de automação maior.

## Casos de Uso Comuns

*   **Provisionamento Dinâmico:** Criar ambientes de staging ou teste sob demanda.
*   **Atualizações de Infraestrutura:** Automatizar o deploy de novas versões da sua infraestrutura.
*   **Gerenciamento de Ambientes:** Destruir ambientes após o uso para economizar custos.
*   **Integração CI/CD:** Executar `pulumi up` ou `preview` como parte de um pipeline de CI/CD.

## Referência da API

### `pulumi.stack(name, options_table)`

Cria uma nova instância de uma stack Pulumi, permitindo que você interaja com ela.

*   `name` (string): O nome completo da stack Pulumi (ex: "my-org/my-project/dev").
*   `options_table` (tabela Lua): Uma tabela de opções para configurar a stack:
    *   `workdir` (string): **Obrigatório.** O caminho para o diretório raiz do projeto Pulumi associado a esta stack.

**Retorna:**
*   `PulumiStack` (userdata): Uma instância do objeto `PulumiStack` para a stack especificada.

### Métodos do Objeto `PulumiStack`

Todos os métodos abaixo são chamados na instância do `PulumiStack` (ex: `my_stack:up(...)`).

#### `stack:up(options)`

Executa o comando `pulumi up` para criar ou atualizar os recursos da stack.

*   `options` (tabela Lua, opcional): Uma tabela de opções para o comando `up`:
    *   `non_interactive` (booleano): Se `true`, adiciona as flags `--non-interactive` e `--yes` ao comando `pulumi up`.
    *   `config` (tabela Lua): Uma tabela de pares chave-valor para passar configurações para a stack (ex: `["my-app:vpcId"] = vpc_id`).
    *   `args` (tabela Lua de strings): Uma lista de argumentos adicionais a serem passados diretamente para o comando `pulumi up`.

**Retorna:**
*   `result` (tabela Lua): Uma tabela contendo:
    *   `success` (booleano): `true` se a operação foi bem-sucedida, `false` caso contrário.
    *   `stdout` (string): A saída padrão do comando Pulumi.
    *   `stderr` (string): A saída de erro padrão do comando Pulumi.
    *   `error` (string ou `nil`): Uma mensagem de erro Go se a execução do comando falhou.

#### `stack:preview(options)`

Executa o comando `pulumi preview` para mostrar uma prévia das alterações que seriam aplicadas.

*   `options` (tabela Lua, opcional): As mesmas opções que para `stack:up()`.

**Retorna:**
*   `result` (tabela Lua): O mesmo formato de retorno que `stack:up()`.

#### `stack:refresh(options)`

Executa o comando `pulumi refresh` para atualizar o estado da stack com os recursos reais na nuvem.

*   `options` (tabela Lua, opcional): As mesmas opções que para `stack:up()`.

**Retorna:**
*   `result` (tabela Lua): O mesmo formato de retorno que `stack:up()`.

#### `stack:destroy(options)`

Executa o comando `pulumi destroy` para destruir todos os recursos da stack.

*   `options` (tabela Lua, opcional): As mesmas opções que para `stack:up()`.

**Retorna:**
*   `result` (tabela Lua): O mesmo formato de retorno que `stack:up()`.

#### `stack:outputs()`

Obtém os outputs da stack Pulumi.

**Retorna:**
*   `outputs` (tabela Lua): Uma tabela Lua onde as chaves são os nomes dos outputs e os valores são os respectivos outputs da stack.
*   `error` (string ou `nil`): Uma mensagem de erro se a operação falhar.

## Exemplos de Uso

### Exemplo Básico de Orquestração Pulumi

Este exemplo demonstra como fazer o deploy de duas stacks Pulumi, passando um output da primeira como input para a segunda.

```lua
-- examples/pulumi_example.lua

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
```

---
[English](../../en/modules/pulumi.md) | [Português](./pulumi.md) | [中文](../../zh/modules/pulumi.md)