# Auditoria pre-publicacao — Mundo Nexus 0.4.5
Data: 20/11/2026

## Alteracoes efetivamente feitas
- Corrigida a falha de indice na rotacao de backups quando existiam menos de sete arquivos validos. O erro foi reproduzido com teste automatizado antes da correcao.
- Backup local agora recebe nome `mundo-nexus-*.zip`, aceitando tanto `mundo-nexus-*.zip` quanto `mundo-na-rede-*.zip` nos mecanismos de listagem, limpeza e exportacao. Nomes antigos de backups remotos foram mantidos por compatibilidade.
- Atualizador tradicional passou a preservar copia recuperavel do executavel antigo antes da troca, com restauracao se a substituicao falhar. Testes cobrem sucesso e falha simulados usando arquivos temporarios.
- Separacao do canal Store: comandos de atualizacao direta nao sao aplicados nas builds compiladas para a Store; teste automatizado adicionado.
- Corrigidas referencias residuais de interface `Mundo Nexo` para `Mundo Nexus`.
- Workflow de tags do GitHub nao publica mais automaticamente executaveis sem assinatura em releases publicas; guarda apenas artefatos temporarios para revisao.
- Politica de privacidade e termos de uso ajustados para descrever verificacoes de atualizacao e canais de distribuicao.
- Versao do codigo, manifesto e instalador atualizados para 0.4.5.
- A geracao do MSIX usa a identidade oficial do Partner Center, separada do Nexus Sites.

## Artefatos locais gerados
- `dist/Mundo-Nexus-0.4.5.0-x64.msix`: pacote para futura submissao Microsoft Store, SEM assinatura local e SEM certificacao. SHA-256: `0FD21A40889489F075A1B468E326635891A8E32BA2C7523434F5FD73CAD47444`.
- `dist/Mundo-Nexus.exe`: compilacao de uso direto sem assinatura.
- `dist/installer/Mundo-Nexus-Setup-v0.4.5.exe`: instalador Inno Setup sem assinatura.

## Verificacoes realizadas
- `go test ./...`: aprovado.
- `go vet ./...`: aprovado.
- Verificacao de sintaxe em `ui/app.js` e `ui/language.js`: aprovada.
- Compilacao e empacotamento MSIX via MakeAppx: concluidos.
- Compilacao Inno Setup: concluida.
- Teste de race detector nao executavel no ambiente atual: requer CGO habilitado e compilador C configurado.
- Cobertura de testes ainda baixa; testes de integracao faltando.

## Bloqueios externos e pendencias
1. Windows Code Integrity bloqueia executaveis nao assinados na maquina: eventos 3077 e 3033; exige validacao com assinatura e politica permitida. Nao foi desativada protecao alguma.
2. Nao foi possivel executar o app para confirmar se a janela abre normalmente ou se o motor Bedrock inicia um mundo. Os registros anteriores de AppHangB1 nao demonstram que o travamento anterior foi resolvido.
3. Ainda faltam testes completos do Windows App Certification Kit, instalacao/atualizacao via MSIX assinado, smoke tests de servidores e verificacao de perfis Java/Bedrock.
4. Dependencias externas podem precisar de atualizacao/analise de seguranca (nao foram atualizadas automaticamente para evitar incompatibilidades).
5. Preco pretendido R$ 5,90, mercados, imagens da ficha e certificacao do Partner Center continuam pendentes.
6. Historico antigo de repositorios e backups foi preservado. Renomeacao remota de repositorios requer migracao coordenada de links, atualizador e pipelines.
7. Alteracoes locais ainda nao commitadas nem enviadas ao GitHub; nao publicar artefatos unsigned em releases ou alterar dados de usuarios sem validacao.

## Roteiro para Work/Codex
1. Revisar `git diff` e testes locais. Criar branch/commit de revisao e CI, sem sobrescrever mudancas anteriores.
2. Confirmar requisitos de certificacao MSIX, identidade oficial e assinatura da loja.
3. Validar runtime WebView2 e motor de servidores no PC quando permitido pela politica de integridade.
4. Testar instalacao, atualizacao, desinstalacao, backups e recuperacao em ambiente limpo.
5. Completar imagens reais, ficha e preco da Microsoft Store; submeter somente apos aprovacao final.
