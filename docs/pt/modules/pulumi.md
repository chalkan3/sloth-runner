# Módulo Pulumi

O módulo `pulumi` fornece uma API fluente para orquestrar stacks do Pulumi, permitindo que você gerencie seus fluxos de trabalho de Infraestrutura como Código (IaC) diretamente do `sloth-runner`.

---

## `pulumi.stack(name, options)`

Cria um objeto de stack do Pulumi.

*   **Parâmetros:**
    *   `name` (string): O nome completo da stack (ex: `"minha-org/meu-projeto/dev"`).
    *   `options` (tabela): Uma tabela de opções.
        *   `workdir` (string): **(Obrigatório)** O caminho para o diretório do projeto Pulumi.
*   **Retorna:**
    *   `stack` (objeto): Um objeto `PulumiStack`.
    *   `error`: Um objeto de erro se a stack não puder ser inicializada.

---

## O Objeto `PulumiStack`

Este objeto representa uma stack específica do Pulumi e fornece métodos para interação.

### `stack:up([options])`

Cria ou atualiza os recursos da stack executando `pulumi up`.

*   **Parâmetros:**
    *   `options` (tabela, opcional):
        *   `yes` (booleano): Se `true`, passa `--yes` para aprovar a atualização automaticamente.
        *   `config` (tabela): Um dicionário de valores de configuração a serem passados para a stack.
        *   `args` (tabela): Uma lista de argumentos de string adicionais a serem passados para o comando.
*   **Retorna:**
    *   `result` (tabela): Uma tabela contendo `success` (booleano), `stdout` (string) e `stderr` (string).

### `stack:preview([options])`

Pré-visualiza as alterações que seriam feitas por uma atualização executando `pulumi preview`.

*   **Parâmetros:** Os mesmos de `stack:up`.
*   **Retorna:** O mesmo de `stack:up`.

### `stack:refresh([options])`

Atualiza o estado da stack executando `pulumi refresh`.

*   **Parâmetros:** Os mesmos de `stack:up`.
*   **Retorna:** O mesmo de `stack:up`.

### `stack:destroy([options])`

Destrói todos os recursos na stack executando `pulumi destroy`.

*   **Parâmetros:** Os mesmos de `stack:up`.
*   **Retorna:** O mesmo de `stack:up`.

### `stack:outputs()`

Recupera os outputs de uma stack implantada.

*   **Retorna:**
    *   `outputs` (tabela): Uma tabela Lua com os outputs da stack.
    *   `error`: Um objeto de erro se a busca dos outputs falhar.

### Exemplo

Este exemplo mostra um padrão comum: implantar uma stack de rede (VPC) e, em seguida, usar seu output (`vpcId`) para configurar e implantar uma stack de aplicação.

```lua
command = function()
  local pulumi = require("pulumi")

  -- 1. Define a stack da VPC
  local vpc_stack = pulumi.stack("minha-org/vpc/prod", { workdir = "./pulumi/vpc" })
  
  -- 2. Implanta a VPC
  log.info("Implantando a stack da VPC...")
  local vpc_result = vpc_stack:up({ yes = true })
  if not vpc_result.success then
    return false, "A implantação da VPC falhou: " .. vpc_result.stderr
  end

  -- 3. Obtém o ID da VPC de seus outputs
  log.info("Buscando outputs da VPC...")
  local vpc_outputs, err = vpc_stack:outputs()
  if err then
    return false, "Falha ao obter os outputs da VPC: " .. err
  end
  local vpc_id = vpc_outputs.vpcId

  -- 4. Define a stack da Aplicação
  local app_stack = pulumi.stack("minha-org/app/prod", { workdir = "./pulumi/app" })

  -- 5. Implanta a Aplicação, passando o vpcId como configuração
  log.info("Implantando a stack da Aplicação na VPC: " .. vpc_id)
  local app_result = app_stack:up({
    yes = true,
    config = { ["my-app:vpcId"] = vpc_id }
  })
  if not app_result.success then
    return false, "A implantação da Aplicação falhou: " .. app_result.stderr
  end

  log.info("Todas as stacks foram implantadas com sucesso.")
  return true, "Orquestração com Pulumi completa."
end
```
