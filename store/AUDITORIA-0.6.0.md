# Auditoria — Mundo Nexus 0.6.0
Data: 20/11/2026
Branch: `feature/pc-only-engines-memory-0.6.0`

## Feito nesta revisão
- Aplicativo de PC, sem fluxo de hospedagem Android na interface ou no pipeline de lançamentos.
- Compatibilidade mantida para perfis antigos; nenhum mundo do usuário foi modificado.
- Biblioteca única para Bedrock, Paper estável e Java 25 portátil (Windows), com aceite dos termos e uso de fontes oficiais.
- Checagem de SHA-256 nos downloads de Paper e Temurin, e arquivo Bedrock obtido pelo mecanismo já existente.
- Seleção do limite de RAM por servidor, cálculo conservador de reserva do PC e Job Object no Windows.
- RAM do servidor apresentada separadamente da memória disponível no PC.
- Menos consultas do painel quando ele está oculto.
- Código organizado por arquivos e responsabilidades documentadas em `docs/ARQUITETURA-PC.md`.
- Código de teste ampliado para cache, endpoints, Java, orçamento de memória e restrição de processo.

## Evidências
- `go vet ./...`: aprovado.
- `go test -c`: testes compilados com sucesso, mas execução local bloqueada pela política de Controle de Aplicativos do Windows.
- Verificações de sintaxe JavaScript e de integridade estrutural do painel: aprovadas.
- Metadados oficiais de PaperMC e Eclipse Adoptium: consultados, formatos confirmados.
- A execução real dos motores continua pendente: nenhum download foi iniciado sem aceite, nenhum firewall foi alterado.

## Antes de publicar
1. Executar testes em ambiente de CI e corrigir as falhas.
2. Validar Java 25 e Paper juntos após aceite dos termos; testar Bedrock, crossplay e consumo de memória real.
3. Validar o executável 0.6.0 instalado, atalhos, recuperação de dados e atualização para MSIX assinado.
4. Rodar os testes de certificação Windows App Certification Kit e revisar a ficha da Store.
5. Confirmar mercados e preço previsto de R$ 5,90. Não publicar o MSIX sem assinatura e certificação.

O instalador 0.6.0 é apenas artefato de revisão até a validação real. A edição atualmente instalada no PC não deve ser substituída automaticamente.
