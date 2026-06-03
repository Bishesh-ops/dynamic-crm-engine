package database

import (
	"context"
	"encoding/json"
	"fmt"
)

func (db *DB) SaveSchema(ctx context.Context, name string, structure any) (int, error) {
	query := `
		INSERT INTO schemas (name, structure)
		VALUES ($1, $2)
		RETURNING id;
	`

	structureJSON, err := json.Marshal(structure)
	if err != nil {
		return 0, fmt.Errorf("failed to marshal schema structure: %w", err)
	}
	var id int
	err = db.Pool.QueryRow(ctx, query, name, structureJSON).Scan(&id)
	if err != nil {
		return 0, fmt.Errorf("failed to insert schema: %w", err)
	}
	return id, nil
}
