# REPL Interativo

O comando `sloth-runner repl` inicia uma sessão interativa de Read-Eval-Print Loop (REPL). Esta é uma ferramenta poderosa para depuração, exploração e experimentação rápida com os módulos do sloth-runner.

## Iniciando o REPL

Para iniciar uma sessão, simplesmente execute:
```bash
sloth-runner repl
```

Você também pode pré-carregar um arquivo de workflow para ter suas `TaskDefinitions` e quaisquer funções auxiliares disponíveis na sessão. Isso é incrivelmente útil para depurar uma pipeline existente.

```bash
sloth-runner repl -f /caminho/para/sua/pipeline.lua
```

## Funcionalidades

### Ambiente ao Vivo
O REPL fornece um ambiente Lua ao vivo onde você pode executar qualquer código Lua. Todos os módulos embutidos do sloth-runner (`aws`, `docker`, `fs`, `log`, etc.) são pré-carregados e prontos para uso.

```
sloth> log.info("Olá do REPL!")
sloth> resultado = fs.read("README.md")
sloth> print(string.sub(resultado, 1, 50))
```

### Autocompletar
O REPL possui um sistema sofisticado de autocompletar.
- Comece a digitar o nome de uma variável global ou módulo (ex: `aws`) e pressione `Tab` para ver as sugestões.
- Digite o nome de um módulo seguido por um ponto (ex: `docker.`) e pressione `Tab` para ver todas as funções disponíveis naquele módulo.

### Histórico
O REPL mantém um histórico de seus comandos. Use as setas para cima e para baixo para navegar pelos comandos anteriores.

## Exemplo de Sessão

Aqui está um exemplo de uso do REPL para depurar um comando Docker.

```bash
$ sloth-runner repl
Sloth-Runner Interactive REPL
Digite 'exit' ou 'quit' para sair.
sloth> resultado = docker.exec({"ps", "-a"})
sloth> print(resultado.stdout)
CONTAINER ID   IMAGE     COMMAND   CREATED   STATUS    PORTS     NAMES
sloth> -- Agora vamos tentar construir uma imagem
sloth> resultado_build = docker.build({tag="meu-teste", path="./examples/docker"})
sloth> print(resultado_build.success)
true
sloth> exit
Tchau!
```
