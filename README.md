# Match Eleitoral Analítico

Protótipo de uma ferramenta em que o usuário monta uma **chapa** (um ou mais candidatos por cargo) e recebe, para cada nome, uma medida de **afinidade com um modelo declarado**: o nacionalismo desenvolvimentista associado a Enéas Carneiro.

> **Aviso importante**
> - Todos os candidatos e notas do catálogo são **fictícios**. Nada aqui é avaliação de pessoa ou partido reais.
> - O resultado mede **proximidade com o modelo**. Não mede capacidade de governar, competência técnica nem chance de vitória. "Alta afinidade" não significa "apto a governar".

## Como funciona (resumo)

1. **Etapa 0, chapa:** o usuário informa candidatos por cargo. Campos vazios são ignorados.
2. **Etapa 1, funil:** três filtros eliminatórios (não neoliberal, proíbe o aborto, respeita/incentiva o cristianismo). Vêm ativos por padrão e podem ser desligados. Quem é eliminado aparece com o motivo.
3. **Etapa 2, pontuação:** quem passou recebe nota de 0 a 100 (qualidade das teses e coerência histórica, peso igual).
4. **Resultado:** `Alta afinidade` a partir de 95, `Afinidade parcial` a partir de 70, `Baixa afinidade` abaixo disso.

Detalhes e limites do modelo: [docs/metodologia.md](docs/metodologia.md).

## Estado da implementação

| Parte | Estado |
|---|---|
| Backend (`main.go`): chapa, funil, pontuação, API | Implementado conforme esta documentação. **Escrito sem acesso a um compilador Go: rode `go test ./...` antes de confiar nele.** |
| Testes (`main_test.go`) | Cobrem pontuação, limiares, funil, chapa e handlers. |
| Catálogo | Fictício e em memória. Independe de ano e município. |
| Etapa 2 | Notas **manuais**. A rubrica verificável está proposta, não implementada. |
| Frontend (`static/index.html`) | **Incompatível com este backend.** O arquivo atual chama `/api/v1/auditoria` com outros nomes de campo (`excluir_neoliberal` etc.). É preciso escolher um contrato e alinhar os dois lados. Veja [docs/arquitetura.md](docs/arquitetura.md). |

## Executar

```bash
go test ./...
go run .            # http://localhost:8080  (variável PORT altera a porta)
```

Exemplo rápido da API:

```bash
curl -X POST localhost:8080/api/v1/analise \
  -H 'Content-Type: application/json' \
  -d '{"ano":2026,"chapa":[{"cargo":"Presidente","candidato":"Exemplo Presidente 1"}]}'
```

## Documentação

- [docs/metodologia.md](docs/metodologia.md): o que o modelo mede, parâmetros, limitações.
- [docs/api.md](docs/api.md): endpoints, exemplos e códigos de status.
- [docs/arquitetura.md](docs/arquitetura.md): componentes reais e o que falta no frontend.

## Licença

Veja [LICENSE](LICENSE).
