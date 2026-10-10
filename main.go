package main

import (
	"encoding/json"
	"errors"
	"net/http"
	"sort"
	"strconv"
	"strings"
	"sync"
)

type Candidate struct {
	ID        int    `json:"id"`
	Name      string `json:"name"`
	Party     string `json:"party"`
	Office    string `json:"office"`
	State     string `json:"state"`
	Biography string `json:"biography"`
}

type CandidateStore struct {
	mu         sync.Mutex
	candidates map[int]Candidate
	nextID     int
}

func NewCandidateStore() *CandidateStore {
	return &CandidateStore{
		candidates: make(map[int]Candidate),
	}
}

func (s *CandidateStore) CreateCandidate(candidate Candidate) (Candidate, error) {
	if candidate.Name == "" {
		return Candidate{}, errors.New("candidate name is required")
	}

	s.mu.Lock()
	defer s.mu.Unlock()

	s.nextID++
	candidate.ID = s.nextID
	s.candidates[candidate.ID] = candidate
	return candidate, nil
}

func (s *CandidateStore) ListCandidates() []Candidate {
	s.mu.Lock()
	defer s.mu.Unlock()

	candidates := make([]Candidate, 0, len(s.candidates))
	for _, candidate := range s.candidates {
		candidates = append(candidates, candidate)
	}
	sort.Slice(candidates, func(i, j int) bool { return candidates[i].ID < candidates[j].ID })
	return candidates
}

func (s *CandidateStore) SearchCandidates(query string, filters map[string]string) []Candidate {
	s.mu.Lock()
	defer s.mu.Unlock()

	found := make([]Candidate, 0, len(s.candidates))
	for _, candidate := range s.candidates {
		if matchesCandidate(candidate, query, filters) {
			found = append(found, candidate)
		}
	}
	sort.Slice(found, func(i, j int) bool { return found[i].ID < found[j].ID })
	return found
}

func (s *CandidateStore) GetCandidate(id int) (Candidate, bool) {
	s.mu.Lock()
	defer s.mu.Unlock()

	candidate, ok := s.candidates[id]
	return candidate, ok
}

type Evaluation struct {
	ID                   int    `json:"id"`
	CandidateID          int    `json:"candidate_id"`
	NaoNeoliberal        bool   `json:"nao_neoliberal"`
	ContraAborto         bool   `json:"contra_aborto"`
	RespeitaCristianismo bool   `json:"respeita_cristianismo"`
	Observacoes          string `json:"observacoes"`
}

type EvaluationStore struct {
	mu          sync.Mutex
	evaluations map[int]Evaluation
	nextID      int
}

func NewEvaluationStore() *EvaluationStore {
	return &EvaluationStore{
		evaluations: make(map[int]Evaluation),
	}
}

func (s *EvaluationStore) CreateEvaluation(evaluation Evaluation) (Evaluation, error) {
	if evaluation.CandidateID == 0 {
		return Evaluation{}, errors.New("candidate_id is required")
	}

	s.mu.Lock()
	defer s.mu.Unlock()

	s.nextID++
	evaluation.ID = s.nextID
	s.evaluations[evaluation.ID] = evaluation
	return evaluation, nil
}

func (s *EvaluationStore) ListEvaluations() []Evaluation {
	s.mu.Lock()
	defer s.mu.Unlock()

	evaluations := make([]Evaluation, 0, len(s.evaluations))
	for _, evaluation := range s.evaluations {
		evaluations = append(evaluations, evaluation)
	}
	sort.Slice(evaluations, func(i, j int) bool { return evaluations[i].ID < evaluations[j].ID })
	return evaluations
}

func (s *EvaluationStore) GetLatestEvaluation(candidateID int) (Evaluation, bool) {
	s.mu.Lock()
	defer s.mu.Unlock()

	var latest Evaluation
	found := false
	for _, evaluation := range s.evaluations {
		if evaluation.CandidateID != candidateID {
			continue
		}
		if !found || evaluation.ID > latest.ID {
			latest = evaluation
			found = true
		}
	}
	return latest, found
}

type Source struct {
	ID          int    `json:"id"`
	CandidateID int    `json:"candidate_id"`
	Title       string `json:"title"`
	URL         string `json:"url"`
	Type        string `json:"type"`
	Observation string `json:"observation"`
}

type SourceStore struct {
	mu      sync.Mutex
	sources map[int]Source
	nextID  int
}

func NewSourceStore() *SourceStore {
	return &SourceStore{
		sources: make(map[int]Source),
	}
}

func (s *SourceStore) CreateSource(source Source) (Source, error) {
	if source.CandidateID == 0 {
		return Source{}, errors.New("candidate_id is required")
	}
	if source.Title == "" {
		return Source{}, errors.New("title is required")
	}
	if source.URL == "" {
		return Source{}, errors.New("url is required")
	}

	s.mu.Lock()
	defer s.mu.Unlock()

	s.nextID++
	source.ID = s.nextID
	s.sources[source.ID] = source
	return source, nil
}

func (s *SourceStore) ListSources() []Source {
	s.mu.Lock()
	defer s.mu.Unlock()

	sources := make([]Source, 0, len(s.sources))
	for _, source := range s.sources {
		sources = append(sources, source)
	}
	sort.Slice(sources, func(i, j int) bool { return sources[i].ID < sources[j].ID })
	return sources
}

func matchesCandidate(candidate Candidate, query string, filters map[string]string) bool {
	if len(filters) > 0 {
		for key, expected := range filters {
			expected = strings.TrimSpace(expected)
			if expected == "" {
				continue
			}
			switch key {
			case "name":
				if !strings.Contains(strings.ToLower(candidate.Name), strings.ToLower(expected)) {
					return false
				}
			case "party":
				if !strings.Contains(strings.ToLower(candidate.Party), strings.ToLower(expected)) {
					return false
				}
			case "office":
				if !strings.Contains(strings.ToLower(candidate.Office), strings.ToLower(expected)) {
					return false
				}
			case "state":
				if !strings.Contains(strings.ToLower(candidate.State), strings.ToLower(expected)) {
					return false
				}
			case "id":
				candidateID, err := strconv.Atoi(expected)
				if err != nil || candidate.ID != candidateID {
					return false
				}
			}
		}
	}

	if query = strings.TrimSpace(query); query != "" {
		needle := strings.ToLower(query)
		searchable := strings.ToLower(candidate.Name + " " + candidate.Party + " " + candidate.Office + " " + candidate.State + " " + candidate.Biography)
		if !strings.Contains(searchable, needle) {
			return false
		}
	}
	return true
}

func parseCandidateIDs(raw string) []int {
	if strings.TrimSpace(raw) == "" {
		return nil
	}

	parts := strings.Split(raw, ",")
	ids := make([]int, 0, len(parts))
	for _, part := range parts {
		trimmed := strings.TrimSpace(part)
		if trimmed == "" {
			continue
		}
		id, err := strconv.Atoi(trimmed)
		if err != nil {
			continue
		}
		ids = append(ids, id)
	}
	return ids
}

func candidateComparisonScore(candidate Candidate, evaluationStore *EvaluationStore) float64 {
	evaluation, ok := evaluationStore.GetLatestEvaluation(candidate.ID)
	if !ok {
		return 0
	}

	var total int
	if evaluation.NaoNeoliberal {
		total++
	}
	if evaluation.ContraAborto {
		total++
	}
	if evaluation.RespeitaCristianismo {
		total++
	}
	return float64(total) / 3.0 * 100.0
}

type dashboardSummary struct {
	TotalCandidates     int                         `json:"total_candidates"`
	TotalEvaluations    int                         `json:"total_evaluations"`
	TotalSources        int                         `json:"total_sources"`
	AverageCoherence    float64                     `json:"average_coherence"`
	StrongestCandidates []candidateDashboardSummary `json:"strongest_candidates"`
	CriteriaSummary     map[string]int              `json:"criteria_summary"`
	ByOffice            map[string]int              `json:"by_office"`
}

type candidateDashboardSummary struct {
	CandidateID    int     `json:"candidate_id"`
	Name           string  `json:"name"`
	Office         string  `json:"office"`
	CoherenceScore float64 `json:"coherence_score"`
}

type candidateReport struct {
	Candidate      Candidate       `json:"candidate"`
	Evaluation     Evaluation      `json:"evaluation,omitempty"`
	CoherenceScore float64         `json:"coherence_score"`
	Criteria       map[string]bool `json:"criteria"`
	Sources        []Source        `json:"sources"`
}

func candidateCriteria(evaluation Evaluation) map[string]bool {
	return map[string]bool{
		"nao_neoliberal":        evaluation.NaoNeoliberal,
		"contra_aborto":         evaluation.ContraAborto,
		"respeita_cristianismo": evaluation.RespeitaCristianismo,
	}
}

func candidateReportFor(candidate Candidate, evaluationStore *EvaluationStore, sourceStore *SourceStore) candidateReport {
	report := candidateReport{
		Candidate: candidate,
		Criteria:  map[string]bool{},
		Sources:   make([]Source, 0),
	}

	evaluation, ok := evaluationStore.GetLatestEvaluation(candidate.ID)
	if ok {
		report.Evaluation = evaluation
		report.CoherenceScore = candidateComparisonScore(candidate, evaluationStore)
		report.Criteria = candidateCriteria(evaluation)
	}

	for _, source := range sourceStore.ListSources() {
		if source.CandidateID == candidate.ID {
			report.Sources = append(report.Sources, source)
		}
	}

	if report.CoherenceScore == 0 && report.Evaluation.ID == 0 {
		report.CoherenceScore = 0
	}
	return report
}

func buildDashboardSummary(candidateStore *CandidateStore, evaluationStore *EvaluationStore, sourceStore *SourceStore) dashboardSummary {
	candidates := candidateStore.ListCandidates()
	evaluations := evaluationStore.ListEvaluations()
	sources := sourceStore.ListSources()

	summary := dashboardSummary{
		TotalCandidates:     len(candidates),
		TotalEvaluations:    len(evaluations),
		TotalSources:        len(sources),
		AverageCoherence:    0,
		StrongestCandidates: make([]candidateDashboardSummary, 0),
		CriteriaSummary: map[string]int{
			"nao_neoliberal":        0,
			"contra_aborto":         0,
			"respeita_cristianismo": 0,
		},
		ByOffice: map[string]int{},
	}

	for _, candidate := range candidates {
		score := candidateComparisonScore(candidate, evaluationStore)
		summary.AverageCoherence += score
		summary.ByOffice[candidate.Office]++
		if candidate.Office == "" {
			summary.ByOffice["indefinido"]++
			summary.ByOffice[candidate.Office]--
		}
	}

	if len(candidates) > 0 {
		summary.AverageCoherence = summary.AverageCoherence / float64(len(candidates))
	}

	for _, evaluation := range evaluations {
		if evaluation.NaoNeoliberal {
			summary.CriteriaSummary["nao_neoliberal"]++
		}
		if evaluation.ContraAborto {
			summary.CriteriaSummary["contra_aborto"]++
		}
		if evaluation.RespeitaCristianismo {
			summary.CriteriaSummary["respeita_cristianismo"]++
		}
	}

	for _, candidate := range candidates {
		score := candidateComparisonScore(candidate, evaluationStore)
		summary.StrongestCandidates = append(summary.StrongestCandidates, candidateDashboardSummary{
			CandidateID:    candidate.ID,
			Name:           candidate.Name,
			Office:         candidate.Office,
			CoherenceScore: score,
		})
	}
	sort.Slice(summary.StrongestCandidates, func(i, j int) bool {
		return summary.StrongestCandidates[i].CoherenceScore > summary.StrongestCandidates[j].CoherenceScore
	})
	if len(summary.StrongestCandidates) > 3 {
		summary.StrongestCandidates = summary.StrongestCandidates[:3]
	}

	return summary
}

func newRouter() http.Handler {
	candidateStore := NewCandidateStore()
	evaluationStore := NewEvaluationStore()
	sourceStore := NewSourceStore()
	mux := http.NewServeMux()

	mux.Handle("/static/", http.StripPrefix("/static/", http.FileServer(http.Dir("./static"))))
	mux.HandleFunc("/", func(w http.ResponseWriter, r *http.Request) {
		if r.Method != http.MethodGet {
			http.Error(w, "method not allowed", http.StatusMethodNotAllowed)
			return
		}
		if r.URL.Path != "/" {
			http.NotFound(w, r)
			return
		}
		http.ServeFile(w, r, "./static/index.html")
	})

	listCandidates := func(w http.ResponseWriter, r *http.Request) {
		query := r.URL.Query().Get("q")
		filters := map[string]string{
			"name":   r.URL.Query().Get("name"),
			"party":  r.URL.Query().Get("party"),
			"office": r.URL.Query().Get("office"),
			"state":  r.URL.Query().Get("state"),
			"id":     r.URL.Query().Get("id"),
		}
		candidates := candidateStore.SearchCandidates(query, filters)
		w.Header().Set("Content-Type", "application/json")
		if err := json.NewEncoder(w).Encode(candidates); err != nil {
			http.Error(w, err.Error(), http.StatusInternalServerError)
		}
	}

	mux.HandleFunc("/api/candidates", func(w http.ResponseWriter, r *http.Request) {
		switch r.Method {
		case http.MethodGet:
			listCandidates(w, r)
		case http.MethodPost:
			var candidate Candidate
			if err := json.NewDecoder(r.Body).Decode(&candidate); err != nil {
				http.Error(w, "invalid JSON", http.StatusBadRequest)
				return
			}

			created, err := candidateStore.CreateCandidate(candidate)
			if err != nil {
				http.Error(w, err.Error(), http.StatusBadRequest)
				return
			}

			w.Header().Set("Content-Type", "application/json")
			w.WriteHeader(http.StatusCreated)
			if err := json.NewEncoder(w).Encode(created); err != nil {
				http.Error(w, err.Error(), http.StatusInternalServerError)
			}
		default:
			http.Error(w, "method not allowed", http.StatusMethodNotAllowed)
		}
	})

	mux.HandleFunc("/api/search", func(w http.ResponseWriter, r *http.Request) {
		if r.Method != http.MethodGet {
			http.Error(w, "method not allowed", http.StatusMethodNotAllowed)
			return
		}
		listCandidates(w, r)
	})

	mux.HandleFunc("/api/dashboard", func(w http.ResponseWriter, r *http.Request) {
		if r.Method != http.MethodGet {
			http.Error(w, "method not allowed", http.StatusMethodNotAllowed)
			return
		}
		w.Header().Set("Content-Type", "application/json")
		if err := json.NewEncoder(w).Encode(buildDashboardSummary(candidateStore, evaluationStore, sourceStore)); err != nil {
			http.Error(w, err.Error(), http.StatusInternalServerError)
		}
	})

	mux.HandleFunc("/api/metrics", func(w http.ResponseWriter, r *http.Request) {
		if r.Method != http.MethodGet {
			http.Error(w, "method not allowed", http.StatusMethodNotAllowed)
			return
		}
		w.Header().Set("Content-Type", "application/json")
		if err := json.NewEncoder(w).Encode(buildDashboardSummary(candidateStore, evaluationStore, sourceStore)); err != nil {
			http.Error(w, err.Error(), http.StatusInternalServerError)
		}
	})

	mux.HandleFunc("/api/reports", func(w http.ResponseWriter, r *http.Request) {
		if r.Method != http.MethodGet {
			http.Error(w, "method not allowed", http.StatusMethodNotAllowed)
			return
		}

		reports := make([]candidateReport, 0, len(candidateStore.ListCandidates()))
		for _, candidate := range candidateStore.ListCandidates() {
			reports = append(reports, candidateReportFor(candidate, evaluationStore, sourceStore))
		}
		sort.Slice(reports, func(i, j int) bool {
			return reports[i].CoherenceScore > reports[j].CoherenceScore
		})

		w.Header().Set("Content-Type", "application/json")
		if err := json.NewEncoder(w).Encode(map[string]interface{}{
			"reports": reports,
			"summary": buildDashboardSummary(candidateStore, evaluationStore, sourceStore),
		}); err != nil {
			http.Error(w, err.Error(), http.StatusInternalServerError)
		}
	})

	mux.HandleFunc("/api/reports/", func(w http.ResponseWriter, r *http.Request) {
		if r.Method != http.MethodGet {
			http.Error(w, "method not allowed", http.StatusMethodNotAllowed)
			return
		}
		idString := strings.TrimPrefix(r.URL.Path, "/api/reports/")
		if idString == "" || idString == "/api/reports/" {
			http.NotFound(w, r)
			return
		}
		candidateID, err := strconv.Atoi(idString)
		if err != nil || candidateID <= 0 {
			http.Error(w, "invalid candidate id", http.StatusBadRequest)
			return
		}
		candidate, ok := candidateStore.GetCandidate(candidateID)
		if !ok {
			http.NotFound(w, r)
			return
		}
		w.Header().Set("Content-Type", "application/json")
		if err := json.NewEncoder(w).Encode(candidateReportFor(candidate, evaluationStore, sourceStore)); err != nil {
			http.Error(w, err.Error(), http.StatusInternalServerError)
		}
	})

	mux.HandleFunc("/api/report/", func(w http.ResponseWriter, r *http.Request) {
		if r.Method != http.MethodGet {
			http.Error(w, "method not allowed", http.StatusMethodNotAllowed)
			return
		}
		idString := strings.TrimPrefix(r.URL.Path, "/api/report/")
		if idString == "" {
			http.NotFound(w, r)
			return
		}
		candidateID, err := strconv.Atoi(idString)
		if err != nil || candidateID <= 0 {
			http.Error(w, "invalid candidate id", http.StatusBadRequest)
			return
		}
		candidate, ok := candidateStore.GetCandidate(candidateID)
		if !ok {
			http.NotFound(w, r)
			return
		}
		w.Header().Set("Content-Type", "application/json")
		if err := json.NewEncoder(w).Encode(candidateReportFor(candidate, evaluationStore, sourceStore)); err != nil {
			http.Error(w, err.Error(), http.StatusInternalServerError)
		}
	})

	mux.HandleFunc("/api/candidates/compare", func(w http.ResponseWriter, r *http.Request) {
		if r.Method != http.MethodGet {
			http.Error(w, "method not allowed", http.StatusMethodNotAllowed)
			return
		}

		ids := parseCandidateIDs(r.URL.Query().Get("ids"))
		if len(ids) == 0 {
			http.Error(w, "ids query parameter is required", http.StatusBadRequest)
			return
		}

		results := make([]struct {
			Candidate Candidate `json:"candidate"`
			Score     float64   `json:"score"`
		}, 0, len(ids))
		for _, id := range ids {
			candidate, ok := candidateStore.GetCandidate(id)
			if !ok {
				continue
			}
			results = append(results, struct {
				Candidate Candidate `json:"candidate"`
				Score     float64   `json:"score"`
			}{Candidate: candidate, Score: candidateComparisonScore(candidate, evaluationStore)})
		}
		sort.Slice(results, func(i, j int) bool { return results[i].Score > results[j].Score })

		w.Header().Set("Content-Type", "application/json")
		if err := json.NewEncoder(w).Encode(map[string]interface{}{
			"candidates": results,
			"summary": map[string]interface{}{
				"count": len(results),
				"average_score": func() float64 {
					if len(results) == 0 {
						return 0
					}
					total := 0.0
					for _, result := range results {
						total += result.Score
					}
					return total / float64(len(results))
				}(),
			},
		}); err != nil {
			http.Error(w, err.Error(), http.StatusInternalServerError)
		}
	})

	mux.HandleFunc("/api/evaluations", func(w http.ResponseWriter, r *http.Request) {
		switch r.Method {
		case http.MethodGet:
			w.Header().Set("Content-Type", "application/json")
			if err := json.NewEncoder(w).Encode(evaluationStore.ListEvaluations()); err != nil {
				http.Error(w, err.Error(), http.StatusInternalServerError)
			}
		case http.MethodPost:
			var evaluation Evaluation
			if err := json.NewDecoder(r.Body).Decode(&evaluation); err != nil {
				http.Error(w, "invalid JSON", http.StatusBadRequest)
				return
			}

			created, err := evaluationStore.CreateEvaluation(evaluation)
			if err != nil {
				http.Error(w, err.Error(), http.StatusBadRequest)
				return
			}

			w.Header().Set("Content-Type", "application/json")
			w.WriteHeader(http.StatusCreated)
			if err := json.NewEncoder(w).Encode(created); err != nil {
				http.Error(w, err.Error(), http.StatusInternalServerError)
			}
		default:
			http.Error(w, "method not allowed", http.StatusMethodNotAllowed)
		}
	})

	mux.HandleFunc("/api/sources", func(w http.ResponseWriter, r *http.Request) {
		switch r.Method {
		case http.MethodGet:
			w.Header().Set("Content-Type", "application/json")
			if err := json.NewEncoder(w).Encode(sourceStore.ListSources()); err != nil {
				http.Error(w, err.Error(), http.StatusInternalServerError)
			}
		case http.MethodPost:
			var source Source
			if err := json.NewDecoder(r.Body).Decode(&source); err != nil {
				http.Error(w, "invalid JSON", http.StatusBadRequest)
				return
			}

			created, err := sourceStore.CreateSource(source)
			if err != nil {
				http.Error(w, err.Error(), http.StatusBadRequest)
				return
			}

			w.Header().Set("Content-Type", "application/json")
			w.WriteHeader(http.StatusCreated)
			if err := json.NewEncoder(w).Encode(created); err != nil {
				http.Error(w, err.Error(), http.StatusInternalServerError)
			}
		default:
			http.Error(w, "method not allowed", http.StatusMethodNotAllowed)
		}
	})

	return mux
}

func main() {
	if err := http.ListenAndServe(":8080", newRouter()); err != nil {
		panic(err)
	}
}
