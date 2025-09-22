# Comandos da CLI

A interface de linha de comando (CLI) do `sloth-runner` é a principal forma de interagir com seus pipelines de tarefas. Ela fornece comandos para executar, listar, validar e gerenciar seus fluxos de trabalho.

---

## `sloth-runner run`

Executa tarefas definidas em um arquivo de configuração Lua. Este é o comando mais comum que você usará.

**Uso:**
```bash
sloth-runner run [flags]
```

**Flags:**

*   `-f, --file string`: **(Obrigatório)** Caminho para o arquivo de configuração de tarefas Lua.
*   `-g, --group string`: Executa tarefas apenas de um grupo de tarefas específico. Se não for fornecido, o `sloth-runner` executará tarefas de todos os grupos.
*   `-t, --tasks string`: Uma lista de tarefas específicas a serem executadas, separadas por vírgula (ex: `tarefa1,tarefa2`). Se não for fornecido, todas as tarefas no grupo especificado (ou em todos os grupos) serão consideradas.
*   `-v, --values string`: Caminho para um arquivo YAML com valores a serem passados para seus scripts Lua. Esses valores são acessíveis em Lua através da tabela global `values`.
*   `-d, --dry-run`: Simula a execução das tarefas. Ele imprimirá as tarefas que seriam executadas e em que ordem, mas não executará o `command` delas.
*   `--return`: Imprime a saída final das tarefas executadas como um objeto JSON no stdout. Isso inclui tanto o valor de retorno da última tarefa quanto quaisquer dados passados para a função global `export()`.
*   `-y, --yes`: Ignora o prompt de seleção interativa de tarefas quando nenhuma tarefa específica é fornecida com `-t`.

**Exemplos:**

*   Executar todas as tarefas em um grupo específico:
    ```bash
    sloth-runner run -f examples/basic_pipeline.lua -g meu_grupo
    ```
*   Executar uma única tarefa específica:
    ```bash

    sloth-runner run -f examples/basic_pipeline.lua -g meu_grupo -t minha_tarefa
    ```
*   Executar múltiplas tarefas e obter a saída combinada como JSON:
    ```bash
    sloth-runner run -f examples/export_example.lua -t export-data-task --return
    ```

---

## `sloth-runner list`

Lista todos os grupos de tarefas e tarefas disponíveis definidos em um arquivo de configuração Lua, juntamente com suas descrições e dependências.

**Uso:**
```bash
sloth-runner list [flags]
```

**Flags:**

*   `-f, --file string`: **(Obrigatório)** Caminho para o arquivo de configuração de tarefas Lua.
*   `-v, --values string`: Caminho para um arquivo de valores YAML, caso suas definições de tarefa dependam dele.

---

## `sloth-runner new`

Gera um novo arquivo de definição de tarefas Lua a partir de um modelo.

**Uso:**
```bash
sloth-runner new <nome-do-grupo> [flags]
```

**Argumentos:**

*   `<nome-do-grupo>`: O nome do grupo de tarefas principal a ser criado no arquivo.

**Flags:**

*   `-t, --template string`: O modelo a ser usado. O padrão é `simple`. Execute `sloth-runner template list` para ver todas as opções disponíveis.
*   `-o, --output string`: O caminho para o arquivo de saída. Se não for fornecido, o conteúdo gerado será impresso no stdout.

**Exemplo:**
```bash
sloth-runner new meu-pipeline-python -t python -o meu_pipeline.lua
```

---

## `sloth-runner validate`

Valida a sintaxe e a estrutura básica de um arquivo de tarefas Lua sem executar nenhuma tarefa.

**Uso:**
```bash
sloth-runner validate [flags]
```

**Flags:**

*   `-f, --file string`: **(Obrigatório)** Caminho para o arquivo de configuração de tarefas Lua a ser validado.
*   `-v, --values string`: Caminho para um arquivo de valores YAML, se necessário para a validação.

---

## `sloth-runner test`

Executa um arquivo de teste baseado em Lua para um fluxo de trabalho. (Este é um recurso avançado).

**Uso:**
```bash
sloth-runner test [flags]
```

**Flags:**

*   `-w, --workflow string`: **(Obrigatório)** Caminho para o arquivo de fluxo de trabalho Lua a ser testado.
*   `-f, --file string`: **(Obrigatório)** Caminho para o arquivo de teste Lua.

---

## `sloth-runner template list`

Lista todos os modelos disponíveis que podem ser usados com o comando `sloth-runner new`.

**Uso:**
```bash
sloth-runner template list
```

---

## `sloth-runner version`

Imprime a versão atual do aplicativo `sloth-runner`.

**Uso:**
```bash
sloth-runner version
```
