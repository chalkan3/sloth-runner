# Módulo Log

O módulo `log` fornece uma interface simples e essencial para registrar mensagens de seus scripts Lua no console do `sloth-runner`. Usar este módulo é a maneira padrão de fornecer feedback e informações de depuração durante a execução de uma tarefa.

---

## `log.info(message)`

Registra uma mensagem no nível INFO. Este é o nível padrão para mensagens gerais e informativas.

*   **Parâmetros:**
    *   `message` (string): A mensagem a ser registrada.

---

## `log.warn(message)`

Registra uma mensagem no nível WARN. É adequado para problemas não críticos que devem ser levados à atenção do usuário.

*   **Parâmetros:**
    *   `message` (string): A mensagem a ser registrada.

---

## `log.error(message)`

Registra uma mensagem no nível ERROR. Deve ser usado para erros significativos que podem fazer com que uma tarefa falhe.

*   **Parâmetros:**
    *   `message` (string): A mensagem a ser registrada.

---

## `log.debug(message)`

Registra uma mensagem no nível DEBUG. Essas mensagens geralmente ficam ocultas, a menos que o runner esteja em modo detalhado ou de depuração. São úteis para informações de diagnóstico detalhadas.

*   **Parâmetros:**
    *   `message` (string): A mensagem a ser registrada.

### Exemplo

```lua
command = function()
  -- O módulo log está disponível globalmente e não precisa ser requerido.
  
  log.info("Iniciando a tarefa de exemplo de log.")
  
  local user_name = "Sloth"
  log.debug("O usuário atual é: " .. user_name)

  if user_name ~= "Sloth" then
    log.warn("O usuário não é o esperado.")
  end

  log.info("A tarefa está executando sua ação principal...")
  
  local success = true -- Simula uma operação bem-sucedida
  if not success then
    log.error("A ação principal falhou inesperadamente!")
    return false, "Ação principal falhou"
  end

  log.info("Tarefa de exemplo de log concluída com sucesso.")
  return true, "Log demonstrado."
end
```
