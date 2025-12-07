package main

import (
	"context"
	"log"
	"net/http"
	"os"
	"os/signal"
	"syscall"
	"time"

	"github.com/michaelschlottmann/darts-web/internal/handlers"
	"github.com/michaelschlottmann/darts-web/internal/store"
)

func main() {
	// Configuration from environment
	port := getEnv("PORT", "8080")
	dbPath := getEnv("DB_PATH", "./darts.db")
	corsOrigin := getEnv("CORS_ORIGIN", "*")

	log.Printf("Starting Darts Web Server")
	log.Printf("Port: %s", port)
	log.Printf("Database: %s", dbPath)

	// Database Init
	db, err := store.NewStore(dbPath)
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
		w.Header().Set("Content-Type", "application/json")
		w.WriteHeader(http.StatusOK)
		w.Write([]byte(`{"status":"ok"}`))
	})

	// Add CORS middleware
	handler := enableCORS(mux, corsOrigin)

	// Setup server with timeouts
	server := &http.Server{
		Addr:         ":" + port,
		Handler:      handler,
		ReadTimeout:  15 * time.Second,
		WriteTimeout: 15 * time.Second,
		IdleTimeout:  60 * time.Second,
	}

	// Start server in goroutine
	go func() {
		log.Printf("Server listening on :%s", port)
		if err := server.ListenAndServe(); err != nil && err != http.ErrServerClosed {
			log.Fatalf("Server failed: %v", err)
		}
	}()

	// Graceful shutdown
	quit := make(chan os.Signal, 1)
	signal.Notify(quit, syscall.SIGINT, syscall.SIGTERM)
	<-quit

	log.Println("Shutting down server...")

	ctx, cancel := context.WithTimeout(context.Background(), 30*time.Second)
	defer cancel()

	if err := server.Shutdown(ctx); err != nil {
		log.Fatalf("Server forced to shutdown: %v", err)
	}

	log.Println("Server stopped")
}

func enableCORS(next http.Handler, origin string) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.Header().Set("Access-Control-Allow-Origin", origin)
		w.Header().Set("Access-Control-Allow-Methods", "GET, POST, OPTIONS, PUT, DELETE")
		w.Header().Set("Access-Control-Allow-Headers", "Content-Type, Authorization")

		if r.Method == "OPTIONS" {
			w.WriteHeader(http.StatusOK)
			return
		}

		next.ServeHTTP(w, r)
	})
}

func getEnv(key, defaultValue string) string {
	value := os.Getenv(key)
	if value == "" {
		return defaultValue
	}
	return value
}
