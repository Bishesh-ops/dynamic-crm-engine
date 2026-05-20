package main

import (
	"context"
	"encoding/json"
	"fmt"
	"log"
	"net/http"
	"os"
	"path/filepath"
	"time"

	"github.com/bisheshops/dynamic-crm-engine/internal/database"
	"github.com/bisheshops/dynamic-crm-engine/internal/eventbus"
	"github.com/bisheshops/dynamic-crm-engine/internal/query"
	"github.com/bisheshops/dynamic-crm-engine/internal/response"
	"github.com/bisheshops/dynamic-crm-engine/internal/schema"
	"github.com/bisheshops/dynamic-crm-engine/internal/workflow"

	"github.com/go-chi/chi/v5"
	"github.com/go-chi/chi/v5/middleware"
)

type API struct {
	DB  *database.DB
	Bus *eventbus.Bus
}

func main() {
	dsn := "postgres://admin:rootpassword@localhost:5432/dynamic_crm?sslmode=disable"

	db, err := database.New(dsn)
	if err != nil {
		log.Fatalf("Fatal database error: %v", err)
	}
	defer db.Close()

	migrationPath := filepath.Join("cmd", "api", "migrations", "001_init.sql")
	log.Printf("Reading database schema from: %s", migrationPath)

	initSQL, err := os.ReadFile(migrationPath)
	if err != nil {
		log.Fatalf("Failed to read schema initialization file: %v", err)
	}

	ctx, cancel := context.WithTimeout(context.Background(), 10*time.Second)
	_, err = db.Pool.Exec(ctx, string(initSQL))
	cancel()
	if err != nil {
		log.Fatalf("Failed to run schema initialization: %v", err)
	}
	log.Println("Database tables verified/initialized successfully.")

	bus := eventbus.New(db, 5, 500, 2*time.Second, []workflow.Workflow{})

	api := &API{
		DB:  db,
		Bus: bus,
	}

	r := chi.NewRouter()
	r.Use(middleware.Logger)
	r.Use(middleware.Recoverer)

	r.Get("/health", func(w http.ResponseWriter, r *http.Request) {
		response.JSON(w, http.StatusOK, map[string]string{"status": "engine and database are running"})
	})

	r.Post("/schemas", api.CreateSchemaHandler)
	r.Post("/entities/{schema_name}", api.CreateEntitiesHandler)
	r.Post("/query", api.SearchEntitiesHandler)

	port := ":8080"
	fmt.Printf("Starting Dynamic CRM Engine on port %s\n", port)

	if err := http.ListenAndServe(port, r); err != nil {
		log.Fatalf("Server failed to start: %v", err)
	}
}

func (api *API) CreateSchemaHandler(w http.ResponseWriter, r *http.Request) {
	var req schema.Schema

	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		response.Error(w, http.StatusBadRequest, fmt.Sprintf("Invalid JSON payload: %v", err))
		return
	}
	if req.Name == "" || len(req.Fields) == 0 {
		response.Error(w, http.StatusBadRequest, "Schema name and at least one field are required")
		return
	}

	id, err := api.DB.SaveSchema(r.Context(), req.Name, req)
	if err != nil {
		log.Printf("DB Error: %v\n", err)
		response.Error(w, http.StatusInternalServerError, "Failed to save schema to database")
		return
	}

	response.JSON(w, http.StatusCreated, map[string]any{
		"message": fmt.Sprintf("Schema '%s' created successfully", req.Name),
		"id":      id,
	})
}
func (api *API) CreateEntitiesHandler(w http.ResponseWriter, r *http.Request) {
	schemaName := chi.URLParam(r, "schema_name")

	var payload map[string]any
	if err := json.NewDecoder(r.Body).Decode(&payload); err != nil {
		response.Error(w, http.StatusBadRequest, fmt.Sprintf("Invalid JSON payload: %v", err))
		return
	}
	s, schemaID, err := api.DB.GetSchemaByName(r.Context(), schemaName)
	if err != nil {
		response.Error(w, http.StatusNotFound, err.Error())
		return
	}
	if err := s.Validate(payload); err != nil {
		response.Error(w, http.StatusUnprocessableEntity, fmt.Sprintf("Validation failed: %v", err))
		return
	}

	err = api.Bus.Publish(eventbus.Event{
		SchemaID:   schemaID,
		SchemaName: schemaName,
		Payload:    payload,
	})

	if err != nil {
		if err == eventbus.ErrQueueFull {
			response.Error(w, http.StatusServiceUnavailable, "Engine is overloaded. Please retry later.")
			return
		}
		response.Error(w, http.StatusInternalServerError, "Failed to publish event")
		return
	}

	response.JSON(w, http.StatusAccepted, map[string]string{"message": "Entity queued for processing"})
}

func (api *API) SearchEntitiesHandler(w http.ResponseWriter, r *http.Request) {
	var req query.Request
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		response.Error(w, http.StatusBadRequest, "Invalid query payload")
		return
	}

	results, err := api.DB.QueryEntities(r.Context(), req)
	if err != nil {
		log.Printf("Query error: %v", err)
		response.Error(w, http.StatusInternalServerError, "Internal server error")
		return
	}

	if results == nil {
		results = []database.BatchEntity{}
	}

	response.JSON(w, http.StatusOK, map[string]any{
		"count": len(results),
		"data":  results,
	})
}
