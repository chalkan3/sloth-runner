-- examples/python_venv_lifecycle_example.lua
--
-- ARQUITETURA DA DSL (v2):
-- Esta versão introduz um ciclo de vida explícito para o 'workdir' no nível do grupo de tarefas.
-- O runner que processa esta DSL agora deve interpretar dois novos campos:
--
-- 1. create_workdir_before_run (boolean, opcional):
--    - true: Usa um workdir de caminho FIXO (/tmp/<group_name>) e o limpa (recria) antes da execução.
--    - false/omitido (padrão): Cria um workdir ÚNICO e temporário (/tmp/<group_name>-<uuid>) para cada execução.
--
-- 2. clean_workdir_after_run (function, opcional):
--    - Recebe o resultado da última tarefa executada.
--    - Retorna 'true' para remover o workdir após a execução, ou 'false' para mantê-lo (útil para depuração).
--    - Se omitido, o runner pode ter um comportamento padrão (ex: sempre limpar).

TaskDefinitions = {
  -- CASO DE USO 1: Workdir Efêmero e Limpeza Condicional (Ideal para Desenvolvimento e Debug)
  python_app_ephemeral = {
    description = "Executa a app Python em um workdir novo e único a cada vez. O workdir só é limpo se a execução for bem-sucedida.",

    -- Omitido, então o runner usará o padrão: criar um workdir único como /tmp/python_app_ephemeral-<uuid>
    -- create_workdir_before_run = false,

    -- Função de limpeza: mantém o diretório em caso de falha para permitir a inspeção dos artefatos.
    clean_workdir_after_run = function(last_task_result)
      log.info("Avaliando limpeza do workdir efêmero...")
      if last_task_result.success then
        log.info("A última tarefa foi bem-sucedida. O workdir será removido.")
        return true
      else
        log.error("A última tarefa falhou. O workdir será mantido para depuração.")
        return false
      end
    end,

    tasks = {
      {
        name = "run_python_app",
        description = "Configura e executa a aplicação em um workdir efêmero.",
        command = function(params, workdir)
          log.info("Executando em workdir efêmero: " .. workdir)
          
          local venv_path = workdir .. "/.venv"
          -- O runner seria responsável por popular este diretório com os arquivos necessários.
          local requirements_path = workdir .. "/requirements.txt" 
          local app_path = workdir .. "/app.py"

          local python = require("python")
          local my_venv = python.venv(venv_path)

          my_venv:create()
          my_venv:pip("install -r " .. requirements_path)
          local exec_result = my_venv:exec(app_path)

          return exec_result.success, "Execução no workdir efêmero concluída.", exec_result
        end
      }
    }
  },

  -- CASO DE USO 2: Workdir Fixo e Limpeza Garantida (Ideal para Ambientes de CI/CD)
  python_app_fixed_and_clean = {
    description = "Executa a app Python em um workdir com caminho fixo, garantindo que ele esteja limpo antes e que seja removido depois.",

    -- Garante que o workdir seja sempre /tmp/python_app_fixed_and_clean e que esteja zerado.
    create_workdir_before_run = true,

    -- Função de limpeza: sempre retorna true, garantindo que o workdir seja removido, não importando o resultado.
    clean_workdir_after_run = function(last_task_result)
      log.info("Política de limpeza para workdir fixo: sempre remover.")
      return true
    end,

    tasks = {
      {
        name = "run_python_app_fixed",
        description = "Configura e executa a aplicação em um workdir fixo e limpo.",
        command = function(params, workdir)
          log.info("Executando em workdir fixo e limpo: " .. workdir)
          
          local venv_path = workdir .. "/.venv"
          local requirements_path = workdir .. "/requirements.txt"
          local app_path = workdir .. "/app.py"

          local python = require("python")
          local my_venv = python.venv(venv_path)

          my_venv:create()
          my_venv:pip("install -r " .. requirements_path)
          local exec_result = my_venv:exec(app_path)

          return exec_result.success, "Execução no workdir fixo concluída.", exec_result
        end
      }
    }
  }
}
