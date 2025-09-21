# Módulo Git

O módulo `git` fornece uma API fluente para interagir com repositórios Git, permitindo que você automatize operações comuns de controle de versão como clonar, commitar e enviar (push).

---

## `git.clone(url, path)`

Clona um repositório Git para um caminho local.

*   **Parâmetros:**
    *   `url` (string): A URL do repositório a ser clonado.
    *   `path` (string): O diretório local para onde clonar.
*   **Retorna:**
    *   `repo` (objeto): Um objeto `GitRepo` em caso de sucesso.
    *   `error`: Um objeto de erro se a clonagem falhar.

---

## `git.repo(path)`

Abre um repositório Git local existente.

*   **Parâmetros:**
    *   `path` (string): O caminho para o repositório local existente.
*   **Retorna:**
    *   `repo` (objeto): Um objeto `GitRepo` em caso de sucesso.
    *   `error`: Um objeto de erro se o caminho não for um repositório Git válido.

---

## O Objeto `GitRepo`

Este objeto representa um repositório local e fornece métodos encadeáveis para realizar operações Git.

### `repo:checkout(ref)`

Faz checkout de um branch, tag ou commit específico.

*   **Parâmetros:** `ref` (string).

### `repo:pull(remote, branch)`

Puxa (pull) as alterações de um repositório remoto.

*   **Parâmetros:** `remote` (string), `branch` (string).

### `repo:add(pattern)`

Adiciona arquivos à área de preparação (staging) para um commit.

*   **Parâmetros:** `pattern` (string), ex: `"."` ou `"caminho/para/arquivo.txt"`.

### `repo:commit(message)`

Cria um commit.

*   **Parâmetros:** `message` (string).

### `repo:tag(name, [message])`

Cria uma nova tag.

*   **Parâmetros:** `name` (string), `message` (string, opcional).

### `repo:push(remote, branch, [options])`

Envia (push) commits para um repositório remoto.

*   **Parâmetros:**
    *   `remote` (string).
    *   `branch` (string).
    *   `options` (tabela, opcional): ex: `{ follow_tags = true }`.

### `repo:result()`

Este método é chamado no final de uma cadeia para obter o resultado da *última* operação.

*   **Retorna:**
    *   `result` (tabela): Uma tabela contendo `success` (booleano), `stdout` (string) e `stderr` (string).

### Exemplo

Este exemplo demonstra um fluxo de trabalho completo semelhante a CI/CD: clonar, criar um arquivo de versão, adicionar, commitar, criar uma tag e enviar (push).

```lua
command = function()
  local git = require("git")
  local repo_path = "/tmp/git-example-repo"
  
  -- Limpa execuções anteriores
  fs.rm_r(repo_path)

  -- 1. Clona o repositório
  log.info("Clonando repositório...")
  local repo, err = git.clone("https://github.com/chalkan3/sloth-runner.git", repo_path)
  if err then
    return false, "Falha ao clonar: " .. err
  end

  -- 2. Cria e escreve um arquivo de versão
  fs.write(repo_path .. "/VERSION", "1.2.3")

  -- 3. Encadear comandos Git: add -> commit -> tag -> push
  log.info("Adicionando, commitando, criando tag e enviando...")
  repo:add("."):commit("ci: Bump version to 1.2.3"):tag("v1.2.3"):push("origin", "main", { follow_tags = true })

  -- 4. Obtém o resultado da operação final (push)
  local result = repo:result()

  if not result.success then
    log.error("O push do Git falhou: " .. result.stderr)
    return false, "O push do Git falhou."
  end

  log.info("Tag da nova versão enviada com sucesso.")
  return true, "Operações Git bem-sucedidas."
end
```
