-- examples/pulumi_git_combined_example.lua
-- Demonstrates the new, more powerful method-style API

-- Etapa 1: Clonar o repositório que contém a infraestrutura
log.info("Iniciando o clone do repositório de infraestrutura...")
local repo = git:clone("https://github.com/user/example-infra.git", "./tmp/infra-checkout")

-- Etapa 2: Usar os dados do objeto 'repo' retornado
-- Acessamos o caminho do repositório clonado para usá-lo em outro módulo.
log.info("Repositório clonado com sucesso em: " .. repo.local.path)
log.info("URL remota: " .. repo.remote.url)
log.info("Branch atual: " .. repo.current.branch)

-- Opcional: Realizar mais operações git no objeto repo
repo:pull()

-- Etapa 3: Definir e provisionar a infraestrutura com Pulumi
-- O 'workdir' do Pulumi agora aponta dinamicamente para o diretório clonado.
log.info("Configurando a stack Pulumi...")
local stack = pulumi:stack("dev-stack", { workdir = repo.local.path })

-- Etapa 4: Executar o 'up' da stack e obter os resultados
log.info("Provisionando a infraestrutura com 'pulumi up'...")
stack:up()

log.info("Infraestrutura provisionada. Obtendo as saídas...")
local outputs = stack:outputs()

-- Etapa 5: Usar as saídas da infraestrutura
log.info("A URL da aplicação é: " .. outputs.url)
log.info("O nome do bucket S3 é: " .. outputs.bucket_name)

log.success("Workflow de Git e Pulumi concluído com sucesso!")