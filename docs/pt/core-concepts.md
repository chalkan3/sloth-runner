# Conceitos Essenciais

Este documento explica os conceitos fundamentais do Sloth-Runner, ajudando você a entender como as tarefas são definidas e executadas.

## Definição de Tarefas em Lua

As tarefas no Sloth-Runner são definidas em arquivos Lua, tipicamente dentro de uma tabela global chamada `TaskDefinitions`. Esta tabela é um mapa onde as chaves são os nomes dos grupos de tarefas e os valores são tabelas de grupo.

### Estrutura de um Grupo de Tarefas

Cada grupo de tarefas possui:
*   `description`: Uma descrição textual do grupo.
*   `tasks`: Uma tabela contendo as definições das tarefas individuais.

### Estrutura de uma Tarefa Individual

Cada tarefa individual pode ter os seguintes campos:

*   `name` (string): O nome único da tarefa dentro do seu grupo.
*   `description` (string): Uma breve descrição do que a tarefa faz.
*   `command` (string ou função Lua):
    *   Se for uma `string`, será executada como um comando de shell.
    *   Se for uma `função Lua`, esta função será executada. Ela pode receber `params` (parâmetros da tarefa) e `deps` (outputs de tarefas das quais ela depende). A função deve retornar `true` para sucesso, `false` para falha, e opcionalmente uma mensagem e uma tabela de outputs.
*   `async` (booleano, opcional): Se `true`, a tarefa será executada assincronamente. Padrão é `false`.
*   `pre_exec` (função Lua, opcional): Uma função Lua a ser executada antes do `command` principal da tarefa.
*   `post_exec` (função Lua, opcional): Uma função Lua a ser executada após o `command` principal da tarefa.
*   `depends_on` (string ou tabela de strings, opcional): Nomes de tarefas que devem ser concluídas com sucesso antes que esta tarefa possa ser executada.
*   `retries` (número, opcional): O número de vezes que a tarefa será tentada novamente em caso de falha. Padrão é `0`.
*   `timeout` (string, opcional): Uma duração (ex: "10s", "1m") após a qual a tarefa será terminada se ainda estiver em execução.
*   `run_if` (string ou função Lua, opcional): A tarefa só será executada se esta condição for verdadeira. Pode ser um comando shell (código de saída 0 para sucesso) ou uma função Lua (retorna `true` para sucesso).
*   `abort_if` (string ou função Lua, opcional): Se esta condição for verdadeira, toda a execução do workflow será abortada. Pode ser um comando shell (código de saída 0 para sucesso) ou uma função Lua (retorna `true` para sucesso).
*   `next_if_fail` (string ou tabela de strings, opcional): Nomes de tarefas a serem executadas se esta tarefa falhar.

### Exemplo de Estrutura `TaskDefinitions`

```lua
TaskDefinitions = {
    my_first_group = {
        description = "Um grupo de tarefas de exemplo.",
        tasks = {
            my_first_task = {
                name = "my_first_task",
                description = "Uma tarefa simples que executa um comando shell.",
                command = "echo 'Hello from Sloth-Runner!'"
            },
            my_second_task = {
                name = "my_second_task",
                description = "Uma tarefa que depende da primeira e usa uma função Lua.",
                depends_on = "my_first_task",
                command = function(params, deps)
                    log.info("Executando a segunda tarefa.")
                    -- Você pode acessar outputs de tarefas anteriores via 'deps'
                    -- local output_from_first = deps.my_first_task.some_output
                    return true, "echo 'Second task completed!'"
                end
            }
        }
    }
}
```

## Parâmetros e Outputs

*   **Parâmetros (`params`):** Podem ser passados para as tarefas via linha de comando ou definidos na própria tarefa. A função `command` e as funções `run_if`/`abort_if` podem acessá-los.
*   **Outputs (`deps`):** As funções Lua de `command` podem retornar uma tabela de outputs. Tarefas que dependem desta tarefa podem acessar esses outputs através do argumento `deps`.

## Exportando Dados para a CLI

Além dos outputs de tarefas, o `sloth-runner` fornece uma função global `export()` que permite passar dados de dentro de um script diretamente para a saída da linha de comando.

### `export(tabela)`

*   **`tabela`**: Uma tabela Lua cujos pares de chave-valor serão exportados.

Quando você executa uma tarefa com a flag `--return`, os dados passados para a função `export()` serão mesclados com o output da tarefa final e impressos como um único objeto JSON. Se houver chaves duplicadas, o valor da função `export()` terá precedência.

Isso é útil para extrair informações importantes de qualquer ponto do seu script, não apenas do valor de retorno da última tarefa.

**Exemplo:**

```lua
command = function(params, deps)
  -- Lógica da tarefa...
  local some_data = {
    info = "Este é um dado importante",
    timestamp = os.time()
  }
  
  -- Exporta a tabela
  export(some_data)
  
  -- A tarefa pode continuar e retornar seu próprio output
  return true, "Tarefa concluída", { status = "ok" }
end
```

Executando com `--return` resultaria em uma saída JSON como:
```json
{
  "info": "Este é um dado importante",
  "timestamp": 1678886400,
  "status": "ok"
}
```

## Módulos Built-in

O Sloth-Runner expõe várias funcionalidades Go como módulos Lua, permitindo que suas tarefas interajam com o sistema e serviços externos. Além dos módulos básicos (`exec`, `fs`, `net`, `data`, `log`, `import`, `parallel`), o Sloth-Runner agora inclui módulos avançados para Git, Pulumi e Salt.

Esses módulos oferecem uma API fluente e intuitiva para automação complexa.

*   **`exec` module:** Para executar comandos de shell arbitrários.
*   **`fs` module:** Para operações de sistema de arquivos (leitura, escrita, etc.).
*   **`net` module:** Para fazer requisições HTTP e downloads.
*   **`data` module:** Para parsear e serializar JSON e YAML.
*   **`log` module:** Para registrar mensagens no console do Sloth-Runner.
*   **`import` function:** Para importar outros arquivos Lua e reutilizar tarefas.
*   **`parallel` function:** Para executar tarefas em paralelo.
*   **`git` module:** Para interagir com repositórios Git.
*   **`pulumi` module:** Para orquestrar stacks do Pulumi.
*   **`salt` module:** Para executar comandos SaltStack.

Para detalhes sobre cada módulo, consulte suas respectivas seções na documentação.

---
[English](../en/core-concepts.md) | [Português](./core-concepts.md) | [中文](../zh/core-concepts.md)