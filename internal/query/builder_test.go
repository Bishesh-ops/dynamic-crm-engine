package query

import (
	"strings"
	"testing"
)

func TestBuildSQL_Base(t *testing.T) {
	req := Request{
		SchemaID: 1,
		Conditions: []Condition{
			{Field: "status", Op: "==", Value: "active"},
			{Field: "profile.score", Op: ">", Value: 90.0},
		},
		Limit:  10,
		Offset: 0,
	}

	sqlStr, _, err := BuildSQL(req)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}

	expectedSQL := `
		SELECT e.id, e.schema_id, e.data
		FROM entities e
		WHERE e.schema_id = $1 AND e.data @> $2 AND (e.data->'profile'->>'score')::numeric > $3
		ORDER BY e.id DESC LIMIT $4 OFFSET $5`

	cleanSQL := strings.Join(strings.Fields(sqlStr), " ")
	cleanExpected := strings.Join(strings.Fields(expectedSQL), " ")

	if cleanSQL != cleanExpected {
		t.Errorf("\nExpected: %s\nGot:      %s", cleanExpected, cleanSQL)
	}
}

func TestBuildSQL_WithJoins(t *testing.T) {
	req := Request{
		SchemaID: 1, // Target: Tickets
		Conditions: []Condition{
			{Field: "status", Op: "==", Value: "open"},
		},
		Joins: []JoinRule{
			{
				RelationField: "customer", // The JSON key holding the UUID
				TargetSchema:  2,          // Target: Contacts
				Conditions: []Condition{
					{Field: "region", Op: "==", Value: "APAC"},
				},
			},
		},
		Limit:  25,
		Offset: 0,
	}

	sqlStr, args, err := BuildSQL(req)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}

	expectedSQL := `
		SELECT e.id, e.schema_id, e.data
		FROM entities e
		JOIN entities j1 ON (e.data->>'customer')::uuid = j1.id AND j1.schema_id = $3
		WHERE e.schema_id = $1 AND e.data @> $2 AND j1.data @> $4
		ORDER BY e.id DESC LIMIT $5 OFFSET $6`

	cleanSQL := strings.Join(strings.Fields(sqlStr), " ")
	cleanExpected := strings.Join(strings.Fields(expectedSQL), " ")

	if cleanSQL != cleanExpected {
		t.Errorf("\nExpected: %s\nGot:      %s", cleanExpected, cleanSQL)
	}

	expectedArgs := []any{1, `{"status":"open"}`, 2, `{"region":"APAC"}`, 25, 0}
	if len(args) != len(expectedArgs) {
		t.Fatalf("Expected %d args, got %d", len(expectedArgs), len(args))
	}
}

func TestBuildSQL_SQLInjectionPrevention(t *testing.T) {
	req := Request{
		SchemaID: 1,
		Conditions: []Condition{
			{Field: "profile.name'; DROP TABLE entities;--", Op: "==", Value: "hacker"},
		},
	}

	_, _, err := BuildSQL(req)
	if err == nil {
		t.Fatal("Expected error when providing malicious field characters, got nil")
	}
}
