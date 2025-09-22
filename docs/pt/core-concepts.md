# Conceitos Essenciais

Este documento explica os conceitos fundamentais do `sloth-runner`, ajudando você a entender como definir e orquestrar fluxos de trabalho complexos.

---

## A Tabela `TaskDefinitions`

O ponto de entrada para qualquer fluxo de trabalho do `sloth-runner` é um arquivo Lua que retorna uma tabela global chamada `TaskDefinitions`. Esta tabela é um dicionário onde cada chave é o nome de um **Grupo de Tarefas**.

```lua
-- meu_pipeline.lua
TaskDefinitions = {
  -- Grupos de Tarefas são definidos aqui
}
```

---

## Grupos de Tarefas

Um Grupo de Tarefas é uma coleção de tarefas relacionadas. Ele também pode definir propriedades que afetam todas as tarefas dentro dele.

**Propriedades do Grupo:**

*   `description` (string): Uma descrição do que o grupo faz.
*   `tasks` (tabela): Uma lista de tabelas de tarefas individuais.
*   `create_workdir_before_run` (booleano): Se `true`, um diretório de trabalho temporário é criado para o grupo antes que qualquer tarefa seja executada. Este diretório é passado para cada tarefa.
*   `clean_workdir_after_run` (função): Uma função Lua que decide se o diretório de trabalho temporário deve ser excluído após a conclusão do grupo. Ela recebe o resultado final do grupo (`{success = true/false, ...}`). Retornar `true` exclui o diretório.

**Exemplo:**
```lua
TaskDefinitions = {
  meu_grupo = {
    description = "Um grupo que gerencia seu próprio diretório temporário.",
    create_workdir_before_run = true,
    clean_workdir_after_run = function(result)
      if not result.success then
        log.warn("O grupo falhou. O diretório de trabalho será mantido para depuração.")
      end
      return result.success -- Limpa apenas se tudo foi bem-sucedido
    end,
    tasks = {
      -- Tarefas aqui
    }
  }
}
```

---

## Tarefas Individuais

Uma tarefa é uma única unidade de trabalho. É definida como uma tabela com várias propriedades disponíveis para controlar seu comportamento.

### Propriedades Básicas

*   `name` (string): O nome único da tarefa dentro de seu grupo.
*   `description` (string): Uma breve descrição do que a tarefa faz.
*   `command` (string ou função): A ação principal da tarefa.
    *   **Como string:** É executada como um comando de shell.
    *   **Como função:** A função Lua é executada. Ela recebe dois argumentos: `params` (uma tabela com seus parâmetros) e `deps` (uma tabela contendo os outputs de suas dependências). A função deve retornar:
        1.  `booleano`: `true` para sucesso, `false` para falha.
        2.  `string`: Uma mensagem descrevendo o resultado.
        3.  `tabela` (opcional): Uma tabela de outputs da qual outras tarefas podem depender.

### Dependência e Fluxo de Execução

*   `depends_on` (string ou tabela): Uma lista de nomes de tarefas que devem ser concluídas com sucesso antes que esta tarefa possa ser executada.
*   `next_if_fail` (string ou tabela): Uma lista de nomes de tarefas a serem executadas *apenas se* esta tarefa falhar. Útil para tarefas de limpeza ou notificação.
*   `async` (booleano): Se `true`, a tarefa é executada em segundo plano, e o runner não espera que ela termine para iniciar a próxima tarefa na ordem de execução.

### Tratamento de Erros e Robustez

*   `retries` (número): O número de vezes que uma tarefa será tentada novamente se falhar. O padrão é `0`.
*   `timeout` (string): Uma duração (ex: `"10s"`, `"1m"`) após a qual a tarefa será encerrada se ainda estiver em execução.

### Execução Condicional

*   `run_if` (string ou função): A tarefa será pulada a menos que esta condição seja atendida.
    *   **Como string:** Um comando de shell. Um código de saída `0` significa que a condição foi atendida.
    *   **Como função:** Uma função Lua que retorna `true` se a tarefa deve ser executada.
*   `abort_if` (string ou função): Todo o fluxo de trabalho será abortado se esta condição for atendida.
    *   **Como string:** Um comando de shell. Um código de saída `0` significa abortar.
    *   **Como função:** Uma função Lua que retorna `true` para abortar.

### Hooks de Ciclo de Vida

*   `pre_exec` (função): Uma função Lua que é executada *antes* do `command` principal.
*   `post_exec` (função): Uma função Lua que é executada *após* o `command` principal ter sido concluído com sucesso.

### Reutilização

*   `uses` (tabela): Especifica uma tarefa pré-definida de outro arquivo (carregado via `import`) para usar como base. A definição da tarefa atual pode então sobrescrever propriedades como `params` ou `description`.
*   `params` (tabela): Um dicionário de pares chave-valor que podem ser passados para a função `command` da tarefa.
*   `artifacts` (string ou tabela): Um padrão de arquivo (glob) ou uma lista de padrões que especificam quais arquivos do `workdir` da tarefa devem ser salvos como artefatos após uma execução bem-sucedida.
*   `consumes` (string ou tabela): O nome de um artefato (ou uma lista de nomes) de uma tarefa anterior que deve ser copiado para o `workdir` desta tarefa antes que ela seja executada.

---

## Gerenciamento de Artefatos

O Sloth-Runner permite que as tarefas compartilhem arquivos entre si através de um mecanismo de artefatos. Uma tarefa pode "produzir" um ou mais arquivos como artefatos, e tarefas subsequentes podem "consumir" esses artefatos.

Isso é útil para pipelines de CI/CD, onde uma etapa de compilação pode gerar um binário (o artefato), que é então usado por uma etapa de teste ou de implantação.

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
        command = "echo 'conteudo_binario' > app.bin",
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
          if content == "conteudo_binario" then
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

---

## Funções Globais

O `sloth-runner` fornece funções globais no ambiente Lua para ajudar a orquestrar os fluxos de trabalho.

### `import(path)`

Carrega outro arquivo Lua e retorna o valor que ele retorna. Este é o principal mecanismo para criar módulos de tarefas reutilizáveis. O caminho é relativo ao arquivo que chama `import`.

**Exemplo (`reusable_tasks.lua`):**
```lua
-- Importa um módulo que retorna uma tabela de definições de tarefas
local docker_tasks = import("shared/docker.lua")

TaskDefinitions = {
  main = {
    tasks = {
      {
        -- Usa a tarefa 'build' do módulo importado
        uses = docker_tasks.build,
        params = { image_name = "my-app" }
      }
    }
  }
}
```

### `parallel(tasks)`

Executa uma lista de tarefas concorrentemente e espera que todas terminem.

*   `tasks` (tabela): Uma lista de tabelas de tarefas para executar em paralelo.

**Exemplo:**
```lua
command = function()
  log.info("Iniciando 3 tarefas em paralelo...")
  local results, err = parallel({
    { name = "short_task", command = "sleep 1" },
    { name = "medium_task", command = "sleep 2" },
    { name = "long_task", command = "sleep 3" }
  })
  if err then
    return false, "Execução paralela falhou"
  end
  return true, "Todas as tarefas paralelas terminaram."
end
```

### `export(table)`

Exporta dados de qualquer ponto de um script para a CLI. Quando a flag `--return` é usada, todas as tabelas exportadas são mescladas com o output da tarefa final em um único objeto JSON.

*   `table`: Uma tabela Lua a ser exportada.

**Exemplo:**
```lua
command = function()
  export({ valor_importante = "dado do meio da tarefa" })
  return true, "Tarefa concluída", { output_final = "algum resultado" }
end
```
Executar com `--return` produziria:
```json
{
  "valor_importante": "dado do meio da tarefa",
  "output_final": "algum resultado"
}
```