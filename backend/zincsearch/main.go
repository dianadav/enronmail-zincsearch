package main

import (
	"log"
	"net/http"

	"github.com/go-chi/chi/v5"
	"github.com/go-chi/chi/v5/middleware"
	"github.com/go-chi/cors"

	"zincsearch-backend/handlers"
)

func main() {
	r := chi.NewRouter()

	// Middleware
	r.Use(middleware.Logger)
	r.Use(middleware.Recoverer)

	// Configuración de CORS
	r.Use(cors.Handler(cors.Options{
		AllowedOrigins:   []string{"http://localhost:5173", "http://localhost:8081", "http://localhost:3001"}, // Permite solicitudes desde el frontend
		AllowedMethods:   []string{"GET", "POST", "PUT", "DELETE", "OPTIONS"},
		AllowedHeaders:   []string{"Accept", "Authorization", "Content-Type", "X-CSRF-Token"},
		ExposedHeaders:   []string{"Link"},
		AllowCredentials: true,
		MaxAge:           300, // Cache preflight request por 5 minutos
	}))

	// Routes
	r.Get("/api/documents", handlers.GetDocuments) // Ruta para obtener documentos

	// Start server
	log.Println("Starting server on :8090...")
	log.Fatal(http.ListenAndServe(":8090", r))
}

//curl "http://localhost:8090/api/documents?index=emails&from=0&size=5"

//curl -X POST "http://localhost:4080/api/emails/_search" -d '{"query": { "match_all": {} }}'
