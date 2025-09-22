# Módulo AWS

O módulo `aws` fornece uma interface abrangente para interagir com a Amazon Web Services usando o AWS CLI. Ele foi projetado para funcionar perfeitamente com as cadeias de credenciais padrão da AWS e também possui suporte de primeira classe para o `aws-vault` para maior segurança.

## Configuração

Nenhuma configuração específica no `values.yaml` é necessária. O módulo depende de seu ambiente estar configurado para interagir com a AWS. Isso pode ser alcançado através de:
- Perfis IAM para instâncias EC2 ou tarefas ECS/EKS.
- Variáveis de ambiente padrão (`AWS_ACCESS_KEY_ID`, `AWS_SECRET_ACCESS_KEY`, etc.).
- Um arquivo `~/.aws/credentials` configurado.
- Usando o `aws-vault` com um perfil nomeado.

## Executor Genérico

### `aws.exec(args, opts)`

Esta é a função principal do módulo. Ela executa qualquer comando do AWS CLI e retorna o resultado.

**Parâmetros:**

- `args` (tabela): **Obrigatório.** Uma tabela de strings representando o comando e os argumentos a serem passados para o AWS CLI (ex: `{"s3", "ls", "--recursive"}`).
- `opts` (tabela): **Opcional.** Uma tabela de opções para a execução.
    - `profile` (string): Se fornecido, o comando será executado usando `aws-vault exec <profile> -- aws ...`. Se omitido, ele executará `aws ...` diretamente.

**Retornos:**

Uma tabela contendo os seguintes campos:
- `stdout` (string): A saída padrão do comando.
- `stderr` (string): O erro padrão do comando.
- `exit_code` (número): O código de saída do comando. `0` normalmente indica sucesso.

**Exemplo:**

```lua
-- Usando credenciais padrão
local result = aws.exec({"sts", "get-caller-identity"})
if result.exit_code == 0 then
  print(result.stdout)
end

-- Usando um perfil do aws-vault
local result_with_profile = aws.exec({"ec2", "describe-instances"}, {profile = "meu-perfil-prod"})
```

## Ajudantes do S3

### `aws.s3.sync(params)`

Um wrapper de alto nível para o comando `aws s3 sync`, útil para sincronizar diretórios com o S3.

**Parâmetros:**

- `params` (tabela): Uma tabela contendo os seguintes campos:
    - `source` (string): **Obrigatório.** O diretório de origem ou caminho S3.
    - `destination` (string): **Obrigatório.** O diretório de destino ou caminho S3.
    - `profile` (string): **Opcional.** O perfil do `aws-vault` a ser usado.
    - `delete` (boolean): **Opcional.** Se `true`, adiciona a flag `--delete` ao comando de sincronização.

**Retornos:**

- `true` em caso de sucesso.
- `false, error_message` em caso de falha.

**Exemplo:**

```lua
local ok, err = aws.s3.sync({
  source = "./build",
  destination = "s3://meu-bucket-app/static",
  profile = "perfil-deploy",
  delete = true
})
if not ok then
  log.error("Falha na sincronização com o S3: " .. err)
end
```

## Ajudantes do Secrets Manager

### `aws.secretsmanager.get_secret(params)`

Recupera o valor de um segredo do AWS Secrets Manager. Esta função simplifica o processo, retornando diretamente a `SecretString`.

**Parâmetros:**

- `params` (tabela): Uma tabela contendo os seguintes campos:
    - `secret_id` (string): **Obrigatório.** O nome ou ARN do segredo a ser recuperado.
    - `profile` (string): **Opcional.** O perfil do `aws-vault` a ser usado.

**Retornos:**

- `secret_string` (string) em caso de sucesso.
- `nil, error_message` em caso de falha.

**Exemplo:**

```lua
local db_password, err = aws.secretsmanager.get_secret({
  secret_id = "producao/database/password",
  profile = "meu-perfil-app"
})

if not db_password then
  log.error("Falha ao obter o segredo: " .. err)
  return false, "Configuração falhou."
end

-- Agora você pode usar a variável db_password
```
