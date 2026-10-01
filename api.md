# API

Base: `/api/v1`. Todas as respostas são JSON (exceto erros de método, em texto puro).
CORS está aberto (`Access-Control-Allow-Origin: *`), adequado a protótipo local, não a produção.

## `GET /api/v1/health`

```json
{"status": "ok"}
```

## `GET /api/v1/candidatos?categoria=<cargo>&ano=<ano>`

Lista os candidatos do catálogo para o cargo. Não expõe notas nem marcadores do funil.

| Status | Quando |
|---|---|
| 200 | Lista de `{nome, partido}` (pode ser vazia) |
| 400 | `ano` ausente ou não numérico |
| 422 | Não há eleição para o cargo no ano, ou cargo desconhecido |

Calendário: Presidente, Governador, Senador, Deputado Federal e Deputado Estadual nos anos gerais (2022, 2026, 2030). Prefeito e Vereador nos municipais (2024, 2028).

## `POST /api/v1/analise`

### Request

```json
{
  "ano": 2026,
  "filtros": {
    "nao_neoliberal": true,
    "proibe_aborto": true,
    "respeita_cristianismo": true
  },
  "chapa": [
    {"cargo": "Presidente", "candidato": "Exemplo Presidente 1"},
    {"cargo": "Presidente", "candidato": "Exemplo Presidente 3"},
    {"cargo": "Prefeito",   "candidato": "Exemplo Prefeito 1"}
  ]
}
```

- `filtros` é opcional. Cada filtro omitido fica **ativo**.
- `chapa`: pode ter vários candidatos por cargo. Itens com `candidato` vazio são ignorados.
- Corpo limitado a 1 MB.

### Response 200

```json
{
  "ano": 2026,
  "dados_demonstrativos": true,
  "limiar_alta_afinidade": 95,
  "resultados": [
    {
      "cargo": "Presidente",
      "nome": "Exemplo Presidente 1",
      "partido": "Partido Fictício A",
      "status": "Alta afinidade",
      "eliminado": false,
      "score": 96,
      "criterios": {"qualidade_teses": 97, "coerencia_historica": 96},
      "justificativa": "Alta afinidade com os critérios do modelo (qualidade das teses e coerência histórica). Isto mede proximidade com o modelo, não capacidade de governar nem chance de vitória."
    },
    {
      "cargo": "Presidente",
      "nome": "Exemplo Presidente 3",
      "partido": "Partido Fictício C",
      "status": "Eliminado no funil",
      "eliminado": true,
      "motivos": ["classificado como neoliberal"],
      "score": null,
      "justificativa": "Eliminado na Etapa 1 (funil): classificado como neoliberal. A Etapa 2 não foi aplicada."
    }
  ],
  "ignorados": [
    {
      "cargo": "Prefeito",
      "candidato": "Exemplo Prefeito 1",
      "motivo": "Em 2026 não haverá eleição para Prefeito. As eleições municipais ocorrem em anos como 2024 e 2028."
    }
  ]
}
```

Campos:

- `resultados`: na ordem da chapa. `score` e `criterios` só existem para quem passou no funil (eliminados têm `score: null` e `motivos`).
- `ignorados`: itens com problema, com o motivo. Nada some em silêncio.
- `dados_demonstrativos`: `true` enquanto o catálogo for fictício. O frontend deve exibir esse aviso junto às notas.

### Erros

| Status | Quando | Corpo |
|---|---|---|
| 400 | JSON inválido | `{"error": "Payload inválido"}` |
| 400 | `ano` fora de 1900 a 9999 | `{"error": "Informe um ano eleitoral válido."}` |
| 400 | Nenhum candidato preenchido | `{"error": "Preencha ao menos um candidato na chapa."}` |

Cargo sem eleição no ano **não** gera erro HTTP: o item vai para `ignorados`, para que o resto da chapa seja analisado.
