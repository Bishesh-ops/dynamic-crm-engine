package schema

import (
	"strings"
	"testing"
)

func intPtr(i int) *int           { return &i }
func floatPtr(f float64) *float64 { return &f }

func TestSchemaValidator(t *testing.T) {
	testSchema := &Schema{
		Name: "test_entity",
		Fields: map[string]FieldDefinition{
			"username": {Type: "string", Required: true, MaxLength: intPtr(10)},
			"age":      {Type: "int", Required: false, Min: floatPtr(18)},
			"profile": {
				Type:     "object",
				Required: true,
				Properties: map[string]FieldDefinition{
					"bio":    {Type: "string", Required: false},
					"rating": {Type: "float", Required: true},
				},
			},
		},
	}

	tests := []struct {
		name        string
		payload     map[string]any
		expectErr   bool
		errContains string
	}{
		{
			name: "Valid Payload (Complete)",
			payload: map[string]any{
				"username": "bishesh",
				"age":      23.0,
				"profile": map[string]any{
					"bio":    "Engine builder",
					"rating": 9.5,
				},
			},
			expectErr: false,
		},
		{
			name: "Missing Required Root Field",
			payload: map[string]any{
				"age": 25.0,
				"profile": map[string]any{
					"rating": 5.0,
				},
			},
			expectErr:   true,
			errContains: "missing required field: username",
		},
		{
			name: "String Max Length Violation",
			payload: map[string]any{
				"username": "this_is_way_too_long",
				"profile": map[string]any{
					"rating": 5.0,
				},
			},
			expectErr:   true,
			errContains: "exceeds max length",
		},
		{
			name: "Integer Min Constraint Violation",
			payload: map[string]any{
				"username": "valid",
				"age":      17.0,
				"profile": map[string]any{
					"rating": 5.0,
				},
			},
			expectErr:   true,
			errContains: "must be at least 18",
		},
		{
			name: "Strict Type Check (Float provided for Int)",
			payload: map[string]any{
				"username": "valid",
				"age":      22.5,
				"profile": map[string]any{
					"rating": 5.0,
				},
			},
			expectErr:   true,
			errContains: "expected int",
		},
		{
			name: "Missing Required Nested Field",
			payload: map[string]any{
				"username": "valid",
				"profile": map[string]any{
					"bio": "missing rating",
				},
			},
			expectErr:   true,
			errContains: "missing required field: profile.rating",
		},
		{
			name: "Invalid Type in Nested Object",
			payload: map[string]any{
				"username": "valid",
				"profile": map[string]any{
					"rating": "five",
				},
			},
			expectErr:   true,
			errContains: "expected float",
		},
	}
	for _, tc := range tests {
		t.Run(tc.name, func(t *testing.T) {
			err := testSchema.Validate(tc.payload)

			if tc.expectErr {
				if err == nil {
					t.Fatalf("Expected error containing '%s', got nil", tc.errContains)
				}
				if !strings.Contains(err.Error(), tc.errContains) {
					t.Errorf("Expected error to contain '%s', got: '%v'", tc.errContains, err)
				}
			} else {
				if err != nil {
					t.Fatalf("Expected no error, got: %v", err)
				}
			}
		})
	}
}
