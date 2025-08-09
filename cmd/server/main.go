package main

import (
	"log"
	"net/http"
	"os"
	"path/filepath"
	"time"

	"github.com/go-chi/chi/v5"
	"github.com/go-chi/chi/v5/middleware"

	"github.com/rooseveltrp/go-url-shortener/internal/storage"
	"github.com/rooseveltrp/go-url-shortener/internal/transport/httpapi"
)

func getenv(key, def string) string {
	if v := os.Getenv(key); v != "" {
		return v
	}
	return def
}

func main() {
	port := getenv("PORT", "8080")
	baseURL := getenv("BASE_URL", "http://localhost:"+port)
	dbPath := getenv("DB_PATH", "/data/urls.db")

	// Ensure data dir exists
	if err := os.MkdirAll(filepath.Dir(dbPath), 0o755); err != nil {
		log.Fatalf("failed to create data dir: %v", err)
	}

	store, err := storage.New(dbPath)
	if err != nil {
		log.Fatalf("failed to open db: %v", err)
	}
	defer store.Close()

	srv := httpapi.NewServer(store, baseURL)

	r := chi.NewRouter()
	r.Use(middleware.RequestID)
	r.Use(middleware.RealIP)
	r.Use(middleware.Logger)
	r.Use(middleware.Recoverer)
	r.Use(middleware.Timeout(30 * time.Second))

	r.Get("/healthz", func(w http.ResponseWriter, r *http.Request) { w.WriteHeader(http.StatusOK) })

	// API routes
	r.Route("/api", func(api chi.Router) {
		api.Post("/shorten", srv.HandleShorten)
		api.Get("/urls/{code}", srv.HandleGetURL)
	})

	// Redirect route
	r.Get("/{code}", srv.HandleRedirect)

	addr := ":" + port
	log.Printf("listening on %s (baseURL=%s)", addr, baseURL)
	if err := http.ListenAndServe(addr, r); err != nil {
		log.Fatalf("server error: %v", err)
	}
}
