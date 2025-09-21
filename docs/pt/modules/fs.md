# Módulo FS

O módulo `fs` fornece funções essenciais para interagir com o sistema de arquivos diretamente de seus scripts Lua.

---\n

## `fs.read(path)`

Lê todo o conteúdo de um arquivo.

*   **Parâmetros:**
    *   `path` (string): O caminho para o arquivo.
*   **Retorna:**
    *   `string`: O conteúdo do arquivo.
    *   `error`: Um objeto de erro se a leitura falhar.

---\n

## `fs.write(path, content)`

Escreve conteúdo em um arquivo, sobrescrevendo-o se ele já existir.

*   **Parâmetros:**
    *   `path` (string): O caminho para o arquivo.
    *   `content` (string): O conteúdo a ser escrito.
*   **Retorna:**
    *   `error`: Um objeto de erro se a escrita falhar.

---\n

## `fs.append(path, content)`

Adiciona conteúdo ao final de um arquivo. Cria o arquivo se ele não existir.

*   **Parâmetros:**
    *   `path` (string): O caminho para o arquivo.
    *   `content` (string): O conteúdo a ser adicionado.
*   **Retorna:**
    *   `error`: Um objeto de erro se a operação falhar.

---\n

## `fs.exists(path)`

Verifica se um arquivo ou diretório existe no caminho fornecido.

*   **Parâmetros:**
    *   `path` (string): O caminho a ser verificado.
*   **Retorna:**
    *   `boolean`: `true` se o caminho existir, `false` caso contrário.

---\n

## `fs.mkdir(path)`

Cria um diretório no caminho fornecido, incluindo quaisquer diretórios pais necessários (como `mkdir -p`).

*   **Parâmetros:**
    *   `path` (string): O caminho do diretório a ser criado.
*   **Retorna:**
    *   `error`: Um objeto de erro se a criação falhar.

---\n

## `fs.rm(path)`

Remove um único arquivo.

*   **Parâmetros:**
    *   `path` (string): O caminho para o arquivo a ser removido.
*   **Retorna:**
    *   `error`: Um objeto de erro se a remoção falhar.

---\n

## `fs.rm_r(path)`

Remove um arquivo ou diretório recursivamente (como `rm -rf`).

*   **Parâmetros:**
    *   `path` (string): O caminho a ser removido.
*   **Retorna:**
    *   `error`: Um objeto de erro se a remoção falhar.

---\n

## `fs.ls(path)`

Lista o conteúdo de um diretório.

*   **Parâmetros:**
    *   `path` (string): O caminho para o diretório.
*   **Retorna:**
    *   `tabela`: Uma tabela contendo os nomes dos arquivos e subdiretórios.
    *   `error`: Um objeto de erro se a listagem falhar.

---\n

## `fs.tmpname()`

Gera um caminho de diretório temporário único. Nota: Esta função apenas retorna o nome; ela não cria o diretório.

*   **Retorna:**
    *   `string`: Um caminho único adequado para um diretório temporário.
    *   `error`: Um objeto de erro se um nome não puder ser gerado.

### Exemplo

```lua
command = function()
  local fs = require("fs")
  
  local tmp_dir = "/tmp/fs-example"
  log.info("Criando diretório: " .. tmp_dir)
  fs.mkdir(tmp_dir)

  local file_path = tmp_dir .. "/meu_arquivo.txt"
  log.info("Escrevendo no arquivo: " .. file_path)
  fs.write(file_path, "Olá, Sloth Runner!\n")

  log.info("Adicionando ao arquivo...")
  fs.append(file_path, "Esta é uma nova linha.")

  if fs.exists(file_path) then
    log.info("Conteúdo do arquivo: " .. fs.read(file_path))
  end

  log.info("Listando conteúdo de " .. tmp_dir)
  local contents = fs.ls(tmp_dir)
  for i, name in ipairs(contents) do
    print("- " .. name)
  end

  log.info("Limpando...")
  fs.rm_r(tmp_dir)
  
  return true, "Operações do módulo FS bem-sucedidas."
end
```

```