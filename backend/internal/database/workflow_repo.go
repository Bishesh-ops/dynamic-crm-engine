package database

import (
	"context"
	"encoding/json"
	"fmt"

	"github.com/bisheshops/dynamic-crm-engine/internal/workflow"
)

func (db *DB) SaveWorkflow(ctx context.Context, wf workflow.Workflow) (int, error) {
	query := `
		INSERT INTO workflows (name, target_schema, is_active, condition, actions)
		VALUES ($1, $2, $3, $4, $5)
		RETURNING id;
	`
	condJSON, err := json.Marshal(wf.Condition)
	if err != nil {
		return 0, fmt.Errorf("failed to marshal workflow condition: %w", err)
	}
	actionsJSON, err := json.Marshal(wf.Actions)
	if err != nil {
		return 0, fmt.Errorf("failed to marshal workflow action: %w", err)
	}
	var id int
	err = db.Pool.QueryRow(ctx, query, wf.Name, wf.TargetSchema, wf.IsActive, condJSON, actionsJSON).Scan(&id)
	if err != nil {
		return 0, fmt.Errorf("failed to insert workflow actions: %w", err)
	}
	return id, nil
}

func (db *DB) LoadWorkflows(ctx context.Context) ([]workflow.Workflow, error) {
	query := `SELECT id, name, target_schema, is_active, condition, actions FROM workflows WHERE is_active = true`

	rows, err := db.Pool.Query(ctx, query)
	if err != nil {
		return nil, fmt.Errorf("failed to query workflows: %w", err)
	}
	defer rows.Close()

	var workflows []workflow.Workflow
	for rows.Next() {
		var wf workflow.Workflow
		var condBytes, actionsBytes []byte

		if err := rows.Scan(&wf.ID, &wf.Name, &wf.TargetSchema, &wf.IsActive, &condBytes, &actionsBytes); err != nil {
			return nil, fmt.Errorf("failed to scan workflow row: %w", err)
		}

		if err := json.Unmarshal(condBytes, &wf.Condition); err != nil {
			return nil, fmt.Errorf("failed to unmarshal condition for workflow %s: %w", wf.Name, err)
		}

		if err := json.Unmarshal(actionsBytes, &wf.Actions); err != nil {
			return nil, fmt.Errorf("failed to unmarshal actions for workflow %s: %w", wf.Name, err)
		}

		workflows = append(workflows, wf)
	}

	if err := rows.Err(); err != nil {
		return nil, fmt.Errorf("row iteration error: %w", err)
	}

	return workflows, nil
}
