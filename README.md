# Mundo Nexus

Gerenciador de servidores Minecraft para **computador**. Interface simples, ferramentas avançadas opcionais e motores compartilhados entre os mundos. O projeto é independente de Mojang, Microsoft, PaperMC e Eclipse Adoptium.

## Plataformas

- **Windows x64:** aplicativo de desktop com WebView2, instalador Inno Setup e pacote MSIX para futura publicação na Microsoft Store.
- **Linux x64:** compilação experimental com painel no navegador local.
- **Celular:** nenhum aplicativo móvel é distribuído nesta edição. Jogadores Bedrock podem participar pelo celular; o servidor continua hospedado no PC.

## Primeira utilização

1. Instale o Mundo Nexus e abra pelo menu Iniciar.
2. Na **Biblioteca de motores**, leia os termos e autorize a preparação dos componentes necessários.
3. Crie um mundo e escolha **Bedrock**, **Java (Paper)** ou **Java + Bedrock** (opção avançada).
4. Selecione o limite de RAM sugerido e clique em **Iniciar**.

O aplicativo baixa os componentes de suas fontes oficiais **somente após autorização**. Depois do primeiro download, reutiliza as cópias locais: Bedrock, Paper e uma versão portátil do Eclipse Temurin Java 25 no Windows. Não instala o Java globalmente. Componentes extras de crossplay e motores importados podem precisar de preparação adicional.

Os downloads não são incluídos indiscriminadamente no instalador: cada software possui licença, tamanho e requisitos próprios. A instalação inicial precisa de internet. As cópias dos motores ficam na pasta de dados, separadas dos executáveis do Mundo Nexus, para não serem apagadas numa atualização.

## Memória e desempenho

O limite de RAM é definido **por servidor**, não para o painel. O Mundo Nexus calcula uma recomendação baseada na memória livre, reserva capacidade para o sistema e verifica novamente antes de iniciar. No Windows, o processo do servidor também recebe um limite por Job Object. O painel apresenta RAM livre do PC e o uso medido do servidor separadamente.

O limite é uma proteção, não garantia absoluta contra falhas: mods, mundos e programas externos podem consumir mais recursos; o servidor pode encerrar se precisar de memória acima do limite. Reduza distância de simulação e complementos quando necessário.

## Áreas

**Painel**, **Mundo**, **Complementos**, **Jogadores**, **Console** e **Backups**. Outras opções ficam em **Mais → Rede, Automação, Conta e Ajuda**. O modo avançado está desligado inicialmente. O painel administrativo usa endereço local e chave de sessão; nenhum acesso remoto é habilitado automaticamente.

## Arquivos

- `ui/`: interface, estilos, traduções, imagens.
- `*.go`: núcleo do aplicativo e módulos nomeados por função; o pacote principal permanece na raiz para manter a compilação estável.
- `installer/`: instalador Windows; `store/`: identidade, scripts MSIX e relatórios.
- `scripts/`: validações; `.github/workflows/`: integração contínua.
- `dist/`: saídas de compilação, não incluídas nos commits.
- `%APPDATA%\MundoNexus\` (Windows): perfis, mundos, motores e backups; a migração de pastas antigas é preservada.

## Desenvolvimento

Requer Go 1.25 ou superior e Node.js para as verificações da interface. Executar:

```powershell
go test ./...
go vet ./...
node scripts/verify-ui.cjs
go build -o dist/Mundo-Nexus.exe .
```

Para a Microsoft Store, o MSIX usa a identidade registrada no Partner Center; o pacote local ainda precisa de certificação e aprovação. Consulte `store/IDENTIDADE-STORE.md` e `store/PUBLICACAO.md`.

Política de privacidade: [PRIVACY.md](PRIVACY.md). As cópias manuais são preservadas até exclusão pelo usuário. Os dados e motores nunca devem ser removidos automaticamente pelo desinstalador. O código das versões anteriores continua no histórico Git.
