-- examples/python_venv_example.lua (Refatorado)
--
-- NOTA: Este script foi refatorado para ser processado por um runner que
-- gerencia e injeta o diretório de trabalho ('workdir') automaticamente.
-- A lógica do runner deve:
-- 1. Garantir um 'workdir' para o grupo, usando '/tmp/<group_name>' como padrão.
-- 2. Injetar o caminho final do 'workdir' como um segundo argumento para a função 'command'.

TaskDefinitions = {
  python_app = {
    -- O campo 'workdir' foi intencionalmente omitido.
    -- O runner irá agora calcular o padrão: /tmp/python_app
    description = "A task group to manage and run the Python application.",
    tasks = {
      {
        name = "run_python_app",
        description = "Configura o ambiente virtual e executa a aplicação Python dentro de um workdir gerenciado.",
        
        -- A assinatura da função agora aceita 'workdir' como um argumento injetado pelo runner.
        command = function(params, workdir)
          -- Todos os caminhos são construídos de forma segura dentro do workdir fornecido.
          local venv_path = workdir .. "/.venv"
          local requirements_path = workdir .. "/requirements.txt"
          local app_path = workdir .. "/app.py"

          -- Assume-se que o runner também é responsável por popular o workdir
          -- com os arquivos necessários (ex: app.py, requirements.txt) antes da execução.
          log.info("Executando tarefa no workdir gerenciado: " .. workdir)

          local python = require("python")
          local my_venv = python.venv(venv_path)

          if not my_venv:exists() then
            log.info("Criando ambiente virtual Python em: " .. venv_path)
            local create_result = my_venv:create()
            if not create_result.success then
              log.error("Falha ao criar o venv: " .. create_result.stderr)
              return false, "venv creation failed"
            end
          else
            log.info("Ambiente virtual Python já existe em: " .. venv_path)
          end

          log.info("Instalando dependências de " .. requirements_path)
          local pip_result = my_venv:pip("install -r " .. requirements_path)
          if not pip_result.success then
            log.error("Falha ao instalar dependências: " .. pip_result.stderr)
            return false, "pip install failed"
          end

          log.info("Executando o script " .. app_path)
          local exec_result = my_venv:exec(app_path)
          if not exec_result.success then
            log.error("Falha ao executar app.py: " .. exec_result.stderr)
            return false, "python exec failed"
          end

          log.info("Saída do app.py:")
          print(exec_result.stdout)

          return true, "Python app executed successfully from " .. workdir
        end
      }
    }
  }
}