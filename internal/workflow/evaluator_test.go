package workflow

import (
	"testing"
)

func TestEvaluateAST(t *testing.T) {
	payload := map[string]any{
		"status": "new",
		"score":  85.5,
		"profile": map[string]any{
			"age":  25,
			"tags": []string{"b2b", "enterprise"},
		},
	}

	tests := []struct {
		name     string
		node     ASTNode
		expected bool
	}{
		{
			name:     "Simple String Match",
			node:     ASTNode{Op: "==", Field: "status", Value: "new"},
			expected: true,
		},
		{
			name:     "Simple String Mismatch",
			node:     ASTNode{Op: "==", Field: "status", Value: "contacted"},
			expected: false,
		},
		{
			name:     "Numeric Greater Than",
			node:     ASTNode{Op: ">", Field: "score", Value: 80.0},
			expected: true,
		},
		{
			name:     "Dot Notation Extraction (Nested Int)",
			node:     ASTNode{Op: ">=", Field: "profile.age", Value: 18.0},
			expected: true,
		},
		{
			name:     "Missing Field Returns False",
			node:     ASTNode{Op: "==", Field: "profile.salary", Value: 100000},
			expected: false,
		},
		{
			name: "AND Condition (Both True)",
			node: ASTNode{
				Op: "AND",
				Args: []ASTNode{
					{Op: "==", Field: "status", Value: "new"},
					{Op: ">", Field: "score", Value: 80.0},
				},
			},
			expected: true,
		},
		{
			name: "AND Condition (One False)",
			node: ASTNode{
				Op: "AND",
				Args: []ASTNode{
					{Op: "==", Field: "status", Value: "new"},
					{Op: ">", Field: "score", Value: 90.0},
				},
			},
			expected: false,
		},
		{
			name: "OR Condition (One True)",
			node: ASTNode{
				Op: "OR",
				Args: []ASTNode{
					{Op: "==", Field: "status", Value: "archived"},
					{Op: "<", Field: "profile.age", Value: 30.0},
				},
			},
			expected: true,
		},
	}

	for _, tc := range tests {
		t.Run(tc.name, func(t *testing.T) {
			result := tc.node.Evaluate(payload)
			if result != tc.expected {
				t.Errorf("Expected %v, got %v for AST: %+v", tc.expected, result, tc.node)
			}
		})
	}
}
