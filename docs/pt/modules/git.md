# Módulo Git

O módulo `git` do Sloth-Runner fornece uma API fluente e de alto nível para interagir com repositórios Git diretamente de seus scripts Lua. Isso permite automatizar operações comuns do Git, como clonagem, pull, adição, commit, tag e push, facilitando fluxos de trabalho de CI/CD e automação de versionamento.

## Casos de Uso Comuns

*   **Automação de CI/CD:** Clonar repositórios, atualizar código, commitar alterações geradas por scripts e empurrar para o controle de versão.
*   **Gerenciamento de Configuração:** Puxar as últimas configurações de um repositório Git antes de aplicar mudanças.
*   **Versionamento Automático:** Criar tags e commits para novas versões de software.

## Referência da API

### `git.clone(url, path)`

Clona um repositório Git de uma URL para um caminho local. Se o caminho já contiver um repositório Git, a função retornará `nil` e uma mensagem de erro.

*   `url` (string): A URL do repositório Git a ser clonado.
*   `path` (string): O caminho local onde o repositório será clonado.

**Retorna:**
*   `GitRepo` (userdata): Uma instância do objeto `GitRepo` se o clone for bem-sucedido.
*   `error` (string): Uma mensagem de erro se o clone falhar ou o caminho já for um repositório.

### `git.repo(path)`

Abre uma referência a um repositório Git local existente.

*   `path` (string): O caminho local para o diretório raiz do repositório Git.

**Retorna:**
*   `GitRepo` (userdata): Uma instância do objeto `GitRepo` se o caminho for um repositório Git válido.
*   `error` (string): Uma mensagem de erro se o caminho não for um repositório Git.

### Métodos do Objeto `GitRepo` (Encadeáveis)

Todos os métodos abaixo são chamados na instância do `GitRepo` (ex: `repo:checkout(...)`) e retornam a própria instância do `GitRepo` para permitir o encadeamento de chamadas. Para obter o resultado da última operação, use o método `:result()`.

#### `repo:checkout(ref)`

Muda o branch ou commit atual do repositório.

*   `ref` (string): O branch, tag ou hash do commit para o qual fazer o checkout.

#### `repo:pull(remote, branch)`

Puxa as últimas alterações de um repositório remoto.

*   `remote` (string): O nome do remoto (ex: "origin").
*   `branch` (string): O nome do branch a ser puxado.

#### `repo:add(pattern)`

Adiciona arquivos ao índice (staging area) do Git.

*   `pattern` (string): O padrão de arquivo a ser adicionado (ex: ".", "path/to/file.txt").

#### `repo:commit(message)`

Cria um novo commit com as alterações no índice.

*   `message` (string): A mensagem do commit.

#### `repo:tag(name, message)`

Cria uma nova tag no repositório.

*   `name` (string): O nome da tag (ex: "v1.0.0").
*   `message` (string, opcional): Uma mensagem opcional para a tag.

#### `repo:push(remote, branch, options)`

Empurra commits e tags para um repositório remoto.

*   `remote` (string): O nome do remoto (ex: "origin").
*   `branch` (string): O nome do branch a ser empurrado.
*   `options` (tabela Lua, opcional): Uma tabela de opções para flags adicionais:
    *   `follow_tags` (booleano): Se `true`, adiciona a flag `--follow-tags` ao comando `git push`.

#### `repo:result()`

Retorna o resultado da última operação Git executada na instância do `GitRepo`.

**Retorna:**
*   `result` (tabela Lua): Uma tabela contendo:
    *   `success` (booleano): `true` se a operação foi bem-sucedida, `false` caso contrário.
    *   `stdout` (string): A saída padrão do comando Git.
    *   `stderr` (string): A saída de erro padrão do comando Git.
    *   `error` (string ou `nil`): Uma mensagem de erro Go se a execução do comando falhou.

## Exemplos de Uso

### Exemplo Básico de Automação Git

Este exemplo demonstra como clonar um repositório, fazer um pull, simular uma alteração, commitar e empurrar as mudanças.

```lua
-- examples/git_example.lua

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

    log.info("Starting git operations on " .. repo.RepoPath .. "...")

    -- Executa uma sequência de comandos de forma fluente e encadeada
    -- Nota: Cada operação retorna o objeto 'repo' para encadeamento.
    -- Para verificar o sucesso de cada passo, você deve chamar :result() após cada um,
    -- ou no final da cadeia para o último comando.

    log.info("Checking out main branch and pulling latest changes...")
    repo:checkout("main"):pull("origin", "main")
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
        :push("origin", "main", { follow_tags = true })

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
```

---
[English](../../en/modules/git.md) | [Português](./git.md) | [中文](../../zh/modules/git.md)