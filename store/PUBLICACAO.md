# Microsoft Store — Mundo Nexus

## Estado verificado em 20/11/2026
- O codigo local declara a versao 0.4.4. Existem alteracoes locais ainda nao enviadas ao GitHub; preserve-as.
- Testes `go test ./...` e `go vet ./...` passaram.
- Windows SDK 10.0.26100.0 e Go estao instalados.
- `store/build-msix.ps1` compilou e gerou com sucesso o MSIX de validacao com identidade ficticia.
- **O MSIX de validacao NAO e publicavel.** Nao foi assinado, instalado, certificado pelo WACK ou submetido.

## Identidade do Partner Center
Abra a reserva do aplicativo **Mundo Nexus** em https://partner.microsoft.com/dashboard/home e obtenha, exatamente como aparecem em Product identity:
1. Package/Identity Name
2. Publisher (valor completo do DN, por exemplo CN=..., sem adicionar CN= outra vez)
3. Publisher display name

Nao tente adivinhar os valores. O modelo `AppxManifest.template.xml` tem marcadores intencionais.

## Gerar o MSIX final
Execute PowerShell no diretorio do projeto (politica de execucao apenas para este processo):

```powershell
powershell.exe -NoProfile -ExecutionPolicy Bypass -File .\store\build-msix.ps1 -IdentityName 'IDENTITY_NAME_EXATO' -Publisher 'PUBLISHER_DN_EXATO' -PublisherDisplayName 'NOME_DO_PUBLICADOR'
```

O script executa testes e analise estatica, compila o EXE diretamente do codigo atual, gera os icones exigidos a partir de `ui/icon.png`, atualiza a identidade e a versao no manifesto, empacota com MakeAppx e imprime o SHA-256. A saida fica em `dist/`. O pacote da Microsoft Store pode ser enviado sem assinatura local, desde que se use a identidade correta e sejam cumpridas as exigencias da Store.

## Antes de enviar
- Confirmar que a reserva pertence a esta conta do Partner Center; verificar o campo Publisher completo.
- Verificar o numero de versao e o nome do executavel.
- Fazer instalacao em ambiente de testes e executar o Windows App Certification Kit; corrigir falhas e repetir. A geracao MSIX sozinha nao substitui certificacao.
- Verificar se o aplicativo escreve apenas em caminhos permitidos no ambiente MSIX, inclusive dados do usuario, logs, downloads e arquivos de mundo.
- Testar na pratica: inicializacao, criacao/abertura de mundo, caminhos de dados, networking, atualizacoes e desinstalacao.
- Revisar ficha comercial, classificacao etaria, screenshots reais, descricao e politica de privacidade.
- Confirmar preco e territorios antes da submissao.
- Registrar aprovacao do proprietario antes de enviar para certificacao.

## Atualizacoes
- Instalacoes feitas pela Store recebem as atualizacoes distribuidas pela Store; nao depender do atualizador GitHub da versao instalavel fora da Store.
- Em cada lancamento: atualizar `appVersion` em `version.go`; executar testes; gerar novo MSIX com **mesmo Identity Name e Publisher** e com versao maior; fazer testes de instalacao/atualizacao; criar nova submissao de pacote no Partner Center.
- Os workflows GitHub existentes continuam destinados a compilacoes/testes e ao canal de distribuicao fora da Store. A simples geracao de novos arquivos no GitHub nao publica uma atualizacao na Microsoft Store.
- Nunca incluir credenciais do Partner Center ou certificados privados no repositorio.
