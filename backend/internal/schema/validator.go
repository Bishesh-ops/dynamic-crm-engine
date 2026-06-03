package schema

import (
	"fmt"
	"regexp"
)

type Schema struct {
	Name   string                     `json:"name"`
	Fields map[string]FieldDefinition `json:"fields"`
}

type FieldDefinition struct {
	Type         string                     `json:"type"`
	Required     bool                       `json:"required"`
	Properties   map[string]FieldDefinition `json:"properties,omitempty"`
	Min          *float64                   `json:"min,omitempty"`
	MaxLength    *int                       `json:"max_length,omitempty"`
	TargetSchema string                     `json:"target_schema,omitempty"`
}

func (s *Schema) Validate(payload map[string]any) error {
	return validateObject(payload, s.Fields, "")
}

func validateObject(payload map[string]any, fields map[string]FieldDefinition, path string) error {
	for key, def := range fields {
		val, exists := payload[key]

		currentPath := key

		if path != "" {
			currentPath = path + "." + key
		}

		if !exists {
			if def.Required {
				return fmt.Errorf("missing required field: %s", currentPath)
			}
			continue
		}

		if err := validateField(val, def, currentPath); err != nil {
			return err
		}
	}
	return nil
}

func validateField(value any, def FieldDefinition, path string) error {
	switch def.Type {
	case "string":
		strVal, ok := value.(string)
		if !ok {
			return fmt.Errorf("invalid type for %s: expected string", path)
		}
		if def.MaxLength != nil && len(strVal) > *def.MaxLength {
			return fmt.Errorf("field %s exceeds max length of %d", path, *def.MaxLength)
		}
	case "float":
		switch v := value.(type) {
		case float64:
			if def.Min != nil && v < *def.Min {
				return fmt.Errorf("field %s must be at least %v", path, *def.Min)
			}
		case int:
			if def.Min != nil && float64(v) < *def.Min {
				return fmt.Errorf("field %s must be at least %v", path, *def.Min)
			}
		default:
			return fmt.Errorf("invalid type for %s: expected float", path)
		}

	case "int":
		switch v := value.(type) {
		case float64:
			if v != float64(int(v)) {
				return fmt.Errorf("invalid tupe for %s: expected int", path)
			}
			if def.Min != nil && v < *def.Min {
				return fmt.Errorf("field %s must be at least %v", path, *def.Min)
			}
		case int:
			if def.Min != nil && float64(v) < *def.Min {
				return fmt.Errorf("field %s must be at least %v", path, *def.Min)
			}
		default:
			return fmt.Errorf("invalid type for %s: expected int", path)

		}
	case "relation":
		strVal, ok := value.(string)
		if !ok {
			return fmt.Errorf("invalid type for  %s: expected relation UUID string", path)
		}
		uuidRegex := regexp.MustCompile(`^[a-fA-F0-9]{8}-[a-fA-F0-9]{4}-[a-fA-F0-9]{4}-[a-fA-F0-9]{4}-[a-fA-F0-9]{12}$`)
		if !uuidRegex.MatchString(strVal) {
			return fmt.Errorf("field %s must be a valid UUID pointing to a %s record", path, def.TargetSchema)
		}
	case "object":
		objMap, ok := value.(map[string]any)
		if !ok {
			return fmt.Errorf("invalid type for %s: expected object", path)
		}
		return validateObject(objMap, def.Properties, path)
	default:
		return fmt.Errorf("unknown schema type '%s' for field %s", def.Type, path)
	}
	return nil

}
