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

func jsonReader(data []byte) *strings.Reader {
	return strings.NewReader(string(data))
}
