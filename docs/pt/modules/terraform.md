# Módulo Terraform

O módulo `terraform` fornece uma interface de alto nível para orquestrar comandos da CLI `terraform`, permitindo que você gerencie o ciclo de vida de sua infraestrutura diretamente de dentro de uma pipeline do Sloth-Runner.

## Configuração

Este módulo requer que a CLI `terraform` esteja instalada e disponível no PATH do sistema. Todos os comandos devem ser executados dentro de um `workdir` específico onde seus arquivos `.tf` estão localizados.

## Funções

### `terraform.init(params)`

Inicializa um diretório de trabalho do Terraform.

- `params` (tabela):
    - `workdir` (string): **Obrigatório.** O caminho para o diretório que contém os arquivos do Terraform.
- **Retorna:** Uma tabela de resultados com `success`, `stdout`, `stderr` e `exit_code`.

### `terraform.plan(params)`

Cria um plano de execução do Terraform.

- `params` (tabela):
    - `workdir` (string): **Obrigatório.** O caminho para o diretório.
    - `out` (string): **Opcional.** O nome do arquivo para salvar o plano gerado.
- **Retorna:** Uma tabela de resultados.

### `terraform.apply(params)`

Aplica um plano do Terraform.

- `params` (tabela):
    - `workdir` (string): **Obrigatório.** O caminho para o diretório.
    - `plan` (string): **Opcional.** O caminho para um arquivo de plano a ser aplicado.
    - `auto_approve` (boolean): **Opcional.** Se `true`, aplica as alterações sem aprovação interativa.
- **Retorna:** Uma tabela de resultados.

### `terraform.destroy(params)`

Destrói a infraestrutura gerenciada pelo Terraform.

- `params` (tabela):
    - `workdir` (string): **Obrigatório.** O caminho para o diretório.
    - `auto_approve` (boolean): **Opcional.** Se `true`, destrói os recursos sem aprovação interativa.
- **Retorna:** Uma tabela de resultados.

### `terraform.output(params)`

Lê uma variável de saída de um arquivo de estado do Terraform.

- `params` (tabela):
    - `workdir` (string): **Obrigatório.** O caminho para o diretório.
    - `name` (string): **Opcional.** O nome de uma saída específica para ler. Se omitido, todas as saídas são retornadas como uma tabela.
- **Retorna:**
    - Em caso de sucesso: O valor JSON analisado da saída (pode ser uma string, tabela, etc.).
    - Em caso de falha: `nil, error_message`.

## Exemplo de Ciclo de Vida Completo

```lua
local tf_workdir = "./examples/terraform"

-- Tarefa 1: Init
local result_init = terraform.init({workdir = tf_workdir})
if not result_init.success then return false, "Init falhou" end

-- Tarefa 2: Plan
local result_plan = terraform.plan({workdir = tf_workdir})
if not result_plan.success then return false, "Plan falhou" end

-- Tarefa 3: Apply
local result_apply = terraform.apply({workdir = tf_workdir, auto_approve = true})
if not result_apply.success then return false, "Apply falhou" end

-- Tarefa 4: Get Output
local filename, err = terraform.output({workdir = tf_workdir, name = "report_filename"})
if not filename then return false, "Output falhou: " .. err end
log.info("Arquivo criado pelo Terraform: " .. filename)

-- Tarefa 5: Destroy
local result_destroy = terraform.destroy({workdir = tf_workdir, auto_approve = true})
if not result_destroy.success then return false, "Destroy falhou" end
```
