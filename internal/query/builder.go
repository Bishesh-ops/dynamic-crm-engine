package query

import (
	"encoding/json"
	"fmt"
	"strings"
)

type Request struct {
	SchemaID   int         `json:"schema_id"`
	Conditions []Condition `json:"conditions"`
	Limit      int         `json:"limit"`
	Offset     int         `json:"offset"`
}

type Condition struct {
	Field string `json:"field"`
	Op    string `json:"op"` // ==, !=, >, <, >=, <=
	Value any    `json:"value"`
}

func BuildSQL(req Request) (string, []any, error) {
	if req.Limit == 0 || req.Limit > 100 {
		req.Limit = 50 // Safe default
	}

	var whereClauses []string
	args := []any{req.SchemaID}
	argCounter := 2

	whereClauses = append(whereClauses, "schema_id = $1")

	for _, c := range req.Conditions {
		if strings.ContainsAny(c.Field, "';\"\\") {
			return "", nil, fmt.Errorf("invalid characters in field name: %s", c.Field)
		}

		keys := strings.Split(c.Field, ".")

		switch c.Op {
		case "==":
			containmentJSON, err := buildNestedJSON(keys, c.Value)
			if err != nil {
				return "", nil, err
			}
			whereClauses = append(whereClauses, fmt.Sprintf("data @> $%d", argCounter))
			args = append(args, string(containmentJSON))
			argCounter++

		case ">", "<", ">=", "<=":
			jsonPath := buildExtractPath(keys)
			whereClauses = append(whereClauses, fmt.Sprintf("(%s)::numeric %s $%d", jsonPath, c.Op, argCounter))
			args = append(args, c.Value)
			argCounter++

		case "!=":
			jsonPath := buildExtractPath(keys)
			whereClauses = append(whereClauses, fmt.Sprintf("%s != $%d", jsonPath, argCounter))
			args = append(args, c.Value)
			argCounter++

		default:
			return "", nil, fmt.Errorf("unsupported operator: %s", c.Op)
		}
	}

	whereSQL := strings.Join(whereClauses, " AND ")

	q := fmt.Sprintf(`
		SELECT id, schema_id, data 
		FROM entities 
		WHERE %s 
		ORDER BY id DESC 
		LIMIT $%d OFFSET $%d`,
		whereSQL, argCounter, argCounter+1)

	args = append(args, req.Limit, req.Offset)

	return q, args, nil
}
func buildNestedJSON(keys []string, value any) ([]byte, error) {
	var result any = value
	for i := len(keys) - 1; i >= 0; i-- {
		result = map[string]any{keys[i]: result}
	}
	return json.Marshal(result)
}

func buildExtractPath(keys []string) string {
	var pb strings.Builder
	pb.WriteString("data")

	for i, key := range keys {
		if i == len(keys)-1 {
			fmt.Fprintf(&pb, "->>'%s'", key)
		} else {
			fmt.Fprintf(&pb, "->'%s'", key)
		}
	}

	return pb.String()
}
