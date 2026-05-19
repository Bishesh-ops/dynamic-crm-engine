package database

import (
	"context"
	"encoding/json"
	"errors"
	"fmt"
	"github.com/bisheshops/dynamic-crm-engine/internal/query"
	"github.com/bisheshops/dynamic-crm-engine/internal/schema"
	"github.com/jackc/pgx/v5"
)

type BatchEntity struct {
	ID       string         `json:"id,omitempty"`
	SchemaID int            `json:"schema_id"`
	Data     map[string]any `json:"data"`
}

func (db *DB) SaveEntityBatch(ctx context.Context, entities []BatchEntity) error {
	if len(entities) == 0 {
		return nil
	}
	batch := &pgx.Batch{}
	query := `INSERT INTO entities (schema_id, data) VALUES ($1, $2)`

	for _, e := range entities {
		dataJSON, err := json.Marshal(e.Data)
		if err != nil {
			return fmt.Errorf("failed to marshal the batch entity data: %w", err)
		}
		batch.Queue(query, e.SchemaID, dataJSON)
	}

	br := db.Pool.SendBatch(ctx, batch)

	defer br.Close()
	for i := range len(entities) {
		_, err := br.Exec()
		if err != nil {
			return fmt.Errorf("batch insert failed at row %d: %w", i, err)
		}
	}

	return nil
}

func (db *DB) GetSchemaByName(ctx context.Context, name string) (*schema.Schema, int, error) {
	var s schema.Schema
	var id int
	var structuredBytes []byte

	query := `SELECT id, structure FROM schemas WHERE name = $1`

	err := db.Pool.QueryRow(ctx, query, name).Scan(&id, &structuredBytes)
	if err != nil {
		if errors.Is(err, pgx.ErrNoRows) {
			return nil, 0, fmt.Errorf("schema '%s' is not found", name)
		}
		return nil, 0, fmt.Errorf("failed to fetch schema: '%w'", err)
	}

	if err := json.Unmarshal(structuredBytes, &s); err != nil {
		return nil, 0, fmt.Errorf("failed to parse schema structure: %w", err)
	}
	s.Name = name
	return &s, id, nil
}

func (db *DB) SaveEntity(ctx context.Context, schemaID int, data map[string]any) (string, error) {
	dataJSON, err := json.Marshal(data)
	if err != nil {
		return "", fmt.Errorf("failed to marshal entity data: %w", err)
	}

	var entityID string
	query := `
		INSERT INTO entities (schema_id, data)
		VALUES ($1, $2)
		RETURNING id;
	`

	err = db.Pool.QueryRow(ctx, query, schemaID, dataJSON).Scan(&entityID)
	if err != nil {
		return "", fmt.Errorf("failed to insert entity: %w", err)
	}

	return entityID, nil
}

func (db *DB) QueryEntities(ctx context.Context, req query.Request) ([]BatchEntity, error) {
	sqlStr, args, err := query.BuildSQL(req)
	if err != nil {
		return nil, fmt.Errorf("failed to build query: %w", err)
	}

	rows, err := db.Pool.Query(ctx, sqlStr, args...)
	if err != nil {
		return nil, fmt.Errorf("query execution failed: %w", err)
	}
	defer rows.Close()

	var results []BatchEntity
	for rows.Next() {
		var e BatchEntity

		if err := rows.Scan(&e.ID, &e.SchemaID, &e.Data); err != nil {
			return nil, err
		}
		results = append(results, e)
	}

	if err := rows.Err(); err != nil {
		return nil, err
	}

	return results, nil
}
