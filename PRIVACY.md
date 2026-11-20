# Política de Privacidade — Mundo Nexus

Última atualização: 20 de novembro de 2026.

## Dados locais

O Mundo Nexus funciona principalmente no dispositivo do anfitrião. Perfis de servidores, configurações, mundos, logs e backups ficam na pasta local de dados do aplicativo. O projeto não possui um servidor próprio para receber esses dados e não usa publicidade ou telemetria comportamental.

## Conexões iniciadas pelo usuário

Na edição tradicional, o aplicativo verifica automaticamente se há atualizações no canal público do GitHub quando o painel é iniciado. A verificação pode transmitir dados técnicos de conexão, como o endereço IP, ao serviço consultado. A edição distribuída pela Microsoft Store utiliza o mecanismo de atualização da própria loja, sem acionar o atualizador direto do aplicativo.

Outras conexões externas ocorrem quando o usuário utiliza funções que dependem delas:

- baixar motores ou componentes de servidor;
- consultar o catálogo Modrinth;
- conectar uma conta Google Drive e enviar/restaurar backups (inclusive as cópias automáticas após a autorização);
- abrir links oficiais no navegador.

Na integração com Google Drive, a autorização é feita pelo fluxo OAuth do Google. O token de acesso e os dados básicos de identificação devolvidos pela conta ficam somente na área local de dados do aplicativo. Backups só são enviados para a conta que o próprio usuário autorizar.

## Rede do jogo

Quando um mundo é iniciado, o dispositivo do anfitrião aceita conexões de jogadores na rede escolhida pelo anfitrião. O endereço de jogo é mostrado localmente no painel. O Mundo Nexus não oferece, por si só, um serviço intermediário de hospedagem, túnel ou coleta dos dados dos jogadores.

## Controle do usuário

O usuário pode remover perfis, backups locais e a conexão do Drive pelo aplicativo ou apagar a pasta local de dados. Cópias já enviadas ao Drive permanecem sob o controle da conta Google do usuário e podem ser excluídas pelo Drive.

## Contato

Para questões sobre esta política, abra uma issue no [repositório do projeto](https://github.com/emanuelfaustino-spec/mundo-na-rede/issues).
