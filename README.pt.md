[English](./README.md) | [Português](./README.pt.md) | [中文](./README.zh.md)

# 🦥 Sloth Runner 🚀

Uma aplicação de execução de tarefas flexível e extensível, escrita em Go e impulsionada por scripts Lua. O `sloth-runner` permite que você defina fluxos de trabalho complexos, gerencie dependências de tarefas e integre com sistemas externos, tudo através de scripts Lua simples.

[![Go CI](https://github.com/chalkan3/sloth-runner/actions/workflows/go.yml/badge.svg)](https://github.com/chalkan3/sloth-runner/actions/workflows/go.yml)

---

## ✨ Funcionalidades

*   **📜 Scripts em Lua:** Defina tarefas e fluxos de trabalho usando o poder e a flexibilidade dos scripts Lua.
*   **🔗 Gerenciamento de Dependências:** Especifique dependências entre tarefas para garantir a execução ordenada de pipelines complexos.
*   **⚡ Execução Assíncrona de Tarefas:** Execute tarefas concorrentemente para melhor desempenho.
*   **🪝 Hooks de Pré/Pós-Execução:** Defina funções Lua customizadas para serem executadas antes e depois dos comandos das tarefas.
*   **⚙️ API Lua Rica:** Acesse funcionalidades do sistema diretamente das suas tarefas Lua:
    *   **Módulo `exec`:** Execute comandos de shell.
    *   **Módulo `fs`:** Realize operações de sistema de arquivos (ler, escrever, anexar, verificar existência, criar diretório, remover, remover recursivamente, listar).
    *   **Módulo `net`:** Faça requisições HTTP (GET, POST) e baixe arquivos.
    *   **Módulo `data`:** Analise e serialize dados em formato JSON e YAML.
    *   **Módulo `log`:** Registre mensagens com diferentes níveis de severidade (info, warn, error, debug).
    *   **Módulo `salt`:** Execute comandos do SaltStack (`salt`, `salt-call`) diretamente.
*   **📝 Integração com `values.yaml`:** Passe valores de configuração para suas tarefas Lua através de um arquivo `values.yaml`, de forma semelhante ao Helm.
*   **💻 Interface de Linha de Comando (CLI):**
    *   `run`: Execute tarefas de um arquivo de configuração Lua.
    *   `list`: Liste todos os grupos de tarefas e tarefas disponíveis com suas descrições e dependências.


## 📚 Documentação Completa

Para obter a documentação mais detalhada, guias de uso e exemplos avançados, visite nossa [Documentação Completa](./docs/pt/index.md).

---

## 🚀 Começando

### Instalação

Para instalar o `sloth-runner` no seu sistema, você pode usar o script `install.sh` fornecido. Este script detecta automaticamente seu sistema operacional e arquitetura, baixa a versão mais recente do GitHub e coloca o executável `sloth-runner` em `/usr/local/bin`.

```bash
bash <(curl -sL https://raw.githubusercontent.com/chalkan3/sloth-runner/master/install.sh)
```

**Nota:** O script `install.sh` requer privilégios de `sudo` para mover o executável para `/usr/local/bin`.

### Uso Básico

Para executar um arquivo de tarefas Lua:

```bash
sloth-runner run -f examples/basic_pipeline.lua
```

Para listar as tarefas em um arquivo:

```bash
sloth-runner list -f examples/basic_pipeline.lua
```

---

## 📜 Definindo Tarefas em Lua

As tarefas são definidas em arquivos Lua, tipicamente dentro de uma tabela `TaskDefinitions`. Cada tarefa pode ter um `name`, `description`, `command` (seja uma string para um comando de shell ou uma função Lua), `async` (booleano), `pre_exec` (hook de função Lua), `post_exec` (hook de função Lua) e `depends_on` (uma string ou uma tabela de strings).

Exemplo (`examples/basic_pipeline.lua`):

```lua
-- Importa tarefas reutilizáveis de outro arquivo. O caminho é relativo.
local docker_tasks = import("examples/shared/docker.lua")

TaskDefinitions = {
    full_pipeline_demo = {
        description = "Um pipeline abrangente demonstrando várias funcionalidades.",
        tasks = {
            -- Tarefa 1: Busca dados, executa de forma assíncrona.
            fetch_data = {
                name = "fetch_data",
                description = "Busca dados brutos de uma API.",
                async = true,
                command = function(params)
                    log.info("Buscando dados...")
                    -- Simula uma chamada de API
                    return true, "echo 'Buscou dados brutos'", { raw_data = "dados_da_api" }
                end,
            },

            -- Tarefa 2: Uma tarefa instável que tenta novamente em caso de falha.
            flaky_task = {
                name = "flaky_task",
                description = "Esta tarefa falha intermitentemente e tentará novamente.",
                retries = 3,
                command = function()
                    if math.random() > 0.5 then
                        log.info("Tarefa instável bem-sucedida.")
                        return true, "echo 'Sucesso!'"
                    else
                        log.error("Tarefa instável falhou, tentará novamente...")
                        return false, "Falha aleatória"
                    end
                end,
            },

            -- Tarefa 3: Processa dados, depende da conclusão bem-sucedida de fetch_data e flaky_task.
            process_data = {
                name = "process_data",
                description = "Processa os dados buscados.",
                depends_on = { "fetch_data", "flaky_task" },
                command = function(params, deps)
                    local raw_data = deps.fetch_data.raw_data
                    log.info("Processando dados: " .. raw_data)
                    return true, "echo 'Dados processados'", { processed_data = "processado_" .. raw_data }
                end,
            },

            -- Tarefa 4: Uma tarefa de longa duração com um tempo limite.
            long_running_task = {
                name = "long_running_task",
                description = "Uma tarefa que será encerrada se demorar muito.",
                timeout = "5s",
                command = "echo 'Iniciando tarefa longa...'; sleep 10; echo 'Isso não será impresso.';",
            },

            -- Tarefa 5: Uma tarefa de limpeza que é executada se a long_running_task falhar.
            cleanup_on_fail = {
                name = "cleanup_on_fail",
                description = "Executa apenas se a tarefa de longa duração falhar.",
                next_if_fail = "long_running_task",
                command = "echo 'Tarefa de limpeza executada devido a falha anterior.'",
            },

            -- Tarefa 6: Usa uma tarefa reutilizável do módulo importado docker.lua.
            build_image = {
                uses = docker_tasks.build,
                description = "Constrói a imagem Docker da aplicação.",
                params = {
                    image_name = "meu-app-incrivel",
                    tag = "v1.2.3",
                    context = "./app_context"
                }
            },

            -- Tarefa 7: Uma tarefa condicional que só é executada se um arquivo existir.
            conditional_deploy = {
                name = "conditional_deploy",
                description = "Implanta o aplicativo apenas se o artefato de construção existir.",
                depends_on = "build_image",
                run_if = "test -f ./app_context/artifact.txt", -- Condição de comando de shell
                command = "echo 'Implantando aplicação...'",
            },

            -- Tarefa 8: Esta tarefa abortará todo o fluxo de trabalho se uma condição for atendida.
            gatekeeper_check = {
                name = "gatekeeper_check",
                description = "Aborta o fluxo de trabalho se uma condição crítica não for atendida.",
                abort_if = function(params, deps)
                    -- Condição de função Lua
                    log.warn("Verificando condição do gatekeeper...")
                    if params.force_proceed ~= "true" then
                        log.error("Verificação do gatekeeper falhou. Abortando fluxo de trabalho.")
                        return true -- Abortar
                    end
                    return false -- Não abortar
                end,
                command = "echo 'Este comando não será executado se for abortado.'"
            }
        }
    }
}
```

---

## Funcionalidades Avançadas

O `sloth-runner` oferece várias funcionalidades avançadas para um controle refinado sobre a execução das tarefas.

### Tentativas e Tempos Limite de Tarefas

Você pode tornar seus fluxos de trabalho mais robustos especificando tentativas para tarefas instáveis e tempos limite para as de longa duração.

*   `retries`: O número de vezes para tentar novamente uma tarefa se ela falhar.
*   `timeout`: Uma string de duração (ex: "10s", "1m") após a qual uma tarefa será encerrada.

<details>
<summary>Exemplo (`examples/retries_and_timeout.lua`):</summary>

```lua
TaskDefinitions = {
    robust_workflow = {
        description = "Um fluxo de trabalho para demonstrar tentativas e tempos limite",
        tasks = {
            {
                name = "flaky_task",
                description = "Esta tarefa falha 50% das vezes",
                retries = 3,
                command = function()
                    if math.random() < 0.5 then
                        log.error("Simulando uma falha aleatória!")
                        return false, "Ocorreu uma falha aleatória"
                    end
                    return true, "echo 'Tarefa instável bem-sucedida!'", { result = "sucesso" }
                end
            },
            {
                name = "long_running_task",
                description = "Esta tarefa simula um processo longo que excederá o tempo limite",
                timeout = "2s",
                command = "sleep 5 && echo 'Isso não deve ser impresso'"
            }
        }
    }
}
```
</details>

### Execução Condicional: `run_if` e `abort_if`

Você pode controlar a execução de tarefas com base em condições usando `run_if` e `abort_if`. Estas podem ser um comando de shell ou uma função Lua.

*   `run_if`: A tarefa só será executada se a condição for atendida.
*   `abort_if`: A execução inteira será abortada se a condição for atendida.

#### Usando Comandos de Shell

O comando de shell é executado, e seu código de saída determina o resultado. Um código de saída `0` significa que a condição foi atendida (sucesso).

<details>
<summary>Exemplo (`examples/conditional_execution.lua`):</summary>

```lua
TaskDefinitions = {
    conditional_workflow = {
        description = "Um fluxo de trabalho para demonstrar execução condicional com run_if e abort_if.",
        tasks = {
            {
                name = "check_condition_for_run",
                description = "Esta tarefa cria um arquivo que a próxima tarefa verifica.",
                command = "touch /tmp/sloth_runner_run_condition"
            },
            {
                name = "conditional_task",
                description = "Esta tarefa só é executada se o arquivo de condição existir.",
                depends_on = "check_condition_for_run",
                run_if = "test -f /tmp/sloth_runner_run_condition",
                command = "echo 'A tarefa condicional está sendo executada porque a condição foi atendida.'"
            },
            {
                name = "check_abort_condition",
                description = "Esta tarefa será abortada se um arquivo específico existir.",
                abort_if = "test -f /tmp/sloth_runner_abort_condition",
                command = "echo 'Isso não será executado se a condição de abortar for atendida.'"
            }
        }
    }
}
```
</details>

#### Usando Funções Lua

Para uma lógica mais complexa, você pode usar uma função Lua. A função recebe os `params` da tarefa e os `deps` (saídas das dependências). Ela deve retornar `true` para que a condição seja atendida.

<details>
<summary>Exemplo (`examples/conditional_functions.lua`):</summary>

```lua
TaskDefinitions = {
    conditional_functions_workflow = {
        description = "Um fluxo de trabalho para demonstrar execução condicional com funções Lua.",
        tasks = {
            {
                name = "setup_task",
                description = "Esta tarefa fornece a saída para a tarefa condicional.",
                command = function()
                    return true, "Configuração completa", { should_run = true }
                end
            },
            {
                name = "conditional_task_with_function",
                description = "Esta tarefa só é executada se a função run_if retornar true.",
                depends_on = "setup_task",
                run_if = function(params, deps)
                    log.info("Verificando condição run_if para conditional_task_with_function...")
                    if deps.setup_task and deps.setup_task.should_run == true then
                        log.info("Condição atendida, a tarefa será executada.")
                        return true
                    end
                    log.info("Condição não atendida, a tarefa será pulada.")
                    return false
                end,
                command = "echo 'A tarefa condicional está sendo executada porque a função retornou true.'"
            },
            {
                name = "abort_task_with_function",
                description = "Esta tarefa abortará a execução se a função abort_if retornar true.",
                params = {
                    abort_execution = "true"
                },
                abort_if = function(params, deps)
                    log.info("Verificando condição abort_if para abort_task_with_function...")
                    if params.abort_execution == "true" then
                        log.info("Condição de abortar atendida, a execução será interrompida.")
                        return true
                    end
                    log.info("Condição de abortar não atendida.")
                    return false
                end,
                command = "echo 'Isso não deve ser executado.'"
            }
        }
    }
}
```
</details>

### Módulos de Tarefas Reutilizáveis com `import`

Você pode criar bibliotecas de tarefas reutilizáveis e importá-las para o seu arquivo de fluxo de trabalho principal. Isso é útil para compartilhar tarefas comuns (como construir imagens Docker, implantar aplicações, etc.) entre múltiplos projetos.

A função global `import()` carrega outro arquivo Lua e retorna o valor que ele retorna. O caminho é resolvido relativamente ao arquivo que chama `import`.

**Como funciona:**
1.  Crie um módulo (ex: `shared/docker.lua`) que define uma tabela de tarefas e a retorna.
2.  No seu arquivo principal, chame `import("shared/docker.lua")` para carregar o módulo.
3.  Referencie as tarefas importadas na sua tabela `TaskDefinitions` principal usando o campo `uses`. O `sloth-runner` irá mesclar automaticamente a tarefa importada com quaisquer sobreposições locais que você fornecer (como `description` ou `params`).

<details>
<summary>Exemplo de Módulo (`examples/shared/docker.lua`):</summary>

```lua
-- examples/shared/docker.lua
-- Um módulo reutilizável para tarefas Docker.

local TaskDefinitions = {
    build = {
        name = "build",
        description = "Constrói uma imagem Docker",
        params = {
            tag = "latest",
            dockerfile = "Dockerfile",
            context = "."
        },
        command = function(params)
            local image_name = params.image_name or "minha-imagem-padrao"
            -- ... lógica do comando de construção ...
            local cmd = string.format("docker build -t %s:%s -f %s %s", image_name, params.tag, params.dockerfile, params.context)
            return true, cmd
        end
    },
    push = {
        name = "push",
        description = "Envia uma imagem Docker para um registro",
        -- ... lógica da tarefa de envio ...
    }
}

return TaskDefinitions
```
</details>

<details>
<summary>Exemplo de Uso (`examples/reusable_tasks.lua`):</summary>

```lua
-- examples/reusable_tasks.lua

-- Importa as tarefas Docker reutilizáveis.
local docker_tasks = import("shared/docker.lua")

TaskDefinitions = {
    app_deployment = {
        description = "Um fluxo de trabalho que usa um módulo Docker reutilizável.",
        tasks = {
            -- Usa a tarefa 'build' do módulo e sobrepõe seus parâmetros.
            build = {
                uses = docker_tasks.build,
                description = "Constrói a imagem Docker da aplicação principal",
                params = {
                    image_name = "meu-app",
                    tag = "v1.0.0",
                    context = "./app"
                }
            },
            
            -- Uma tarefa regular que depende da tarefa 'build' importada.
            deploy = {
                name = "deploy",
                description = "Implanta a aplicação",
                depends_on = "build",
                command = "echo 'Implantando...'"
            }
        }
    }
}
```
</details>

---

## 💻 Comandos da CLI

O `sloth-runner` fornece uma interface de linha de comando simples e poderosa.

### `sloth-runner run`

Executa tarefas definidas em um arquivo de modelo Lua.

**Flags:**

*   `-f, --file string`: Caminho para o arquivo de configuração de tarefas Lua.
*   `-t, --tasks string`: Lista de tarefas específicas para executar, separadas por vírgula.
*   `-g, --group string`: Executa tarefas apenas de um grupo de tarefas específico.
*   `-v, --values string`: Caminho para um arquivo YAML com valores a serem passados para as tarefas Lua.
*   `-d, --dry-run`: Simula a execução de tarefas sem realmente executá-las.

### `sloth-runner list`

Lista todos os grupos de tarefas e tarefas disponíveis definidos em um arquivo de modelo Lua.

**Flags:**

*   `-f, --file string`: Caminho para o arquivo de configuração de tarefas Lua.
*   `-v, --values string`: Caminho para um arquivo YAML com valores.

---

## ⚙️ API Lua

O `sloth-runner` expõe várias funcionalidades do Go como módulos Lua, permitindo que suas tarefas interajam com o sistema e serviços externos.

*   **Módulo `exec`:** Execute comandos de shell.
*   **Módulo `fs`:** Realize operações de sistema de arquivos.
*   **Módulo `net`:** Faça requisições HTTP e baixe arquivos.
*   **Módulo `data`:** Analise e serialize dados em formato JSON e YAML.
*   **Módulo `log`:** Registre mensagens com diferentes níveis de severidade.
*   **Módulo `salt`:** Execute comandos do SaltStack.

Para uso detalhado da API, por favor, consulte os exemplos no diretório `/examples`.
