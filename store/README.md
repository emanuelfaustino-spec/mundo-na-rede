# Publicação na Microsoft Store

O Mundo Nexus será distribuído pela Microsoft Store como um pacote **MSIX x64**. Assim, a Microsoft assina o pacote da loja e entrega atualizações pelo próprio Windows; não é necessário manter o atualizador GitHub para quem instalou pela Store.

## Antes de empacotar

1. Entre no [Partner Center](https://partner.microsoft.com/dashboard/home) com a conta Microsoft que será dona do aplicativo.
2. A reserva **Mundo Nexus** já foi identificada no Partner Center com a ID `9N0WMH45BN3C`.
3. Copie da página de identidade do produto os três valores abaixo:
   - `Package/Identity name`;
   - `Publisher`;
   - `Publisher display name`.
4. Execute `store/build-msix.ps1` com os três valores oficiais. O script preenche os marcadores `__STORE_*__` durante a geração, sem modificar o manifesto modelo. Consulte `IDENTIDADE-STORE.md` para os valores verificados.

Os valores da identidade são atribuídos pela Microsoft. Não devem ser inventados nem reutilizados de outro produto.

## Conteúdo do pacote

- `Mundo-Nexus.exe` — aplicativo desktop Win32, executado como aplicativo de confiança total dentro do MSIX;
- os quatro logotipos obrigatórios em `Assets/`, exportados do ícone mestre em resolução nítida;
- `AppxManifest.xml` preenchido com a identidade reservada.

## Checklist da ficha da loja

- Nome: **Mundo Nexus**;
- Categoria: produtividade/utilitários (a seleção final é feita no Partner Center);
- Idioma principal: Português (Brasil), com descrição em inglês adicionada em seguida;
- Política de privacidade: `https://github.com/emanuelfaustino-spec/mundo-na-rede/blob/main/PRIVACY.md`;
- Site de suporte: `https://github.com/emanuelfaustino-spec/mundo-na-rede/issues`;
- Aviso de não afiliação à Mojang Studios e à Microsoft;
- capturas reais do aplicativo e uma descrição que não prometa hospedagem externa;
- preço a confirmar antes da submissão. A Store converte o valor para as moedas dos demais mercados.

## Validação local

Depois de receber a identidade, gerar o pacote com `makeappx.exe`, instalar apenas para teste local e executar o Windows App Certification Kit (`appcert.exe`). Não enviar o pacote se essa validação acusar erro.

## Publicação

A criação da submissão, definição do preço e o clique de envio são ações públicas/financeiras. Só devem ocorrer após revisão final da ficha, dos arquivos e confirmação explícita do proprietário da conta.
