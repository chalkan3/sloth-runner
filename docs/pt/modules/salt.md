# Módulo Salt

O módulo `salt` fornece uma API fluente para interagir com o SaltStack, permitindo que você execute comandos de execução remota e gerencie configurações a partir de seus fluxos de trabalho do `sloth-runner`.

---

## `salt.client([options])`

Cria um objeto de cliente Salt.

*   **Parâmetros:**
    *   `options` (tabela, opcional): Uma tabela de opções.
        *   `config_path` (string): Caminho para o arquivo de configuração do Salt master.
*   **Retorna:**
    *   `client` (objeto): Um objeto `SaltClient`.

---

## O Objeto `SaltClient`

Este objeto representa um cliente para um Salt master e fornece métodos para direcionar minions.

### `client:target(target_string, [expr_form])`

Especifica o(s) minion(s) alvo para um comando.

*   **Parâmetros:**
    *   `target_string` (string): A expressão de alvo (ex: `"*"` para todos os minions, `"web-server-1"`, ou um valor de grain).
    *   `expr_form` (string, opcional): O tipo de direcionamento a ser usado (ex: `"glob"`, `"grain"`, `"list"`). O padrão é glob.
*   **Retorna:**
    *   `target` (objeto): Um objeto `SaltTarget`.

---

## O Objeto `SaltTarget`

Este objeto representa um alvo específico e fornece métodos encadeáveis para executar funções do Salt.

### `target:cmd(function, [arg1, arg2, ...])`

Executa uma função do módulo de execução do Salt no alvo.

*   **Parâmetros:**
    *   `function` (string): O nome da função a ser executada (ex: `"test.ping"`, `"state.apply"`, `"cmd.run"`).
    *   `arg1`, `arg2`, ... (qualquer): Argumentos adicionais a serem passados para a função do Salt.
*   **Retorna:**
    *   `result` (tabela): Uma tabela contendo `success` (booleano), `stdout` (string ou tabela) e `stderr` (string). Se o comando Salt retornar JSON, `stdout` será uma tabela Lua analisada.

### Exemplo

Este exemplo demonstra como direcionar minions para pingá-los e aplicar um estado do Salt.

```lua
command = function()
  local salt = require("salt")

  -- 1. Cria um cliente Salt
  local client = salt.client()

  -- 2. Direciona todos os minions e os pinga
  log.info("Pingando todos os minions...")
  local ping_result = client:target("*"):cmd("test.ping")
  if not ping_result.success then
    return false, "Falha ao pingar minions: " .. ping_result.stderr
  end
  print("Resultados do Ping:")
  print(data.to_yaml(ping_result.stdout)) -- stdout é uma tabela

  -- 3. Direciona um servidor web específico e aplica um estado
  log.info("Aplicando o estado 'nginx' em web-server-1...")
  local apply_result = client:target("web-server-1", "glob"):cmd("state.apply", "nginx")
  if not apply_result.success then
    return false, "Falha ao aplicar o estado: " .. apply_result.stderr
  end
  
  log.info("Estado aplicado com sucesso.")
  return true, "Operações do Salt concluídas."
end
```
