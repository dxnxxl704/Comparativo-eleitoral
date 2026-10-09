# CLAUDE.md

# Comparativo Eleitoral

## Visão Geral
O Comparativo Eleitoral é um projeto Build to Learn criado para analisar candidatos políticos utilizando critérios objetivos e observações documentadas.

O sistema busca avaliar se um candidato possui coerência entre discurso, histórico, posicionamentos públicos e capacidade de exercer funções políticas relevantes.

A análise é baseada em critérios binários (Sim/Não), observações complementares, fontes documentadas e histórico público do candidato.

---

## Objetivo
Criar uma plataforma de consulta e análise eleitoral que permita:

- Registrar candidatos.
- Armazenar evidências e fontes.
- Avaliar candidatos segundo critérios pré-definidos.
- Registrar observações e teses.
- Comparar posicionamentos políticos.
- Facilitar uma análise estruturada e auditável.

---

## Critérios de Avaliação
Cada candidato recebe avaliação nos seguintes critérios:

### 1. Não é neoliberal
Avalia se o candidato possui alinhamento ou não com políticas econômicas neoliberais.

Resultado:

- Sim
- Não

---

### 2. Não apoia o aborto
Avalia posicionamentos públicos, votações, declarações e histórico.

Resultado:

- Sim
- Não

---

### 3. Respeita o cristianismo
Avalia discursos, ações, propostas e histórico público.

Resultado:

- Sim
- Não

---

## Observações Complementares
Além dos critérios binários, cada candidato possui:

- Teses defendidas
- Principais argumentos
- Histórico político
- Contradições identificadas
- Coerência entre discurso e prática
- Avaliação qualitativa

---

## Problema que o Projeto Resolve
Existe dificuldade em centralizar informações políticas de forma organizada e comparável.

O projeto busca reunir:

- Fontes públicas
- Histórico dos candidatos
- Posicionamentos políticos
- Informações eleitorais
Permitindo uma análise mais estruturada sobre a capacidade do candidato exercer um cargo público e contribuir para um projeto nacional consistente.

---

# Responsabilidade pelas Decisões

## Product Owner
Responsável:

- Daniel B. Stiwers
Todas as decisões técnicas e de produto são centralizadas no responsável do projeto.

---

# Arquitetura
Arquitetura em camadas:

```
┌────────────────────┐
│      Frontend      │
├────────────────────┤
│       API Go       │
├────────────────────┤
│   Serviços Go      │
├────────────────────┤
│ Regras de Negócio  │
├────────────────────┤
│ Banco Relacional   │
└────────────────────┘
```
Fluxo:

```
Usuário
   ↓
Frontend Estático
   ↓
API REST (Go)
   ↓
Serviços
   ↓
Banco SQL
   ↓
Fontes Externas
   ↓
Google Storage
```

---

# Tecnologias

## Frontend
Tecnologias:

- HTML
- CSS
- JavaScript
- Tailwind CSS
Objetivos:

- Simplicidade
- Clareza visual
- Baixa complexidade
- Fácil manutenção
Características:

- Frontend estático
- Servido pelo backend
- Estilo rápido e consistente com Tailwind
- Interface focada em consulta e análise

---

## Backend
Tecnologia principal:

- Golang
Motivação:

- Performance
- Concorrência
- Simplicidade operacional
- Facilidade para processamento e busca de dados
Responsabilidades:

- API REST
- Regras de negócio
- Integração com banco
- Integração com armazenamento externo
- Busca de fontes públicas

---

## Banco de Dados
Tipo:

- SQL Relacional
Motivação:

- Estrutura bem definida
- Integridade dos dados
- Relacionamentos claros
- Consultas analíticas consistentes
Exemplos possíveis:

- PostgreSQL
- MySQL
Preferência inicial:

- PostgreSQL

---

## Armazenamento
Serviço:

- Google Cloud Storage
Finalidade:

- Evidências
- Documentos
- Capturas
- Arquivos de apoio
- Material coletado em pesquisas

---

## Infraestrutura
Containerização:

- Docker
Objetivos:

- Isolamento do ambiente
- Reprodutibilidade
- Facilidade de deploy
- Consistência entre ambientes

---

# Estrutura do Monorepo

```
comparativo-eleitoral/
│
├── backend/
│   │
│   ├── cmd/
│   │   └── api/
│   │
│   ├── internal/
│   │   ├── candidate/
│   │   ├── evaluation/
│   │   ├── source/
│   │   ├── storage/
│   │   └── database/
│   │
│   ├── pkg/
│   │
│   ├── migrations/
│   │
│   └── tests/
│
├── static/
│   │
│   ├── css/
│   ├── js/
│   ├── images/
│   └── index.html
│
├── database/
│   │
│   ├── schema/
│   ├── seed/
│   └── views/
│
├── docs/
│
├── docker/
│
├── scripts/
│
├── .env.example
├── docker-compose.yml
├── Dockerfile
├── go.mod
└── README.md
```

---

# Modelo Conceitual Inicial

## Candidate
Representa um candidato político.

Campos sugeridos:

- id
- nome
- partido
- cargo
- estado
- biografia
- data_criacao

---

## Evaluation
Representa avaliações.

Campos sugeridos:

- id
- candidate_id
- nao_neoliberal
- contra_aborto
- respeita_cristianismo
- observacoes
- data_avaliacao

---

## Source
Representa fontes e evidências.

Campos sugeridos:

- id
- candidate_id
- titulo
- url
- tipo
- observacao

---

# APIs

## API Interna
Principais endpoints:

### Candidatos

```
GET /api/candidates
GET /api/candidates/{id}
POST /api/candidates
PUT /api/candidates/{id}
DELETE /api/candidates/{id}
```

### Avaliações

```
GET /api/evaluations
POST /api/evaluations
```

### Fontes

```
GET /api/sources
POST /api/sources
```

---

# Comandos de Desenvolvimento

## Backend
Executar aplicação:

```
go run ./backend/cmd/api
```
Build:

```
go build ./backend/cmd/api
```
Testes:

```
go test ./...
```

---

## Docker
Subir ambiente:

```
docker compose up -d
```
Visualizar logs:

```
docker compose logs -f
```
Parar ambiente:

```
docker compose down
```

---

## Banco de Dados
Executar migrações:

```
make migrate-up
```
Reverter migração:

```
make migrate-down
```

---

# Princípios do Projeto

- Clareza acima de complexidade.
- Dados devem possuir fonte verificável.
- Avaliações devem possuir justificativa documentada.
- Critérios devem ser explícitos.
- Manter simplicidade arquitetural.
- Evitar dependências desnecessárias.
- Priorizar aprendizado sobre escalabilidade prematura.

---

# Roadmap Inicial

## Fase 1

- Estrutura do monorepo
- API básica
- Banco relacional
- CRUD de candidatos

## Fase 2

- Cadastro de avaliações
- Cadastro de fontes
- Upload de evidências

## Fase 3

- Busca avançada
- Filtros
- Comparação entre candidatos

## Fase 4

- Relatórios
- Dashboard
- Métricas de coerência

---

# Metadados
Projeto: Comparativo Eleitoral

Tipo: Build to Learn

Modelo de desenvolvimento: Monorepo

Data de início: 09/10/2026

Responsável técnico e validação final:

- Autor do projeto
- Programador principal

**Observação importante:** como os critérios envolvem julgamentos políticos e ideológicos (por exemplo, "não é neoliberal" ou "respeita o cristianismo"), recomendo que a análise seja sempre suportada por evidências, fontes e justificativas documentadas para reduzir subjetividade e manter rastreabilidade das avaliações. Isso também facilita futuras auditorias dos dados cadastrados.
