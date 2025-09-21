# Módulo Data

O módulo `data` fornece funções para analisar (parse) e serializar dados entre tabelas Lua e formatos de dados comuns como JSON e YAML.

---\n

## `data.parse_json(json_string)`

Analisa uma string JSON e a converte em uma tabela Lua.

*   **Parâmetros:**
    *   `json_string` (string): A string formatada em JSON para analisar.
*   **Retorna:**
    *   `tabela`: A tabela Lua resultante.
    *   `error`: Um objeto de erro se a análise falhar.

---\n

## `data.to_json(lua_table)`

Serializa uma tabela Lua em uma string JSON.

*   **Parâmetros:**
    *   `lua_table` (tabela): A tabela Lua a ser serializada.
*   **Retorna:**
    *   `string`: A string JSON resultante.
    *   `error`: Um objeto de erro se a serialização falhar.

---\n

## `data.parse_yaml(yaml_string)`

Analisa uma string YAML e a converte em uma tabela Lua.

*   **Parâmetros:**
    *   `yaml_string` (string): A string formatada em YAML para analisar.
*   **Retorna:**
    *   `tabela`: A tabela Lua resultante.
    *   `error`: Um objeto de erro se a análise falhar.

---\n

## `data.to_yaml(lua_table)`

Serializa uma tabela Lua em uma string YAML.

*   **Parâmetros:**
    *   `lua_table` (tabela): A tabela Lua a ser serializada.
*   **Retorna:**
    *   `string`: A string YAML resultante.
    *   `error`: Um objeto de erro se a serialização falhar.

### Exemplo

```lua
command = function()
  local data = require("data")

  -- Exemplo JSON
  log.info("Testando serialização JSON...")
  local minha_tabela = { name = "sloth-runner", version = 1.0, features = { "tasks", "lua" } }
  local json_str, err = data.to_json(minha_tabela)
  if err then
    return false, "Falha ao serializar para JSON: " .. err
  end
  print("JSON Serializado: " .. json_str)

  log.info("Testando análise de JSON...")
  local tabela_parseada, err = data.parse_json(json_str)
  if err then
    return false, "Falha ao analisar JSON: " .. err
  end
  log.info("Nome extraído do JSON: " .. tabela_parseada.name)

  -- Exemplo YAML
  log.info("Testando serialização YAML...")
  local yaml_str, err = data.to_yaml(minha_tabela)
  if err then
    return false, "Falha ao serializar para YAML: " .. err
  end
  print("YAML Serializado:\n" .. yaml_str)
  
  log.info("Testando análise de YAML...")
  tabela_parseada, err = data.parse_yaml(yaml_str)
  if err then
    return false, "Falha ao analisar YAML: " .. err
  end
  log.info("Versão extraída do YAML: " .. tabela_parseada.version)

  return true, "Operações do módulo Data bem-sucedidas."
end
```

```