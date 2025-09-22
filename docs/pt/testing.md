# Testando Workflows

O sloth-runner inclui um framework de testes embutido que permite escrever testes unitários e de integração para seus workflows de tarefas. Escrever testes para sua automação é crucial para garantir a confiabilidade, prevenir regressões e ter confiança ao fazer alterações.

## O Comando `test`

Você pode executar um arquivo de teste usando o comando `sloth-runner test`. Ele requer dois arquivos principais: o workflow que você quer testar e o próprio script de teste.

```bash
sloth-runner test -w <caminho_para_workflow.lua> -f <caminho_para_arquivo_de_teste.lua>
```

-   `-w, --workflow`: Especifica o caminho para o arquivo principal de `TaskDefinitions` que você quer testar.
-   `-f, --file`: Especifica o caminho para o seu arquivo de teste.

## Escrevendo Testes

Os testes são escritos em Lua e usam dois novos módulos globais fornecidos pelo executor de testes: `test` e `assert`.

### O Módulo `test`

O módulo `test` é usado para estruturar seus testes e para executar tarefas específicas do seu workflow.

-   `test.describe(suite_name, function)`: Agrupa testes relacionados em uma "suíte". Serve para organização.
-   `test.it(function)`: Define um caso de teste individual. A descrição do teste deve ser incluída nas mensagens de asserção dentro desta função.
-   `test.run_task(task_name)`: Esta é a função principal do framework de testes. Ela executa uma única tarefa pelo seu nome a partir do arquivo de workflow carregado. Ela retorna uma tabela de `result` contendo os detalhes da execução.

A tabela `result` retornada por `run_task` tem a seguinte estrutura:

```lua
{
  success = true, -- booleano: true se a tarefa foi bem-sucedida, false caso contrário
  message = "Tarefa executada com sucesso", -- string: A mensagem retornada pela tarefa
  duration = "1.23ms", -- string: A duração da execução
  output = { ... }, -- tabela: A tabela de output retornada pela tarefa
  error = nil -- string: A mensagem de erro se a tarefa falhou
}
```

### O Módulo `assert`

O módulo `assert` fornece funções para verificar os resultados das execuções de suas tarefas.

-   `assert.is_true(value, message)`: Verifica se o `value` é verdadeiro.
-   `assert.equals(actual, expected, message)`: Verifica se o valor `actual` é igual ao valor `expected`.

### Mocking de Módulos

Para testar a lógica de suas pipelines sem fazer chamadas externas reais (ex: para AWS, Docker ou Terraform), o framework de testes inclui um poderoso recurso de mocking.

#### Política de Mocking Estrito

O executor de testes impõe uma **política de mocking estrito**. Ao rodar em modo de teste, qualquer chamada a uma função de módulo (como `aws.exec` ou `docker.build`) que **não** tenha sido explicitamente mockada fará com que o teste falhe imediatamente. Isso garante que seus testes sejam totalmente autocontidos, determinísticos e não tenham efeitos colaterais indesejados.

#### `test.mock(function_name, mock_definition)`

Esta função permite que você defina um valor de retorno falso para qualquer função de módulo que possa ser mockada.

-   `function_name` (string): O nome completo da função a ser mockada (ex: `"aws.s3.sync"`, `"docker.build"`).
-   `mock_definition` (tabela): Uma tabela que define o que a função mockada deve retornar. Ela **deve** conter uma chave `returns`, que é uma lista dos valores que a função retornará.

A lista `returns` é crucial porque funções Lua podem retornar múltiplos valores.

**Exemplo:**

```lua
-- Mock de uma função que retorna uma única tabela de resultado
test.mock("docker.build", {
  returns = {
    { success = true, stdout = "Imagem construída com sucesso" }
  }
})

-- Mock de uma função que retorna dois valores (ex: um valor e um erro)
-- Isto simula uma chamada bem-sucedida a terraform.output
test.mock("terraform.output", {
  returns = { "meu_arquivo.txt", nil }
})

-- Isto simula uma chamada com falha
test.mock("terraform.output", {
  returns = { nil, "output não encontrado" }
})
```

### Exemplo Completo de Mocking

Digamos que você tenha uma tarefa que chama `aws.exec` e possui uma lógica que depende do resultado.

**Tarefa em `meu_workflow.lua`:**
```lua
-- ...
{
  name = "verificar-conta",
  command = function()
    local result = aws.exec({"sts", "get-caller-identity"})
    local data = data.parse_json(result.stdout)
    if data.Account == "123456789012" then
      return true, "Conta correta."
    else
      return false, "Conta errada."
    end
  end
}
-- ...
```

**Teste em `meu_teste.lua`:**
```lua
test.describe("Lógica de Verificação de Conta", function()
  test.it(function()
    -- Mock do valor de retorno de aws.exec
    test.mock("aws.exec", {
      returns = {
        {
          success = true,
          stdout = '{"Account": "123456789012"}'
        }
      }
    })

    -- Executa a tarefa que usa o mock
    local result = test.run_task("verificar-conta")

    -- Afirma que a lógica da tarefa funcionou corretamente com os dados mockados
    assert.is_true(result.success, "A tarefa deve ser bem-sucedida com o ID de conta correto")
    assert.equals(result.message, "Conta correta.", "A mensagem deve ser correta")
  end)
end)
```
