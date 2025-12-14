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
	port := os.Getenv("PORT")
	if port == "" {
		port = "8080"
	}

	dbPath := os.Getenv("DB_PATH")
	if dbPath == "" {
		dbPath = "./darts.db"
	}

	basePath := os.Getenv("BASE_PATH") // e.g., "/darts"
	corsOrigin := getEnv("CORS_ORIGIN", "*")

	log.Printf("Starting Darts Web Server")
	log.Printf("Port: %s", port)
	log.Printf("Database: %s", dbPath)
	log.Printf("Base Path: %s", basePath)

	// Database Init
	db, err := store.NewStore(dbPath)
	if err != nil {
		log.Fatalf("Failed to open database: %v", err)
	}
	defer db.Close()

	// Handlers Init
	h := handlers.NewHandler(db)

	mux := http.NewServeMux()

	// API Routes (ensure they work with or without prefix)
	apiPrefix := basePath + "/api"
	mux.HandleFunc("GET "+apiPrefix+"/users", h.ListUsers)
	mux.HandleFunc("POST "+apiPrefix+"/users", h.CreateUser)
	mux.HandleFunc("DELETE "+apiPrefix+"/users/{id}", h.DeleteUser)
	mux.HandleFunc("POST "+apiPrefix+"/games", h.CreateGame)
	mux.HandleFunc("GET "+apiPrefix+"/games/{id}", h.GetGame)
	mux.HandleFunc("POST "+apiPrefix+"/games/{id}/throw", h.HandleThrow)
	mux.HandleFunc("GET "+apiPrefix+"/users/{id}/stats", h.GetUserStats)

	// Health Check
	mux.HandleFunc("GET "+apiPrefix+"/health", func(w http.ResponseWriter, r *http.Request) {
		w.Header().Set("Content-Type", "application/json")
		w.WriteHeader(http.StatusOK)
		w.Write([]byte(`{"status":"ok"}`))
	})

	// Static Files Serving
	// We serve files from "dist" directory which will be copied into the container
	fs := http.FileServer(http.Dir("./dist"))

	// Strip the prefix if one is set, so /darts/assets/x.js -> /assets/x.js
	var fileHandler http.Handler
	if basePath != "" {
		fileHandler = http.StripPrefix(basePath, fs)
	} else {
		fileHandler = fs
	}

	// Handle root and everything else with file server (SPA support would need more logic ideally,
	// but for now, we just serve files. Accessing /darts/ should serve index.html)
	mux.Handle(basePath+"/", fileHandler)

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
