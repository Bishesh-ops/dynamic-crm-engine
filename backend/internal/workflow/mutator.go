package workflow

import (
	"fmt"
	"strings"
)

func ApplyActions(payload map[string]any, actions []ActionDef) error {
	for _, action := range actions {
		if err := applyAction(payload, action); err != nil {
			return fmt.Errorf("failed to apply action '%s' on field '%s': %w", action.Type, action.Field, err)
		}
	}
	return nil
}

func applyAction(payload map[string]any, action ActionDef) error {
	keys := strings.Split(action.Field, ".")
	if len(keys) == 0 {
		return fmt.Errorf("empty field path")
	}

	var current map[string]any = payload
	for i := 0; i < len(keys)-1; i++ {
		key := keys[i]
		next, exists := current[key]
		if !exists {
			newMap := make(map[string]any)
			current[key] = newMap
			current = newMap
		} else {
			nextMap, ok := next.(map[string]any)
			if !ok {
				return fmt.Errorf("path '%s' is not an object", strings.Join(keys[:i+1], "."))
			}
			current = nextMap
		}
	}

	finalKey := keys[len(keys)-1]

	switch strings.ToUpper(action.Type) {
	case "SET":
		current[finalKey] = action.Value

	case "INCREMENT":
		existing, exists := current[finalKey]
		if !exists {
			current[finalKey] = action.Value
			return nil
		}

		existingFloat, ok1 := toFloat(existing)
		addFloat, ok2 := toFloat(action.Value)

		if ok1 && ok2 {
			current[finalKey] = existingFloat + addFloat
		} else {
			return fmt.Errorf("cannot increment non-numeric values")
		}

	case "APPEND":
		existing, exists := current[finalKey]
		if !exists {
			current[finalKey] = []any{action.Value}
			return nil
		}

		slice, ok := existing.([]any)
		if !ok {
			return fmt.Errorf("cannot append to non-array type")
		}
		current[finalKey] = append(slice, action.Value)

	default:
		return fmt.Errorf("unknown action type: %s", action.Type)
	}

	return nil
}
