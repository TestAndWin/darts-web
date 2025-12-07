package main

import (
	"log"
	"net/http"
	"os"

	"github.com/michaelschlottmann/darts-web/internal/handlers"
	"github.com/michaelschlottmann/darts-web/internal/store"
)

func main() {
	port := os.Getenv("PORT")
	if port == "" {
		port = "8080"
	}

	log.Printf("Starting server on port %s", port)

	// Database Init
	db, err := store.NewStore("./darts.db")
	if err != nil {
		log.Fatalf("Failed to open database: %v", err)
	}
	defer db.Close()

	// Handlers Init
	h := handlers.NewHandler(db)

	mux := http.NewServeMux()

	// API Routes
	mux.HandleFunc("GET /api/users", h.ListUsers)
	mux.HandleFunc("POST /api/users", h.CreateUser)
	mux.HandleFunc("POST /api/games", h.CreateGame)
	mux.HandleFunc("GET /api/games/{id}", h.GetGame)
	mux.HandleFunc("POST /api/games/{id}/throw", h.HandleThrow)
	mux.HandleFunc("GET /api/users/{id}/stats", h.GetUserStats)

	// Health Check
	mux.HandleFunc("GET /api/health", func(w http.ResponseWriter, r *http.Request) {
		w.WriteHeader(http.StatusOK)
		w.Write([]byte("OK"))
	})

	// Add CORS middleware for dev (Vite runs on different port)
	handler := enableCORS(mux)

	log.Printf("Starting server on port %s", port)
	if err := http.ListenAndServe(":"+port, handler); err != nil {
		log.Fatalf("Server failed: %v", err)
	}
}

func enableCORS(next http.Handler) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.Header().Set("Access-Control-Allow-Origin", "*") // Allow all for dev
		w.Header().Set("Access-Control-Allow-Methods", "GET, POST, OPTIONS, PUT, DELETE")
		w.Header().Set("Access-Control-Allow-Headers", "Content-Type, Authorization")

		if r.Method == "OPTIONS" {
			w.WriteHeader(http.StatusOK)
			return
		}

		next.ServeHTTP(w, r)
	})
}
