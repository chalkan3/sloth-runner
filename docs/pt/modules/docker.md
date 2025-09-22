# Módulo Docker

O módulo `docker` fornece uma interface conveniente para interagir com o daemon do Docker, permitindo que você construa, execute e envie imagens Docker como parte de suas pipelines.

## Configuração

Este módulo requer que a CLI `docker` esteja instalada e que o daemon do Docker esteja em execução e acessível.

## Funções

### `docker.exec(args)`

Executa qualquer comando `docker` bruto.

- `args` (tabela): **Obrigatório.** Uma lista de argumentos a serem passados para o comando `docker` (ex: `{"ps", "-a"}`).
- **Retorna:** Uma tabela de resultados com `success`, `stdout`, `stderr` e `exit_code`.

### `docker.build(params)`

Constrói uma imagem Docker usando `docker build`.

- `params` (tabela):
    - `tag` (string): **Obrigatório.** A tag para a imagem (ex: `meu-app:latest`).
    - `path` (string): **Obrigatório.** O caminho do contexto de construção.
    - `dockerfile` (string): **Opcional.** O caminho para o Dockerfile.
    - `build_args` (tabela): **Opcional.** Uma tabela de argumentos de construção (ex: `{VERSION = "1.0"}`).
- **Retorna:** Uma tabela de resultados.

### `docker.push(params)`

Envia uma imagem Docker para um registro usando `docker push`.

- `params` (tabela):
    - `tag` (string): **Obrigatório.** A tag da imagem a ser enviada.
- **Retorna:** Uma tabela de resultados.

### `docker.run(params)`

Executa um contêiner Docker usando `docker run`.

- `params` (tabela):
    - `image` (string): **Obrigatório.** A imagem a ser executada.
    - `name` (string): **Opcional.** O nome para o contêiner.
    - `detach` (booleano): **Opcional.** Se `true`, executa o contêiner em segundo plano (`-d`).
    - `ports` (tabela): **Opcional.** Uma lista de mapeamentos de portas (ex: `{"8080:80"}`).
    - `env` (tabela): **Opcional.** Uma tabela de variáveis de ambiente (ex: `{MINHA_VAR = "valor"}`).
- **Retorna:** Uma tabela de resultados.

## Exemplo

```lua
local image_tag = "minha-imagem-teste:latest"

-- Tarefa 1: Build
local result_build = docker.build({
  tag = image_tag,
  path = "./app"
})
if not result_build.success then return false, "Build falhou" end

-- Tarefa 2: Run
local result_run = docker.run({
  image = image_tag,
  name = "meu-container-teste",
  ports = {"8080:80"}
})
if not result_run.success then return false, "Run falhou" end

-- Tarefa 3: Push (após teste bem-sucedido)
local result_push = docker.push({tag = image_tag})
if not result_push.success then return false, "Push falhou" end
```
