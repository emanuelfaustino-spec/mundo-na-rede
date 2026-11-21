# Mundo Nexus — organização da versão para PC

## Pastas

| Área | Arquivos |
| --- | --- |
| Núcleo | `main.go`, `version.go`, `desktop*.go`, `process*.go` |
| Motores e downloads | `engine_library.go`, `java_runtime*.go`, `downloads.go`, `bedrock_setup.go`, `crossplay.go` |
| Servidores e mundo | `world_panel.go`, `pregeneration.go`, `packs.go`, `players*.go`, `schedules.go` |
| RAM e desempenho | `memory*.go`, `performance*.go`, `diagnostics.go` |
| Arquivos e backups | `files.go`, `backup_history.go`, `restore_preview.go`, `drive*.go` |
| Atualizações | `update.go`, `store/`, `installer/` |
| Interface | `ui/` |
| Testes e CI | `*_test.go`, `scripts/`, `.github/workflows/` |

Os arquivos Go continuam na raiz porque pertencem ao mesmo pacote `main`; movê-los sem separar as dependências criaria erros e dificultaria o suporte. Quando houver testes suficientes, separar pacotes internos por domínio será uma melhoria opcional.

## Biblioteca de motores

Os componentes são baixados **após aceite dos termos** na primeira configuração, não incluídos no instalador de forma indiscriminada. O cache fica em:

- `MundoNexus/engines/bedrock/<versão>`: ZIP oficial BDS extraído.
- `MundoNexus/engines/paper/<versão>/server.jar`: Paper estável com SHA-256 verificado.
- `MundoNexus/engines/java/25/`: Eclipse Temurin 25 portátil para Windows, sem instalação global.

Perfis diferentes reutilizam a biblioteca; mundos permanecem separados em `servers/<id>/game/`. Não apagar a biblioteca ou os mundos na atualização do aplicativo. A preparação da combinação Java + Bedrock requer componentes adicionais conforme o motor.

## Política de RAM

O orçamento do servidor usa a RAM total, a memória disponível e uma reserva para o sistema. O limite escolhido é verificado no momento da partida. No Windows, Job Objects restringem a memória comprometida do motor; Java usa `-Xmx` menor que o limite do processo para manter espaço para o próprio Java. A RAM do painel fica separada. O limite é uma proteção, não uma garantia de execução sem falhas.

## Compatibilidade

O aplicativo é para computador: Windows x64 primeiro, Linux experimental. Scripts móveis foram removidos das distribuições atuais. Perfis legados continuam legíveis e não são apagados automaticamente. A administração remota e um futuro aplicativo móvel estão fora do escopo desta versão.

## Próximas melhorias

Verificar motores em execução real, testar Paper + Temurin 25, Bedrock e crossplay; medir RAM e CPU com carga; cobrir restauração e atualizações do MSIX; ampliar testes de arquivos, instalação e versões incompatíveis. Complementos automáticos para Forge, Fabric, NeoForge, Quilt e Purpur exigem validação por motor antes de serem anunciados como prontos.
