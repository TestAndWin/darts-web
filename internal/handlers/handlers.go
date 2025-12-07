package handlers

import (
	"encoding/json"
	"log"
	"net/http"
	"strconv"

	"github.com/michaelschlottmann/darts-web/internal/game"
	"github.com/michaelschlottmann/darts-web/internal/store"
)

type Handler struct {
	store  *store.Store
	engine *game.Engine
}

func NewHandler(s *store.Store) *Handler {
	return &Handler{
		store:  s,
		engine: game.NewEngine(),
	}
}

// User Handlers
func (h *Handler) ListUsers(w http.ResponseWriter, r *http.Request) {
	users, err := h.store.ListUsers()
	if err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}
	json.NewEncoder(w).Encode(users)
}

func (h *Handler) CreateUser(w http.ResponseWriter, r *http.Request) {
	var req struct {
		Name string `json:"name"`
	}
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		http.Error(w, err.Error(), http.StatusBadRequest)
		return
	}

	user, err := h.store.CreateUser(req.Name)
	if err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}
	json.NewEncoder(w).Encode(user)
}

// Game Handlers
func (h *Handler) CreateGame(w http.ResponseWriter, r *http.Request) {
	var req struct {
		TotalPoints int   `json:"total_points"`
		BestOf      int   `json:"best_of"`
		PlayerIDs   []int `json:"player_ids"`
	}
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		http.Error(w, err.Error(), http.StatusBadRequest)
		return
	}

	g, err := h.store.CreateGame(req.TotalPoints, req.BestOf, req.PlayerIDs)
	if err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}
	json.NewEncoder(w).Encode(g)
}

func (h *Handler) GetGame(w http.ResponseWriter, r *http.Request) {
	idStr := r.URL.Path[len("/api/games/"):] // Simple path parsing, use a router in prod
	id, err := strconv.Atoi(idStr)
	if err != nil {
		http.Error(w, "Invalid ID", http.StatusBadRequest)
		return
	}

	g, err := h.store.GetGame(id)
	if err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}
	if g == nil {
		http.Error(w, "Game not found", http.StatusNotFound)
		return
	}

	json.NewEncoder(w).Encode(g)
}

func (h *Handler) HandleThrow(w http.ResponseWriter, r *http.Request) {
	// Parse ID manually for now or via library
	// Assuming /api/games/{id}/throw
	// Split by /
	// This is messy without a router, I'll assume I have a router in main or simple parsing here.
	// Let's implement path parsing helper or just rely on main.go to strip prefix and pass ID?
	// I'll keep it simple: The handler will decode body, lookup game from URL logic or passed context if using router.
	// Standard library mux 1.22 has regex but user might be on older Go?
	// I'll use simple standard library patterns.
	// But `HandleThrow` needs the ID.
	// Let's assume the router in `main.go` parses it or we parse from URL.

	// Temporarily: expecting ID in body for simplicity? No, RESTful URL is better.
	// I will parse URL here assuming standard mux behavior (path includes ID).

	// Helper to extract ID from URL /api/games/{id}/throw
	// path: /api/games/1/throw
	// We can use regex or `http.StripPrefix` in main.

	// Just use a query param `?game_id=1` is easiest for vanilla generic handlers,
	// but I want `/api/games/1/throw`.
	// I will just use `ServeMux` with pattern `/api/games/{id}/throw` (Go 1.22+).
	// assuming user has Go 1.22+.

	idStr := r.PathValue("id")
	id, err := strconv.Atoi(idStr)
	if err != nil {
		http.Error(w, "Invalid Game ID", http.StatusBadRequest)
		return
	}

	var req struct {
		UserID     int `json:"user_id"`
		Points     int `json:"points"`
		Multiplier int `json:"multiplier"`
	}
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		http.Error(w, err.Error(), http.StatusBadRequest)
		return
	}

	// 1. Load Game
	g, err := h.store.GetGame(id)
	if err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}

	// 2. Process Throw via Engine
	throw, err := h.engine.ProcessThrow(g, req.UserID, req.Points, req.Multiplier)
	if err != nil {
		http.Error(w, err.Error(), http.StatusBadRequest)
		return
	}

	// 3. Save Throw
	if err := h.store.SaveThrow(throw); err != nil {
		// Log error but game state is in memory, dangerous if partial failure.
		log.Printf("Failed to save throw: %v", err)
	}

	// 4. Update Game State in DB
	if err := h.store.UpdateGame(g); err != nil {
		http.Error(w, "Failed to save game state", http.StatusInternalServerError)
		return
	}

	// 5. Notify Stream (SSE) - optional for now

	json.NewEncoder(w).Encode(g) // Return updated game state
}

func (h *Handler) GetUserStats(w http.ResponseWriter, r *http.Request) {
	idStr := r.PathValue("id")
	id, err := strconv.Atoi(idStr)
	if err != nil {
		http.Error(w, "Invalid User ID", http.StatusBadRequest)
		return
	}

	stats, err := h.store.GetUserStats(id)
	if err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}

	json.NewEncoder(w).Encode(stats)
}
