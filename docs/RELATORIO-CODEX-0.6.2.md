# Mundo Nexus — repasse objetivo para o Codex
**Data:** 20/11/2026 · **Versão de desenvolvimento:** 0.6.2  
**Repositório:** https://github.com/emanuelfaustino-spec/mundo-na-rede  
**Branch:** `feature/pc-only-engines-memory-0.6.0` · **PR:** https://github.com/emanuelfaustino-spec/mundo-na-rede/pull/1 (rascunho, não mesclado)

## 1. O que já existia e foi preservado
- App de **PC**, principalmente Windows, em **Go + WebView2 + interface HTML/CSS/JS**, com dez áreas: Painel, Mundo, Complementos, Jogadores, Console, Backups, Rede, Automação, Conta e Ajuda. Cabeçalho limitado a cinco controles, opções avançadas recolhidas.
- Perfis de servidores, mundos, importação de complementos, console, controles de jogadores, backups, indicadores de desempenho, verificações de pré-início e interface de gerenciamento. Nem todas as funções avançadas estão completas.
- Versões 0.6.0/0.6.1: biblioteca local de **Bedrock, Paper e Java 25 portátil**; reserva e limite de RAM por servidor (Job Object Windows); barra nativa do Windows combinando com o tema; foco em computador e remoção das opções atuais de hospedagem Android. Perfis antigos não foram apagados.
- A instalação que o usuário já tinha **não foi substituída**; a 0.6.2 foi testada como prévia com dados separados.

## 2. Corrigido nesta rodada (0.6.2)
- **Causa do erro de download Bedrock:** a API da biblioteca exigia o cabeçalho `X-Accept-Minecraft-Terms`, mas `installOfficial()` exigia `X-Accept-BDS-Terms`. Agora `acceptedBDSTerms()` reconhece ambos, após confirmação expressa do usuário. Arquivos: `downloads.go`, `bedrock_terms.go`, `bedrock_terms_test.go`.
- **Resultado real:** no Windows, o ZIP oficial Bedrock 1.26.66.1 foi baixado, extraído, validado e reconhecido: `bedrockReady=true`, `paperReady=true`, `javaReady=true`. Preparar novamente reutilizou o cache em ~118 ms, sem download repetido. Testes usaram a pasta temporária, não mundos reais.
- **Diálogos integrados:** todas as **15 confirmações nativas** encontradas em `ui/app.js` e `ui/engine-library.js` foram convertidas para `await nexusConfirm()`. Criados `ui/dialogs.js` e o modal `#nexus-dialog` em `ui/index.html`, com estilos em `ui/style.css`. API também oferece `nexusAlert()` e `nexusPrompt()`; suporte a confirmar, cancelar, Escape, foco e temas. Não deve mais aparecer `127.0.0.1:porta diz` nos fluxos convertidos.
- **Biblioteca:** prepara apenas motores ausentes, mostra status verdadeiro e erro específico por motor, sem sobrescrever sucesso com mensagem genérica. A verificação de estrutura de UI em `scripts/verify-ui.cjs` detecta diálogos nativos remanescentes.
- **Testes:** `go vet ./...`, build Windows, sintaxe JS e verificação da UI passaram localmente. Chrome headless confirmou abertura e resultado dos diálogos de confirmação, cancelamento, alerta e entrada de texto. Teste do endpoint confirmou a preparação real e o reaproveitamento. O CI Windows do GitHub deve confirmar os testes Go após o commit.

## 3. O que continua pendente (não afirmar que está pronto)
1. **Entrega Windows/Store:** instalar e inspecionar a 0.6.2 na janela WebView2 de forma completa, testar upgrades e recuperação de dados, assinar o MSIX, executar WACK e finalizar a ficha da Microsoft Store. Identidade: `Emanuel2229.MundoNexus`, produto `9N0WMH45BN3C`; preço pretendido **R$ 5,90**, ainda não configurado/publicado.
2. **Motores e compatibilidade:** testar Paper + Java 25 com mundo real, Geyser/Floodgate (crossplay) e complementos; concluir motores Java adicionais (Purpur, Fabric, Forge, NeoForge, Quilt, Spigot) conforme suporte/licenças. Validar dependências, incompatibilidades, ativação e versões de mods, plugins e add-ons.
3. **Estabilidade e proteção de dados:** testes de RAM/CPU sob carga e vários servidores, limites reais e otimização; agendamento/retenção dos sete backups automáticos, recuperação, integridade e prévias de restauração; validar conta/Google Drive com credenciais apropriadas.
4. **Rede e QA:** diagnosticar conectividade externa em redes distintas, testar Windows/Linux e acessibilidade visual de tema claro/escuro, teclado e mensagens. Não liberar portas ou firewall automaticamente.
5. **Escopo:** aplicativo móvel/controle pelo celular está **adiado**. Manter tudo de hospedagem e administração no PC; jogadores Bedrock pelo celular continuam compatíveis.

## 4. Como continuar com segurança
- Revisar a **PR #1** e a branch de desenvolvimento; rodar `go test ./...`, `go vet ./...`, `node scripts/verify-ui.cjs` e testar o executável em dados isolados. O Windows às vezes bloqueia testes temporários não assinados; CI GitHub é alternativa, **não desativar as proteções**.
- Prévia local: `%TEMP%\MundoNexus-061-Preview\Mundo-Nexus-0.6.2-Preview.exe`, com dados em `%TEMP%\MundoNexus-061-Preview\user-preview-data`. A versão antiga instalada fica em `%LOCALAPPDATA%\Programs\Mundo Nexus\`. Não sobrescrever os perfis ou motores da instalação atual.
- Não mesclar na `main`, publicar Release/Store ou autorizar cobranças sem validação final. Continuar com interface simples, sem botões duplicados e com no máximo cinco controles no cabeçalho.
