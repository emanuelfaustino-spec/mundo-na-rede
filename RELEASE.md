# Processo de lançamento

O canal de atualização público permanece na versão estável até que uma compilação seja assinada com um certificado confiável. Não publique manualmente um executável não assinado como atualização automática.

## Antes do primeiro lançamento assinado

1. Crie uma tag `vX.Y.Z` a partir da `main` revisada.
2. Confirme que a execução **Compilar Windows** do GitHub Actions terminou com sucesso.
3. Quando o SignPath Foundation aprovar o projeto, conecte a política de assinatura ao workflow do GitHub Actions. A assinatura deve aceitar somente builds da `main` ou de uma tag de lançamento e exigir aprovação manual.
4. Gere o instalador com `iscc /DAppVersion=X.Y.Z /DSourceExe=caminho\Mundo-na-Rede.exe installer\MundoNaRede.iss`.
5. Verifique a assinatura com `Get-AuthenticodeSignature` antes de publicar.

## Publicação

Publique o instalador assinado e o executável assinado em um GitHub Release. Atualize então o `manifest.json` do repositório `mundo-na-rede-updates` com a versão, notas, URL HTTPS e SHA-256 do executável. O aplicativo só oferecerá uma atualização se a versão for maior e o hash corresponder.

## Reversão

Se houver qualquer problema, retire a versão do manifesto de atualização ou aponte o manifesto para a última versão assinada conhecida. Nunca substitua o arquivo de uma versão já publicada.
