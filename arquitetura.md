# Arquitetura

## Componentes reais

```text
Navegador ──(HTML/CSS/JS puro)──> static/index.html
    │
    └── fetch /api/v1/* ──> main.go (net/http, Go 1.22)
                              ├── catálogo fictício em memória
                              ├── Etapa 1: funil (3 filtros)
                              └── Etapa 2: pontuação (2 critérios)
```

- **Backend:** um único arquivo, `main.go`, só com a biblioteca padrão. Sem banco, sem autenticação, sem serviços externos.
- **Frontend:** `static/index.html`, servido pelo próprio backend.
- **Sem Google Cloud Storage, Google Search API ou concorrência.** Versões anteriores da documentação citavam esses itens, mas eles nunca existiram no código.

## Frontend e contrato da API: pendência

Existem dois contratos diferentes no repositório, e é preciso escolher um:

- **Este backend (`main.go`, `docs/api.md`):** `POST /api/v1/analise`, com `filtros.nao_neoliberal`, `proibe_aborto`, `respeita_cristianismo` e `chapa[{cargo, candidato}]`.
- **O `static/index.html` e o `docs/api-especificacao.md` encontrados na pasta:** `POST /api/v1/auditoria`, com `filtros.excluir_neoliberal`, `exigir_proibicao_aborto`, `exigir_respeito_cristianismo`, e uma resposta com outros campos (`afinidade`, `limiar`, status como `Abaixo do limiar de afinidade`).

Enquanto isso não for resolvido, o frontend não funciona com este backend. Qualquer um dos lados pode ceder, e o importante é que sobre **um** contrato documentado. O que o frontend precisa mostrar, seja qual for o contrato: a chapa por cargo, os três filtros (padrão ligados), os eliminados com motivo, os itens ignorados com motivo, e o aviso de dados fictícios junto de cada nota.

## Decisões de projeto

- **Backend decide, frontend apresenta.** Limiares e status vivem só no backend para não divergirem.
- **Sem notas na listagem de candidatos.** `GET /candidatos` devolve nome e partido, para o usuário não ver a nota antes de escolher.
- **Eliminar sem sumir.** Todo candidato retirado (funil ou erro de entrada) aparece com motivo.
