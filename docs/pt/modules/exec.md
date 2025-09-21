# Módulo Exec

O módulo `exec` é um dos módulos mais fundamentais do `sloth-runner`. Ele fornece uma função poderosa para executar comandos de shell arbitrários, dando a você controle total sobre o ambiente de execução.

## `exec.run(command, [options])`

Executa um comando de shell usando `bash -c`.

### Parâmetros

*   `command` (string): O comando de shell a ser executado.
*   `options` (tabela, opcional): Uma tabela de opções para controlar a execução.
    *   `workdir` (string): O diretório de trabalho onde o comando deve ser executado. Se não for fornecido, ele é executado no diretório temporário do grupo de tarefas (se disponível) ou no diretório atual.
    *   `env` (tabela): Um dicionário de variáveis de ambiente (pares chave-valor) a serem definidas para a execução do comando. Elas são adicionadas ao ambiente existente.

### Retorna

Uma tabela contendo o resultado da execução do comando:

*   `success` (booleano): `true` se o comando saiu com o código `0`, caso contrário `false`.
*   `stdout` (string): A saída padrão do comando.
*   `stderr` (string): A saída de erro padrão do comando.

### Exemplo

Este exemplo demonstra como usar `exec.run` com um diretório de trabalho e variáveis de ambiente personalizados.

```lua
-- examples/exec_module_example.lua

TaskDefinitions = {
  main = {
    description = "Uma tarefa para demonstrar o módulo exec.",
    tasks = {
      {
        name = "run-with-options",
        description = "Executa um comando com um workdir e ambiente personalizados.",
        command = function()
          log.info("Preparando para executar um comando personalizado...")
          
          local exec = require("exec")
          
          -- Cria um diretório temporário para o exemplo
          local temp_dir = "/tmp/sloth-exec-test"
          fs.mkdir(temp_dir)
          fs.write(temp_dir .. "/test.txt", "olá do arquivo de teste")

          -- Define as opções
          local options = {
            workdir = temp_dir,
            env = {
              MINHA_VAR = "SlothRunner",
              OUTRA_VAR = "e_incrivel"
            }
          }

          -- Executa o comando
          local result = exec.run("echo 'MINHA_VAR é $MINHA_VAR' && ls -l && cat test.txt", options)

          -- Limpa o diretório temporário
          fs.rm_r(temp_dir)

          if result.success then
            log.info("Comando executado com sucesso!")
            print("--- STDOUT ---")
            print(result.stdout)
            print("--------------")
            return true, "Comando exec bem-sucedido."
          else
            log.error("Comando exec falhou.")
            log.error("Stderr: " .. result.stderr)
            return false, "Comando exec falhou."
          end
        end
      }
    }
  }
}
```
