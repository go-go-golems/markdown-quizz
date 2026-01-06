package quizdsl

import (
	"encoding/json"

	pkgerrors "github.com/pkg/errors"
	"gopkg.in/yaml.v3"
)

func ParseYAMLDefinition(yamlContent string) (any, error) {
	var v any
	if err := yaml.Unmarshal([]byte(yamlContent), &v); err != nil {
		return nil, pkgerrors.Wrap(err, "unmarshal yaml")
	}
	return normalizeYAMLValue(v), nil
}

func MarshalDefinitionJSON(definition any) (string, error) {
	b, err := json.Marshal(definition)
	if err != nil {
		return "", pkgerrors.Wrap(err, "marshal definition json")
	}
	return string(b), nil
}

func normalizeYAMLValue(v any) any {
	switch t := v.(type) {
	case map[any]any:
		m := make(map[string]any, len(t))
		for k, v2 := range t {
			m[toStringKey(k)] = normalizeYAMLValue(v2)
		}
		return m
	case map[string]any:
		m := make(map[string]any, len(t))
		for k, v2 := range t {
			m[k] = normalizeYAMLValue(v2)
		}
		return m
	case []any:
		out := make([]any, 0, len(t))
		for _, it := range t {
			out = append(out, normalizeYAMLValue(it))
		}
		return out
	default:
		return v
	}
}

func toStringKey(k any) string {
	switch t := k.(type) {
	case string:
		return t
	default:
		return yamlKeyString(t)
	}
}

func yamlKeyString(v any) string {
	b, err := json.Marshal(v)
	if err != nil {
		return ""
	}
	return string(b)
}
