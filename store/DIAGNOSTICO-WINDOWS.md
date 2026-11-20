# Diagnostico Windows — Mundo Nexus — 20/11/2026

## Resultado
- O instalador Inno Setup 0.4.4 foi recompilado e seu registro aparece em HKCU com DisplayName "Mundo Nexus versao 0.4.4".
- O arquivo instalado e o executavel compilado em dist/Mundo-Nexus.exe tinham o mesmo SHA-256.
- O sistema bloqueia a execucao do instalador e do executavel sem assinatura aceita: Microsoft-Windows-CodeIntegrity/Operational, eventos **3077** e **3033**, politica {0283ac0f-fff1-49ae-ada1-8a933130cad6}.
- Alem disso, eventos de 20/11/2026 registram travamentos AppHangB1 (Application Hang 1002) de Mundo-Nexus.exe em versoes anteriores. Causa daqueles travamentos nao confirmada.
- O teste automatizado do executavel instalado retornou explicitamente: "Uma politica de Controle de Aplicativo bloqueou este arquivo". Portanto nao foi possivel validar a interface nem iniciar um servidor de jogo neste teste.
- O codigo WebView2 passou a fixar a thread de interface com runtime.LockOSThread. A mudanca compila, mas o teste grafico permanece pendente.
- O comando MakeAppx criou o MSIX 0.4.4.0 de validacao, sem assinatura e com identidade ficticia. Nao e produto publicavel.

## Proximos passos seguros
1. Obter os campos exatos de identidade do aplicativo reservado no Partner Center. Nunca adivinhar o Publisher DN.
2. Empacotar com a identidade real; testar e validar com Windows App Certification Kit e recursos do Partner Center.
3. Distribuir via Microsoft Store assinada e checar se a politica de integridade de codigo do computador aceita o pacote. A Store nao garante excecao de uma politica corporativa administrada.
4. Se necessario, consultar o administrador da politica de Controle de Aplicativo sobre a assinatura/publisher permitido. **Nao desativar o controle nem criar excecoes inseguras apenas para executar o app.**
5. Repetir testes do aplicativo e de inicializacao de servidores com logs coletados e perfil real, apos o executavel estar autorizado a iniciar.

## Nomes legados mantidos intencionalmente
- URL dos dois repositorios atuais do GitHub (mundo-na-rede e mundo-na-rede-updates): exigem mudanca coordenada com links, releases, atualizador e configuracoes externas.
- Formatos antigos de backups, marcadores internos .mnr-*, pastas historicas de dados e referencias a backups do Google Drive: alteracao cega pode fazer dados antigos deixarem de ser reconhecidos.
- Executaveis e instaladores antigos arquivados: conservados como historico/rollback, nao usados no novo instalador.

## Alteracoes nao publicadas
As modificacoes locais de codigo e documentacao nao foram enviadas ao GitHub. Os trabalhos anteriores sem commit foram preservados.
