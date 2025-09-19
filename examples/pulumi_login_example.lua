-- examples/pulumi_login_example.lua

-- Carrega o módulo pulumi
local pulumi = require("pulumi")

-- Exemplo de login no Pulumi
-- Substitua "s3://my-pulumi-state-bucket" pelo seu backend de estado, se necessário.
-- Se você usa o serviço Pulumi, pode não precisar de um argumento.
local login_result = pulumi.login("s3://my-pulumi-state-bucket")

if login_result.success then
  print("Login no Pulumi realizado com sucesso!")
  print("Saída padrão:", login_result.stdout)
else
  print("Erro ao fazer login no Pulumi.")
  print("Erro padrão:", login_result.stderr)
end
