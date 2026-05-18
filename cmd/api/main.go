package main

import (
	"encoding/json"
	"fmt"
	"io"
	"log"
	"net/http"

	"github.com/bisheshops/dynamic-crm-engine/internal/database"
	"github.com/bisheshops/dynamic-crm-engine/internal/schema"
	"github.com/go-chi/chi/v5"
	"github.com/go-chi/chi/v5/middleware"
)

type API struct {
	DB *database.DB
}

func main() {
	dsn := "postgres://admin:rootpassword@localhost:5432/dynamic_crm?sslmode=disable"
	db, err := database.New(dsn)
	if err != nil {
		log.Fatalf("Fatal database error: %v", err)
	}
	defer db.Close()

	api := &API{DB: db}

	r := chi.NewRouter()
	r.Use(middleware.Logger)
	r.Use(middleware.Recoverer)

	r.Get("/health", func(w http.ResponseWriter, r *http.Request) {
		w.Header().Set("Content-Type", "application/json")
		w.WriteHeader(http.StatusOK)
		io.WriteString(w, `{"status": "engine and database are running"}`)
	})

	r.Post("/schemas", api.CreateSchemaHandler)
	r.Post("/entities/{schema_name}", api.CreateEntitiesHandler)

	port := ":8080"
	fmt.Printf("Starting Dynamic CRM Engine on port %s\n", port)

	if err := http.ListenAndServe(port, r); err != nil {
		log.Fatalf("Server failed to start: %v", err)
	}
}

func (api *API) CreateSchemaHandler(w http.ResponseWriter, r *http.Request) {
	var req schema.Schema

	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		w.Header().Set("Content-Type", "application/json")
		w.WriteHeader(http.StatusBadRequest)
		fmt.Fprintf(w, `{"error": "Invalid JSON payload: %v"}`, err)
		return
	}
	if req.Name == "" || len(req.Fields) == 0 {
		w.Header().Set("Content-Type", "application/json")
		w.WriteHeader(http.StatusBadRequest)
		io.WriteString(w, `{"error": "Schema name and at least one field are required"}`)
		return
	}

	id, err := api.DB.SaveSchema(r.Context(), req.Name, req)
	if err != nil {
		log.Printf("DB Error: %v\n", err)
		w.Header().Set("Content-Type", "application/json")
		w.WriteHeader(http.StatusInternalServerError)
		io.WriteString(w, `{"error": "Failed to save schema to database"}`)
		return
	}

	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(http.StatusCreated)
	fmt.Fprintf(w, `{"message": "Schema '%s' created successfully", "id": %d}`, req.Name, id)
}

func (api *API) CreateEntitiesHandler(w http.ResponseWriter, r *http.Request) {
	schemaName := chi.URLParam(r, "schema_name")

	var payload map[string]any
	if err := json.NewDecoder(r.Body).Decode(&payload); err != nil {
		w.Header().Set("Content-Type", "application/json")
		w.WriteHeader(http.StatusBadRequest)
		fmt.Fprintf(w, `{"error": "Invalid JSON payload: %v}`, err)
		return
	}
	s, schemaID, err := api.DB.GetSchemaByName(r.Context(), schemaName)
	if err != nil {
		w.Header().Set("Content-Type", "application/json")
		w.WriteHeader(http.StatusNotFound)
		fmt.Fprintf(w, `{"error": "%v"}`, err)
		return
	}
	if err := s.Validate(payload); err != nil {
		w.Header().Set("Content-Type", "application/json")
		w.WriteHeader(http.StatusUnprocessableEntity)
		fmt.Fprintf(w, `{"error": "Validation failed: %v"}`, err)
		return
	}
	entityID, err := api.DB.SaveEntity(r.Context(), schemaID, payload)
	if err != nil {
		log.Printf("DB Error: %v\n", err)
		w.Header().Set("Content-Type", "application/json")
		w.WriteHeader(http.StatusInternalServerError)
		io.WriteString(w, `{"error": "Failed to save entity to database"}`)
		return
	}
	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(http.StatusCreated)
	fmt.Fprintf(w, `{"message": "Entity created successfully", "id": "%s"}`, entityID)
}
