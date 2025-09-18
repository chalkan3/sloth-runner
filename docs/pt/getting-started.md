# Início Rápido

Bem-vindo ao Sloth-Runner! Este guia o ajudará a começar a usar a ferramenta rapidamente.

## Instalação

Para instalar o `sloth-runner` em seu sistema, você pode usar o script `install.sh` fornecido. Este script detecta automaticamente seu sistema operacional e arquitetura, baixa a versão mais recente do GitHub e coloca o executável `sloth-runner` em `/usr/local/bin`.

```bash
bash <(curl -sL https://raw.githubusercontent.com/chalkan3/sloth-runner/master/install.sh)
```

**Nota:** O script `install.sh` requer privilégios de `sudo` para mover o executável para `/usr/local/bin`.

## Uso Básico

Para executar um arquivo de tarefa Lua:

```bash
sloth-runner run -f examples/basic_pipeline.lua
```

Para listar as tarefas em um arquivo:

```bash
sloth-runner list -f examples/basic_pipeline.lua
```

## Próximos Passos

Agora que você tem o Sloth-Runner instalado e funcionando, explore os [Conceitos Essenciais](./core-concepts.md) para entender como definir suas tarefas, ou mergulhe diretamente nos novos [Módulos Built-in](../index.md#módulos-built-in) para automação avançada com Git, Pulumi e Salt.

---
**Idiomas Disponíveis:**
[English](../en/getting-started.md) | [Português](./getting-started.md) | [中文](../zh/getting-started.md)