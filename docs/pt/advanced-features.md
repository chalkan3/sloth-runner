# Funcionalidades Avançadas

Este documento aborda algumas das funcionalidades mais avançadas do `sloth-runner`, projetadas para aprimorar seus fluxos de trabalho de desenvolvimento, depuração e configuração.

## Executor de Tarefas Interativo

Para fluxos de trabalho complexos, pode ser útil percorrer as tarefas uma a uma, inspecionar suas saídas e decidir se deve prosseguir, pular ou tentar novamente uma tarefa. O executor de tarefas interativo fornece uma maneira poderosa de depurar e desenvolver seus pipelines de tarefas.

Para usar o executor interativo, adicione a flag `--interactive` ao comando `sloth-runner run`:

```bash
sloth-runner run -f examples/basic_pipeline.lua --yes --interactive
```

Quando habilitado, o executor pausará antes de executar cada tarefa e solicitará uma ação:

```
? Tarefa: fetch_data (Simula a busca de dados brutos)
> executar
  pular
  abortar
  continuar
```

**Ações:**

*   **executar:** (Padrão) Prossegue com a execução da tarefa atual.
*   **pular:** Pula a tarefa atual e passa para a próxima na ordem de execução.
*   **abortar:** Aborta imediatamente toda a execução da tarefa.
*   **continuar:** Executa a tarefa atual e todas as subsequentes sem mais prompts, desativando efetivamente o modo interativo para o resto da execução.

## Modelagem Aprimorada de `values.yaml`

Você pode tornar seus arquivos `values.yaml` mais dinâmicos usando a sintaxe de modelo Go para injetar variáveis de ambiente. Isso é particularmente útil para fornecer informações sensíveis (como tokens ou chaves) ou configurações específicas do ambiente sem codificá-las.

O `sloth-runner` processa o `values.yaml` como um modelo Go, disponibilizando quaisquer variáveis de ambiente no mapa `.Env`.

**Exemplo:**

1.  **Crie um arquivo `values.yaml` com um placeholder de modelo:**

    ```yaml
    # values.yaml
    api_key: "{{ .Env.MY_API_KEY }}"
    region: "{{ .Env.AWS_REGION | default "us-east-1" }}"
    ```
    *Nota: Você pode usar `default` para fornecer um valor de fallback se a variável de ambiente não estiver definida.*

2.  **Crie uma tarefa Lua que use esses valores:**

    ```lua
    -- my_task.lua
    TaskDefinitions = {
      my_group = {
        tasks = {
          {
            name = "deploy",
            command = function()
              log.info("Implantando na região: " .. values.region)
              log.info("Usando a chave de API (primeiros 5 caracteres): " .. string.sub(values.api_key, 1, 5) .. "...")
              return true, "Implantação bem-sucedida."
            end
          }
        }
      }
    }
    ```

3.  **Execute a tarefa com as variáveis de ambiente definidas:**

    ```bash
    export MY_API_KEY="supersecretkey12345"
    export AWS_REGION="us-west-2"

    sloth-runner run -f my_task.lua -v values.yaml --yes
    ```

**Saída:**

A saída mostrará que os valores das variáveis de ambiente foram substituídos corretamente:

```
INFO Implantando na região: us-west-2
INFO Usando a chave de API (primeiros 5 caracteres): super...
```
