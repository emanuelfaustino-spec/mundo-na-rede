# Mundo na Rede

Gerenciador independente de servidores Minecraft. Cada instalação executa o próprio motor no aparelho do dono do mundo. O painel abre localmente e os jogadores se conectam diretamente ao endereço LAN do anfitrião. Cada aparelho define sua porta, limite de jogadores e, quando usa Java, o limite de memória da JVM.

## Política de assinatura de código

As versões oficiais para Windows são compiladas a partir deste repositório público pela [automação de build](.github/workflows/build-windows.yml) e publicadas pela equipe do Mundo na Rede. Quando a inscrição for aprovada, a assinatura de código gratuita será fornecida por SignPath.io, com certificado da SignPath Foundation.

- **Responsável por alterações e revisões:** [emanuelfaustino-spec](https://github.com/emanuelfaustino-spec), proprietário e mantenedor deste repositório. Contribuições externas devem ser revisadas antes de serem incorporadas.
- **Responsável por aprovar assinaturas de lançamento:** [emanuelfaustino-spec](https://github.com/emanuelfaustino-spec). Cada lançamento assinado será aprovado manualmente.
- **Privacidade:** o aplicativo não transfere informações para serviços externos sem ação do usuário. Ele pode baixar motores ou atualizações quando solicitado, consultar o Modrinth quando o catálogo for usado e enviar backups ao Google Drive somente após o usuário conectar e autorizar a própria conta. Os mundos, perfis e backups locais permanecem no dispositivo do anfitrião.

## Windows

1. Abra o instalador oficial e inicie **Mundo na Rede** pelo menu Iniciar ou pelo atalho criado. O painel abre em uma janela própria; o gerenciador e o motor do mundo executam sem exigir uma janela de console aberta.
2. Crie um perfil **Bedrock oficial**. Consulte a EULA e os termos no link **Obter motor**. Use **Instalar BDS oficial** para baixar da Mojang, ou importe o ZIP oficial que você baixou para Windows.
3. Clique em **Iniciar**. No Minecraft Bedrock de outro aparelho conectado à sua LAN, adicione o servidor com o IP e a porta exibidos. Caso o firewall peça permissão, escolha a rede privada. Se houver vários IPs no painel, use o da interface conectada à mesma LAN que o jogador.
4. Para desligar, use **Parar** e aguarde o estado **parado**. Use o console para consultar comandos e logs. Fechar a janela da interface deixa o gerenciador rodando.

O arquivo Windows é para x64. `mundo-na-rede-linux-amd64` também é fornecido para testes no Ubuntu.

## Android ARM64 com Termux

A edição experimental roda no próprio celular mediante Termux. Instale uma versão mantida do [Termux](https://github.com/termux/termux-app#installation). Extraia o ZIP desta edição para a pasta Downloads e, no Termux, execute:

```sh
termux-setup-storage
pkg install openjdk-21
mkdir -p ~/mundo-na-rede
cp ~/storage/downloads/mundo-na-rede-android-arm64 ~/storage/downloads/iniciar-android.sh ~/mundo-na-rede/
cd ~/mundo-na-rede
chmod +x mundo-na-rede-android-arm64 iniciar-android.sh
./iniciar-android.sh
```

Abra o URL impresso no Termux se o painel não abrir automaticamente. Crie um perfil **PowerNukkitX**, baixe o JAR dos [releases oficiais](https://github.com/PowerNukkitX/PowerNukkitX/releases), importe pelo painel e inicie. Outro aparelho deve ingressar pelo IP LAN do Android. O funcionamento do PowerNukkitX e do Java pode variar conforme a compilação e o aparelho. O Android pode suspender o Termux com a tela desligada. Esta edição foi compilada para Android ARM64 e o motor foi testado em Linux, portanto a execução e a conexão em Android real ainda precisam ser verificadas.

## Recursos e dados

- Perfis independentes, status, tempo de execução, console, comandos de jogadores, reinício e tarefas com intervalo enquanto o servidor está online.
- BDS para Bedrock oficial em PCs compatíveis. PowerNukkitX para Bedrock com seu próprio formato e ecossistema. JAR executável para servidores Java compatíveis. Alguns loaders Java exigem instalação prévia ou têm um método próprio de iniciar.
- Importação de `.mcworld` ou ZIP Bedrock com `level.dat` e `db/` na raiz. Importe o motor antes do mundo. O destino padrão do BDS é `worlds/Mundo`. A restauração preserva a pasta anterior ao lado da pasta ativa.
- Backup ZIP do diretório do motor com o servidor parado, para restaurar no mesmo tipo de perfil e plataforma. O backup pode incluir arquivos privados do servidor, como chaves de identidade. Guarde-o para uso próprio. Limite de 2 GiB por ZIP e 6 GiB descompactados.
- Importação de configuração e de extensões. Pacotes do Bedrock oficial são registrados no mundo ativo ao iniciar. No PowerNukkitX, pacotes de recursos são copiados para a pasta do motor. Use apenas executáveis, plugins e backups confiáveis.
- O painel usa `127.0.0.1` e uma chave por sessão. O IP e a porta exibidos para o Minecraft são os do **jogo**, independentes da porta do painel. A instalação não hospeda os mundos de outros donos.
- Os caminhos padrão de dados são `%AppData%\MundoNaRede` no Windows e `~/.config/MundoNaRede` nos demais sistemas. A opção `-data` permite usar outra pasta. O painel usa a porta local 37654 por padrão.

## Situação dos recursos de referência

A versão cobre o fluxo local e o gerenciamento básico no anfitrião. A paridade com o Anvil-MC ainda depende de APK independente, administração por outro dispositivo, catálogo Modrinth e outros instaladores por motor, mapa, métricas em tempo real, validação do backup em nuvem, voz, diagnóstico avançado e acesso pela internet. Um túnel público exigiria que cada anfitrião configurasse seu próprio acesso externo ou usasse um serviço com termos e limites próprios. A LAN funciona sem contratar hospedagem, e o download inicial dos motores precisa de internet.

## Compilação

Requer Go 1.25 ou superior. Execute `go build .` para a plataforma atual. Os binários Windows e Android usam, respectivamente, `GOOS=windows GOARCH=amd64` e `GOOS=android GOARCH=arm64`, com `CGO_ENABLED=0`. Nenhum motor Minecraft está incluído nos arquivos de distribuição.

## Google Drive em desenvolvimento

Na primeira conexão pelo aplicativo desktop, abra **Conta → Configurar Google Drive neste aplicativo**. Crie no [Google Cloud Console](https://console.cloud.google.com/) um projeto seu, ative a **Google Drive API** e crie um cliente OAuth do tipo **Aplicativo para computador**. Cole o ID e o segredo do cliente no app. Eles são salvos somente na pasta local de dados do Mundo na Rede, nunca no repositório e nunca no backup. Esta etapa é exigida pelo Google para que cada distribuição do aplicativo não compartilhe uma credencial privada.

No Windows e no Linux, defina `MUNDO_GOOGLE_CLIENT_ID` e `MUNDO_GOOGLE_CLIENT_SECRET` no ambiente do aplicativo com um cliente OAuth do tipo Desktop em um projeto Google Cloud com Drive API ativada. Depois abra **Conta**, selecione **Conectar Google** e autorize no navegador. A integração usa `drive.file`. O app reconhece a conta pelo nome e e-mail devolvidos pelo Google. Se a conta for perdida, use o link para a recuperação oficial do Google.

Ao encerrar normalmente um servidor que ficou online, o app cria um ZIP e tenta enviá-lo para a conta conectada. Se a rede falhar, a cópia fica em uma fila local associada àquela conta e o aplicativo tenta novamente enquanto estiver aberto. Para restaurar, encontre os arquivos MundoNaRede no Drive e importe o ZIP em um perfil correspondente. Esta configuração é de desenvolvimento, ainda sem credenciais públicas nem validação com uma conta real. No Android, use o fluxo com rclone abaixo.

## Edições, extensões e pausa por inatividade

Em **Quem vai jogar?**, escolha Bedrock, Java ou Java e Bedrock. O modo misto instala Paper 1.21.11, Geyser, Floodgate e ViaVersion. Use Java 21 ou superior, aceite a EULA e crie um perfil novo. Os jogadores compartilham um mundo Java. A porta Java é a do perfil, e a do Bedrock é 19132. A compatibilidade com versões novas do cliente depende dos plugins de tradução. A lista de jogadores, as permissões e os comandos são do servidor Java.

No Bedrock oficial, importe um pacote de recursos, pacote de comportamento ou add-on. O aplicativo lê os manifestos, registra os pacotes importados no mundo e mantém os registros de pacotes que já estavam associados. No Java, os mods e plugins são importados nas respectivas pastas e exigem um motor compatível com o formato. O perfil misto aceita plugins Paper.

Servidores Java compatíveis recebem `pause-when-empty-seconds=900`. A pausa nativa interrompe a simulação de um servidor vazio mantendo a escuta de rede. Ela pode reduzir trabalho de CPU, mantém o mundo na memória e depende do suporte do motor e dos plugins. O Bedrock oficial e o PowerNukkitX ainda não têm essa opção implementada no Mundo na Rede.

## Limite de jogadores, extensões e acesso pela internet

Em **Painel do mundo**, ajuste nome, modo, porta, memória e limite de jogadores de um perfil existente. A sugestão inicial usa a memória disponível e os núcleos do anfitrião como ponto de partida. Cabe ao anfitrião testar o desempenho real, especialmente o upload para jogadores remotos. O limite editável vai de 1 a 100.

O painel lista add-ons, mods e plugins instalados. Para adicionar mais a um mundo existente, pare o servidor e importe a extensão compatível. O catálogo consulta o Modrinth para mods ou plugins com suporte a servidor e abre a página do projeto para baixar a versão correta.

Use **Amigos fora da sua rede** para registrar um endereço de jogo já configurado no roteador ou obtido de um túnel. O painel administrativo continua local. A hospedagem do mundo fica no anfitrião e as condições de rede e de serviços de túnel variam. O aplicativo não provisiona o túnel. No Android, o Google Drive pode ser conectado via rclone no Termux.

### Google Drive no Android via Termux

A edição atual usa o cliente rclone no próprio aparelho. Instale com `pkg install rclone`. Execute `rclone config`, crie um remoto chamado `mndrive`, escolha Google Drive e o escopo `drive.file`, e autorize a conta no navegador. O rclone administra a sessão e pode solicitar autorização num computador dependendo do dispositivo. Inicie o app com `MUNDO_DRIVE_REMOTE=mndrive ./iniciar-android.sh` e selecione **Conta → Vincular pasta do Drive**. Use o mesmo remoto nas próximas vezes que abrir o app.

O app grava na pasta `MundoNaRede` do Drive após uma parada normal. Em **Conta**, selecione uma cópia compatível para restaurar no perfil aberto. Pare o servidor antes de restaurar. A restauração mantém a pasta anterior. A desconexão no app mantém as cópias no Drive e o login do rclone no Termux. Com escopo `drive.file`, clientes diferentes podem ter visibilidade diferente. Use o Drive diretamente para baixar cópias antigas ou feitas com outra integração e importe o ZIP manualmente. A lista de restauração no painel inclui arquivos da versão atual identificados pelo tipo de servidor.

## Primeira partida e restauração no Windows

O botão **Iniciar** abre **Vamos preparar seu servidor** no primeiro uso. Escolha quem pode entrar. Se restringir o acesso, adicione pelo menos uma gamertag. Aceite os termos para instalar o BDS. Quando o Windows pedir autorização, permita o BDS em redes privadas. O app guarda downloads oficiais por versão na biblioteca e reaproveita os arquivos para preparar outros perfis. Cada perfil mantém uma pasta de execução própria para isolar o mundo e as configurações.

Em **Backups**, é possível exportar e consultar as cópias locais. Para recuperar, use **Restaurar backup**, selecione o ZIP e revise a prévia. Se faltar o motor, escolha uma versão na biblioteca, reutilize o motor do próprio perfil ou de outro perfil, ou prepare o motor depois. A restauração conserva os arquivos antigos em uma pasta separada. Para restaurar um ZIP com o executável incluído, use a mesma plataforma em que foi criado.

Em **Mundo e add-ons**, a lista normal exibe os pacotes importados no Bedrock oficial. Use a seção de diagnóstico para ver os arquivos internos. É possível importar um add-on antes do primeiro início e escolher uma capa para o painel. **Avançado** mostra o tamanho do mundo, as distâncias configuradas e o histórico recente de CPU quando disponível. A CPU do servidor é uma fração da capacidade total do anfitrião.
