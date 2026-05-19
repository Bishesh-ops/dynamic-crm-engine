package query

import (
	"strings"
	"testing"
)

func TestBuildSQL(t *testing.T) {
	req := Request{
		SchemaID: 1,
		Conditions: []Condition{
			{Field: "status", Op: "==", Value: "active"},
			{Field: "profile.score", Op: ">", Value: 90.0},
		},
		Limit:  10,
		Offset: 0,
	}

	sqlStr, args, err := BuildSQL(req)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}

	expectedSQL := `
		SELECT id, schema_id, data 
		FROM entities 
		WHERE schema_id = $1 AND data @> $2 AND (data->'profile'->>'score')::numeric > $3 
		ORDER BY id DESC LIMIT $4 OFFSET $5`

	cleanSQL := strings.Join(strings.Fields(sqlStr), " ")
	cleanExpected := strings.Join(strings.Fields(expectedSQL), " ")

	if cleanSQL != cleanExpected {
		t.Errorf("\nExpected: %s\nGot:      %s", cleanExpected, cleanSQL)
	}

	expectedArgs := []any{1, `{"status":"active"}`, 90.0, 10, 0}

	if len(args) != len(expectedArgs) {
		t.Fatalf("Expected %d args, got %d", len(expectedArgs), len(args))
	}

	for i := range args {
		val1 := args[i]
		if b, ok := val1.([]byte); ok {
			val1 = string(b)
		}

		if val1 != expectedArgs[i] {
			t.Errorf("Arg %d: expected %v, got %v", i, expectedArgs[i], val1)
		}
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
