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
*   `artifacts` (string ou tabela de strings, opcional): Um padrão de arquivo (glob) ou uma lista de padrões que especificam quais arquivos do `workdir` da tarefa devem ser salvos como artefatos após a execução bem-sucedida.
*   `consumes` (string ou tabela de strings, opcional): O nome de um artefato (ou uma lista de nomes) de uma tarefa anterior que deve ser copiado para o `workdir` desta tarefa antes de sua execução.

## Gerenciamento de Artefatos

O Sloth-Runner permite que as tarefas compartilhem arquivos entre si através de um mecanismo de artefatos. Uma tarefa pode "produzir" um ou mais arquivos como artefatos, e tarefas subsequentes podem "consumir" esses artefatos.

Isso é útil para pipelines de CI/CD, onde uma etapa de compilação pode gerar um binário (artefato), que é então usado por uma etapa de teste ou de implantação.

### Como Funciona

1.  **Produzindo Artefatos:** Adicione a chave `artifacts` à sua definição de tarefa. O valor pode ser um único padrão de arquivo (ex: `"report.txt"`) ou uma lista (ex: `{"*.log", "app.bin"}`). Após a tarefa ser executada com sucesso, o runner procurará por arquivos no `workdir` da tarefa que correspondam a esses padrões e os copiará para um armazenamento de artefatos compartilhado para a pipeline.

2.  **Consumindo Artefatos:** Adicione a chave `consumes` à definição de outra tarefa (que normalmente `depends_on` da tarefa produtora). O valor deve ser o nome do arquivo do artefato que você deseja usar (ex: `"report.txt"`). Antes que esta tarefa seja executada, o runner copiará o artefato nomeado do armazenamento compartilhado para o `workdir` desta tarefa, tornando-o disponível para o `command`.

### Exemplo de Artefatos

```lua
TaskDefinitions = {
  ["ci-pipeline"] = {
    description = "Demonstra o uso de artefatos.",
    create_workdir_before_run = true,
    tasks = {
      {
        name = "build",
        description = "Cria um binário e o declara como um artefato.",
        command = "echo 'binary_content' > app.bin",
        artifacts = {"app.bin"}
      },
      {
        name = "test",
        description = "Consome o binário para executar testes.",
        depends_on = "build",
        consumes = {"app.bin"},
        command = function(params)
          -- Neste ponto, 'app.bin' existe no workdir desta tarefa
          local content, err = fs.read(params.workdir .. "/app.bin")
          if content == "binary_content" then
            log.info("Artefato consumido com sucesso!")
            return true
          else
            return false, "Conteúdo do artefato incorreto!"
          end
        end
      }
    }
  }
}
```

📜 Defining Tasks in Lua
Tasks are defined in Lua files, typically within a `TaskDefinitions` table. Each task can have a name, description, and a `command` (either a string for a shell command or a Lua function). For modular pipelines, tasks can declare dependencies using `depends_on` and receive outputs from previous tasks via the `inputs` table.

Here's an example using our GCP Hub-and-Spoke orchestration pipeline, demonstrating how tasks are chained and how data flows between them:

```lua
-- examples/gcp_pulumi_orchestration.lua
--
-- This pipeline demonstrates a complete, modular orchestration for deploying a GCP Hub and Spoke network.

TaskDefinitions = {
  gcp_deployment = {
    description = "Orchestrates the deployment of a GCP Hub and Spoke architecture.",
    tasks = {
      {
        name = "setup_workspace",
        command = function()
          log.info("Cleaning up previous run artifacts...")
          fs.rm_r(values.paths.base_workdir)
          fs.mkdir(values.paths.base_workdir)
          return true, "Workspace cleaned and created."
        end
      },
      {
        name = "clone_hub_repo",
        depends_on = "setup_workspace",
        command = function()
          log.info("Cloning Hub repository...")
          local hub_repo = git.clone(values.repos.hub.url, values.repos.hub.path)
          log.info("Hub repo cloned to: " .. hub_repo.path)
          -- Return the cloned repository object to be used by dependent tasks
          return true, "Hub repo cloned.", { repo = hub_repo }
        end
      },
      {
        name = "clone_spoke_repo",
        depends_on = "setup_workspace",
        command = function()
          log.info("Cloning Spoke repository...")
          local spoke_repo = git.clone(values.repos.spoke.url, values.repos.spoke.path)
          log.info("Spoke repo cloned to: " .. spoke_repo.path)
          -- Return the cloned repository object
          return true, "Spoke repo cloned.", { repo = spoke_repo }
        end
      },
      {
        name = "setup_spoke_venv",
        depends_on = "clone_spoke_repo", -- Depends on the spoke repo being cloned
        command = function(inputs) -- Receives inputs from dependent tasks
          log.info("Setting up Python venv for the host manager...")
          local spoke_repo = inputs.clone_spoke_repo.repo -- Access the repo from the 'clone_spoke_repo' task's output
          local spoke_venv = python.venv(values.paths.spoke_venv)
            :create()
            :pip("install setuptools")
            :pip("install -r " .. spoke_repo.path .. "/requirements.txt")
          log.info("Python venv for spoke is ready at: " .. values.paths.spoke_venv)
          -- Return the venv object
          return true, "Spoke venv created.", { venv = spoke_venv, repo = spoke_repo }
        end
      },
      {
        name = "deploy_hub_stack",
        depends_on = "clone_hub_repo", -- Depends on the hub repo being cloned
        command = function(inputs) -- Receives inputs from dependent tasks
          log.info("Deploying GCP Hub Network...")
          local hub_repo = inputs.clone_hub_repo.repo -- Access the repo from the 'clone_hub_repo' task's output
          local hub_stack = pulumi.stack(values.pulumi.hub.stack_name, {
            workdir = hub_repo.path,
            login = values.pulumi.login_url
          })
          hub_stack:select():config_map(values.pulumi.hub.config)
          local hub_result = hub_stack:up({ yes = true })
          if not hub_result.success then
            log.error("Hub stack deployment failed: " .. hub_result.stdout)
            return false, "Hub stack deployment failed."
          end
          log.info("Hub stack deployed successfully.")
          local hub_outputs = hub_stack:outputs()
          -- Return the outputs of the hub stack
          return true, "Hub stack deployed.", { outputs = hub_outputs }
        end
      },
      {
        name = "deploy_spoke_stack",
        depends_on = { "setup_spoke_venv", "deploy_hub_stack" }, -- Depends on venv setup and hub deployment
        command = function(inputs) -- Receives inputs from multiple dependent tasks
          log.info("Deploying GCP Spoke Host...")
          local spoke_repo = inputs.setup_spoke_venv.repo -- Access repo from venv setup task
          local spoke_venv = inputs.setup_spoke_venv.venv -- Access venv from venv setup task
          local hub_outputs = inputs.deploy_hub_stack.outputs -- Access hub outputs from hub deployment task

          local spoke_stack = pulumi.stack(values.pulumi.spoke.stack_name, {
            workdir = spoke_repo.path,
            login = values.pulumi.login_url,
            venv = spoke_venv
          })

          local spoke_config = values.pulumi.spoke.config
          spoke_config.hub_network_self_link = hub_outputs.network_self_link -- Use hub output in spoke config

          spoke_stack:select():config_map(spoke_config)
          local spoke_result = spoke_stack:up({ yes = true })
          if not spoke_result.success then
            log.error("Spoke stack deployment failed: " .. spoke_result.stdout)
            return false, "Spoke stack deployment failed."
          end
          log.info("Spoke stack deployed successfully.")
          local spoke_outputs = spoke_stack:outputs()
          return true, "Spoke stack deployed.", { outputs = spoke_outputs }
        end
      },
      {
          name = "final_summary",
          depends_on = "deploy_spoke_stack", -- Depends on the final deployment task
          command = function(inputs)
              log.info("GCP Hub and Spoke orchestration completed successfully!")
              -- You can access outputs from dependencies like this:
              -- local hub_outputs = inputs.deploy_hub_stack.outputs
              -- local spoke_outputs = inputs.deploy_spoke_stack.outputs
              return true, "Orchestration successful."
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

[Voltar ao Índice](./index.md)
