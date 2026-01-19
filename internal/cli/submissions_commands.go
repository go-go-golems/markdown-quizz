package cli

import (
	"context"
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

type SubmissionsMineCommand struct {
	*cmds.CommandDefinition
}

func NewSubmissionsMineCommand() (*SubmissionsMineCommand, error) {
	apiSection, err := newAPISection()
	if err != nil {
		return nil, err
	}

	s := schema.NewSchema(schema.WithSections(apiSection))
	desc := cmds.NewCommandDefinition(
		"mine",
		cmds.WithShort("List my quiz submissions"),
		cmds.WithLong("List submissions for the current user (no-auth mode: userId=1)."),
		cmds.WithSchema(s),
	)
	return &SubmissionsMineCommand{CommandDefinition: desc}, nil
}

var _ cmds.GlazeCommand = &SubmissionsMineCommand{}

func (c *SubmissionsMineCommand) RunIntoGlazeProcessor(ctx context.Context, parsedLayers *layers.ParsedLayers, gp middlewares.Processor) error {
	api := &APISettings{}
	if err := values.DecodeSectionInto(parsedLayers, "api", api); err != nil {
		return pkgerrors.Wrap(err, "decode api settings")
	}

	client, err := restclient.New(restclient.Options{BaseURL: api.BaseURL, Timeout: time.Duration(api.TimeoutSeconds) * time.Second})
	if err != nil {
		return err
	}

	items, err := client.ListMySubmissions(ctx)
	if err != nil {
		return err
	}

	for _, it := range items {
		s := it.Submission
		var scorePct *float64
		if s.Score != nil && s.MaxScore != nil && *s.MaxScore > 0 {
			v := (float64(*s.Score) / float64(*s.MaxScore)) * 100
			scorePct = &v
		}

		if err := gp.AddRow(ctx, types.NewRow(
			types.MRP("submissionId", s.ID),
			types.MRP("documentId", s.DocumentID),
			types.MRP("documentTitle", it.DocumentTitle),
			types.MRP("documentSlug", it.DocumentSlug),
			types.MRP("formId", s.FormID),
			types.MRP("score", derefIntPtr(s.Score)),
			types.MRP("maxScore", derefIntPtr(s.MaxScore)),
			types.MRP("scorePct", derefFloat64Ptr(scorePct)),
			types.MRP("submittedAt", s.SubmittedAt),
		)); err != nil {
			return err
		}
	}

	return nil
}

type SubmissionsByDocumentSettings struct {
	DocumentID int64 `glazed.parameter:"document-id"`
}

type SubmissionsByDocumentCommand struct {
	*cmds.CommandDefinition
}

func NewSubmissionsByDocumentCommand() (*SubmissionsByDocumentCommand, error) {
	apiSection, err := newAPISection()
	if err != nil {
		return nil, err
	}

	section, err := schema.NewSection(
		"submissions",
		"Submissions",
		schema.WithDescription("List submissions"),
		schema.WithFields(
			fields.New("document-id", fields.TypeInteger,
				fields.WithHelp("Document ID"),
				fields.WithDefault(0),
			),
		),
	)
	if err != nil {
		return nil, err
	}

	s := schema.NewSchema(schema.WithSections(apiSection, section))
	desc := cmds.NewCommandDefinition(
		"by-document",
		cmds.WithShort("List submissions for a document"),
		cmds.WithLong("List quiz submissions for a given document."),
		cmds.WithSchema(s),
	)
	return &SubmissionsByDocumentCommand{CommandDefinition: desc}, nil
}

var _ cmds.GlazeCommand = &SubmissionsByDocumentCommand{}

func (c *SubmissionsByDocumentCommand) RunIntoGlazeProcessor(ctx context.Context, parsedLayers *layers.ParsedLayers, gp middlewares.Processor) error {
	api := &APISettings{}
	if err := values.DecodeSectionInto(parsedLayers, "api", api); err != nil {
		return pkgerrors.Wrap(err, "decode api settings")
	}
	settings := &SubmissionsByDocumentSettings{}
	if err := values.DecodeSectionInto(parsedLayers, "submissions", settings); err != nil {
		return pkgerrors.Wrap(err, "decode submissions settings")
	}
	if settings.DocumentID == 0 {
		return pkgerrors.New("document-id is required")
	}

	client, err := restclient.New(restclient.Options{BaseURL: api.BaseURL, Timeout: time.Duration(api.TimeoutSeconds) * time.Second})
	if err != nil {
		return err
	}
	items, err := client.GetDocumentSubmissions(ctx, settings.DocumentID)
	if err != nil {
		return err
	}

	for _, it := range items {
		s := it.Submission
		if err := gp.AddRow(ctx, types.NewRow(
			types.MRP("submissionId", s.ID),
			types.MRP("documentId", settings.DocumentID),
			types.MRP("userId", s.UserID),
			types.MRP("userName", derefStringPtr(it.UserName)),
			types.MRP("formId", s.FormID),
			types.MRP("score", derefIntPtr(s.Score)),
			types.MRP("maxScore", derefIntPtr(s.MaxScore)),
			types.MRP("submittedAt", s.SubmittedAt),
		)); err != nil {
			return err
		}
	}

	return nil
}

type SubmissionsGetSettings struct {
	ID int64 `glazed.parameter:"id"`
}

type SubmissionsGetCommand struct {
	*cmds.CommandDefinition
}

func NewSubmissionsGetCommand() (*SubmissionsGetCommand, error) {
	apiSection, err := newAPISection()
	if err != nil {
		return nil, err
	}

	section, err := schema.NewSection(
		"submissions",
		"Submissions",
		schema.WithDescription("Fetch submission detail"),
		schema.WithFields(
			fields.New("id", fields.TypeInteger,
				fields.WithHelp("Submission ID"),
				fields.WithDefault(0),
			),
		),
	)
	if err != nil {
		return nil, err
	}

	s := schema.NewSchema(schema.WithSections(apiSection, section))
	desc := cmds.NewCommandDefinition(
		"get",
		cmds.WithShort("Get a submission by ID"),
		cmds.WithLong("Fetch a submission detail (includes formDefinition) via REST."),
		cmds.WithSchema(s),
	)
	return &SubmissionsGetCommand{CommandDefinition: desc}, nil
}

var _ cmds.GlazeCommand = &SubmissionsGetCommand{}

func (c *SubmissionsGetCommand) RunIntoGlazeProcessor(ctx context.Context, parsedLayers *layers.ParsedLayers, gp middlewares.Processor) error {
	api := &APISettings{}
	if err := values.DecodeSectionInto(parsedLayers, "api", api); err != nil {
		return pkgerrors.Wrap(err, "decode api settings")
	}
	settings := &SubmissionsGetSettings{}
	if err := values.DecodeSectionInto(parsedLayers, "submissions", settings); err != nil {
		return pkgerrors.Wrap(err, "decode submissions settings")
	}
	if settings.ID == 0 {
		return pkgerrors.New("id is required")
	}

	client, err := restclient.New(restclient.Options{BaseURL: api.BaseURL, Timeout: time.Duration(api.TimeoutSeconds) * time.Second})
	if err != nil {
		return err
	}

	item, err := client.GetSubmissionByID(ctx, settings.ID)
	if err != nil {
		return err
	}

	respCount := 0
	if item.Submission.Responses != nil {
		respCount = len(item.Submission.Responses)
	}

	row := types.NewRow(
		types.MRP("submissionId", item.Submission.ID),
		types.MRP("documentId", item.Submission.DocumentID),
		types.MRP("documentTitle", item.DocumentTitle),
		types.MRP("documentSlug", item.DocumentSlug),
		types.MRP("formId", item.Submission.FormID),
		types.MRP("score", derefIntPtr(item.Submission.Score)),
		types.MRP("maxScore", derefIntPtr(item.Submission.MaxScore)),
		types.MRP("submittedAt", item.Submission.SubmittedAt),
		types.MRP("responsesCount", respCount),
		types.MRP("responses", item.Submission.Responses),
		types.MRP("formDefinition", item.FormDefinition),
	)

	// Small convenience: emit a short hint in human output modes.
	if strings.TrimSpace(item.DocumentSlug) == "" {
		row.Set("documentUrl", nil)
	} else {
		row.Set("documentUrl", "/documents/"+item.DocumentSlug)
	}

	return gp.AddRow(ctx, row)
}
