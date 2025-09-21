# Módulo Net

O módulo `net` fornece funções para fazer requisições HTTP e baixar arquivos, permitindo que suas tarefas interajam com serviços web e recursos remotos.

---

## `net.http_get(url)`

Realiza uma requisição HTTP GET para a URL especificada.

*   **Parâmetros:**
    *   `url` (string): A URL para a qual enviar a requisição GET.
*   **Retorna:**
    *   `body` (string): O corpo da resposta como uma string.
    *   `status_code` (número): O código de status HTTP da resposta.
    *   `headers` (tabela): Uma tabela contendo os cabeçalhos da resposta.
    *   `error` (string): Uma mensagem de erro se a requisição falhar.

---

## `net.http_post(url, body, [headers])`

Realiza uma requisição HTTP POST para a URL especificada.

*   **Parâmetros:**
    *   `url` (string): A URL para a qual enviar a requisição POST.
    *   `body` (string): O corpo da requisição a ser enviado.
    *   `headers` (tabela, opcional): Uma tabela de cabeçalhos de requisição a serem definidos.
*   **Retorna:**
    *   `body` (string): O corpo da resposta como uma string.
    *   `status_code` (número): O código de status HTTP da resposta.
    *   `headers` (tabela): Uma tabela contendo os cabeçalhos da resposta.
    *   `error` (string): Uma mensagem de erro se a requisição falhar.

---

## `net.download(url, destination_path)`

Baixa um arquivo de uma URL e o salva em um caminho local.

*   **Parâmetros:**
    *   `url` (string): A URL do arquivo a ser baixado.
    *   `destination_path` (string): O caminho do arquivo local para salvar o conteúdo baixado.
*   **Retorna:**
    *   `error`: Um objeto de erro se o download falhar.

### Exemplo

```lua
command = function()
  local net = require("net")
  
  -- Exemplo de requisição GET
  log.info("Realizando requisição GET para httpbin.org...")
  local body, status, headers, err = net.http_get("https://httpbin.org/get")
  if err then
    log.error("Requisição GET falhou: " .. err)
    return false, "Requisição GET falhou"
  end
  log.info("Requisição GET bem-sucedida! Status: " .. status)
  -- print("Corpo da Resposta: " .. body)

  -- Exemplo de requisição POST
  log.info("Realizando requisição POST para httpbin.org...")
  local post_body = '{"name": "sloth-runner", "awesome": true}'
  local post_headers = { ["Content-Type"] = "application/json" }
  body, status, headers, err = net.http_post("https://httpbin.org/post", post_body, post_headers)
  if err then
    log.error("Requisição POST falhou: " .. err)
    return false, "Requisição POST falhou"
  end
  log.info("Requisição POST bem-sucedida! Status: " .. status)
  -- print("Corpo da Resposta: " .. body)

  -- Exemplo de Download
  local download_path = "/tmp/sloth-runner-logo.svg"
  log.info("Baixando arquivo para " .. download_path)
  local err = net.download("https://raw.githubusercontent.com/chalkan3/sloth-runner/master/assets/sloth-runner-logo.svg", download_path)
  if err then
    log.error("Download falhou: " .. err)
    return false, "Download falhou"
  end
  log.info("Arquivo baixado com sucesso.")
  fs.rm(download_path) -- Limpeza

  return true, "Operações do módulo Net bem-sucedidas."
end
```
