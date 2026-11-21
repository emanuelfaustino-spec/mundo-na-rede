# Continuidade — Mundo Nexus

## O que é

Mundo Nexus é um aplicativo Windows para preparar e administrar servidores de Minecraft no próprio computador. Ele reúne criação de perfis, regras de mundo, complementos, console, jogadores, backups, rede e diagnóstico em uma interface local.

## O que funciona hoje

- Hospedagem local para Bedrock e preparação para Paper/Java.
- Biblioteca local de motores, limite de RAM, console, regras de mundo, complementos e backups.
- Interface em português/inglês, dados locais preservados e atualização direta planejada.

## Estado do código

- Versão: `0.6.7`.
- Branch ativa: `feature/pc-only-engines-memory-0.6.0`.
- `main` não deve ser alterada sem revisão e validação.
- A versão de desenvolvimento não é uma publicação da Microsoft Store.

## Pendências importantes

1. Validar instalação/atualização real e ícones em um instalador assinado.
2. Testar Bedrock, Paper e crossplay com jogadores reais.
3. Concluir retenção de sete backups, integridade e Drive em conta real.
4. Executar WACK e completar a ficha/assinatura da Microsoft Store.
5. Para hospedagem externa: projetar um conector de VPS sob conta do próprio usuário; não prometer serviço gratuito ilimitado.

## Regras de manutenção

- Não apagar mundos, motores, backups ou perfis durante atualização/desinstalação.
- Não publicar Release ou Store automaticamente antes de validação, assinatura e WACK.
- Manter `ui/` para interface, `installer/` para instalador, `store/` para MSIX e `docs/` para decisões e status.
