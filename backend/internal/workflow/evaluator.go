package workflow

import (
	"strings"
)

type Workflow struct {
	ID           int         `json:"id,omitempty"`
	Name         string      `json:"workflow_name"`
	TargetSchema string      `json:"target_schema"`
	IsActive     bool        `json:"is_active"`
	Condition    ASTNode     `json:"condition"`
	Actions      []ActionDef `json:"actions"`
}

type ASTNode struct {
	Op    string    `json:"op"`
	Field string    `json:"field,omitempty"`
	Value any       `json:"value,omitempty"`
	Args  []ASTNode `json:"args,omitempty"`
}

type ActionDef struct {
	Type  string `json:"type"`
	Field string `json:"field"`
	Value any    `json:"value"`
}

func (node ASTNode) Evaluate(payload map[string]any) bool {
	switch strings.ToUpper(node.Op) {
	case "AND":
		if len(node.Args) == 0 {
			return false
		}
		for _, arg := range node.Args {
			if !arg.Evaluate(payload) {
				return false
			}
		}
		return true

	case "OR":
		if len(node.Args) == 0 {
			return false
		}
		for _, arg := range node.Args {
			if arg.Evaluate(payload) {
				return true
			}
		}
		return false

	default:
		payloadVal, exists := extractValue(payload, node.Field)
		if !exists {
			return false
		}
		return compareValues(node.Op, payloadVal, node.Value)
	}
}

func extractValue(payload map[string]any, path string) (any, bool) {
	keys := strings.Split(path, ".")
	var current any = payload

	for _, key := range keys {
		currMap, ok := current.(map[string]any)
		if !ok {
			return nil, false
		}

		val, exists := currMap[key]
		if !exists {
			return nil, false
		}
		current = val
	}
	return current, true
}

func toFloat(val any) (float64, bool) {
	switch v := val.(type) {
	case float64:
		return v, true
	case int:
		return float64(v), true
	case float32:
		return float64(v), true
	default:
		return 0, false
	}
}

func compareValues(op string, payloadVal any, targetVal any) bool {
	if pF, pOk := toFloat(payloadVal); pOk {
		if tF, tOk := toFloat(targetVal); tOk {
			switch op {
			case "==":
				return pF == tF
			case "!=":
				return pF != tF
			case ">":
				return pF > tF
			case "<":
				return pF < tF
			case ">=":
				return pF >= tF
			case "<=":
				return pF <= tF
			}
		}
	}

	switch p := payloadVal.(type) {
	case string:
		t, ok := targetVal.(string)
		if !ok {
			return false
		}
		switch op {
		case "==":
			return p == t
		case "!=":
			return p != t
		}
	case bool:
		t, ok := targetVal.(bool)
		if !ok {
			return false
		}
		switch op {
		case "==":
			return p == t
		case "!=":
			return p != t
		}
	}

	return false
}
