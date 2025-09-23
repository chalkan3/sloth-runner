# Comandos da CLI

A interface de linha de comando (CLI) do `sloth-runner` é a principal forma de interagir com seus pipelines de tarefas. Ela fornece comandos para executar, listar, validar e gerenciar seus fluxos de trabalho.

---

## `sloth-runner run`

Executa tarefas definidas em um arquivo de configuração Lua.

**Uso:** `sloth-runner run [flags]`

**Descrição:**
O comando `run` executa tarefas definidas em um arquivo de modelo Lua.
Você pode especificar o arquivo, variáveis de ambiente e direcionar tarefas ou grupos específicos.

**Flags:**

*   `-f, --file string`: Caminho para o arquivo de configuração de tarefas Lua (padrão: "examples/basic_pipeline.lua")
*   `-e, --env string`: Ambiente para as tarefas (ex: Development, Production) (padrão: "Development")
*   `-p, --prod`: Definir como verdadeiro para ambiente de produção (padrão: false)
*   `--shards string`: Lista de números de shard separados por vírgula (ex: 1,2,3) (padrão: "1,2,3")
*   `-t, --tasks string`: Lista de tarefas específicas a serem executadas, separadas por vírgula (ex: tarefa1,tarefa2)
*   `-g, --group string`: Executa tarefas apenas de um grupo de tarefas específico
*   `-v, --values string`: Caminho para um arquivo YAML com valores a serem passados para as tarefas Lua
*   `-d, --dry-run`: Simula a execução das tarefas sem realmente executá-las (padrão: false)
*   `--return`: Retorna a saída das tarefas de destino como JSON (padrão: false)
*   `-y, --yes`: Ignora a seleção interativa de tarefas e executa todas as tarefas (padrão: false)

### `sloth-runner list`

Lista todos os grupos de tarefas e tarefas disponíveis.

**Uso:** `sloth-runner list [flags]`

**Descrição:**
O comando `list` exibe todos os grupos de tarefas e suas respectivas tarefas, juntamente com suas descrições e dependências.

**Flags:**

*   `-f, --file string`: Caminho para o arquivo de configuração de tarefas Lua (padrão: "examples/basic_pipeline.lua")
*   `-e, --env string`: Ambiente para as tarefas (ex: Development, Production) (padrão: "Development")
*   `-p, --prod`: Definir como verdadeiro para ambiente de produção (padrão: false)
*   `--shards string`: Lista de números de shard separados por vírgula (ex: 1,2,3) (padrão: "1,2,3")
*   `-v, --values string`: Caminho para um arquivo YAML com valores a serem passados para as tarefas Lua

### `sloth-runner validate`

Valida a sintaxe e a estrutura de um arquivo de tarefas Lua.

**Uso:** `sloth-runner validate [flags]`

**Descrição:**
O comando `validate` verifica um arquivo de tarefas Lua quanto a erros de sintaxe e garante que a tabela `TaskDefinitions` esteja corretamente estruturada.

**Flags:**

*   `-f, --file string`: Caminho para o arquivo de configuração de tarefas Lua (padrão: "examples/basic_pipeline.lua")
*   `-e, --env string`: Ambiente para as tarefas (ex: Development, Production) (padrão: "Development")
*   `-p, --prod`: Definir como verdadeiro para ambiente de produção (padrão: false)
*   `--shards string`: Lista de números de shard separados por vírgula (ex: 1,2,3) (padrão: "1,2,3")
*   `-v, --values string`: Caminho para um arquivo YAML com valores a serem passados para as tarefas Lua

### `sloth-runner test`

Executa um arquivo de teste Lua para um fluxo de trabalho de tarefas.

**Uso:** `sloth-runner test -w <workflow-file> -f <test-file>`

**Descrição:**
O comando `test` executa um arquivo de teste Lua especificado contra um fluxo de trabalho.
Dentro do arquivo de teste, você pode usar os módulos 'test' e 'assert' para validar os comportamentos das tarefas.

**Flags:**

*   `-f, --file string`: Caminho para o arquivo de teste Lua (obrigatório)
*   `-w, --workflow string`: Caminho para o arquivo de fluxo de trabalho Lua a ser testado (obrigatório)

### `sloth-runner repl`

Inicia uma sessão REPL interativa.

**Uso:** `sloth-runner repl [flags]`

**Descrição:**
O comando `repl` inicia um Loop de Leitura-Avaliação-Impressão interativo que permite
executar código Lua e interagir com todos os módulos sloth-runner integrados.
Você pode opcionalmente carregar um arquivo de fluxo de trabalho para ter seu contexto disponível.

**Flags:**

*   `-f, --file string`: Caminho para um arquivo de fluxo de trabalho Lua a ser carregado na sessão REPL

### `sloth-runner version`

Imprime o número da versão do sloth-runner.

**Uso:** `sloth-runner version`

**Descrição:**
Todo software tem versões. Esta é a do sloth-runner.

### `sloth-runner scheduler`

Gerencia o agendador de tarefas do `sloth-runner`, permitindo habilitar, desabilitar, listar e excluir tarefas agendadas.

Para informações detalhadas sobre os comandos e configuração do agendador, consulte a [documentação do Agendador de Tarefas](scheduler.md).

**Subcomandos:**

*   `sloth-runner scheduler enable`: Inicia o agendador como um processo em segundo plano.
*   `sloth-runner scheduler disable`: Para o processo do agendador em execução.
*   `sloth-runner scheduler list`: Lista todas as tarefas agendadas configuradas.
*   `sloth-runner scheduler delete <task_name>`: Exclui uma tarefa agendada específica.

---

### `sloth-runner template list`

Lista todos os modelos disponíveis.

**Uso:** `sloth-runner template list`

**Descrição:**
Exibe uma tabela de todos os modelos disponíveis que podem ser usados com o comando 'new'.

---

## 📄 Modelos

`sloth-runner` oferece vários modelos para criar rapidamente novos arquivos de definição de tarefas.

| Nome do Modelo       | Descrição                                                                    |
| :------------------- | :----------------------------------------------------------------------------- |
| `simple`             | Gera um único grupo com uma tarefa 'hello world'. Ideal para começar.          |
| `python`             | Cria um pipeline para configurar um ambiente Python, instalar dependências e executar um script. |
| `parallel`           | Demonstra como executar várias tarefas simultaneamente.                        |
| `python-pulumi`      | Pipeline para implantar infraestrutura Pulumi gerenciada com Python.           |
| `python-pulumi-salt` | Provisiona infraestrutura com Pulumi e a configura usando SaltStack.           |
| `git-python-pulumi`  | Pipeline CI/CD: Clona um repositório, configura o ambiente e implanta com Pulumi. |
| `dummy`              | Gera uma tarefa fictícia que não faz nada.                                     |

---

### `sloth-runner new <group-name>`

Gera um novo arquivo de definição de tarefas a partir de um modelo.

**Uso:** `sloth-runner new <group-name> [flags]`

**Descrição:**
O comando `new` cria um arquivo de definição de tarefas Lua boilerplate.
Você pode escolher entre diferentes modelos e especificar um arquivo de saída.
Execute `sloth-runner template list` para ver as opções.

**Argumentos:**

*   `<group-name>`: O nome do grupo de tarefas a ser gerado.

**Flags:**

*   `-o, --output string`: Caminho do arquivo de saída (padrão: stdout)
*   `-t, --template string`: Modelo a ser usado. Veja `template list` para opções. (padrão: "simple")
*   `--set key=value`: Passa pares chave-valor para o modelo para geração dinâmica de conteúdo.

### `sloth-runner check dependencies`

Verifica as ferramentas CLI externas necessárias.

**Uso:** `sloth-runner check dependencies`

**Descrição:**
Verifica se todas as ferramentas de linha de comando externas usadas pelos vários módulos (por exemplo, docker, aws, doctl) estão instaladas e disponíveis no PATH do sistema.
