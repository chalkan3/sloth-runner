-- examples/workdir_lifecycle_scenarios.lua
--
-- Este arquivo testa especificamente as funcionalidades da DSL para o ciclo de vida do 'workdir'.
-- Cenários abordados:
-- 1. Workdir Efêmero que é limpo após SUCESSO.
-- 2. Workdir Efêmero que é MANTIDO após FALHA.
-- 3. Workdir Fixo que é SEMPRE limpo, independentemente do resultado.

-- Simulação do que o Runner deveria fazer: pré-processar a tabela de definições.
local function setup_workdirs(task_defs)
  for group_name, group in pairs(task_defs) do
    -- Simula a lógica de criação do workdir
    if group.create_workdir_before_run then
      group.workdir = "/tmp/" .. group_name
    else
      -- Gera um "uuid" simples para o teste
      local uuid = os.time()
      group.workdir = "/tmp/" .. group_name .. "-" .. uuid
    end

    -- Simula a injeção do workdir em cada tarefa
    for _, task in ipairs(group.tasks) do
      local original_command = task.command
      task.command = function(params)
        -- Cria o diretório para a simulação
        fs.mkdir(group.workdir)
        -- Passa o workdir para a função original
        return original_command(params, group.workdir)
      end
    end
  end
  return task_defs
end


TaskDefinitions = setup_workdirs({
  -- =================================================================================
  -- CENÁRIO 1: Workdir efêmero, limpo apenas em caso de sucesso.
  -- =================================================================================
  ephemeral_clean_on_success = {
    description = "Usa um workdir único. Executa um script de SUCESSO e o workdir deve ser REMOVIDO.",
    
    clean_workdir_after_run = function(last_result)
      log.info("Avaliando limpeza para 'ephemeral_clean_on_success'...")
      if last_result.success then
        log.info("Tarefa bem-sucedida. O workdir será removido.")
        return true
      else
        log.warn("Tarefa falhou. O workdir seria mantido.")
        return false
      end
    end,

    tasks = {
      {
        name = "run_success_script",
        command = function(params, workdir)
          log.info("Executando em workdir efêmero (deve ser único): " .. workdir)
          
          exec.command("cp", "succeeding_app.py", workdir .. "/app.py")
          exec.command("cp", "requirements.txt", workdir .. "/requirements.txt")

          local python = require("python")
          local venv = python.venv(workdir .. "/.venv")
          venv:create()
          venv:pip("install -r " .. workdir .. "/requirements.txt")
          local result = venv:exec(workdir .. "/app.py")
          
          return result.success, "Script de sucesso executado.", result
        end
      }
    }
  },

  -- =================================================================================
  -- CENÁRIO 2: Workdir efêmero, mantido em caso de falha (para depuração).
  -- =================================================================================
  ephemeral_preserve_on_failure = {
    description = "Usa um workdir único. Executa um script de FALHA e o workdir deve ser MANTIDO.",

    clean_workdir_after_run = function(last_result)
      log.info("Avaliando limpeza para 'ephemeral_preserve_on_failure'...")
      if last_result.success then
        log.warn("Tarefa bem-sucedida. O workdir seria removido.")
        return true
      else
        -- Adicionamos o workdir ao resultado para poder logá-lo aqui.
        local workdir_path = last_result.output and last_result.output.workdir or "N/A"
        log.error("Tarefa falhou. O workdir será MANTIDO para depuração em: " .. workdir_path)
        return false
      end
    end,

    tasks = {
      {
        name = "run_failure_script",
        command = function(params, workdir)
          log.info("Executando em workdir efêmero (deve ser único): " .. workdir)
          
          exec.command("cp", "failing_app.py", workdir .. "/app.py")
          exec.command("cp", "requirements.txt", workdir .. "/requirements.txt")

          local python = require("python")
          local venv = python.venv(workdir .. "/.venv")
          venv:create()
          venv:pip("install -r " .. workdir .. "/requirements.txt")
          local result = venv:exec(workdir .. "/app.py")
          
          -- Adiciona o workdir ao resultado para a função de limpeza poder logá-lo.
          result.workdir = workdir
          return result.success, "Script de falha executado.", result
        end
      }
    }
  },

  -- =================================================================================
  -- CENÁRIO 3: Workdir de caminho fixo, sempre limpo (para CI/CD).
  -- =================================================================================
  fixed_always_clean = {
    description = "Usa um workdir de caminho fixo. O workdir deve ser SEMPRE REMOVIDO.",
    
    create_workdir_before_run = true,

    clean_workdir_after_run = function(last_result)
      log.info("Política de limpeza para workdir fixo: sempre remover, independentemente do resultado.")
      return true
    end,

    tasks = {
      {
        name = "run_in_fixed_workdir",
        command = function(params, workdir)
          log.info("Executando em workdir fixo (deve ser /tmp/fixed_always_clean): " .. workdir)
          
          exec.command("cp", "succeeding_app.py", workdir .. "/app.py")
          exec.command("cp", "requirements.txt", workdir .. "/requirements.txt")

          local python = require("python")
          local venv = python.venv(workdir .. "/.venv")
          venv:create()
          venv:pip("install -r " .. workdir .. "/requirements.txt")
          local result = venv:exec(workdir .. "/app.py")
          
          return result.success, "Script de sucesso executado em workdir fixo.", result
        end
      }
    }
  }
})
