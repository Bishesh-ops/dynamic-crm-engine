package database

import (
	"context"
	"testing"
	"time"

	"github.com/bisheshops/dynamic-crm-engine/internal/query"
)

func TestQueryEntitiesIntegration(t *testing.T) {
	dsn := "postgres://admin:rootpassword@localhost:5432/dynamic_crm?sslmode=disable"
	db, err := New(dsn)
	if err != nil {
		t.Skipf("Database not running, skipping integration test: %v", err)
	}
	defer db.Close()

	ctx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
	defer cancel()

	schemaID, err := db.SaveSchema(ctx, "test_query_schema_"+time.Now().String(), map[string]any{})
	if err != nil {
		t.Fatalf("Failed to save schema: %v", err)
	}

	entities := []BatchEntity{
		{SchemaID: schemaID, Data: map[string]any{"status": "active", "score": 85.0}},
		{SchemaID: schemaID, Data: map[string]any{"status": "active", "score": 95.0}},
		{SchemaID: schemaID, Data: map[string]any{"status": "inactive", "score": 100.0}},
	}

	err = db.SaveEntityBatch(ctx, entities)
	if err != nil {
		t.Fatalf("Failed to save entities: %v", err)
	}

	req := query.Request{
		SchemaID: schemaID,
		Conditions: []query.Condition{
			{Field: "status", Op: "==", Value: "active"},
			{Field: "score", Op: ">", Value: 90.0},
		},
		Limit:  10,
		Offset: 0,
	}

	results, err := db.QueryEntities(ctx, req)
	if err != nil {
		t.Fatalf("QueryEntities failed: %v", err)
	}

	if len(results) != 1 {
		t.Fatalf("Expected exactly 1 matching result, got %d", len(results))
	}

	returnedScore, ok := results[0].Data["score"].(float64)
	if !ok || returnedScore != 95.0 {
		t.Errorf("Expected score 95.0, got %v", results[0].Data["score"])
	}
}
