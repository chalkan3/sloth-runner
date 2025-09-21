# Módulo GCP

O módulo `gcp` fornece uma interface simples para executar comandos da CLI do Google Cloud (`gcloud`) de dentro de uma tarefa do `sloth-runner`.

## `gcp.exec(args)`

Executa um comando `gcloud` com os argumentos especificados.

### Parâmetros

*   `args` (tabela): Uma tabela Lua (array) de strings representando os argumentos a serem passados para o comando `gcloud`. Por exemplo, `{"compute", "instances", "list"}`.

### Retorna

Uma tabela contendo o resultado da execução do comando com as seguintes chaves:

*   `stdout` (string): A saída padrão do comando.
*   `stderr` (string): A saída de erro padrão do comando.
*   `exit_code` (número): O código de saída do comando. Um código de saída `0` geralmente indica sucesso.

### Exemplo

Este exemplo define uma tarefa que lista todas as instâncias do Compute Engine na região `us-central1` para um projeto específico.

```lua
-- examples/gcp_cli_example.lua

TaskDefinitions = {
  main = {
    description = "Uma tarefa para listar instâncias de computação do GCP.",
    tasks = {
      {
        name = "list-instances",
        description = "Lista instâncias do GCE em us-central1.",
        command = function()
          log.info("Listando instâncias do GCP...")
          
          -- Requer o módulo gcp para torná-lo disponível
          local gcp = require("gcp")

          -- Executa o comando gcloud
          local result = gcp.exec({
            "compute", 
            "instances", 
            "list", 
            "--project", "meu-projeto-gcp-id",
            "--zones", "us-central1-a,us-central1-b"
          })

          -- Verifica o resultado
          if result and result.exit_code == 0 then
            log.info("Instâncias listadas com sucesso.")
            print("--- LISTA DE INSTÂNCIAS ---")
            print(result.stdout)
            print("-------------------------")
            return true, "Comando GCP bem-sucedido."
          else
            log.error("Falha ao listar instâncias do GCP.")
            if result then
              log.error("Stderr: " .. result.stderr)
            end
            return false, "Comando GCP falhou."
          end
        end
      }
    }
  }
}
```
