-- examples/git_conditional_deploy.lua
-- Description: Este exemplo demonstra um pipeline de deploy condicional
-- que utiliza a nova API orientada a objetos. O script inspeciona a
-- branch de um repositório Git local e, em seguida, usa o módulo Salt
-- para "deployar" para ambientes diferentes com base no nome da branch.

-- Pré-requisito: Você deve ter um repositório git local em "./tmp/app-checkout"
-- para que este exemplo funcione.

log.info("Iniciando pipeline de deploy condicional...")

-- Etapa 1: Obter um objeto para um repositório Git local existente.
-- Note que estamos usando 'git:repo()' em vez de 'git:clone()'.
local app_repo = git:repo("./tmp/app-checkout")

log.info("Repositório local em: " .. app_repo.local.path)
log.info("Branch atual detectada: '" .. app_repo.current.branch .. "'")

-- Etapa 2: Lógica condicional baseada nos dados do objeto do repositório.
if app_repo.current.branch == "main" then
  -- Se a branch for 'main', consideramos um deploy de produção.
  log.info("Branch 'main' detectada. Preparando para deploy em produção.")

  -- Garante que o código está atualizado antes do deploy.
  log.info("Executando 'git pull' para obter as últimas atualizações...")
  app_repo:pull()

  -- Usa o módulo Salt para aplicar o estado no ambiente de produção.
  log.info("Executando deploy em produção com Salt...")
  salt:target("prod-servers"):cmd("state.apply", "my-app")

  log.success("Deploy em produção concluído!")

else
  -- Se for qualquer outra branch, consideramos um deploy de homologação/staging.
  log.info("Branch de feature detectada. Preparando para deploy em homologação.")

  -- Usa o módulo Salt para aplicar o estado no ambiente de homologação.
  log.info("Executando deploy em homologação com Salt...")
  salt:target("staging-servers"):cmd("state.apply", "my-app", "test=true")

  log.success("Deploy em homologação concluído!")
end

log.info("Pipeline de deploy condicional finalizado.")
