package main

import (
	"embed"
	"io/fs"
	"log"
	"net/http"
	"os"

	"github.com/kidandcat/hogar/internal/auth"
	"github.com/kidandcat/hogar/internal/db"
	"github.com/kidandcat/hogar/internal/handlers"
)

//go:embed web/*
var webFS embed.FS

func main() {
	port := os.Getenv("HOGAR_PORT")
	if port == "" {
		port = "8080"
	}

	// Initialize database
	database, err := db.New("data/hogar.db")
	if err != nil {
		log.Fatalf("Failed to initialize database: %v", err)
	}
	defer database.Close()

	// Create admin user
	if err := auth.EnsureAdminUser(database); err != nil {
		log.Fatalf("Failed to create admin user: %v", err)
	}

	// Seed initial data
	if err := database.Seed(); err != nil {
		log.Fatalf("Failed to seed database: %v", err)
	}

	// Setup template FS (web/ subdirectory)
	webSubFS, err := fs.Sub(webFS, "web")
	if err != nil {
		log.Fatalf("Failed to create sub FS: %v", err)
	}

	// Setup handler
	h := handlers.New(database, webSubFS)

	// Setup routes
	mux := http.NewServeMux()
	h.RegisterRoutes(mux)

	// Serve static files
	staticFS, err := fs.Sub(webFS, "web/static")
	if err != nil {
		log.Fatalf("Failed to create static sub FS: %v", err)
	}
	mux.Handle("GET /static/", http.StripPrefix("/static/", http.FileServer(http.FS(staticFS))))

	// Wrap with auth middleware
	handler := auth.Middleware(mux)

	log.Printf("Hogar server starting on port %s", port)
	if err := http.ListenAndServe(":"+port, handler); err != nil {
		log.Fatalf("Server failed: %v", err)
	}
}
