package cli

import (
	"encoding/json"
	"os"
	"path/filepath"
	"strings"

	"github.com/go-go-golems/glazed/pkg/cmds/fields"
	"github.com/go-go-golems/glazed/pkg/cmds/schema"
	pkgerrors "github.com/pkg/errors"
)

type APISettings struct {
	BaseURL        string `glazed.parameter:"base-url"`
	TimeoutSeconds int    `glazed.parameter:"timeout-seconds"`
}

func newAPISection() (schema.Section, error) {
	return schema.NewSection(
		"api",
		"API",
		schema.WithDescription("REST API connection settings"),
		schema.WithFields(
			fields.New("base-url", fields.TypeString,
				fields.WithHelp("Base URL for the markdown-quizz REST API"),
				fields.WithDefault("http://127.0.0.1:9092"),
			),
			fields.New("timeout-seconds", fields.TypeInteger,
				fields.WithHelp("HTTP client timeout (seconds)"),
				fields.WithDefault(10),
			),
		),
	)
}

func readFileString(p string) (string, error) {
	b, err := os.ReadFile(p)
	if err != nil {
		return "", pkgerrors.Wrap(err, "read file")
	}
	return string(b), nil
}

func guessTitleFromPath(p string) string {
	base := filepath.Base(p)
	ext := filepath.Ext(base)
	base = strings.TrimSuffix(base, ext)
	base = strings.ReplaceAll(base, "-", " ")
	base = strings.ReplaceAll(base, "_", " ")
	base = strings.TrimSpace(base)
	if base == "" {
		return "Untitled"
	}
	return base
}

func parseJSONMap(s string) (map[string]any, error) {
	if strings.TrimSpace(s) == "" {
		return map[string]any{}, nil
	}
	var out map[string]any
	if err := json.Unmarshal([]byte(s), &out); err != nil {
		return nil, pkgerrors.Wrap(err, "parse JSON object")
	}
	if out == nil {
		out = map[string]any{}
	}
	return out, nil
}

func trimOptionalStringPtr(v *string) *string {
	if v == nil {
		return nil
	}
	s := strings.TrimSpace(*v)
	if s == "" {
		return nil
	}
	return &s
}

func derefStringPtr(v *string) any {
	if v == nil {
		return nil
	}
	return *v
}

func derefIntPtr(v *int) any {
	if v == nil {
		return nil
	}
	return *v
}

func derefFloat64Ptr(v *float64) any {
	if v == nil {
		return nil
	}
	return *v
}
