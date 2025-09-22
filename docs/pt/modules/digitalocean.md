# Módulo DigitalOcean

O módulo `digitalocean` fornece uma interface para interagir com seus recursos da DigitalOcean usando a ferramenta de linha de comando `doctl`.

## Configuração

Este módulo requer que o CLI `doctl` esteja instalado e autenticado. A maneira padrão de fazer isso é gerar um token de acesso pessoal em seu painel de controle da DigitalOcean e defini-lo como a variável de ambiente `DIGITALOCEAN_ACCESS_TOKEN`.

```bash
export DIGITALOCEAN_ACCESS_TOKEN="seu_token_de_api_da_do_aqui"
```

O módulo usará automaticamente este token para todos os comandos.

## Executor Genérico

### `digitalocean.exec(args)`

Executa qualquer comando `doctl`. Esta função adiciona automaticamente a flag `--output json` para garantir que a saída seja analisável por máquina.

**Parâmetros:**

- `args` (tabela): **Obrigatório.** Uma tabela de strings representando o comando e os argumentos a serem passados para o `doctl` (ex: `{"compute", "droplet", "list"}`).

**Retornos:**

Uma tabela contendo os seguintes campos:
- `stdout` (string): A saída padrão do comando (como uma string JSON).
- `stderr` (string): O erro padrão do comando.
- `exit_code` (número): O código de saída do comando. `0` normalmente indica sucesso.

**Exemplo:**

```lua
local result = digitalocean.exec({"account", "get"})
if result.exit_code == 0 then
  local account_info, err = data.parse_json(result.stdout)
  if account_info then
    log.info("Status da conta: " .. account_info.status)
  end
end
```

## Ajudantes de Droplets

### `digitalocean.droplets.list()`

Um wrapper de alto nível para listar todos os Droplets em sua conta.

**Retornos:**

- `droplets` (tabela) em caso de sucesso, onde a tabela é um array JSON analisado de seus objetos Droplet.
- `nil, error_message` em caso de falha.

**Exemplo:**

```lua
local droplets, err = digitalocean.droplets.list()
if droplets then
  for _, droplet in ipairs(droplets) do
    print("Droplet encontrado: " .. droplet.name)
  end
end
```

### `digitalocean.droplets.delete(params)`

Exclui um Droplet específico pelo seu ID.

**Parâmetros:**

- `params` (tabela): Uma tabela contendo os seguintes campos:
    - `id` (string): **Obrigatório.** O ID do Droplet a ser excluído.
    - `force` (boolean): **Opcional.** Se `true`, adiciona a flag `--force` para ignorar a solicitação de confirmação. O padrão é `false`.

**Retornos:**

- `true` em caso de sucesso.
- `false, error_message` em caso de falha.

**Exemplo:**

```lua
local ok, err = digitalocean.droplets.delete({
  id = "123456789",
  force = true
})
if not ok then
  log.error("Falha ao excluir o droplet: " .. err)
end
```
