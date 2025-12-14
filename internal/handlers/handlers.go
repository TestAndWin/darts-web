package handlers

import (
	"database/sql"
	"encoding/json"
	"errors"
	"fmt"
	"log"
	"net/http"
	"strconv"
	"sync"

	"github.com/michaelschlottmann/darts-web/internal/game"
	"github.com/michaelschlottmann/darts-web/internal/store"
)

type Handler struct {
	store      *store.Store
	engine     *game.Engine
	gameLocks  map[int]*sync.Mutex
	locksMutex sync.Mutex
}

func NewHandler(s *store.Store) *Handler {
	return &Handler{
		store:     s,
		engine:    game.NewEngine(),
		gameLocks: make(map[int]*sync.Mutex),
	}
}

// getGameLock returns or creates a mutex for a specific game ID
func (h *Handler) getGameLock(gameID int) *sync.Mutex {
	h.locksMutex.Lock()
	defer h.locksMutex.Unlock()

	if lock, exists := h.gameLocks[gameID]; exists {
		return lock
	}

	lock := &sync.Mutex{}
	h.gameLocks[gameID] = lock
	return lock
}

// writeJSON writes a JSON response with proper content type
func writeJSON(w http.ResponseWriter, status int, data interface{}) {
	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(status)
	json.NewEncoder(w).Encode(data)
}

// writeError writes a JSON error response
func writeError(w http.ResponseWriter, status int, message string) {
	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(status)
	json.NewEncoder(w).Encode(map[string]string{"error": message})
}

// User Handlers
func (h *Handler) ListUsers(w http.ResponseWriter, r *http.Request) {
	users, err := h.store.ListUsers()
	if err != nil {
		writeError(w, http.StatusInternalServerError, "Failed to list users")
		return
	}
	writeJSON(w, http.StatusOK, users)
}

func (h *Handler) CreateUser(w http.ResponseWriter, r *http.Request) {
	var req struct {
		Name string `json:"name"`
	}
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		writeError(w, http.StatusBadRequest, "Invalid request body")
		return
	}

	// Validate input
	if req.Name == "" || len(req.Name) > 100 {
		writeError(w, http.StatusBadRequest, "Name must be between 1 and 100 characters")
		return
	}

	user, err := h.store.CreateUser(req.Name)
	if err != nil {
		if errors.Is(err, store.ErrDuplicateUsername) {
			writeError(w, http.StatusConflict, "Username already exists")
			return
		}
		writeError(w, http.StatusInternalServerError, "Failed to create user")
		return
	}
	writeJSON(w, http.StatusCreated, user)
}

func (h *Handler) DeleteUser(w http.ResponseWriter, r *http.Request) {
	idStr := r.PathValue("id")
	id, err := strconv.Atoi(idStr)
	if err != nil {
		writeError(w, http.StatusBadRequest, "Invalid user ID")
		return
	}

	err = h.store.DeleteUser(id)
	if err != nil {
		if errors.Is(err, sql.ErrNoRows) {
			writeError(w, http.StatusNotFound, "User not found")
			return
		}
		writeError(w, http.StatusInternalServerError, "Failed to delete user")
		return
	}

	w.WriteHeader(http.StatusNoContent)
}

// Game Handlers
func (h *Handler) CreateGame(w http.ResponseWriter, r *http.Request) {
	var req struct {
		TotalPoints int   `json:"total_points"`
		BestOf      int   `json:"best_of"`
		DoubleOut   bool  `json:"double_out"`
		PlayerIDs   []int `json:"player_ids"`
	}
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		writeError(w, http.StatusBadRequest, "Invalid request body")
		return
	}

	// Validate input
	if req.TotalPoints != 301 && req.TotalPoints != 501 {
		writeError(w, http.StatusBadRequest, "Total points must be 301 or 501")
		return
	}
	if req.BestOf != 1 && req.BestOf != 3 && req.BestOf != 5 {
		writeError(w, http.StatusBadRequest, "Best of must be 1, 3, or 5")
		return
	}
	if len(req.PlayerIDs) < 1 || len(req.PlayerIDs) > 4 {
		writeError(w, http.StatusBadRequest, "Number of players must be between 1 and 4")
		return
	}

	g, err := h.store.CreateGame(req.TotalPoints, req.BestOf, req.DoubleOut, req.PlayerIDs)
	if err != nil {
		writeError(w, http.StatusInternalServerError, "Failed to create game")
		return
	}
	writeJSON(w, http.StatusCreated, g)
}

func (h *Handler) GetGame(w http.ResponseWriter, r *http.Request) {
	idStr := r.PathValue("id")
	id, err := strconv.Atoi(idStr)
	if err != nil {
		writeError(w, http.StatusBadRequest, "Invalid game ID")
		return
	}

	g, err := h.store.GetGame(id)
	if err != nil {
		writeError(w, http.StatusInternalServerError, "Failed to get game")
		return
	}
	if g == nil {
		writeError(w, http.StatusNotFound, "Game not found")
		return
	}

	writeJSON(w, http.StatusOK, g)
}

func (h *Handler) HandleThrow(w http.ResponseWriter, r *http.Request) {
	idStr := r.PathValue("id")
	id, err := strconv.Atoi(idStr)
	if err != nil {
		writeError(w, http.StatusBadRequest, "Invalid game ID")
		return
	}

	var req struct {
		UserID     int `json:"user_id"`
		Points     int `json:"points"`
		Multiplier int `json:"multiplier"`
	}
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		writeError(w, http.StatusBadRequest, "Invalid request body")
		return
	}

	// Acquire lock for this game to prevent race conditions
	lock := h.getGameLock(id)
	lock.Lock()
	defer lock.Unlock()

	// 1. Load Game
	g, err := h.store.GetGame(id)
	if err != nil {
		writeError(w, http.StatusInternalServerError, "Failed to load game")
		return
	}
	if g == nil {
		writeError(w, http.StatusNotFound, "Game not found")
		return
	}

	// 2. Process Throw via Engine
	throw, err := h.engine.ProcessThrow(g, req.UserID, req.Points, req.Multiplier)
	if err != nil {
		writeError(w, http.StatusBadRequest, fmt.Sprintf("Invalid throw: %v", err))
		return
	}

	// 3. Save Throw
	if err := h.store.SaveThrow(throw); err != nil {
		log.Printf("Failed to save throw for game %d: %v", id, err)
		writeError(w, http.StatusInternalServerError, "Failed to save throw")
		return
	}

	// 4. Update Game State in DB
	if err := h.store.UpdateGame(g); err != nil {
		log.Printf("Failed to update game %d: %v", id, err)
		writeError(w, http.StatusInternalServerError, "Failed to update game state")
		return
	}

	writeJSON(w, http.StatusOK, g)
}

func (h *Handler) GetUserStats(w http.ResponseWriter, r *http.Request) {
	idStr := r.PathValue("id")
	id, err := strconv.Atoi(idStr)
	if err != nil {
		writeError(w, http.StatusBadRequest, "Invalid user ID")
		return
	}

	stats, err := h.store.GetUserStats(id)
	if err != nil {
		writeError(w, http.StatusInternalServerError, "Failed to get user stats")
		return
	}

	writeJSON(w, http.StatusOK, stats)
}
