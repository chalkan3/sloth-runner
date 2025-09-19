-- examples/salt_accept_and_ping.lua
--
-- Este workflow demonstra como esperar por uma nova chave de minion,
-- aceitá-la automaticamente e, em seguida, usar o ID do novo minion
-- em uma tarefa subsequente para verificá-lo com test.ping.

TaskDefinitions = {
  ["accept_and_ping_workflow"] = {
    description = "Um workflow para aceitar uma nova chave de minion e depois pingá-la.",
    tasks = {
      {
        name = "wait_and_accept_key",
        description = "Espera até que uma nova chave pendente apareça e a aceita.",
        command = function()
          local log = require("log")
          local salt = require("salt")

          log.info("Aguardando uma nova chave de minion... O script ficará bloqueado aqui até que uma nova chave seja detectada.")
          
          local new_minion_id, err = salt.key.wait_and_accept()
          
          if err then
            log.error("Falha ao esperar ou aceitar a chave: " .. err)
            return false, "Falha ao aceitar a chave."
          end
          
          log.info("Chave para o minion '" .. new_minion_id .. "' foi aceita com sucesso.")
          
          -- Retorna o ID do minion para que a próxima tarefa possa usá-lo
          return true, "Chave aceita.", { minion_id = new_minion_id }
        end
      },
      {
        name = "ping_new_minion",
        description = "Executa test.ping no minion que foi recém-aceito.",
        depends_on = "wait_and_accept_key",
        retries = 10,
        command = function(params, inputs)
          local log = require("log")
          local salt = require("salt")
          local data = require("data")

          -- Obtém o minion_id da saída da tarefa anterior
          local minion_id = inputs.wait_and_accept_key.minion_id
          
          if not minion_id then
            log.error("O ID do minion não foi recebido da tarefa anterior.")
            return false, "ID do minion não encontrado."
          end
          
          log.info("Pingando o novo minion: '" .. minion_id .. "'...")
          
          local salt_client = salt.client()
          local success, stdout, stderr, err = salt_client:target(minion_id):cmd("test.ping")
          
          if err or not success then
            local error_message = "Falha ao pingar o minion '" .. minion_id .. "'."
            if err then error_message = error_message .. " Erro: " .. err end
            if stderr then error_message = error_message .. " Stderr: " .. stderr end
            log.error(error_message)
            return false, "Ping falhou."
          end
          
          log.info("Resultado do Ping para '" .. minion_id .. "': " .. stdout)
          log.info("Minion '" .. minion_id .. "' está online e respondendo.")
          
          return true, "Ping bem-sucedido."
        end
      }
    }
  }
}