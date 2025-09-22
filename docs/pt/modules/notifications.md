# Módulo de Notificações

O módulo `notifications` fornece uma maneira simples de enviar mensagens para vários serviços de notificação a partir de suas pipelines. Isso é particularmente útil para relatar o sucesso ou a falha de um fluxo de trabalho de CI/CD.

Atualmente, os seguintes serviços são suportados:
- [Slack](#slack)
- [ntfy](#ntfy)

## Configuração

Antes de usar o módulo, você precisa adicionar as credenciais ou URLs necessárias ao seu arquivo `configs/values.yaml`. O módulo lerá esses valores em tempo de execução.

```yaml
# configs/values.yaml

notifications:
  slack:
    # Sua URL de Webhook de Entrada do Slack
    webhook_url: "https://hooks.slack.com/services/..."
  ntfy:
    # O servidor ntfy a ser usado. Pode ser o público ou auto-hospedado.
    server: "https://ntfy.sh"
    # O tópico para publicar a notificação.
    topic: "seu-topico-sloth-runner"
```

## Slack

### `notifications.slack.send(params)`

Envia uma mensagem para um canal do Slack através de um Webhook de Entrada.

**Parâmetros:**

- `params` (tabela): Uma tabela contendo os seguintes campos:
    - `webhook_url` (string): **Obrigatório.** A URL do Webhook de Entrada do Slack. Recomenda-se obter isso do módulo `values`.
    - `message` (string): **Obrigatório.** O texto principal da mensagem.
    - `pipeline` (string): **Opcional.** O nome da pipeline, que será exibido no anexo da mensagem para contexto.
    - `error_details` (string): **Opcional.** Quaisquer detalhes de erro a serem incluídos no anexo da mensagem. Isso é útil para notificações de falha.

**Retornos:**

- `true` em caso de sucesso.
- `false, error_message` em caso de falha.

**Exemplo:**

```lua
local values = require("values")

local slack_webhook = values.get("notifications.slack.webhook_url")

if slack_webhook and slack_webhook ~= "" then
  -- Em caso de sucesso
  notifications.slack.send({
    webhook_url = slack_webhook,
    message = "✅ Pipeline executada com sucesso!",
    pipeline = "minha-pipeline-incrivel"
  })

  -- Em caso de falha
  notifications.slack.send({
    webhook_url = slack_webhook,
    message = "❌ Falha na execução da pipeline!",
    pipeline = "minha-pipeline-incrivel",
    error_details = "Não foi possível conectar ao banco de dados."
  })
end
```

## ntfy

### `notifications.ntfy.send(params)`

Envia uma mensagem para um tópico do [ntfy.sh](https://ntfy.sh/).

**Parâmetros:**

- `params` (tabela): Uma tabela contendo os seguintes campos:
    - `server` (string): **Obrigatório.** A URL do servidor ntfy.
    - `topic` (string): **Obrigatório.** O tópico para o qual a mensagem será enviada.
    - `message` (string): **Obrigatório.** O corpo da notificação.
    - `title` (string): **Opcional.** O título da notificação.
    - `priority` (string): **Opcional.** Prioridade da notificação (ex: `high`, `default`, `low`).
    - `tags` (tabela): **Opcional.** Uma lista de tags (emojis) para adicionar à notificação.

**Retornos:**

- `true` em caso de sucesso.
- `false, error_message` em caso de falha.

**Exemplo:**

```lua
local values = require("values")

local ntfy_server = values.get("notifications.ntfy.server")
local ntfy_topic = values.get("notifications.ntfy.topic")

if ntfy_topic and ntfy_topic ~= "" then
  -- Em caso de sucesso
  notifications.ntfy.send({
    server = ntfy_server,
    topic = ntfy_topic,
    title = "Pipeline com Sucesso",
    message = "A pipeline terminou sem erros.",
    priority = "default",
    tags = {"tada"}
  })

  -- Em caso de falha
  notifications.ntfy.send({
    server = ntfy_server,
    topic = ntfy_topic,
    title = "Pipeline Falhou!",
    message = "A pipeline falhou com um erro.",
    priority = "high",
    tags = {"skull", "warning"}
  })
end
```
