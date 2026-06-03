package query

import (
	"encoding/json"
	"fmt"
	"regexp"
	"strings"
)

var validFieldRegex = regexp.MustCompile(`^[a-zA-Z0-9_.]+$`)

type Request struct {
	SchemaID   int         `json:"schema_id"`
	Conditions []Condition `json:"conditions"`
	Joins      []JoinRule  `json:"joins,omitempty"` // NEW: Support for relational joins
	Limit      int         `json:"limit"`
	Offset     int         `json:"offset"`
}

type Condition struct {
	Field string `json:"field"`
	Op    string `json:"op"` // ==, !=, >, <, >=, <=
	Value any    `json:"value"`
}

type JoinRule struct {
	RelationField string      `json:"relation_field"`
	TargetSchema  int         `json:"target_schema_id"`
	Conditions    []Condition `json:"conditions"`
}

func BuildSQL(req Request) (string, []any, error) {
	if req.Limit == 0 || req.Limit > 100 {
		req.Limit = 50
	}

	var whereClauses []string
	var joinClauses []string
	args := []any{req.SchemaID}
	argCounter := 2

	whereClauses = append(whereClauses, "e.schema_id = $1")

	for _, c := range req.Conditions {
		clause, err := buildConditionSQL("e", c, &argCounter, &args)
		if err != nil {
			return "", nil, err
		}
		whereClauses = append(whereClauses, clause)
	}

	for i, j := range req.Joins {
		alias := fmt.Sprintf("j%d", i+1)

		if !validFieldRegex.MatchString(j.RelationField) {
			return "", nil, fmt.Errorf("invalid characters in relation field: %s", j.RelationField)
		}

		relationPath := buildExtractPath("e", strings.Split(j.RelationField, "."))

		joinClauses = append(joinClauses, fmt.Sprintf(
			"JOIN entities %s ON (%s)::uuid = %s.id AND %s.schema_id = $%d",
			alias, relationPath, alias, alias, argCounter,
		))
		args = append(args, j.TargetSchema)
		argCounter++

		for _, c := range j.Conditions {
			clause, err := buildConditionSQL(alias, c, &argCounter, &args)
			if err != nil {
				return "", nil, err
			}
			whereClauses = append(whereClauses, clause)
		}
	}

	whereSQL := strings.Join(whereClauses, " AND ")
	joinSQL := strings.Join(joinClauses, " ")

	q := fmt.Sprintf(`
		SELECT e.id, e.schema_id, e.data
		FROM entities e %s
		WHERE %s
		ORDER BY e.id DESC
		LIMIT $%d OFFSET $%d`,
		joinSQL, whereSQL, argCounter, argCounter+1)

	args = append(args, req.Limit, req.Offset)

	return q, args, nil
}

func buildConditionSQL(alias string, c Condition, argCounter *int, args *[]any) (string, error) {
	if !validFieldRegex.MatchString(c.Field) {
		return "", fmt.Errorf("invalid characters in field name: %s", c.Field)
	}
	keys := strings.Split(c.Field, ".")

	switch c.Op {
	case "==":
		containmentJSON, err := buildNestedJSON(keys, c.Value)
		if err != nil {
			return "", err
		}
		clause := fmt.Sprintf("%s.data @> $%d", alias, *argCounter)
		*args = append(*args, string(containmentJSON))
		*argCounter++
		return clause, nil

	case ">", "<", ">=", "<=":
		jsonPath := buildExtractPath(alias, keys)
		clause := fmt.Sprintf("(%s)::numeric %s $%d", jsonPath, c.Op, *argCounter)
		*args = append(*args, c.Value)
		*argCounter++
		return clause, nil

	case "!=":
		jsonPath := buildExtractPath(alias, keys)
		clause := fmt.Sprintf("%s != $%d", jsonPath, *argCounter)
		*args = append(*args, c.Value)
		*argCounter++
		return clause, nil

	default:
		return "", fmt.Errorf("unsupported operator: %s", c.Op)
	}
}

func buildNestedJSON(keys []string, value any) ([]byte, error) {
	var result any = value
	for i := len(keys) - 1; i >= 0; i-- {
		result = map[string]any{keys[i]: result}
	}
	return json.Marshal(result)
}

func buildExtractPath(alias string, keys []string) string {
	var pb strings.Builder
	fmt.Fprintf(&pb, "%s.data", alias)

	for i, key := range keys {
		if i == len(keys)-1 {
			fmt.Fprintf(&pb, "->>'%s'", key)
		} else {
			fmt.Fprintf(&pb, "->'%s'", key)
		}
	}

	return pb.String()
}
