package main

import (
	"bytes"
	"encoding/json"
	"net/http"
	"net/http/httptest"
	"strconv"
	"strings"
	"testing"
)

func boolPtr(b bool) *bool { return &b }

func TestFindCandidateProfileByCategory(t *testing.T) {
	profile, ok := findCandidateProfile("Presidente", "Exemplo Presidente 1")
	if !ok {
		t.Fatalf("esperava encontrar candidato do catálogo")
	}
	if profile.Partido == "" {
		t.Fatalf("esperava partido preenchido")
	}
	if _, ok := findCandidateProfile("Presidente", "Lula"); ok {
		t.Fatalf("nome fora do catálogo não deveria ser encontrado")
	}
}

func TestEveryCargoHasCandidates(t *testing.T) {
	for _, cargo := range []string{"Presidente", "Governador", "Senador", "Prefeito", "Deputado Federal", "Deputado Estadual", "Vereador"} {
		if len(getCandidatesByCategory(cargo)) == 0 {
			t.Fatalf("catálogo vazio para %s", cargo)
		}
	}
}

func TestElectionYearAvailability(t *testing.T) {
	tests := []struct {
		category string
		year     int
		wantErr  bool
	}{
		{category: "Presidente", year: 2026},
		{category: "Governador", year: 2026},
		{category: "Senador", year: 2026},
		{category: "Deputado Federal", year: 2026},
		{category: "Deputado Estadual", year: 2026},
		{category: "Prefeito", year: 2026, wantErr: true},
		{category: "Vereador", year: 2026, wantErr: true},
		{category: "Prefeito", year: 2028},
		{category: "Vereador", year: 2028},
		{category: "Presidente", year: 2028, wantErr: true},
	}

	for _, test := range tests {
		t.Run(test.category+"/"+strconv.Itoa(test.year), func(t *testing.T) {
			message := electionIssueMessage(test.category, test.year)
			if test.wantErr && message == "" {
				t.Fatalf("esperava mensagem de indisponibilidade eleitoral")
			}
			if !test.wantErr && message != "" {
				t.Fatalf("não esperava mensagem de erro, recebeu: %s", message)
			}
			if test.wantErr && !strings.Contains(message, strconv.Itoa(test.year)) {
				t.Fatalf("a mensagem deve identificar o ano solicitado: %s", message)
			}
		})
	}
}

func TestComputeScore(t *testing.T) {
	tests := []struct {
		qualidade, coerencia, want int
	}{
		{97, 96, 96}, // 96,5 trunca para 96
		{60, 58, 59},
		{100, 100, 100},
		{0, 0, 0},
	}
	for _, test := range tests {
		got := computeScore(CandidateProfile{QualidadeTeses: test.qualidade, CoerenciaHistorica: test.coerencia})
		if got != test.want {
			t.Fatalf("computeScore(%d,%d) = %d, esperado %d", test.qualidade, test.coerencia, got, test.want)
		}
	}
}

func TestStatusPorScoreLimiares(t *testing.T) {
	tests := []struct {
		score int
		want  string
	}{
		{100, statusAltaAfinidade},
		{95, statusAltaAfinidade},
		{94, statusAfinidadeParcial},
		{70, statusAfinidadeParcial},
		{69, statusBaixaAfinidade},
		{0, statusBaixaAfinidade},
	}
	for _, test := range tests {
		if got := statusPorScore(test.score); got != test.want {
			t.Fatalf("statusPorScore(%d) = %q, esperado %q", test.score, got, test.want)
		}
	}
}

func TestFunilEliminaComMotivoEPulaEtapa2(t *testing.T) {
	resp := analisarChapa(AnalysisRequest{
		Ano:   2026,
		Chapa: []ChapaItem{{Cargo: "Presidente", Candidato: "Exemplo Presidente 3"}},
	})
	if len(resp.Resultados) != 1 {
		t.Fatalf("esperava 1 resultado, recebeu %d", len(resp.Resultados))
	}
	r := resp.Resultados[0]
	if !r.Eliminado || r.Status != statusEliminado {
		t.Fatalf("esperava candidato eliminado, recebeu %+v", r)
	}
	if r.Score != nil {
		t.Fatalf("candidato eliminado não deve ter score")
	}
	if len(r.Motivos) != 1 || !strings.Contains(r.Motivos[0], "neoliberal") {
		t.Fatalf("esperava motivo 'neoliberal', recebeu %v", r.Motivos)
	}
}

func TestFunilListaTodosOsMotivos(t *testing.T) {
	resp := analisarChapa(AnalysisRequest{
		Ano:   2026,
		Chapa: []ChapaItem{{Cargo: "Presidente", Candidato: "Exemplo Presidente 4"}},
	})
	if got := len(resp.Resultados[0].Motivos); got != 2 {
		t.Fatalf("esperava 2 motivos (aborto e cristianismo), recebeu %d", got)
	}
}

func TestFiltroDesativadoNaoElimina(t *testing.T) {
	resp := analisarChapa(AnalysisRequest{
		Ano:     2026,
		Filtros: &FiltrosRequest{NaoNeoliberal: boolPtr(false)},
		Chapa:   []ChapaItem{{Cargo: "Presidente", Candidato: "Exemplo Presidente 3"}},
	})
	r := resp.Resultados[0]
	if r.Eliminado {
		t.Fatalf("com o filtro desligado o candidato não deveria ser eliminado")
	}
	if r.Score == nil || *r.Score != 90 {
		t.Fatalf("esperava score 90, recebeu %v", r.Score)
	}
}

func TestFiltrosOmitidosFicamAtivos(t *testing.T) {
	// Só um filtro informado: os outros dois continuam ativos.
	resp := analisarChapa(AnalysisRequest{
		Ano:     2026,
		Filtros: &FiltrosRequest{NaoNeoliberal: boolPtr(false)},
		Chapa:   []ChapaItem{{Cargo: "Presidente", Candidato: "Exemplo Presidente 4"}},
	})
	if !resp.Resultados[0].Eliminado {
		t.Fatalf("filtros omitidos devem ficar ativos")
	}
}

func TestChapaComVariosCargos(t *testing.T) {
	resp := analisarChapa(AnalysisRequest{
		Ano: 2026,
		Chapa: []ChapaItem{
			{Cargo: "Presidente", Candidato: "Exemplo Presidente 1"},
			{Cargo: "Governador", Candidato: "Exemplo Governador 1"},
			{Cargo: "Senador", Candidato: ""},          // não preenchido: ignorado em silêncio
			{Cargo: "Prefeito", Candidato: "Exemplo Prefeito 1"}, // 2026 não tem eleição municipal
			{Cargo: "Presidente", Candidato: "Fulano de Tal"},    // fora do catálogo
			{Cargo: "", Candidato: "Exemplo Presidente 2"},       // sem cargo
		},
	})

	if len(resp.Resultados) != 2 {
		t.Fatalf("esperava 2 resultados, recebeu %d", len(resp.Resultados))
	}
	if resp.Resultados[0].Cargo != "Presidente" || resp.Resultados[1].Cargo != "Governador" {
		t.Fatalf("a ordem da chapa deve ser preservada: %+v", resp.Resultados)
	}
	if len(resp.Ignorados) != 3 {
		t.Fatalf("esperava 3 ignorados com motivo, recebeu %d: %+v", len(resp.Ignorados), resp.Ignorados)
	}
	for _, ig := range resp.Ignorados {
		if ig.Motivo == "" {
			t.Fatalf("todo ignorado precisa de motivo: %+v", ig)
		}
	}
}

func TestAltaAfinidadeNoLimiar(t *testing.T) {
	resp := analisarChapa(AnalysisRequest{
		Ano:   2026,
		Chapa: []ChapaItem{{Cargo: "Presidente", Candidato: "Exemplo Presidente 1"}},
	})
	r := resp.Resultados[0]
	if r.Score == nil || *r.Score != 96 || r.Status != statusAltaAfinidade {
		t.Fatalf("esperava 96 / Alta afinidade, recebeu %+v", r)
	}
}

func TestResultadoNaoDeclaraCandidatoApto(t *testing.T) {
	resp := analisarChapa(AnalysisRequest{
		Ano: 2026,
		Chapa: []ChapaItem{
			{Cargo: "Presidente", Candidato: "Exemplo Presidente 1"},
			{Cargo: "Presidente", Candidato: "Exemplo Presidente 2"},
			{Cargo: "Presidente", Candidato: "Exemplo Presidente 5"},
		},
	})
	for _, r := range resp.Resultados {
		texto := strings.ToLower(r.Status + " " + r.Justificativa)
		if strings.Contains(texto, "apto") || strings.Contains(texto, "aprovado") {
			t.Fatalf("o resultado não pode declarar aptidão nem aprovação: %q", texto)
		}
	}
	if !resp.DadosDemonstrativos {
		t.Fatalf("a resposta deve sinalizar dados demonstrativos")
	}
}

func TestHandlerAnalise(t *testing.T) {
	mux := newMux()

	post := func(body string) *httptest.ResponseRecorder {
		req := httptest.NewRequest(http.MethodPost, "/api/v1/analise", bytes.NewBufferString(body))
		rec := httptest.NewRecorder()
		mux.ServeHTTP(rec, req)
		return rec
	}

	rec := post(`{"ano":2026,"chapa":[{"cargo":"Presidente","candidato":"Exemplo Presidente 1"}]}`)
	if rec.Code != http.StatusOK {
		t.Fatalf("esperava 200, recebeu %d: %s", rec.Code, rec.Body.String())
	}
	var resp AnalysisResponse
	if err := json.Unmarshal(rec.Body.Bytes(), &resp); err != nil {
		t.Fatalf("resposta não é JSON válido: %v", err)
	}
	if len(resp.Resultados) != 1 || resp.LimiarAltaAfinidade != 95 {
		t.Fatalf("resposta inesperada: %+v", resp)
	}

	if rec := post(`{"ano":2026,"chapa":[{"cargo":"Presidente","candidato":""}]}`); rec.Code != http.StatusBadRequest {
		t.Fatalf("chapa sem candidatos deve retornar 400, recebeu %d", rec.Code)
	}
	if rec := post(`{"ano":0,"chapa":[{"cargo":"Presidente","candidato":"Exemplo Presidente 1"}]}`); rec.Code != http.StatusBadRequest {
		t.Fatalf("ano inválido deve retornar 400, recebeu %d", rec.Code)
	}
	if rec := post(`isto não é json`); rec.Code != http.StatusBadRequest {
		t.Fatalf("payload inválido deve retornar 400, recebeu %d", rec.Code)
	}
}

func TestHandlerCandidatos(t *testing.T) {
	mux := newMux()

	req := httptest.NewRequest(http.MethodGet, "/api/v1/candidatos?categoria=Senador&ano=2026", nil)
	rec := httptest.NewRecorder()
	mux.ServeHTTP(rec, req)
	if rec.Code != http.StatusOK {
		t.Fatalf("esperava 200, recebeu %d", rec.Code)
	}
	if strings.Contains(rec.Body.String(), "qualidade") {
		t.Fatalf("a listagem não deve expor notas")
	}

	req = httptest.NewRequest(http.MethodGet, "/api/v1/candidatos?categoria=Prefeito&ano=2026", nil)
	rec = httptest.NewRecorder()
	mux.ServeHTTP(rec, req)
	if rec.Code != http.StatusUnprocessableEntity {
		t.Fatalf("Prefeito em 2026 deve retornar 422, recebeu %d", rec.Code)
	}
}
