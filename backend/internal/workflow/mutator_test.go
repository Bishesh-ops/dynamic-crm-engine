package workflow

import (
	"testing"
)

func TestApplyActions(t *testing.T) {
	tests := []struct {
		name           string
		initialPayload map[string]any
		actions        []ActionDef
		validate       func(*testing.T, map[string]any)
		expectErr      bool
	}{
		{
			name:           "SET Top-Level Field",
			initialPayload: map[string]any{"status": "new"},
			actions: []ActionDef{
				{Type: "SET", Field: "status", Value: "hot"},
				{Type: "SET", Field: "priority", Value: 1},
			},
			validate: func(t *testing.T, p map[string]any) {
				if p["status"] != "hot" {
					t.Errorf("Expected status to be 'hot', got %v", p["status"])
				}
				if p["priority"] != 1 {
					t.Errorf("Expected priority to be 1, got %v", p["priority"])
				}
			},
		},
		{
			name:           "SET Nested Field (Auto-Create Maps)",
			initialPayload: map[string]any{},
			actions: []ActionDef{
				{Type: "SET", Field: "profile.tags.primary", Value: "enterprise"},
			},
			validate: func(t *testing.T, p map[string]any) {
				profile := p["profile"].(map[string]any)
				tags := profile["tags"].(map[string]any)
				if tags["primary"] != "enterprise" {
					t.Errorf("Expected nested set to work, got %v", tags["primary"])
				}
			},
		},
		{
			name:           "INCREMENT Existing Value",
			initialPayload: map[string]any{"score": 5.0},
			actions: []ActionDef{
				{Type: "INCREMENT", Field: "score", Value: 2.5},
			},
			validate: func(t *testing.T, p map[string]any) {
				if p["score"] != 7.5 {
					t.Errorf("Expected score to be 7.5, got %v", p["score"])
				}
			},
		},
		{
			name: "APPEND to Array",
			initialPayload: map[string]any{
				"tags": []any{"b2b"},
			},
			actions: []ActionDef{
				{Type: "APPEND", Field: "tags", Value: "saas"},
			},
			validate: func(t *testing.T, p map[string]any) {
				tags := p["tags"].([]any)
				if len(tags) != 2 || tags[1] != "saas" {
					t.Errorf("Expected array to append, got %v", tags)
				}
			},
		},
	}

	for _, tc := range tests {
		t.Run(tc.name, func(t *testing.T) {
			err := ApplyActions(tc.initialPayload, tc.actions)
			if tc.expectErr && err == nil {
				t.Fatalf("Expected error but got nil")
			}
			if !tc.expectErr && err != nil {
				t.Fatalf("Expected no error, got: %v", err)
			}
			if tc.validate != nil {
				tc.validate(t, tc.initialPayload)
			}
		})
	}
}
