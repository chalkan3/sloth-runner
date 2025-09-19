-- examples/comprehensive_scenarios.lua
--
-- Este arquivo demonstra vários cenários de uso do Sloth Runner,
-- servindo como um teste de integração e um guia de funcionalidades.

TaskDefinitions = {
  -- ===================================================================
  -- GRUPO 1: Comandos Básicos, Módulos e Dependências
  -- ===================================================================
  basic_features = {
    description = "Testa a execução de comandos simples, o módulo 'fs' e dependências.",
    tasks = {
      {
        name = "print_message",
        description = "Executa um comando 'echo' simples.",
        command = "echo 'Hello from a simple command!'"
      },
      {
        name = "read_file",
        description = "Usa o módulo 'fs' para ler um arquivo local.",
        command = function()
          local content, err = fs.read("test_file.txt")
          if err then
            log.error("Falha ao ler test_file.txt: " .. err)
            return false, "File read failed"
          end
          log.info("Conteúdo do arquivo: " .. content)
          return true, "File read successfully"
        end
      },
      {
        name = "chained_task",
        description = "Executa somente após 'print_message'.",
        depends_on = "print_message",
        command = "echo 'Esta tarefa executou depois de print_message.'"
      }
    }
  },

  -- ===================================================================
  -- GRUPO 2: Manipulação de Dados com Módulos 'net' e 'data'
  -- ===================================================================
  data_pipeline = {
    description = "Busca dados de uma API, processa com JSON e usa o resultado.",
    tasks = {
      {
        name = "fetch_api_data",
        description = "Busca dados de httpbin.org.",
        command = function()
          local body, status, _, err = net.http_get("https://httpbin.org/json")
          if err or status ~= 200 then
            return false, "API fetch failed", { error = err or "status not 200" }
          end
          return true, "API data fetched", { json_string = body }
        end
      },
      {
        name = "parse_json",
        description = "Analisa o JSON retornado pela API.",
        depends_on = "fetch_api_data",
        command = function(params, inputs)
          local json_str = inputs.fetch_api_data.json_string
          local parsed, err = data.parse_json(json_str)
          if err then
            return false, "JSON parsing failed", { error = err }
          end
          log.info("Título do slideshow: " .. parsed.slideshow.title)
          return true, "JSON parsed successfully", { title = parsed.slideshow.title }
        end
      }
    }
  },

  -- ===================================================================
  -- GRUPO 3: Execução Paralela
  -- ===================================================================
  parallel_tasks = {
    description = "Executa múltiplas tarefas em paralelo.",
    tasks = {
      { name = "sleep_1", command = "sleep 1 && echo 'Slept for 1s'" },
      { name = "sleep_2", command = "sleep 2 && echo 'Slept for 2s'" },
      {
        name = "run_in_parallel",
        description = "Usa a função 'parallel' para executar tarefas concorrentemente.",
        command = function()
          log.info("Iniciando execução paralela...")
          local results, err = parallel({
            { name = "sleep_1" },
            { name = "sleep_2" }
          })
          if err then
            return false, "Parallel execution failed: " .. err
          end
          log.info("Execução paralela concluída.")
          return true, "Parallel tasks finished", results
        end
      }
    }
  },

  -- ===================================================================
  -- GRUPO 4: Ciclo de Vida do Workdir (Sucesso e Falha)
  -- ===================================================================
  python_lifecycle = {
    description = "Testa a criação e limpeza do workdir em cenários de sucesso e falha.",
    
    -- Política de limpeza: manter o workdir se a tarefa falhar.
    clean_workdir_after_run = function(last_result)
      return last_result.success
    end,

    tasks = {
      {
        name = "python_succeeds",
        description = "Executa um script Python que deve ter sucesso.",
        command = function(params, workdir)
          -- O runner deve copiar os arquivos para o workdir antes.
          -- Para este teste, vamos criá-los dinamicamente.
          fs.write(workdir .. "/requirements.txt", "requests==2.28.1")
          fs.write(workdir .. "/app.py", "import sys; print('Success!'); sys.exit(0)")

          local python = require("python")
          local venv = python.venv(workdir .. "/.venv")
          venv:create()
          venv:pip("install -r " .. workdir .. "/requirements.txt")
          local result = venv:exec(workdir .. "/app.py")
          
          log.info("Resultado da execução Python (sucesso): " .. result.stdout)
          return result.success, "Python script finished.", result
        end
      },
      {
        name = "python_fails",
        description = "Executa um script Python que deve falhar.",
        command = function(params, workdir)
          fs.write(workdir .. "/requirements.txt", "requests==2.28.1")
          fs.write(workdir .. "/app.py", "import sys; print('Failure!', file=sys.stderr); sys.exit(1)")

          local python = require("python")
          local venv = python.venv(workdir .. "/.venv")
          venv:create()
          venv:pip("install -r " .. workdir .. "/requirements.txt")
          local result = venv:exec(workdir .. "/app.py")
          
          log.info("Resultado da execução Python (falha): " .. result.stderr)
          return result.success, "Python script finished.", result
        end
      }
    }
  },

  -- ===================================================================
  -- GRUPO 5: Tratamento de Erros (Retries e Next-If-Fail)
  -- ===================================================================
  error_handling = {
    description = "Demonstra as capacidades de retentativas e fluxo de falha.",
    tasks = {
      {
        name = "flaky_task",
        description = "Uma tarefa que falha na primeira vez.",
        retries = 2,
        command = function()
          local marker_file = "/tmp/sloth_marker"
          if not fs.exists(marker_file) then
            log.warn("Tentativa 1: Falhando de propósito.")
            fs.write(marker_file, "exists")
            return false, "Falha simulada"
          else
            log.info("Tentativa 2: Sucesso!")
            fs.rm(marker_file)
            return true, "Sucesso na retentativa"
          end
        end
      },
      {
        name = "always_fail",
        description = "Uma tarefa que sempre falha.",
        command = "exit 1"
      },
      {
        name = "cleanup_on_fail",
        description = "Executa somente se 'always_fail' falhar.",
        next_if_fail = "always_fail",
        command = "echo 'Executando limpeza após a falha!'"
      }
    }
  }
}
