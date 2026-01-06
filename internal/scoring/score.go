package scoring

import (
	stderrors "errors"
	"reflect"
)

type Result struct {
	Score    int
	MaxScore int
}

type missingValue struct{}

func Calculate(definition any, responses map[string]any) (Result, error) {
	fields, err := extractFields(definition)
	if err != nil {
		return Result{}, err
	}

	res := Result{}
	for _, field := range fields {
		correct, ok := field["correct"]
		if !ok {
			continue
		}
		res.MaxScore++

		key := fieldKey(field)
		if key == "" {
			continue
		}

		userAnswer, present := responses[key]
		if !present {
			userAnswer = missingValue{}
		}

		if correctSlice, ok := asSlice(correct); ok {
			userSlice, ok := asSlice(userAnswer)
			if !ok {
				continue
			}
			if len(correctSlice) != len(userSlice) {
				continue
			}
			if sliceContainsAll(userSlice, correctSlice) {
				res.Score++
			}
			continue
		}

		if valuesEqual(userAnswer, correct) {
			res.Score++
		}
	}

	return res, nil
}

func extractFields(definition any) ([]map[string]any, error) {
	m, ok := definition.(map[string]any)
	if !ok {
		return nil, stderrors.New("definition must be an object")
	}

	if f, ok := m["fields"]; ok {
		return coerceFieldsSlice(f), nil
	}

	if form, ok := m["form"].(map[string]any); ok {
		if f, ok := form["fields"]; ok {
			return coerceFieldsSlice(f), nil
		}
	}

	return []map[string]any{}, nil
}

func coerceFieldsSlice(v any) []map[string]any {
	s, ok := v.([]any)
	if !ok {
		return []map[string]any{}
	}
	out := make([]map[string]any, 0, len(s))
	for _, it := range s {
		m, ok := it.(map[string]any)
		if !ok {
			continue
		}
		out = append(out, m)
	}
	return out
}

func fieldKey(field map[string]any) string {
	if v, ok := field["name"]; ok {
		if s, ok := v.(string); ok {
			return s
		}
	}
	if v, ok := field["key"]; ok {
		if s, ok := v.(string); ok {
			return s
		}
	}
	return ""
}

func asSlice(v any) ([]any, bool) {
	if _, ok := v.(missingValue); ok {
		return nil, false
	}
	s, ok := v.([]any)
	return s, ok
}

func sliceContainsAll(haystack []any, needles []any) bool {
	for _, n := range needles {
		found := false
		for _, h := range haystack {
			if valuesEqual(h, n) {
				found = true
				break
			}
		}
		if !found {
			return false
		}
	}
	return true
}

func valuesEqual(a any, b any) bool {
	if _, ok := a.(missingValue); ok {
		return false
	}
	if _, ok := b.(missingValue); ok {
		return false
	}

	na, aIsNumber := normalizeNumber(a)
	nb, bIsNumber := normalizeNumber(b)
	if aIsNumber && bIsNumber {
		return na == nb
	}

	return reflect.DeepEqual(a, b)
}

func normalizeNumber(v any) (float64, bool) {
	switch t := v.(type) {
	case int:
		return float64(t), true
	case int8:
		return float64(t), true
	case int16:
		return float64(t), true
	case int32:
		return float64(t), true
	case int64:
		return float64(t), true
	case uint:
		return float64(t), true
	case uint8:
		return float64(t), true
	case uint16:
		return float64(t), true
	case uint32:
		return float64(t), true
	case uint64:
		return float64(t), true
	case float32:
		return float64(t), true
	case float64:
		return t, true
	default:
		return 0, false
	}
}
