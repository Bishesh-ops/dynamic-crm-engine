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
type DLQEvent struct {
	SchemaID   int
	SchemaName string
	Payload    map[string]any
}
type DLQRecord struct {
	ID         int
	SchemaID   int
	SchemaName string
	Payload    map[string]any
}

func (db *DB) GetUnresolvedDLQEvents(ctx context.Context, limit int) ([]DLQRecord, error) {
	query := `SELECT id, schema_id, schema_name, payload FROM dead_letter_queue WHERE resolved = FALSE LIMIT $1`

	rows, err := db.Pool.Query(ctx, query, limit)
	if err != nil {
		return nil, fmt.Errorf("failed to fetch DLQ events: %w", err)
	}
	defer rows.Close()

	var results []DLQRecord
	for rows.Next() {
		var r DLQRecord
		if err := rows.Scan(&r.ID, &r.SchemaID, &r.SchemaName, &r.Payload); err != nil {
			return nil, err
		}
		results = append(results, r)
	}
	return results, rows.Err()
}

func (db *DB) MarkDLQEventsResolved(ctx context.Context, ids []int) error {
	if len(ids) == 0 {
		return nil
	}
	query := `UPDATE dead_letter_queue SET resolved = TRUE WHERE id = ANY($1)`
	_, err := db.Pool.Exec(ctx, query, ids)
	return err
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

func (db *DB) GetEntityByID(ctx context.Context, id string) (*BatchEntity, error) {
	query := `SELECT id, schema_id, data FROM entities WHERE id = $1`
	var e BatchEntity
	err := db.Pool.QueryRow(ctx, query, id).Scan(&e.ID, &e.SchemaID, &e.Data)
	if err != nil {
		if err == pgx.ErrNoRows {
			return nil, fmt.Errorf("entity not found")
		}
		return nil, fmt.Errorf("database error: %w", err)
	}
	return &e, nil
}

func (db *DB) UpdateEntityByID(ctx context.Context, id string, data map[string]any) error {
	query := `UPDATE entities SET data = $1, updated_at = CURRENT_TIMESTAMP WHERE id = $2`

	cmdTag, err := db.Pool.Exec(ctx, query, data, id)
	if err != nil {
		return fmt.Errorf("failed to update entity: %w", err)
	}

	if cmdTag.RowsAffected() == 0 {
		return fmt.Errorf("entity not found")
	}

	return nil
}

func (db *DB) DeleteEntityByID(ctx context.Context, id string) error {
	query := `DELETE FROM entities WHERE id = $1`

	cmdTag, err := db.Pool.Exec(ctx, query, id)
	if err != nil {
		return fmt.Errorf("failed to delete entity: %w", err)
	}

	if cmdTag.RowsAffected() == 0 {
		return fmt.Errorf("entity not found")
	}

	return nil
}

func (db *DB) GetRecentEntities(ctx context.Context, limit int) ([]BatchEntity, error) {
	query := `SELECT id, schema_id, data FROM entities ORDER BY created_at DESC LIMIT $1`

	rows, err := db.Pool.Query(ctx, query, limit)
	if err != nil {
		return nil, fmt.Errorf("failed to fetch recent entities: %w", err)
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

func (db *DB) SaveToDLQ(ctx context.Context, events []DLQEvent, reason string) error {
	if len(events) == 0 {
		return nil
	}

	batch := &pgx.Batch{}
	query := `INSERT INTO dead_letter_queue (schema_id, schema_name, payload, error_reason) VALUES ($1, $2, $3, $4)`

	for _, ev := range events {
		dataJSON, err := json.Marshal(ev.Payload)
		if err != nil {
			continue
		}
		batch.Queue(query, ev.SchemaID, ev.SchemaName, dataJSON, reason)
	}

	br := db.Pool.SendBatch(ctx, batch)
	defer br.Close()
	for i := range events {
		if _, err := br.Exec(); err != nil {
			return fmt.Errorf("DLQ batch insert failed at row %d: %w", i, err)
		}
	}
	return nil
}
