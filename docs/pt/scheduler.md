# Agendador de Tarefas

O `sloth-runner` agora inclui um agendador de tarefas integrado, permitindo automatizar a execução de suas tarefas definidas em Lua em intervalos específicos usando a sintaxe cron.

## Funcionalidades

*   **Processo em Segundo Plano:** O agendador é executado como um processo persistente em segundo plano, independente da sua sessão de terminal.
*   **Agendamento Baseado em Cron:** Defina agendamentos de tarefas usando strings cron flexíveis.
*   **Persistência:** As tarefas agendadas são carregadas de um arquivo de configuração, garantindo que sejam retomadas após reinícios.
*   **Integração com Tarefas Existentes:** O agendador utiliza o comando `sloth-runner run` existente para executar suas tarefas.

## Configuração: `scheduler.yaml`

As tarefas agendadas são definidas em um arquivo YAML, tipicamente chamado `scheduler.yaml`. Este arquivo especifica as tarefas a serem executadas, seu agendamento e o arquivo Lua, grupo e nome da tarefa.

```yaml
scheduled_tasks:
  - name: "my_daily_backup"
    schedule: "0 0 * * *" # Todo dia à meia-noite
    task_file: "examples/my_workflow.lua"
    task_group: "backup_group"
    task_name: "perform_backup"
  - name: "hourly_report_generation"
    schedule: "0 * * * *" # Toda hora
    task_file: "examples/reporting.lua"
    task_group: "reports"
    task_name: "generate_report"
```

**Campos:**

*   `name` (string, obrigatório): Um nome único para a tarefa agendada.
*   `schedule` (string, obrigatório): A string cron que define quando a tarefa deve ser executada. Suporta a sintaxe cron padrão e alguns agendamentos predefinidos (ex: `@every 1h`, `@daily`). Consulte a [documentação do robfig/cron](https://pkg.go.dev/github.com/robfig/cron/v3#hdr-CRON_Expression_Format) para detalhes.
*   `task_file` (string, obrigatório): O caminho para o arquivo de definição da tarefa Lua.
*   `task_group` (string, obrigatório): O nome do grupo de tarefas dentro do arquivo Lua.
*   `task_name` (string, obrigatório): O nome da tarefa específica a ser executada dentro do grupo de tarefas.

## Comandos CLI

### `sloth-runner scheduler enable`

Inicia o agendador do `sloth-runner` como um processo em segundo plano. Este comando garante que o agendador esteja em execução e pronto para processar tarefas agendadas.

```bash
sloth-runner scheduler enable --scheduler-config scheduler.yaml
```

*   `--scheduler-config` (ou `-c`): Especifica o caminho para o seu arquivo de configuração `scheduler.yaml`. O padrão é `scheduler.yaml` no diretório atual.

Após a execução, o comando imprimirá o PID do processo do agendador em segundo plano. O agendador continuará a ser executado mesmo que sua sessão de terminal seja fechada.

### `sloth-runner scheduler disable`

Para o processo em segundo plano do agendador do `sloth-runner` em execução.

```bash
sloth-runner scheduler disable
```

Este comando tentará encerrar o processo do agendador de forma graciosa. Se bem-sucedido, ele removerá o arquivo PID criado pelo comando `enable`.

### `sloth-runner scheduler list`

Lista todas as tarefas agendadas definidas no arquivo de configuração `scheduler.yaml`. Este comando fornece uma visão geral de suas tarefas configuradas, seus agendamentos e detalhes da tarefa Lua associada.

```bash
sloth-runner scheduler list --scheduler-config scheduler.yaml
```

*   `--scheduler-config` (ou `-c`): Especifica o caminho para o seu arquivo de configuração `scheduler.yaml`. O padrão é `scheduler.yaml` no diretório atual.

**Exemplo de Saída:**

```
# Configured Scheduled Tasks

NAME                     | SCHEDULE    | FILE                     | GROUP        | TASK
my_daily_backup          | 0 0 * * *   | examples/my_workflow.lua | backup_group | perform_backup
hourly_report_generation | 0 * * * *   | examples/reporting.lua   | reports      | generate_report
```

### `sloth-runner scheduler delete <task_name>`

Exclui uma tarefa agendada específica do arquivo de configuração `scheduler.yaml`. Este comando remove a definição da tarefa, e o agendador não a executará mais.

```bash
sloth-runner scheduler delete my_daily_backup --scheduler-config scheduler.yaml
```

*   `<task_name>` (string, obrigatório): O nome único da tarefa agendada a ser excluída.
*   `--scheduler-config` (ou `-c`): Especifica o caminho para o seu arquivo de configuração `scheduler.yaml`. O padrão é `scheduler.yaml` no diretório atual.

**Importante:** Este comando modifica seu arquivo `scheduler.yaml`. Certifique-se de ter um backup, se necessário. Se o agendador estiver em execução, pode ser necessário desativá-lo e reativá-lo para que as alterações entrem em vigor imediatamente.

## Registro e Tratamento de Erros

O agendador registra suas atividades e o status de execução das tarefas agendadas na saída padrão e no erro padrão. Recomenda-se redirecionar essas saídas para um arquivo de log ao executar em um ambiente de produção.

Se uma tarefa agendada falhar, o agendador registrará o erro e continuará com outras tarefas agendadas. Ele não será interrompido devido a falhas de tarefas individuais.

## Exemplo

1.  Crie um arquivo `scheduler.yaml`:

    ```yaml
    scheduled_tasks:
      - name: "my_test_task"
        schedule: "@every 1m"
        task_file: "examples/basic_pipeline.lua"
        task_group: "basic_pipeline"
        task_name: "fetch_data"
    ```

2.  Habilite o agendador:

    ```bash
    sloth-runner scheduler enable --scheduler-config scheduler.yaml
    ```

3.  Observe a saída. A cada minuto, você deverá ver mensagens indicando a execução de `my_test_task`.

4.  Para parar o agendador:

    ```bash
    sloth-runner scheduler disable
    ```
