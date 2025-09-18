# Módulo Salt

O módulo `salt` do Sloth-Runner fornece uma API fluente para interagir com o SaltStack diretamente de seus scripts Lua. Isso permite automatizar a orquestração e configuração de servidores, integrando o poder do Salt em seus fluxos de trabalho do Sloth-Runner.

## Casos de Uso Comuns

*   **Automação de Configuração:** Aplicar estados Salt (`state.apply`) em minions específicos.
*   **Verificação de Status:** Realizar pings (`test.ping`) para verificar a conectividade com minions.
*   **Execução Remota de Comandos:** Executar comandos arbitrários (`cmd.run`) em um ou mais minions.
*   **Orquestração de Deployments:** Coordenar a implantação de aplicações usando funções Salt.

## Referência da API

### `salt.target(target_string)`

Define o alvo (minion ou grupo de minions) para as operações Salt subsequentes.

*   `target_string` (string): O ID do minion, glob, lista, ou outro tipo de alvo suportado pelo Salt.

**Retorna:**
*   `SaltTargeter` (userdata): Uma instância do objeto `SaltTargeter` para o alvo especificado.

### Métodos do Objeto `SaltTargeter` (Encadeáveis)

Todos os métodos abaixo são chamados na instância do `SaltTargeter` (ex: `minion:ping()`) e retornam a própria instância do `SaltTargeter` para permitir o encadeamento de chamadas. Para obter o resultado da última operação, use o método `:result()`.

#### `target:ping()`

Executa o comando `test.ping` no alvo definido.

#### `target:cmd(function, ...args)`

Executa uma função Salt arbitrária no alvo.

*   `function` (string): O nome da função Salt a ser executada (ex: "state.apply", "cmd.run", "pkg.upgrade").
*   `...args` (variadic): Argumentos adicionais a serem passados para a função Salt.

#### `target:result()`

Retorna o resultado da última operação Salt executada na instância do `SaltTargeter`.

**Retorna:**
*   `result` (tabela Lua): Uma tabela contendo:
    *   `success` (booleano): `true` se a operação foi bem-sucedida, `false` caso contrário.
    *   `stdout` (string ou tabela Lua): A saída padrão do comando Salt. Se o Salt retornar JSON válido, será uma tabela Lua.
    *   `stderr` (string): A saída de erro padrão do comando Salt.
    *   `error` (string ou `nil`): Uma mensagem de erro Go se a execução do comando falhou.

## Exemplos de Uso

### Exemplo Básico de Orquestração Salt

Este exemplo demonstra como usar a API fluente do Salt para realizar pings e executar comandos em minions.

```lua
-- examples/fluent_salt_api_test.lua

command = function()
    log.info("Iniciando teste da API fluente do Salt...")

    -- Teste 1: Executando comandos no minion 'keiteguica'
    log.info("Testando alvo único: keiteguica")
    -- Encadeia o comando ping() para o alvo 'keiteguica'
    salt.target('keiteguica'):ping()

    log.info("--------------------------------------------------")

    -- Teste 2: Executando comandos em múltiplos minions usando globbing
    log.info("Testando alvo com glob: vm-gcp-squid-proxy*")
    -- Encadeia os comandos ping() e cmd() para alvos que correspondem ao padrão
    salt.target('vm-gcp-squid-proxy*'):ping():cmd('pkg.upgrade')

    log.info("Teste da API fluente do Salt concluído.")

    log.info("Executando 'ls -la' via Salt e tratando a saída...")
    local result_stdout, result_stderr, result_err = salt.target('keiteguica'):cmd('cmd.run', 'ls -la'):result()

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

TaskDefinitions = {
    test_fluent_salt = {
        description = "Demonstrates using the 'salt' module for SaltStack orchestration.",
        tasks = {
            {
                name = "run_salt_orchestration",
                command = command
            }
        }
    }
}
```

---

[Voltar aos Módulos](../index.md#módulos-built-in) | [Voltar ao Índice](../../index.md)
