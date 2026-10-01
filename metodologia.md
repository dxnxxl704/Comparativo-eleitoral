# Metodologia

## 1. O que o sistema mede e o que não mede

**Mede:** a proximidade de um candidato com um modelo declarado (nacionalismo desenvolvimentista inspirado em Enéas Carneiro), por meio de três filtros e dois critérios descritos abaixo.

**Não mede:**

- capacidade técnica de governar;
- viabilidade eleitoral (chance de vitória);
- mérito pessoal ou probidade.

Por isso o sistema nunca usa as palavras "apto", "aprovado" ou "eleito". Ter alta afinidade com uma tese não prova competência para executá-la, e a história eleitoral mostra que votos também não medem mérito.

## 2. Fluxo

### Etapa 0: chapa

O usuário informa um ou mais candidatos por cargo (Presidente, Governador, Senador, Prefeito, Deputado Federal, Deputado Estadual, Vereador). Cada candidato é auditado de forma independente. Entradas com candidato vazio são ignoradas. Entradas com problema (cargo vazio, cargo sem eleição no ano, nome fora do catálogo) aparecem em `ignorados`, com o motivo.

### Etapa 1: funil eliminatório

Três filtros booleanos. Os três começam **ativos** (padrão da especificação) e cada um pode ser desligado na requisição. O candidato é eliminado se falhar em **qualquer** filtro ativo, e todos os motivos aparecem.

| Filtro (request) | Campo no catálogo | Elimina se | Definição operacional |
|---|---|---|---|
| `nao_neoliberal` | `Neoliberal` | `true` | **A definir.** Sugestão: votos ou propostas verificáveis a favor de privatização ampla de setores estratégicos ou de Estado mínimo. |
| `proibe_aborto` | `ProibeAborto` | `false` | **A definir.** Sugestão: declaração ou voto em projetos contra a ampliação das hipóteses legais de aborto. |
| `respeita_cristianismo` | `RespeitaCristianismo` | `false` | **A definir.** Sugestão: conduta pública verificável (votos, declarações), nunca filiação religiosa do candidato. |

Observações:

- Os filtros expressam **valores**, não fatos. Servem como preferência de quem usa o sistema, por isso são ligáveis e desligáveis.
- Sem definição operacional escrita e fonte por candidato, o valor de cada booleano é só a opinião de quem preencheu o catálogo.

### Etapa 2: pontuação

Só quem passou no funil é pontuado, em dois critérios de 0 a 100:

| Critério | Pergunta | Peso |
|---|---|---|
| `qualidade_teses` | As teses e argumentos são válidos, coesos e de boa qualidade? | 50 |
| `coerencia_historica` | O perfil e a história do candidato são coerentes com as propostas? | 50 |

`score = (qualidade × 50 + coerência × 50) / 100`, com divisão inteira (trunca). Exemplo: 97 e 96 resultam em 96.

### Etapa 3: processamento

Sequencial, na ordem da chapa. A conta é aritmética sobre dados em memória, então paralelismo (goroutines) só aumentaria a complexidade. Se um dia a Etapa 2 depender de buscas externas, aí vale reavaliar.

### Etapa 4: resultado

| Condição | Status |
|---|---|
| Falhou em algum filtro ativo | `Eliminado no funil` (com `motivos`; `score` é `null`) |
| `score ≥ 95` | `Alta afinidade` |
| `70 ≤ score < 95` | `Afinidade parcial` |
| `score < 70` | `Baixa afinidade` |

O limiar de **95** vem da especificação. O de **70** é uma escolha de protótipo, sem fundamento empírico.

## 3. Parâmetros

Ficam como constantes no início do `main.go`: `limiarAltaAfinidade`, `limiarAfinidadeParcial`, `pesoQualidadeTeses`, `pesoCoerenciaHistorica` e `catalogoDemonstrativo`.

## 4. Rubrica proposta para a Etapa 2 (não implementada)

Hoje as notas são números digitados à mão, o que as torna subjetivas e irreproduzíveis. Para que dois avaliadores cheguem à mesma nota, cada critério deveria virar itens de **sim/não verificáveis**, cada um com fonte. Exemplos:

- `qualidade_teses`: cita a fonte de financiamento da proposta; define meta mensurável com prazo; indica indicador de acompanhamento; reconhece custos ou riscos.
- `coerencia_historica`: votou ou agiu de forma compatível com a proposta que defende; manteve a mesma posição em projetos relacionados; as promessas anteriores têm registro de cumprimento.

Princípio: só entra na nota o que é um **fato verificável** (voto, projeto, meta, financiamento). Impressões, popularidade, tratamento recebido na imprensa e trechos de entrevista editados não entram.

## 5. Limitações conhecidas

- **Catálogo fictício e manual.** Nenhum número tem fonte.
- **Resultado majoritariamente vazio.** Três filtros mais limiar de 95 eliminam quase todos os nomes. Se esse efeito não é o desejado, desligue filtros ou revise os limiares.
- **O catálogo não depende de ano nem de município.** O mesmo nome vale para qualquer eleição.
- **Um único modelo.** O sistema mede afinidade com uma tese, e outros eleitores com outras preferências terão resultados pouco úteis.
- **Rótulos políticos são contestados.** "Neoliberal" e "respeita o cristianismo" não têm consenso de definição, e por isso a definição operacional precisa ser escrita antes de usar dados reais.
