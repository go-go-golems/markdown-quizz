package cli

import (
	"context"
	"encoding/json"
	"strings"
	"time"

	"github.com/go-go-golems/XXX/internal/restclient"
	"github.com/go-go-golems/glazed/pkg/cmds"
	"github.com/go-go-golems/glazed/pkg/cmds/fields"
	"github.com/go-go-golems/glazed/pkg/cmds/layers"
	"github.com/go-go-golems/glazed/pkg/cmds/schema"
	"github.com/go-go-golems/glazed/pkg/cmds/values"
	"github.com/go-go-golems/glazed/pkg/middlewares"
	"github.com/go-go-golems/glazed/pkg/types"
	pkgerrors "github.com/pkg/errors"
)

type QuizSubmitSettings struct {
	DocumentID    int64  `glazed.parameter:"document-id"`
	FormID        string `glazed.parameter:"form-id"`
	ResponsesJSON string `glazed.parameter:"responses-json"`
	ResponsesFile string `glazed.parameter:"responses-file"`
}

type QuizSubmitCommand struct {
	*cmds.CommandDefinition
}

func NewQuizSubmitCommand() (*QuizSubmitCommand, error) {
	apiSection, err := newAPISection()
	if err != nil {
		return nil, err
	}

	section, err := schema.NewSection(
		"quiz",
		"Quiz",
		schema.WithDescription("Quiz submission"),
		schema.WithFields(
			fields.New("document-id", fields.TypeInteger,
				fields.WithHelp("Document ID"),
				fields.WithDefault(0),
			),
			fields.New("form-id", fields.TypeString,
				fields.WithHelp("Form ID"),
				fields.WithDefault(""),
			),
			fields.New("responses-json", fields.TypeString,
				fields.WithHelp("Responses JSON object (e.g. '{\"q1\":\"a\"}')"),
				fields.WithDefault(""),
			),
			fields.New("responses-file", fields.TypeString,
				fields.WithHelp("Path to a JSON file containing the responses object"),
				fields.WithDefault(""),
			),
		),
	)
	if err != nil {
		return nil, err
	}

	s := schema.NewSchema(schema.WithSections(apiSection, section))
	desc := cmds.NewCommandDefinition(
		"submit",
		cmds.WithShort("Submit a quiz response (single form)"),
		cmds.WithLong("Submit responses for a single form via REST."),
		cmds.WithSchema(s),
	)
	return &QuizSubmitCommand{CommandDefinition: desc}, nil
}

var _ cmds.GlazeCommand = &QuizSubmitCommand{}

func (c *QuizSubmitCommand) RunIntoGlazeProcessor(ctx context.Context, parsedLayers *layers.ParsedLayers, gp middlewares.Processor) error {
	api := &APISettings{}
	if err := values.DecodeSectionInto(parsedLayers, "api", api); err != nil {
		return pkgerrors.Wrap(err, "decode api settings")
	}
	settings := &QuizSubmitSettings{}
	if err := values.DecodeSectionInto(parsedLayers, "quiz", settings); err != nil {
		return pkgerrors.Wrap(err, "decode quiz settings")
	}

	if settings.DocumentID == 0 {
		return pkgerrors.New("document-id is required")
	}
	formID := strings.TrimSpace(settings.FormID)
	if formID == "" {
		return pkgerrors.New("form-id is required")
	}
	if strings.TrimSpace(settings.ResponsesJSON) != "" && strings.TrimSpace(settings.ResponsesFile) != "" {
		return pkgerrors.New("responses-json and responses-file are mutually exclusive")
	}

	responsesJSON := settings.ResponsesJSON
	if strings.TrimSpace(settings.ResponsesFile) != "" {
		s, err := readFileString(settings.ResponsesFile)
		if err != nil {
			return err
		}
		responsesJSON = s
	}
	responses, err := parseJSONMap(responsesJSON)
	if err != nil {
		return err
	}

	client, err := restclient.New(restclient.Options{BaseURL: api.BaseURL, Timeout: time.Duration(api.TimeoutSeconds) * time.Second})
	if err != nil {
		return err
	}

	res, err := client.SubmitQuiz(ctx, restclient.SubmitQuizRequest{
		DocumentID: settings.DocumentID,
		FormID:     formID,
		Responses:  responses,
	})
	if err != nil {
		return err
	}

	return gp.AddRow(ctx, types.NewRow(
		types.MRP("id", res.ID),
		types.MRP("documentId", settings.DocumentID),
		types.MRP("formId", formID),
		types.MRP("score", res.Score),
		types.MRP("maxScore", res.MaxScore),
	))
}

type QuizSubmitBatchSettings struct {
	DocumentID      int64  `glazed.parameter:"document-id"`
	SubmissionsJSON string `glazed.parameter:"submissions-json"`
	SubmissionsFile string `glazed.parameter:"submissions-file"`
}

type QuizSubmitBatchCommand struct {
	*cmds.CommandDefinition
}

func NewQuizSubmitBatchCommand() (*QuizSubmitBatchCommand, error) {
	apiSection, err := newAPISection()
	if err != nil {
		return nil, err
	}

	section, err := schema.NewSection(
		"quiz",
		"Quiz",
		schema.WithDescription("Quiz batch submission"),
		schema.WithFields(
			fields.New("document-id", fields.TypeInteger,
				fields.WithHelp("Document ID"),
				fields.WithDefault(0),
			),
			fields.New("submissions-json", fields.TypeString,
				fields.WithHelp("JSON array of submissions (e.g. '[{\"formId\":\"f1\",\"responses\":{...}}]')"),
				fields.WithDefault(""),
			),
			fields.New("submissions-file", fields.TypeString,
				fields.WithHelp("Path to a JSON file containing the submissions array"),
				fields.WithDefault(""),
			),
		),
	)
	if err != nil {
		return nil, err
	}

	s := schema.NewSchema(schema.WithSections(apiSection, section))
	desc := cmds.NewCommandDefinition(
		"submit-batch",
		cmds.WithShort("Submit quiz responses (batch / submitMultiple)"),
		cmds.WithLong("Submit responses for multiple forms in one request via REST."),
		cmds.WithSchema(s),
	)
	return &QuizSubmitBatchCommand{CommandDefinition: desc}, nil
}

var _ cmds.GlazeCommand = &QuizSubmitBatchCommand{}

func (c *QuizSubmitBatchCommand) RunIntoGlazeProcessor(ctx context.Context, parsedLayers *layers.ParsedLayers, gp middlewares.Processor) error {
	api := &APISettings{}
	if err := values.DecodeSectionInto(parsedLayers, "api", api); err != nil {
		return pkgerrors.Wrap(err, "decode api settings")
	}
	settings := &QuizSubmitBatchSettings{}
	if err := values.DecodeSectionInto(parsedLayers, "quiz", settings); err != nil {
		return pkgerrors.Wrap(err, "decode quiz settings")
	}

	if settings.DocumentID == 0 {
		return pkgerrors.New("document-id is required")
	}
	if strings.TrimSpace(settings.SubmissionsJSON) != "" && strings.TrimSpace(settings.SubmissionsFile) != "" {
		return pkgerrors.New("submissions-json and submissions-file are mutually exclusive")
	}

	subJSON := settings.SubmissionsJSON
	if strings.TrimSpace(settings.SubmissionsFile) != "" {
		s, err := readFileString(settings.SubmissionsFile)
		if err != nil {
			return err
		}
		subJSON = s
	}
	if strings.TrimSpace(subJSON) == "" {
		return pkgerrors.New("submissions-json is required (or use submissions-file)")
	}

	var raw []restclient.SubmitQuizBatchItem
	if err := json.Unmarshal([]byte(subJSON), &raw); err != nil {
		return pkgerrors.Wrap(err, "parse submissions JSON array")
	}
	if len(raw) == 0 {
		return pkgerrors.New("submissions array is empty")
	}
	for i := range raw {
		raw[i].FormID = strings.TrimSpace(raw[i].FormID)
		if raw[i].FormID == "" {
			return pkgerrors.New("each submission requires formId")
		}
		if raw[i].Responses == nil {
			raw[i].Responses = map[string]any{}
		}
	}

	client, err := restclient.New(restclient.Options{BaseURL: api.BaseURL, Timeout: time.Duration(api.TimeoutSeconds) * time.Second})
	if err != nil {
		return err
	}

	res, err := client.SubmitQuizBatch(ctx, restclient.SubmitQuizBatchRequest{
		DocumentID:  settings.DocumentID,
		Submissions: raw,
	})
	if err != nil {
		return err
	}

	for _, r := range res.Results {
		if err := gp.AddRow(ctx, types.NewRow(
			types.MRP("documentId", settings.DocumentID),
			types.MRP("formId", r.FormID),
			types.MRP("score", r.Score),
			types.MRP("maxScore", r.MaxScore),
		)); err != nil {
			return err
		}
	}

	return nil
}
