# Mundo Nexus — experiência simples e integração v0.5.0
Data: 20/11/2026

## Regra de produto
Uma única versão integrada, sem publicar módulos isolados. O primeiro uso deve permitir criar um perfil, preparar um motor compatível e iniciar o servidor sem conhecimento técnico. Ações comuns ficam visíveis; opções complexas ficam sob **Mais** ou em painéis recolhíveis. Textos devem explicar a finalidade em uma frase. Não duplicar o mesmo comando em duas áreas, exceto atalhos úteis no Painel.

## Navegação unificada
- **Painel:** situação e controles rápidos, RAM/CPU, verificação antes de iniciar.
- **Mundo:** instalação do motor, mundo, seed e regras, pré-geração avançada, configuração.
- **Complementos:** add-ons Bedrock ou mods/plugins Java conforme motor, catálogo opcional.
- **Jogadores:** lista e permissões disponíveis para Bedrock; Java usa Console.
- **Console:** comandos, últimas mensagens, pesquisa e explicações básicas de erros.
- **Backups:** cópias locais e recuperação; backups manuais não são removidos automaticamente.
- **Mais → Rede:** endereço/convite, testes locais sem abrir portas, ajustes externos explícitos.
- **Mais → Automação:** métricas, modo econômico e comandos programados.
- **Mais → Conta:** Google Drive; controle remoto indisponível até existir autenticação segura.
- **Mais → Ajuda:** respostas curtas e guia opcional.

## Funcionamento da revisão
- A interface preserva identificadores e listeners atuais, movendo botões existentes em vez de copiá-los.
- O diagnóstico local em `/api/servers/:id/preflight` verifica pasta, motor, portas e memória e não abre firewall.
- O início do motor exige confirmação do diagnóstico local para evitar erros comuns; preparação Bedrock continua guiada.
- Avançado permanece opcional, desligado por padrão e armazenado só no dispositivo.
- O suporte a backups manuais é permanente até exclusão pelo usuário. Formatos legados continuam legíveis.
- A versão 0.5.0 existe em branch de desenvolvimento até testes de instalação e compatibilidade.

## Funcionalidades ainda NÃO concluídas
Não anunciar como prontas: painel remoto entre dispositivos, identidade/autenticação compartilhada, notificações remotas, papéis granulares, túnel de saída seguro, instalações automáticas de todos os motores Java, busca automática de dependências completas, reorganização visual com resolução de conflitos de todos os add-ons, testes de conectividade externa em diferentes operadoras, backup diário/semanal com retenção automática, lixeira recuperável, certificação de instalação Store e testes de carga/TPS reais. Essas funcionalidades precisam de backend, compatibilidade de versões e testes adicionais. O produto ainda não está pronto para publicação comercial.

## Testes
- Go `go test ./...` e `go vet ./...`;
- Node `node --check ui/{app,language,unified}.js`;
- `node scripts/verify-ui.cjs` cobre dez abas, IDs e controles críticos;
- Chrome headless validou o DOM carregado com um servidor criado em diretório temporário;
- Endpoint de pré-início validado em testes de integração. Nenhum mundo real foi alterado.

## Implantação
Não substituir automaticamente a instalação atual no PC. Não enviar MSIX não assinado a usuários, nem colocar a venda na Store. Publicação final após testes reais e confirmação do proprietário.
