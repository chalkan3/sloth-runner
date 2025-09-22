# Módulo Azure

O módulo `azure` fornece uma interface para interagir com o Microsoft Azure usando a ferramenta de linha de comando `az`.

## Configuração

Este módulo requer que o CLI `az` esteja instalado e autenticado. Antes de executar pipelines que usam este módulo, você deve fazer login em sua conta do Azure:

```bash
az login
```

O módulo usará suas credenciais de login para todos os comandos.

## Executor Genérico

### `azure.exec(args)`

Executa qualquer comando `az`. Esta função adiciona automaticamente a flag `--output json` (se ainda não estiver presente) para garantir que a saída seja analisável por máquina.

**Parâmetros:**

- `args` (tabela): **Obrigatório.** Uma tabela de strings representando o comando e os argumentos a serem passados para o `az` (ex: `{"group", "list", "--location", "eastus"}`).

**Retornos:**

Uma tabela contendo os seguintes campos:
- `stdout` (string): A saída padrão do comando (como uma string JSON).
- `stderr` (string): O erro padrão do comando.
- `exit_code` (número): O código de saída do comando. `0` normalmente indica sucesso.

**Exemplo:**

```lua
local result = azure.exec({"account", "show"})
if result.exit_code == 0 then
  local account_info, err = data.parse_json(result.stdout)
  if account_info then
    log.info("Logado como: " .. account_info.user.name)
  end
end
```

## Ajudantes de Grupo de Recursos (RG)

### `azure.rg.delete(params)`

Exclui um grupo de recursos.

**Parâmetros:**

- `params` (tabela): Uma tabela contendo os seguintes campos:
    - `name` (string): **Obrigatório.** O nome do grupo de recursos a ser excluído.
    - `yes` (boolean): **Opcional.** Se `true`, adiciona a flag `--yes` para ignorar a solicitação de confirmação.

**Retornos:**

- `true` em caso de sucesso.
- `false, error_message` em caso de falha.

**Exemplo:**

```lua
local ok, err = azure.rg.delete({
  name = "meu-rg-de-teste",
  yes = true
})
if not ok then
  log.error("Falha ao excluir o grupo de recursos: " .. err)
end
```

## Ajudantes de Máquina Virtual (VM)

### `azure.vm.list(params)`

Lista máquinas virtuais.

**Parâmetros:**

- `params` (tabela): **Opcional.** Uma tabela contendo os seguintes campos:
    - `resource_group` (string): O nome de um grupo de recursos para limitar a lista. Se omitido, lista as VMs em toda a assinatura.

**Retornos:**

- `vms` (tabela) em caso de sucesso, onde a tabela é um array JSON analisado de seus objetos VM.
- `nil, error_message` em caso de falha.

**Exemplo:**

```lua
-- Lista todas as VMs na assinatura
local all_vms, err1 = azure.vm.list()

-- Lista VMs em um grupo de recursos específico
local specific_vms, err2 = azure.vm.list({resource_group = "meu-rg-de-producao"})
if specific_vms then
  for _, vm in ipairs(specific_vms) do
    print("VM encontrada: " .. vm.name)
  end
end
```
