# Estado atual — Mundo Nexus

Atualizado em 20/11/2026. Versão de desenvolvimento: **0.6.3**.

## Onde fica cada coisa

| Área | Local | Responsabilidade |
|---|---|---|
| Aplicativo e API local | `*.go` na raiz | Perfis, mundos, motores, rede, backups e atualizações. |
| Interface | `ui/` | Páginas, estilos, textos, diálogos e marcas visuais. |
| Instalador tradicional | `installer/` | Instalação por usuário do Windows. |
| Microsoft Store | `store/` | Manifesto MSIX, empacotamento e auditorias. |
| Testes e validação | `*_test.go`, `scripts/` | Regras de negócio, qualidade e estrutura da interface. |
| Decisões e relatórios | `docs/` | Arquitetura, estado, publicação e pendências. |

## Funcionando nesta versão

- Biblioteca de Bedrock, Paper e Java portátil; perfis, mundo, complementos, console, jogadores, backups e diagnósticos.
- Memória em passos de 0,5 GiB. A recomendação protege o PC; reserva maior exige confirmação explícita.
- Favicon do Windows e marca interna separados: o ícone do aplicativo é grande e legível; a marca do cabeçalho é uma variação de três blocos.
- Janela nativa WebView2, sem recarregar com F5, e processos de servidores sem janela de CMD visível.

## Entrega e publicação

1. Gerar e testar o executável/instalador da versão.
2. Validar a instalação e a atualização sem alterar dados reais.
3. Executar WACK e finalizar informações da Microsoft Store.
4. Submeter o MSIX com a identidade reservada; a Store assina a versão distribuída.

**Regra:** não publicar Release/Store automaticamente enquanto os testes de instalação, assinatura e WACK não estiverem concluídos. A branch de desenvolvimento é `feature/pc-only-engines-memory-0.6.0`; `main` permanece protegida.

## Próximo ponto de trabalho

Validar visualmente o instalador 0.6.3, depois concluir os testes de motores reais e a publicação na Store. Dados de mundos e backups nunca devem ser apagados por atualização ou desinstalação.
