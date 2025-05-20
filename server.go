package zen

import (
	"encoding/json"
	"net/http"

	"github.com/go-chi/chi/v5"
	"github.com/go-chi/chi/v5/middleware"
)

type Server struct {
	engine Engine
}

type EvaluateRequest struct {
	Context any `json:"context"`
}

type CreateDecisionRequest struct {
	Content []byte `json:"content"`
}

func NewServer(engine Engine) *Server {
	return &Server{
		engine: engine,
	}
}

func (s *Server) Routes() chi.Router {
	r := chi.NewRouter()

	// Middleware
	r.Use(middleware.Logger)
	r.Use(middleware.Recoverer)
	r.Use(middleware.RequestID)
	r.Use(middleware.RealIP)
	r.Use(middleware.AllowContentType("application/json"))

	// Routes
	r.Route("/api/v1", func(r chi.Router) {
		// Evaluate a decision by key
		r.Post("/evaluate/{key}", s.handleEvaluate)

		// Create a new decision
		r.Post("/decisions", s.handleCreateDecision)

		// Get a decision by key
		r.Get("/decisions/{key}", s.handleGetDecision)
	})

	return r
}

func (s *Server) handleEvaluate(w http.ResponseWriter, r *http.Request) {
	key := chi.URLParam(r, "key")
	if key == "" {
		http.Error(w, "key is required", http.StatusBadRequest)
		return
	}

	var req EvaluateRequest
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		http.Error(w, "invalid request body", http.StatusBadRequest)
		return
	}

	result, err := s.engine.Evaluate(key, req.Context)
	if err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}

	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(result)
}

func (s *Server) handleCreateDecision(w http.ResponseWriter, r *http.Request) {
	var req CreateDecisionRequest
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		http.Error(w, "invalid request body", http.StatusBadRequest)
		return
	}

	decision, err := s.engine.CreateDecision(req.Content)
	if err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}
	defer decision.Dispose()

	// Evaluate the decision with empty context
	result, err := decision.Evaluate(nil)
	if err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}

	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(result)
}

func (s *Server) handleGetDecision(w http.ResponseWriter, r *http.Request) {
	key := chi.URLParam(r, "key")
	if key == "" {
		http.Error(w, "key is required", http.StatusBadRequest)
		return
	}

	decision, err := s.engine.GetDecision(key)
	if err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}
	defer decision.Dispose()

	// Evaluate the decision with empty context
	result, err := decision.Evaluate(nil)
	if err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}

	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(result)
}
