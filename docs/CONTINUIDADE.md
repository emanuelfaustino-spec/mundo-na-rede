# Continuidade — Mundo Nexus

## O que é

Mundo Nexus é um aplicativo Windows para preparar e administrar servidores de Minecraft no próprio computador. Ele reúne criação de perfis, regras de mundo, complementos, console, jogadores, backups, rede e diagnóstico em uma interface local.

## O que funciona hoje

- Hospedagem local para Bedrock e preparação para Paper/Java.
- Biblioteca local de motores, limite de RAM, console, regras de mundo, complementos e backups.
- Interface em português/inglês, dados locais preservados e atualização direta planejada.

## Estado do código

- Versão: `0.6.8`.
- Branch ativa: `feature/pc-only-engines-memory-0.6.0`.
- `main` não deve ser alterada sem revisão e validação.
- A versão de desenvolvimento não é uma publicação da Microsoft Store.

## Pendências importantes

1. Validar instalação/atualização real e ícones em um instalador assinado.
2. Testar Bedrock, Paper e crossplay com jogadores reais.
3. Concluir retenção de sete backups, integridade e Drive em conta real.
4. Executar WACK e completar a ficha/assinatura da Microsoft Store.
5. Para hospedagem externa: implementar o conector de VPS por SSH sob conta do próprio usuário; não prometer serviço gratuito ilimitado.

## Hospedagem externa planejada

O aplicativo terá duas rotas claramente separadas:

- **Servidor local:** o computador do usuário executa o mundo, como funciona hoje.
- **VPS própria:** o Mundo Nexus prepara e acompanha uma máquina remota da conta do usuário por SSH. A conta, a chave e a liberação de portas continuam sob controle dessa pessoa.

A primeira compatibilidade planejada é Oracle Cloud Always Free, por ser uma opção de baixo custo, mas a criação da conta e a escolha de região exigem ação do proprietário. Não automatizar firewall, roteador, conta ou cobrança sem confirmação explícita.

## Regras de manutenção

- Não apagar mundos, motores, backups ou perfis durante atualização/desinstalação.
- Não publicar Release ou Store automaticamente antes de validação, assinatura e WACK.
- Manter `ui/` para interface, `installer/` para instalador, `store/` para MSIX e `docs/` para decisões e status.
