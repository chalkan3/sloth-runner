-- examples/pulumi_multi_stack_dependencies.lua
-- Description: Este exemplo demonstra um workflow avançado com Pulumi,
-- onde a saída (output) de uma stack é usada como entrada (input) para
-- outra. Isso é comum para separar o gerenciamento de diferentes camadas
-- da infraestrutura, como rede e aplicação.

log.info("Iniciando workflow de stacks Pulumi com dependências...")

-- --- Etapa 1: Provisionar a Stack de Rede ---
log.info("Configurando e provisionando a stack de rede base...")

-- Cria um objeto para a stack que gerencia a infraestrutura de rede (VPC, subnets, etc.)
local network_stack = pulumi:stack("infra-network-stack", { workdir = "./pulumi/network" })

-- Executa o 'up' para criar ou atualizar os recursos de rede.
network_stack:up()

-- Após o provisionamento, obtemos as saídas da stack.
log.info("Coletando saídas da stack de rede...")
local network_outputs = network_stack:outputs()

-- A saída 'vpc_id' agora está disponível para ser usada em outras partes do nosso script.
log.info("VPC ID obtido: " .. network_outputs.vpc_id)
log.info("Subnet ID obtida: " .. network_outputs.public_subnet_id)

-- --- Etapa 2: Provisionar a Stack da Aplicação ---
log.info("Configurando e provisionando a stack da aplicação...")

-- Agora, criamos a stack da aplicação.
-- Em um cenário real, passaríamos o ID da VPC e da subnet como configuração
-- para que os servidores da aplicação sejam criados na rede correta.
-- A lógica abaixo simula como esses dados seriam preparados.
local app_config = {
  vpc_id = network_outputs.vpc_id,
  subnet_id = network_outputs.public_subnet_id
}

log.info("A stack da aplicação será provisionada na VPC '" .. app_config.vpc_id .. "'")

-- Cria o objeto da stack da aplicação.
-- O workdir aponta para o projeto Pulumi que define os servidores.
local app_stack = pulumi:stack("app-server-stack", { workdir = "./pulumi/app" })

-- Executa o 'up' da aplicação. O código Pulumi (em ./pulumi/app)
-- seria responsável por ler a configuração (app_config) para usar a VPC correta.
app_stack:up()

-- Coleta as saídas da stack da aplicação.
log.info("Coletando saídas da stack da aplicação...")
local app_outputs = app_stack:outputs()

log.info("URL do load balancer da aplicação: " .. app_outputs.app_url)

log.success("Workflow de múltiplas stacks Pulumi concluído com sucesso!")
