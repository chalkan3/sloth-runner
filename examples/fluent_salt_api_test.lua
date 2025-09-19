-- examples/fluent_salt_api_test.lua
--
-- Este arquivo de exemplo demonstra o uso da nova API fluente do Salt.
-- Ele executa comandos em alvos específicos usando encadeamento de métodos.

-- A função 'command' é o ponto de entrada para a execução da tarefa.
command = function()
    log.info("Iniciando teste da API fluente do Salt...")

    -- Teste 1: Executando comandos no minion 'keiteguica'
    log.info("Testando alvo único: keiteguica")
    -- Encadeia o comando ping() para o alvo 'keiteguica'
    salt.target('keiteguica', 'glob'):ping()

    log.info("--------------------------------------------------")

    -- Teste 2: Executando comandos em múltiplos minions usando globbing
    log.info("Testando alvo com glob: vm-gcp-squid-proxy*")
    -- Encadeia os comandos ping() e cmd() para alvos que correspondem ao padrão
    salt.target('vm-gcp-squid-proxy*', 'glob'):ping():cmd('pkg.upgrade')

    log.info("Teste da API fluente do Salt concluído.")

    log.info("Executando 'ls -la' via Salt e tratando a saída...")
    local result_stdout, result_stderr, result_err = salt.target('keiteguica', 'glob'):cmd('cmd.run', 'ls -la'):result()

    if result_err ~= nil then
        log.error("Erro ao executar 'ls -la' via Salt: " .. result_err)
        log.error("Stderr: " .. result_stderr)
    else
        log.info("Saída de 'ls -la' via Salt:")
        -- Se a saída for uma tabela (JSON), você pode iterar sobre ela ou convertê-la para string
        if type(result_stdout) == "table" then
            log.info("Saída JSON (tabela): " .. data.to_json(result_stdout))
        else
            log.info(result_stdout)
        end
    end
    log.info("Tratamento da saída de 'ls -la' via Salt concluído.")

    return true, "Comandos da API fluente do Salt e 'ls -la' executados com sucesso."
end

-- Definição da tarefa para o sloth-runner
TaskDefinitions = {
    test_fluent_salt = {
        description = "Testa a nova API fluente do Salt com múltiplos alvos e comandos.",
        tasks = {
            {
                name = "run_salt_tests",
                command = command
            }
        }
    }
}
