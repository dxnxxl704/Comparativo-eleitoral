package main

import (
	"encoding/json"
	"errors"
	"net/http"
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
	return candidates
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
	return evaluations
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
	return sources
}

func newRouter() http.Handler {
	candidateStore := NewCandidateStore()
	evaluationStore := NewEvaluationStore()
	sourceStore := NewSourceStore()
	mux := http.NewServeMux()

	mux.HandleFunc("/api/candidates", func(w http.ResponseWriter, r *http.Request) {
		switch r.Method {
		case http.MethodGet:
			w.Header().Set("Content-Type", "application/json")
			if err := json.NewEncoder(w).Encode(candidateStore.ListCandidates()); err != nil {
				http.Error(w, err.Error(), http.StatusInternalServerError)
			}
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
