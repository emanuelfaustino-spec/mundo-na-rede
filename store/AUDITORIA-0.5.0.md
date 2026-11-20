# Auditoria integrada — Mundo Nexus 0.5.0
Data: 20/11/2026
Branch: feature/experiencia-unificada-0.5.0

## Evidência de testes
- `go test ./...`: passou.
- `go vet ./...`: passou.
- `node --check`: app.js, language.js e unified.js passaram.
- `node scripts/verify-ui.cjs`: 10 abas, IDs únicos e controles críticos confirmados.
- Compilação Windows, Linux amd64 e Android ARM64: aprovada após isolar o WebView2 em build Windows.
- Aplicativo Win32 de desenvolvimento iniciado com dados isolados e porta 37679; HTML respondeu HTTP 200.
- Chromium headless renderizou o app, com servidor de teste, menu unificado, seção do motor e sem aba "details" antiga.
- Endpoint preflight retornou pronto=false corretamente para servidor Bedrock sem motor; porta e LAN local foram verificadas.
- Código do backup manual e da retenção de automáticos testado com arquivos temporários.
- A instalação de 0.4.4 do usuário não foi modificada, e nenhuma cópia real de mundo foi tocada.

## Correções
1. Dez áreas reorganizadas, com Painel/Mundo/Complementos/Jogadores/Console/Backups e menu Mais para Rede/Automação/Conta/Ajuda.
2. Guia com respostas curtas e controle avançado sob demanda. Ações de uso frequente permanecem visíveis.
3. Reutilização dos controles antigos sem duplicá-los.
4. Verificação segura antes de iniciar e mensagens de diagnóstico simples.
5. Backups manuais preservados sem limpeza automática; formatos antigos aceitos.
6. Seletores de complementos por motor, mensagens simples e área própria de Console.
7. Novo teste estrutural da interface no fluxo de compilação do GitHub Actions.

## Pacotes locais
- `dist/Mundo-Nexus-0.5.0.0-x64.msix`: MSIX com identidade Microsoft Store oficial (não assinado/certificado).
- SHA256: `04C31CD69D382FC8B9A754103568585E3104B5290E2D9DD7F3B98CBA979627DB`.
- `dist/installer/Mundo-Nexus-Setup-v0.5.0.exe`: instalador tradicional não assinado.
- Não disponibilizar esses instaladores ao público antes dos testes, assinatura e certificação.

## Pendências importantes
- Controle remoto seguro entre dispositivos, autenticação, papéis e serviço de conexão de saída.
- Suporte automático a todos os motores Java, incluindo Forge/NeoForge/Fabric/Quilt com verificação de versões.
- Detecção completa de conflitos entre mods e dependências e rearranjo visual de packs.
- Agendamento automático de backups completos, lixeira, teste aprofundado de integridade e restauração.
- Testes com motor Minecraft em execução real, inclusive servidor com mods, addons e carga de jogadores.
- Testes gráficos WebView2 no computador com versão 0.5.0 instalada, tempo de abertura e responsividade.
- Regras de Store, imagens reais, WACK, instalação MSIX assinada, preço R$ 5,90 e submissão aprovada.

Não apresentar esta branch como versão completa de todas as funcionalidades solicitadas.
