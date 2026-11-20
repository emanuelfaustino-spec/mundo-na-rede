# Mundo Nexus — matriz de funcionalidades e próximos marcos
Data: 20/11/2026. Avaliação preliminar por inspeção do código, não equivale a testes de ponta a ponta.

## 1. Navegação e visual
**Parcial.** Existe navegação Início/Jogadores/Mundo/Backups/Detalhes, painel de performance e fundo espacial. Não existem ainda dez módulos independentes completos; redesenhar a arquitetura da navegação sem perder recursos atuais, filtros por motor, atalhos de teclado e acessibilidade. Revisar tema claro/escuro e responsividade.

## 2. Painel do servidor
**Parcial.** Iniciar/parar/reiniciar, lista de jogadores, LAN, QR Code, medições locais e console já existem em alguma forma. Criar diagnóstico pré-início, notificações resumidas, checagem de porta e histórico persistente; testar desempenho real.

## 3. Mundo e regras
**Parcial.** Perfil Bedrock, presets, opções de regras, importação de mundo e pré-geração existem. Revisar compatibilidade com versões atuais e risco de experimentos; ampliar suporte Java por motor (Paper/Purpur/Spigot/Fabric/Forge/NeoForge/Quilt) sem prometer integração não implementada.

## 4. Complementos
**Parcial.** Importação Bedrock, plugins/mods Java, validações pontuais de extensão e catálogo Modrinth. Faltam classificação abrangente, dependências, precedência por arrastar, desativação reversível e detecção de conflitos. Nunca instalar automaticamente arquivos de origem não confiável.

## 5. Jogadores e permissões
**Parcial.** Jogadores, lista de permitidos e comandos existem. Faltam papéis completos, auditoria persistente por administrador e controle remoto delegado. Separar recursos por tipo de motor.

## 6. Console e diagnóstico
**Parcial.** Console, logs recentes e comandos estão presentes. Faltam análise estruturada de falhas, guia por erro, coleta de diagnóstico redigido sem segredos e histórico pesquisável persistente.

## 7. Backups
**Parcial com correção 0.4.5.** ZIP, histórico local, restauração e Drive foram encontrados. Rotação corrigida para 0..7 cópias e nomes novo/antigo aceitos. Faltam distinção entre backup manual e automático na retenção, lixeira temporária, validação independente de integridade e testes completos de recuperação.

## 8. Rede
**Parcial.** Endereços LAN, porta e opções para acesso externo já existem. Testes de acessibilidade externa, latência e qualidade de conexão precisam de metodologia específica, sem expor portas ou recursos automaticamente.

## 9. Automação
**Parcial.** Agendas e captura de backup ao parar aparecem no código. Validar em cenários reais; iniciar com Windows, alertas, reinício seguro e pause/resume exigem trabalho adicional.

## 10. Conta e painel remoto
**Não implementado como sistema completo.** Há integração Google Drive, mas não há evidência de autenticação de dispositivos, controle remoto seguro, convites por permissões, auditoria de ações ou serviço de conexão de saída com autenticação. Requer arquitetura e revisão de segurança próprias antes da publicação.

## Sequência recomendada
**Marco A — estabilidade e segurança (antes da Store):**
1. Testar servidor Bedrock real, Java, importação e restauração em ambiente limpo.
2. Corrigir erros de execução e bloqueio por política de assinatura, sem desabilitar segurança.
3. Testar backup automático/manual, rollback, carga e conflitos de porta.
4. Certificar MSIX assinado, atualizar instalação e testar desinstalação sem perda de dados.
5. Aumentar testes em fluxos críticos. Cobertura registrada de 9,9% no último ensaio.

**Marco B — experiência de uso:** implementar dez áreas de navegação progressivamente, priorizando responsividade, filtros por motor, console e ajuda.

**Marco C — recursos avançados:** complementos, loaders Java, diagnóstico de rede, automações e backups aperfeiçoados.

**Marco D — painel remoto:** desenhar identidade, consentimento, papéis, revogação, autenticação e comunicação cifrada, com revisão de segurança independente.

## Limites da revisão atual
O aplicativo instalado e aberto é 0.4.4. A versão 0.4.5 foi compilada para MSIX e Inno Setup, sem assinatura, instalação ou teste completo de motores. A publicação deste código no GitHub não publica automaticamente a versão na Microsoft Store.
