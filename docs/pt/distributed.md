# Execução de Tarefas Distribuídas

`sloth-runner` suporta a execução de tarefas distribuídas, permitindo que você execute tarefas em agentes remotos. Isso possibilita fluxos de trabalho escaláveis e distribuídos, onde diferentes partes do seu pipeline podem ser executadas em máquinas distintas.

## Como Funciona

O modelo de execução distribuída no `sloth-runner` segue uma arquitetura mestre-agente:

1.  **Mestre:** A instância principal do `sloth-runner` atua como o mestre. Ela analisa a definição do fluxo de trabalho, identifica as tarefas configuradas para serem executadas em agentes remotos e as despacha.
2.  **Agente:** Uma instância do `sloth-runner` executando no modo `agent` em uma máquina remota. Ela escuta as solicitações de execução de tarefas recebidas do mestre, executa as tarefas e envia os resultados de volta.

## Configurando Tarefas Remotas

Para executar uma tarefa em um agente remoto, você precisa especificar o campo `delegate_to` no grupo de tarefas ou na definição da tarefa individual.

### 1. Delegar a um Agente no Nível do Grupo de Tarefas

Você pode definir o agente diretamente dentro do seu grupo `TaskDefinitions` usando o campo `delegate_to`. Todas as tarefas dentro deste grupo serão então delegadas a este agente, a menos que sejam substituídas por um `delegate_to` específico da tarefa.

```lua
TaskDefinitions = {
  my_distributed_group = {
    description = "Um grupo de tarefas com tarefas distribuídas.",
    delegate_to = { address = "localhost:50051" }, -- Define o agente para todo o grupo
    tasks = {
      {
        name = "remote_hello",
        description = "Executa uma tarefa hello world em um agente remoto.",
        -- Não é necessário o campo 'delegate_to' aqui, ele herda do grupo
        command = function(params)
          log.info("Olá do agente remoto!")
          return true, "Tarefa remota executada."
        end
      }
    }
  }
}
```

### 2. Delegar a um Agente no Nível da Tarefa

Alternativamente, você pode especificar o campo `delegate_to` diretamente em uma tarefa individual. Isso substituirá qualquer delegação em nível de grupo ou permitirá a execução remota ad-hoc.

```lua
TaskDefinitions = {
  my_group = {
    description = "Um grupo de tarefas com uma tarefa remota específica.",
    tasks = {
      {
        name = "specific_remote_task",
        description = "Executa esta tarefa em um agente remoto específico.",
        delegate_to = { address = "192.168.1.100:50051" }, -- Define o agente apenas para esta tarefa
        command = function(params)
          log.info("Olá de um agente remoto específico!")
          return true, "Tarefa remota específica executada."
        end
      },
      {
        name = "local_task",
        description = "Esta tarefa é executada localmente.",
        command = "echo 'Olá da máquina local!'"
      }
    }
  }
}
```

## Executando um Agente

Para iniciar uma instância do `sloth-runner` no modo agente, use o comando `agent`:

```bash
sloth-runner agent -p 50051
```

*   `-p, --port`: Especifica a porta em que o agente deve escutar. O padrão é `50051`.

Quando um agente é iniciado, ele escutará as solicitações gRPC recebidas da instância mestre do `sloth-runner`. Ao receber uma tarefa, ele a executará em seu ambiente local e retornará o resultado, juntamente com quaisquer arquivos de espaço de trabalho atualizados, de volta ao mestre.

## Sincronização do Espaço de Trabalho

Quando uma tarefa é despachada para um agente remoto, o `sloth-runner` lida automaticamente com a sincronização do espaço de trabalho da tarefa:

1.  **Mestre para Agente:** O mestre cria um tarball do diretório de trabalho atual da tarefa e o envia para o agente.
2.  **Execução do Agente:** O agente extrai o tarball para um diretório temporário, executa a tarefa dentro desse diretório e quaisquer alterações feitas nos arquivos no diretório temporário são capturadas.
3.  **Agente para Mestre:** Após a conclusão da tarefa, o agente cria um tarball do diretório temporário modificado e o envia de volta ao mestre. O mestre então extrai esse tarball, atualizando seu espaço de trabalho local com quaisquer alterações feitas pela tarefa remota.