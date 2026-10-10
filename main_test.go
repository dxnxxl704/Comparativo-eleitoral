package main

import (
	"encoding/json"
	"net/http"
	"net/http/httptest"
	"strings"
	"testing"
)

func TestCandidateStoreCRUD(t *testing.T) {
	store := NewCandidateStore()

	candidate := Candidate{
		Name:      "João da Silva",
		Party:     "Partido Exemplo",
		Office:    "Prefeito",
		State:     "SP",
		Biography: "Histórico público e propostas",
	}

	created, err := store.CreateCandidate(candidate)
	if err != nil {
		t.Fatalf("CreateCandidate() error = %v", err)
	}
	if created.ID == 0 {
		t.Fatal("CreateCandidate() deve atribuir um ID válido")
	}

	all := store.ListCandidates()
	if len(all) != 1 {
		t.Fatalf("ListCandidates() = %d candidatos; want 1", len(all))
	}

	found, ok := store.GetCandidate(created.ID)
	if !ok {
		t.Fatal("GetCandidate() não encontrou o candidato criado")
	}
	if found.Name != candidate.Name {
		t.Fatalf("GetCandidate().Name = %q; want %q", found.Name, candidate.Name)
	}
}

func TestCandidateAPI(t *testing.T) {
	server := newRouter()

	payload := Candidate{
		Name:      "Maria Souza",
		Party:     "Partido Teste",
		Office:    "Governadora",
		State:     "RJ",
		Biography: "Perfil de teste",
	}
	body, _ := json.Marshal(payload)

	req := httptest.NewRequest(http.MethodPost, "/api/candidates", jsonReader(body))
	res := httptest.NewRecorder()
	server.ServeHTTP(res, req)

	if res.Code != http.StatusCreated {
		t.Fatalf("POST /api/candidates returns %d; want %d", res.Code, http.StatusCreated)
	}

	var created Candidate
	if err := json.Unmarshal(res.Body.Bytes(), &created); err != nil {
		t.Fatalf("json.Unmarshal() error = %v", err)
	}
	if created.ID == 0 {
		t.Fatal("POST /api/candidates não retornou um ID válido")
	}

	getReq := httptest.NewRequest(http.MethodGet, "/api/candidates", nil)
	getRes := httptest.NewRecorder()
	server.ServeHTTP(getRes, getReq)

	if getRes.Code != http.StatusOK {
		t.Fatalf("GET /api/candidates returns %d; want %d", getRes.Code, http.StatusOK)
	}
}

func TestEvaluationAndSourceAPI(t *testing.T) {
	server := newRouter()

	candidateBody, _ := json.Marshal(Candidate{
		Name:      "Candidato Fase 2",
		Party:     "Partido Fase 2",
		Office:    "Senador",
		State:     "SP",
		Biography: "Biografia de exemplo",
	})

	candidateReq := httptest.NewRequest(http.MethodPost, "/api/candidates", jsonReader(candidateBody))
	candidateRes := httptest.NewRecorder()
	server.ServeHTTP(candidateRes, candidateReq)

	if candidateRes.Code != http.StatusCreated {
		t.Fatalf("POST /api/candidates returns %d; want %d", candidateRes.Code, http.StatusCreated)
	}

	var candidate Candidate
	if err := json.Unmarshal(candidateRes.Body.Bytes(), &candidate); err != nil {
		t.Fatalf("json.Unmarshal(candidate) error = %v", err)
	}

	evaluationBody, _ := json.Marshal(Evaluation{
		CandidateID:          candidate.ID,
		NaoNeoliberal:        true,
		ContraAborto:         true,
		RespeitaCristianismo: true,
		Observacoes:          "Avaliação de exemplo",
	})

	evaluationReq := httptest.NewRequest(http.MethodPost, "/api/evaluations", jsonReader(evaluationBody))
	evaluationRes := httptest.NewRecorder()
	server.ServeHTTP(evaluationRes, evaluationReq)

	if evaluationRes.Code != http.StatusCreated {
		t.Fatalf("POST /api/evaluations returns %d; want %d", evaluationRes.Code, http.StatusCreated)
	}

	getEvaluationsReq := httptest.NewRequest(http.MethodGet, "/api/evaluations", nil)
	getEvaluationsRes := httptest.NewRecorder()
	server.ServeHTTP(getEvaluationsRes, getEvaluationsReq)

	if getEvaluationsRes.Code != http.StatusOK {
		t.Fatalf("GET /api/evaluations returns %d; want %d", getEvaluationsRes.Code, http.StatusOK)
	}

	sourceBody, _ := json.Marshal(Source{
		CandidateID: candidate.ID,
		Title:       "Entrevista pública",
		URL:         "https://example.com/fonte",
		Type:        "video",
		Observation: "Fonte citada na análise",
	})

	sourceReq := httptest.NewRequest(http.MethodPost, "/api/sources", jsonReader(sourceBody))
	sourceRes := httptest.NewRecorder()
	server.ServeHTTP(sourceRes, sourceReq)

	if sourceRes.Code != http.StatusCreated {
		t.Fatalf("POST /api/sources returns %d; want %d", sourceRes.Code, http.StatusCreated)
	}

	getSourcesReq := httptest.NewRequest(http.MethodGet, "/api/sources", nil)
	getSourcesRes := httptest.NewRecorder()
	server.ServeHTTP(getSourcesRes, getSourcesReq)

	if getSourcesRes.Code != http.StatusOK {
		t.Fatalf("GET /api/sources returns %d; want %d", getSourcesRes.Code, http.StatusOK)
	}
}

func TestCandidateSearchAndFilters(t *testing.T) {
	server := newRouter()

	for _, payload := range []Candidate{
		{Name: "Ana Silva", Party: "Partido A", Office: "Prefeito", State: "SP", Biography: "Experiência em gestão municipal"},
		{Name: "Bruno Lima", Party: "Partido B", Office: "Governador", State: "SP", Biography: "Foco em infraestrutura"},
		{Name: "Carlos Rocha", Party: "Partido C", Office: "Prefeito", State: "RJ", Biography: "Prefeito do Rio"},
	} {
		body, _ := json.Marshal(payload)
		req := httptest.NewRequest(http.MethodPost, "/api/candidates", jsonReader(body))
		res := httptest.NewRecorder()
		server.ServeHTTP(res, req)
		if res.Code != http.StatusCreated {
			t.Fatalf("POST /api/candidates returns %d; want %d", res.Code, http.StatusCreated)
		}
	}

	searchReq := httptest.NewRequest(http.MethodGet, "/api/candidates?q=Silva", nil)
	searchRes := httptest.NewRecorder()
	server.ServeHTTP(searchRes, searchReq)
	if searchRes.Code != http.StatusOK {
		t.Fatalf("GET /api/candidates?q=Silva returns %d; want %d", searchRes.Code, http.StatusOK)
	}

	var searchCandidates []Candidate
	if err := json.Unmarshal(searchRes.Body.Bytes(), &searchCandidates); err != nil {
		t.Fatalf("json.Unmarshal(searchCandidates) error = %v", err)
	}
	if len(searchCandidates) != 1 || searchCandidates[0].Name != "Ana Silva" {
		t.Fatalf("searchCandidates = %#v; want only Ana Silva", searchCandidates)
	}

	filterReq := httptest.NewRequest(http.MethodGet, "/api/candidates?office=Prefeito&state=SP", nil)
	filterRes := httptest.NewRecorder()
	server.ServeHTTP(filterRes, filterReq)
	if filterRes.Code != http.StatusOK {
		t.Fatalf("GET /api/candidates?office=Prefeito&state=SP returns %d; want %d", filterRes.Code, http.StatusOK)
	}

	var filtered []Candidate
	if err := json.Unmarshal(filterRes.Body.Bytes(), &filtered); err != nil {
		t.Fatalf("json.Unmarshal(filtered) error = %v", err)
	}
	if len(filtered) != 1 || filtered[0].Name != "Ana Silva" {
		t.Fatalf("filtered candidates = %#v; want only Ana Silva", filtered)
	}
}

func TestCandidateComparisonAPI(t *testing.T) {
	server := newRouter()

	for _, payload := range []Candidate{
		{Name: "Candidato A", Party: "Partido A", Office: "Prefeito", State: "SP", Biography: "Perfil A"},
		{Name: "Candidato B", Party: "Partido B", Office: "Prefeito", State: "SP", Biography: "Perfil B"},
	} {
		body, _ := json.Marshal(payload)
		req := httptest.NewRequest(http.MethodPost, "/api/candidates", jsonReader(body))
		res := httptest.NewRecorder()
		server.ServeHTTP(res, req)
		if res.Code != http.StatusCreated {
			t.Fatalf("POST /api/candidates returns %d; want %d", res.Code, http.StatusCreated)
		}
	}

	for _, payload := range []Evaluation{
		{CandidateID: 1, NaoNeoliberal: true, ContraAborto: true, RespeitaCristianismo: true, Observacoes: "Muito alinhado"},
		{CandidateID: 2, NaoNeoliberal: false, ContraAborto: true, RespeitaCristianismo: false, Observacoes: "Parcialmente alinhado"},
	} {
		body, _ := json.Marshal(payload)
		req := httptest.NewRequest(http.MethodPost, "/api/evaluations", jsonReader(body))
		res := httptest.NewRecorder()
		server.ServeHTTP(res, req)
		if res.Code != http.StatusCreated {
			t.Fatalf("POST /api/evaluations returns %d; want %d", res.Code, http.StatusCreated)
		}
	}

	compareReq := httptest.NewRequest(http.MethodGet, "/api/candidates/compare?ids=1,2", nil)
	compareRes := httptest.NewRecorder()
	server.ServeHTTP(compareRes, compareReq)
	if compareRes.Code != http.StatusOK {
		t.Fatalf("GET /api/candidates/compare?ids=1,2 returns %d; want %d", compareRes.Code, http.StatusOK)
	}

	var payload compareResponse
	if err := json.Unmarshal(compareRes.Body.Bytes(), &payload); err != nil {
		t.Fatalf("json.Unmarshal(compare) error = %v", err)
	}
	if len(payload.Candidates) != 2 {
		t.Fatalf("compareResponse.Candidates = %d; want 2", len(payload.Candidates))
	}
	if payload.Candidates[0].Score < payload.Candidates[1].Score {
		t.Fatalf("unexpected order in comparison output: %#v", payload.Candidates)
	}
}

func TestPhaseFourDashboardAndReportAPI(t *testing.T) {
	server := newRouter()

	candidateBody, _ := json.Marshal(Candidate{
		Name:      "Dashboard Candidate",
		Party:     "Partido Dashboard",
		Office:    "Prefeito",
		State:     "SP",
		Biography: "Perfil para métricas",
	})
	candidateReq := httptest.NewRequest(http.MethodPost, "/api/candidates", jsonReader(candidateBody))
	candidateRes := httptest.NewRecorder()
	server.ServeHTTP(candidateRes, candidateReq)
	if candidateRes.Code != http.StatusCreated {
		t.Fatalf("POST /api/candidates returns %d; want %d", candidateRes.Code, http.StatusCreated)
	}

	var candidate Candidate
	if err := json.Unmarshal(candidateRes.Body.Bytes(), &candidate); err != nil {
		t.Fatalf("json.Unmarshal(candidate) error = %v", err)
	}

	evalBody, _ := json.Marshal(Evaluation{
		CandidateID:          candidate.ID,
		NaoNeoliberal:        true,
		ContraAborto:         true,
		RespeitaCristianismo: false,
		Observacoes:          "Análise de coerência",
	})
	evalReq := httptest.NewRequest(http.MethodPost, "/api/evaluations", jsonReader(evalBody))
	evalRes := httptest.NewRecorder()
	server.ServeHTTP(evalRes, evalReq)
	if evalRes.Code != http.StatusCreated {
		t.Fatalf("POST /api/evaluations returns %d; want %d", evalRes.Code, http.StatusCreated)
	}

	dashboardReq := httptest.NewRequest(http.MethodGet, "/api/dashboard", nil)
	dashboardRes := httptest.NewRecorder()
	server.ServeHTTP(dashboardRes, dashboardReq)
	if dashboardRes.Code != http.StatusOK {
		t.Fatalf("GET /api/dashboard returns %d; want %d", dashboardRes.Code, http.StatusOK)
	}

	var dashboard dashboardSummary
	if err := json.Unmarshal(dashboardRes.Body.Bytes(), &dashboard); err != nil {
		t.Fatalf("json.Unmarshal(dashboard) error = %v", err)
	}
	if dashboard.TotalCandidates != 1 {
		t.Fatalf("dashboard.TotalCandidates = %d; want 1", dashboard.TotalCandidates)
	}
	if dashboard.AverageCoherence <= 0 {
		t.Fatalf("dashboard.AverageCoherence = %v; want > 0", dashboard.AverageCoherence)
	}

	reportReq := httptest.NewRequest(http.MethodGet, "/api/reports/1", nil)
	reportRes := httptest.NewRecorder()
	server.ServeHTTP(reportRes, reportReq)
	if reportRes.Code != http.StatusOK {
		t.Fatalf("GET /api/reports/1 returns %d; want %d", reportRes.Code, http.StatusOK)
	}

	var report candidateReport
	if err := json.Unmarshal(reportRes.Body.Bytes(), &report); err != nil {
		t.Fatalf("json.Unmarshal(report) error = %v", err)
	}
	if report.Candidate.ID != candidate.ID {
		t.Fatalf("report.Candidate.ID = %d; want %d", report.Candidate.ID, candidate.ID)
	}
	if report.CoherenceScore <= 0 {
		t.Fatalf("report.CoherenceScore = %v; want > 0", report.CoherenceScore)
	}
}

type compareResponse struct {
	Candidates []candidateComparison `json:"candidates"`
}

type candidateComparison struct {
	Candidate Candidate `json:"candidate"`
	Score     float64   `json:"score"`
}

func jsonReader(data []byte) *strings.Reader {
	return strings.NewReader(string(data))
}
