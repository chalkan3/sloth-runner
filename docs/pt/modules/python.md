# Módulo Python

O módulo `python` fornece uma maneira conveniente de gerenciar ambientes virtuais Python (`venv`) e executar scripts de dentro de suas tarefas do `sloth-runner`. Isso é particularmente útil para fluxos de trabalho que envolvem ferramentas ou scripts baseados em Python.

---

## `python.venv(path)`

Cria um objeto de ambiente virtual Python. Note que isso apenas cria o objeto em Lua; o ambiente em si não é criado no sistema de arquivos até que você chame `:create()`.

*   **Parâmetros:**
    *   `path` (string): O caminho no sistema de arquivos onde o ambiente virtual deve ser criado (ex: `./.venv`).
*   **Retorna:**
    *   `venv` (objeto): Um objeto de ambiente virtual com métodos para interagir com ele.

---

### `venv:create()`

Cria o ambiente virtual no sistema de arquivos no caminho especificado.

*   **Retorna:**
    *   `error`: Um objeto de erro se a criação falhar.

---

### `venv:pip(command)`

Executa um comando `pip` dentro do contexto do ambiente virtual.

*   **Parâmetros:**
    *   `command` (string): Os argumentos a serem passados para o `pip` (ex: `install -r requirements.txt`).
*   **Retorna:**
    *   `result` (tabela): Uma tabela contendo `stdout`, `stderr` e `exit_code` do comando `pip`.

---

### `venv:exec(script_path)`

Executa um script Python usando o interpretador Python do ambiente virtual.

*   **Parâmetros:**
    *   `script_path` (string): O caminho para o script Python a ser executado.
*   **Retorna:**
    *   `result` (tabela): Uma tabela contendo `stdout`, `stderr` e `exit_code` da execução do script.

### Exemplo

Este exemplo demonstra um ciclo de vida completo: criar um ambiente virtual, instalar dependências de um arquivo `requirements.txt` e executar um script Python.

```lua
-- examples/python_venv_lifecycle_example.lua

TaskDefinitions = {
  main = {
    description = "Uma tarefa para demonstrar o ciclo de vida de um venv Python.",
    create_workdir_before_run = true, -- Usa um diretório de trabalho temporário
    tasks = {
      {
        name = "run-python-script",
        description = "Cria um venv, instala dependências e executa um script.",
        command = function(params)
          local python = require("python")
          local workdir = params.workdir -- Obtém o diretório de trabalho temporário do grupo
          
          -- 1. Escreve nosso script Python e dependências no workdir
          fs.write(workdir .. "/requirements.txt", "requests==2.28.1")
          fs.write(workdir .. "/main.py", "import requests\nprint(f'Olá do Python! Usando a versão do requests: {requests.__version__}')")

          -- 2. Cria um objeto venv
          local venv_path = workdir .. "/.venv"
          log.info("Configurando ambiente virtual em: " .. venv_path)
          local venv = python.venv(venv_path)

          -- 3. Cria o venv no sistema de arquivos
          venv:create()

          -- 4. Instala as dependências usando pip
          log.info("Instalando dependências do requirements.txt...")
          local pip_result = venv:pip("install -r " .. workdir .. "/requirements.txt")
          if pip_result.exit_code ~= 0 then
            log.error("A instalação com pip falhou: " .. pip_result.stderr)
            return false, "Falha ao instalar dependências Python."
          end

          -- 5. Executa o script
          log.info("Executando o script Python...")
          local exec_result = venv:exec(workdir .. "/main.py")
          if exec_result.exit_code ~= 0 then
            log.error("O script Python falhou: " .. exec_result.stderr)
            return false, "A execução do script Python falhou."
          end

          log.info("Script Python executado com sucesso.")
          print("--- Saída do Script Python ---")
          print(exec_result.stdout)
          print("----------------------------")

          return true, "Ciclo de vida do venv Python completo."
        end
      }
    }
  }
}
```

