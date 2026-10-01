package main

import (
	"encoding/json"
	"fmt"
	"log"
	"net/http"
	"os"
	"path/filepath"
	"strconv"
	"strings"
)

// Parâmetros do modelo. Veja docs/metodologia.md.
// Apenas limiarAltaAfinidade vem da especificação original (95%).
// Os demais valores são escolhas de protótipo e podem ser alterados.
const (
	limiarAltaAfinidade    = 95
	limiarAfinidadeParcial = 70

	pesoQualidadeTeses     = 50
	pesoCoerenciaHistorica = 50

	// Todo o catálogo é fictício. Enquanto isso for verdade, a API avisa em cada resposta.
	catalogoDemonstrativo = true
)

const (
	statusEliminado        = "Eliminado no funil"
	statusAltaAfinidade    = "Alta afinidade"
	statusAfinidadeParcial = "Afinidade parcial"
	statusBaixaAfinidade   = "Baixa afinidade"
)

// ---------- Contrato da API ----------

type ChapaItem struct {
	Cargo     string `json:"cargo"`
	Candidato string `json:"candidato"`
}

// FiltrosRequest: cada filtro é opcional e, se omitido, fica ATIVO.
type FiltrosRequest struct {
	NaoNeoliberal        *bool `json:"nao_neoliberal"`
	ProibeAborto         *bool `json:"proibe_aborto"`
	RespeitaCristianismo *bool `json:"respeita_cristianismo"`
}

type AnalysisRequest struct {
	Ano     int             `json:"ano"`
	Filtros *FiltrosRequest `json:"filtros"`
	Chapa   []ChapaItem     `json:"chapa"`
}

type Result struct {
	Cargo         string         `json:"cargo"`
	Nome          string         `json:"nome"`
	Partido       string         `json:"partido"`
	Status        string         `json:"status"`
	Eliminado     bool           `json:"eliminado"`
	Motivos       []string       `json:"motivos,omitempty"`
	Score         *int           `json:"score"` // nil (null) quando eliminado no funil
	Criterios     map[string]int `json:"criterios,omitempty"`
	Justificativa string         `json:"justificativa"`
}

type Ignorado struct {
	Cargo     string `json:"cargo"`
	Candidato string `json:"candidato"`
	Motivo    string `json:"motivo"`
}

type AnalysisResponse struct {
	Ano                 int        `json:"ano"`
	DadosDemonstrativos bool       `json:"dados_demonstrativos"`
	LimiarAltaAfinidade int        `json:"limiar_alta_afinidade"`
	Resultados          []Result   `json:"resultados"`
	Ignorados           []Ignorado `json:"ignorados"`
}

type CandidateCatalogEntry struct {
	Nome    string `json:"nome"`
	Partido string `json:"partido"`
}

// ---------- Catálogo ----------

// CandidateProfile reúne o que o modelo usa:
//   - três marcadores para o funil da Etapa 1 (booleanos);
//   - dois critérios da Etapa 2 (0 a 100).
//
// Hoje todos os valores são fictícios e preenchidos à mão.
type CandidateProfile struct {
	Nome    string
	Partido string

	// Etapa 1 (funil)
	Neoliberal           bool
	ProibeAborto         bool
	RespeitaCristianismo bool

	// Etapa 2 (pontuação, 0 a 100)
	QualidadeTeses     int
	CoerenciaHistorica int
}

func perfil(nome, partido string, neoliberal, proibeAborto, respeitaCristianismo bool, qualidade, coerencia int) CandidateProfile {
	return CandidateProfile{
		Nome:                 nome,
		Partido:              partido,
		Neoliberal:           neoliberal,
		ProibeAborto:         proibeAborto,
		RespeitaCristianismo: respeitaCristianismo,
		QualidadeTeses:       qualidade,
		CoerenciaHistorica:   coerencia,
	}
}

// Catálogo FICTÍCIO. Não representa nenhuma pessoa ou partido real.
// Os perfis cobrem os caminhos do modelo: alta afinidade, parcial, baixa,
// eliminado por um filtro e eliminado por vários filtros.
var catalog = map[string][]CandidateProfile{
	"Presidente": {
		perfil("Exemplo Presidente 1", "Partido Fictício A", false, true, true, 97, 96),
		perfil("Exemplo Presidente 2", "Partido Fictício B", false, true, true, 82, 78),
		perfil("Exemplo Presidente 3", "Partido Fictício C", true, true, true, 90, 90),
		perfil("Exemplo Presidente 4", "Partido Fictício D", false, false, false, 70, 70),
		perfil("Exemplo Presidente 5", "Partido Fictício E", false, true, true, 60, 58),
	},
	"Governador": {
		perfil("Exemplo Governador 1", "Partido Fictício A", false, true, true, 91, 92),
		perfil("Exemplo Governador 2", "Partido Fictício C", true, true, true, 88, 85),
	},
	"Senador": {
		perfil("Exemplo Senador 1", "Partido Fictício B", false, true, true, 96, 97),
		perfil("Exemplo Senador 2", "Partido Fictício D", false, false, true, 75, 72),
	},
	"Prefeito": {
		perfil("Exemplo Prefeito 1", "Partido Fictício A", false, true, true, 84, 86),
		perfil("Exemplo Prefeito 2", "Partido Fictício C", true, false, true, 80, 80),
	},
	"Deputado Federal": {
		perfil("Exemplo Deputado Federal 1", "Partido Fictício B", false, true, true, 78, 74),
		perfil("Exemplo Deputado Federal 2", "Partido Fictício E", false, true, false, 85, 83),
	},
	"Deputado Estadual": {
		perfil("Exemplo Deputado Estadual 1", "Partido Fictício A", false, true, true, 95, 95),
		perfil("Exemplo Deputado Estadual 2", "Partido Fictício D", true, true, true, 70, 68),
	},
	"Vereador": {
		perfil("Exemplo Vereador 1", "Partido Fictício B", false, true, true, 66, 70),
		perfil("Exemplo Vereador 2", "Partido Fictício E", false, false, true, 90, 88),
	},
}

// ---------- Cargos e calendário eleitoral ----------

func normalizeCategory(category string) string {
	normalized := strings.TrimSpace(category)
	if normalized == "" {
		return "Presidente"
	}

	switch strings.ToLower(normalized) {
	case "presidente":
		return "Presidente"
	case "governador":
		return "Governador"
	case "senador":
		return "Senador"
	case "prefeito":
		return "Prefeito"
	case "deputado federal", "deputado-federal":
		return "Deputado Federal"
	case "deputado estadual", "deputado-estadual":
		return "Deputado Estadual"
	case "vereador":
		return "Vereador"
	default:
		for key := range catalog {
			if strings.EqualFold(key, normalized) {
				return key
			}
		}
		return normalized
	}
}

// electionIssueMessage devolve texto vazio quando há eleição para o cargo no ano.
// Eleições gerais: anos com resto 2 na divisão por 4 (2022, 2026, 2030).
// Eleições municipais: anos múltiplos de 4 (2024, 2028).
func electionIssueMessage(category string, year int) string {
	if year < 1900 || year > 9999 {
		return "Informe um ano eleitoral válido."
	}

	normalizedCategory := normalizeCategory(category)
	switch normalizedCategory {
	case "Presidente", "Governador", "Senador", "Deputado Federal", "Deputado Estadual":
		if year%4 != 2 {
			return fmt.Sprintf("Em %d não haverá eleição para %s. As eleições gerais ocorrem em anos como 2022, 2026 e 2030.", year, normalizedCategory)
		}
	case "Prefeito", "Vereador":
		if year%4 != 0 {
			return fmt.Sprintf("Em %d não haverá eleição para %s. As eleições municipais ocorrem em anos como 2024 e 2028.", year, normalizedCategory)
		}
	default:
		return "Cargo eleitoral não reconhecido."
	}

	return ""
}

func getCandidatesByCategory(category string) []CandidateProfile {
	profiles := catalog[normalizeCategory(category)]
	result := make([]CandidateProfile, len(profiles))
	copy(result, profiles)
	return result
}

func findCandidateProfile(category string, name string) (CandidateProfile, bool) {
	target := strings.TrimSpace(name)
	for _, profile := range getCandidatesByCategory(category) {
		if strings.EqualFold(strings.TrimSpace(profile.Nome), target) {
			return profile, true
		}
	}
	return CandidateProfile{}, false
}

// ---------- Modelo: Etapa 1 (funil) e Etapa 2 (pontuação) ----------

type filtrosAtivos struct {
	naoNeoliberal        bool
	proibeAborto         bool
	respeitaCristianismo bool
}

func boolOuPadrao(valor *bool, padrao bool) bool {
	if valor == nil {
		return padrao
	}
	return *valor
}

// resolverFiltros aplica o padrão da especificação: os três filtros começam ativos.
func resolverFiltros(req *FiltrosRequest) filtrosAtivos {
	if req == nil {
		return filtrosAtivos{naoNeoliberal: true, proibeAborto: true, respeitaCristianismo: true}
	}
	return filtrosAtivos{
		naoNeoliberal:        boolOuPadrao(req.NaoNeoliberal, true),
		proibeAborto:         boolOuPadrao(req.ProibeAborto, true),
		respeitaCristianismo: boolOuPadrao(req.RespeitaCristianismo, true),
	}
}

// motivosDeEliminacao lista TODOS os filtros ativos que o candidato não atende.
func motivosDeEliminacao(p CandidateProfile, f filtrosAtivos) []string {
	motivos := []string{}
	if f.naoNeoliberal && p.Neoliberal {
		motivos = append(motivos, "classificado como neoliberal")
	}
	if f.proibeAborto && !p.ProibeAborto {
		motivos = append(motivos, "não defende a proibição do aborto")
	}
	if f.respeitaCristianismo && !p.RespeitaCristianismo {
		motivos = append(motivos, "não respeita/incentiva o cristianismo")
	}
	return motivos
}

// computeScore devolve a média ponderada (0 a 100) dos dois critérios da Etapa 2,
// com divisão inteira (trunca).
func computeScore(p CandidateProfile) int {
	soma := p.QualidadeTeses*pesoQualidadeTeses + p.CoerenciaHistorica*pesoCoerenciaHistorica
	return soma / (pesoQualidadeTeses + pesoCoerenciaHistorica)
}

func statusPorScore(score int) string {
	switch {
	case score >= limiarAltaAfinidade:
		return statusAltaAfinidade
	case score >= limiarAfinidadeParcial:
		return statusAfinidadeParcial
	default:
		return statusBaixaAfinidade
	}
}

func generateJustificativa(score int) string {
	const aviso = " Isto mede proximidade com o modelo, não capacidade de governar nem chance de vitória."
	switch statusPorScore(score) {
	case statusAltaAfinidade:
		return "Alta afinidade com os critérios do modelo (qualidade das teses e coerência histórica)." + aviso
	case statusAfinidadeParcial:
		return "Afinidade parcial com os critérios do modelo." + aviso
	default:
		return "Baixa afinidade com os critérios do modelo." + aviso
	}
}

func avaliar(cargo string, p CandidateProfile, f filtrosAtivos) Result {
	r := Result{Cargo: cargo, Nome: p.Nome, Partido: p.Partido}

	if motivos := motivosDeEliminacao(p, f); len(motivos) > 0 {
		r.Status = statusEliminado
		r.Eliminado = true
		r.Motivos = motivos
		r.Justificativa = "Eliminado na Etapa 1 (funil): " + strings.Join(motivos, "; ") + ". A Etapa 2 não foi aplicada."
		return r
	}

	score := computeScore(p)
	r.Score = &score
	r.Status = statusPorScore(score)
	r.Criterios = map[string]int{
		"qualidade_teses":     p.QualidadeTeses,
		"coerencia_historica": p.CoerenciaHistorica,
	}
	r.Justificativa = generateJustificativa(score)
	return r
}

func contarPreenchidos(chapa []ChapaItem) int {
	total := 0
	for _, item := range chapa {
		if strings.TrimSpace(item.Candidato) != "" {
			total++
		}
	}
	return total
}

// analisarChapa audita cada item da chapa de forma independente, na ordem recebida.
// Itens com candidato vazio são ignorados em silêncio (campo não preenchido).
// Itens com problema (cargo vazio, fora do ano, nome fora do catálogo) vão para
// "ignorados", com o motivo, e não somem.
func analisarChapa(req AnalysisRequest) AnalysisResponse {
	filtros := resolverFiltros(req.Filtros)
	resp := AnalysisResponse{
		Ano:                 req.Ano,
		DadosDemonstrativos: catalogoDemonstrativo,
		LimiarAltaAfinidade: limiarAltaAfinidade,
		Resultados:          []Result{},
		Ignorados:           []Ignorado{},
	}

	for _, item := range req.Chapa {
		nome := strings.TrimSpace(item.Candidato)
		if nome == "" {
			continue
		}

		if strings.TrimSpace(item.Cargo) == "" {
			resp.Ignorados = append(resp.Ignorados, Ignorado{Candidato: nome, Motivo: "Informe o cargo do candidato."})
			continue
		}
		cargo := normalizeCategory(item.Cargo)

		if message := electionIssueMessage(cargo, req.Ano); message != "" {
			resp.Ignorados = append(resp.Ignorados, Ignorado{Cargo: cargo, Candidato: nome, Motivo: message})
			continue
		}

		profile, ok := findCandidateProfile(cargo, nome)
		if !ok {
			resp.Ignorados = append(resp.Ignorados, Ignorado{
				Cargo:     cargo,
				Candidato: nome,
				Motivo:    "Candidato não encontrado no catálogo de " + cargo + ".",
			})
			continue
		}

		resp.Resultados = append(resp.Resultados, avaliar(cargo, profile, filtros))
	}

	return resp
}

// ---------- HTTP ----------

func writeJSON(w http.ResponseWriter, status int, payload any) {
	w.Header().Set("Content-Type", "application/json")
	w.Header().Set("Access-Control-Allow-Origin", "*")
	w.Header().Set("Access-Control-Allow-Methods", "GET, POST, OPTIONS")
	w.Header().Set("Access-Control-Allow-Headers", "Content-Type")
	w.WriteHeader(status)
	_ = json.NewEncoder(w).Encode(payload)
}

func newMux() *http.ServeMux {
	mux := http.NewServeMux()

	mux.HandleFunc("/api/v1/health", func(w http.ResponseWriter, r *http.Request) {
		if r.Method != http.MethodGet {
			http.Error(w, "Method not allowed", http.StatusMethodNotAllowed)
			return
		}
		writeJSON(w, http.StatusOK, map[string]string{"status": "ok"})
	})

	mux.HandleFunc("/api/v1/candidatos", func(w http.ResponseWriter, r *http.Request) {
		if r.Method == http.MethodOptions {
			writeJSON(w, http.StatusOK, map[string]string{"status": "ok"})
			return
		}
		if r.Method != http.MethodGet {
			http.Error(w, "Method not allowed", http.StatusMethodNotAllowed)
			return
		}

		categoria := r.URL.Query().Get("categoria")
		ano, err := strconv.Atoi(r.URL.Query().Get("ano"))
		if err != nil {
			writeJSON(w, http.StatusBadRequest, map[string]string{"error": "Informe um ano eleitoral válido."})
			return
		}
		if message := electionIssueMessage(categoria, ano); message != "" {
			writeJSON(w, http.StatusUnprocessableEntity, map[string]string{"error": message})
			return
		}

		profiles := getCandidatesByCategory(categoria)
		entries := make([]CandidateCatalogEntry, 0, len(profiles))
		for _, profile := range profiles {
			entries = append(entries, CandidateCatalogEntry{Nome: profile.Nome, Partido: profile.Partido})
		}
		writeJSON(w, http.StatusOK, entries)
	})

	mux.HandleFunc("/api/v1/analise", func(w http.ResponseWriter, r *http.Request) {
		if r.Method == http.MethodOptions {
			writeJSON(w, http.StatusOK, map[string]string{"status": "ok"})
			return
		}
		if r.Method != http.MethodPost {
			http.Error(w, "Method not allowed", http.StatusMethodNotAllowed)
			return
		}

		r.Body = http.MaxBytesReader(w, r.Body, 1<<20)
		var req AnalysisRequest
		if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
			writeJSON(w, http.StatusBadRequest, map[string]string{"error": "Payload inválido"})
			return
		}
		if req.Ano < 1900 || req.Ano > 9999 {
			writeJSON(w, http.StatusBadRequest, map[string]string{"error": "Informe um ano eleitoral válido."})
			return
		}
		if contarPreenchidos(req.Chapa) == 0 {
			writeJSON(w, http.StatusBadRequest, map[string]string{"error": "Preencha ao menos um candidato na chapa."})
			return
		}

		writeJSON(w, http.StatusOK, analisarChapa(req))
	})

	staticDir := "static"
	if _, err := os.Stat(staticDir); err == nil {
		mux.Handle("/", http.FileServer(http.Dir(staticDir)))
	} else {
		mux.HandleFunc("/", func(w http.ResponseWriter, r *http.Request) {
			if r.URL.Path != "/" {
				http.NotFound(w, r)
				return
			}
			w.Header().Set("Content-Type", "text/html; charset=utf-8")
			w.WriteHeader(http.StatusOK)
			_, _ = w.Write([]byte(`<!doctype html><html><body><h1>Match Eleitoral</h1><p>Servidor em execução.</p></body></html>`))
		})
	}

	return mux
}

func main() {
	if staticRoot, err := filepath.Abs("static"); err == nil {
		log.Printf("Diretório estático: %s", staticRoot)
	}

	port := os.Getenv("PORT")
	if port == "" {
		port = "8080"
	}

	addr := fmt.Sprintf(":%s", port)
	log.Printf("Servidor rodando em http://localhost%s", addr)
	log.Fatal(http.ListenAndServe(addr, newMux()))
}
