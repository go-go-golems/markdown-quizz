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

type DocumentsListSettings struct {
	Scope string `glazed.parameter:"scope"`
}

type DocumentsListCommand struct {
	*cmds.CommandDefinition
}

func NewDocumentsListCommand() (*DocumentsListCommand, error) {
	apiSection, err := newAPISection()
	if err != nil {
		return nil, err
	}

	section, err := schema.NewSection(
		"documents",
		"Documents",
		schema.WithDescription("Documents queries"),
		schema.WithFields(
			fields.New("scope", fields.TypeChoice,
				fields.WithHelp("List scope"),
				fields.WithChoices("all", "mine"),
				fields.WithDefault("all"),
			),
		),
	)
	if err != nil {
		return nil, err
	}

	s := schema.NewSchema(schema.WithSections(apiSection, section))

	desc := cmds.NewCommandDefinition(
		"list",
		cmds.WithShort("List documents"),
		cmds.WithLong("List documents from the REST API."),
		cmds.WithSchema(s),
	)

	return &DocumentsListCommand{CommandDefinition: desc}, nil
}

var _ cmds.GlazeCommand = &DocumentsListCommand{}

func (c *DocumentsListCommand) RunIntoGlazeProcessor(ctx context.Context, parsedLayers *layers.ParsedLayers, gp middlewares.Processor) error {
	api := &APISettings{}
	if err := values.DecodeSectionInto(parsedLayers, "api", api); err != nil {
		return pkgerrors.Wrap(err, "decode api settings")
	}
	settings := &DocumentsListSettings{}
	if err := values.DecodeSectionInto(parsedLayers, "documents", settings); err != nil {
		return pkgerrors.Wrap(err, "decode documents settings")
	}

	client, err := restclient.New(restclient.Options{BaseURL: api.BaseURL, Timeout: time.Duration(api.TimeoutSeconds) * time.Second})
	if err != nil {
		return err
	}
	docs, err := client.ListDocuments(ctx, settings.Scope)
	if err != nil {
		return err
	}

	for _, d := range docs {
		if err := gp.AddRow(ctx, types.NewRow(
			types.MRP("id", d.ID),
			types.MRP("title", d.Title),
			types.MRP("slug", d.Slug),
			types.MRP("category", derefStringPtr(d.Category)),
			types.MRP("isPublished", d.IsPublished),
			types.MRP("updatedAt", d.UpdatedAt),
		)); err != nil {
			return err
		}
	}

	return nil
}

type DocumentsCreateSettings struct {
	Title       string  `glazed.parameter:"title"`
	Content     string  `glazed.parameter:"content"`
	ContentFile string  `glazed.parameter:"content-file"`
	Description *string `glazed.parameter:"description"`
	Category    *string `glazed.parameter:"category"`
	IsPublished bool    `glazed.parameter:"is-published"`
}

type DocumentsCreateCommand struct {
	*cmds.CommandDefinition
}

func NewDocumentsCreateCommand() (*DocumentsCreateCommand, error) {
	apiSection, err := newAPISection()
	if err != nil {
		return nil, err
	}

	section, err := schema.NewSection(
		"documents",
		"Documents",
		schema.WithDescription("Create documents"),
		schema.WithFields(
			fields.New("title", fields.TypeString,
				fields.WithHelp("Document title (required unless --content-file is a real path)"),
				fields.WithDefault(""),
			),
			fields.New("content", fields.TypeString,
				fields.WithHelp("Document content (mutually exclusive with --content-file)"),
				fields.WithDefault(""),
			),
			fields.New("content-file", fields.TypeString,
				fields.WithHelp("Read content from a file path, or '-' for stdin (mutually exclusive with --content)"),
				fields.WithDefault(""),
			),
			fields.New("description", fields.TypeString,
				fields.WithHelp("Optional description"),
				fields.WithDefault(""),
			),
			fields.New("category", fields.TypeString,
				fields.WithHelp("Optional category"),
				fields.WithDefault(""),
			),
			fields.New("is-published", fields.TypeBool,
				fields.WithHelp("Whether to publish the document"),
				fields.WithDefault(false),
			),
		),
	)
	if err != nil {
		return nil, err
	}

	s := schema.NewSchema(schema.WithSections(apiSection, section))
	desc := cmds.NewCommandDefinition(
		"create",
		cmds.WithShort("Create a document (extracts <form> quizzes)"),
		cmds.WithLong("Create a document via REST; any <form id=\"...\"> YAML </form> blocks are extracted into quiz forms by the backend."),
		cmds.WithSchema(s),
	)

	return &DocumentsCreateCommand{CommandDefinition: desc}, nil
}

var _ cmds.GlazeCommand = &DocumentsCreateCommand{}

func (c *DocumentsCreateCommand) RunIntoGlazeProcessor(ctx context.Context, parsedLayers *layers.ParsedLayers, gp middlewares.Processor) error {
	api := &APISettings{}
	if err := values.DecodeSectionInto(parsedLayers, "api", api); err != nil {
		return pkgerrors.Wrap(err, "decode api settings")
	}
	settings := &DocumentsCreateSettings{}
	if err := values.DecodeSectionInto(parsedLayers, "documents", settings); err != nil {
		return pkgerrors.Wrap(err, "decode documents settings")
	}

	title := strings.TrimSpace(settings.Title)

	content := strings.TrimSpace(settings.Content)
	contentFile := strings.TrimSpace(settings.ContentFile)
	if content != "" && contentFile != "" {
		return pkgerrors.New("content and content-file are mutually exclusive")
	}
	if contentFile != "" {
		s, err := readFileOrStdin(contentFile)
		if err != nil {
			return err
		}
		content = s
	}

	if strings.TrimSpace(content) == "" {
		return pkgerrors.New("content is required (use --content or --content-file)")
	}

	if title == "" {
		if contentFile != "" && contentFile != "-" {
			title = guessTitleFromPath(contentFile)
		} else {
			return pkgerrors.New("title is required")
		}
	}

	var description *string
	if settings.Description != nil && strings.TrimSpace(*settings.Description) != "" {
		v := strings.TrimSpace(*settings.Description)
		description = &v
	}
	var category *string
	if settings.Category != nil && strings.TrimSpace(*settings.Category) != "" {
		v := strings.TrimSpace(*settings.Category)
		category = &v
	}

	client, err := restclient.New(restclient.Options{BaseURL: api.BaseURL, Timeout: time.Duration(api.TimeoutSeconds) * time.Second})
	if err != nil {
		return err
	}
	res, err := client.CreateDocument(ctx, restclient.CreateDocumentRequest{
		Title:       title,
		Content:     content,
		Description: description,
		Category:    category,
		IsPublished: settings.IsPublished,
	})
	if err != nil {
		return err
	}

	return gp.AddRow(ctx, types.NewRow(
		types.MRP("id", res.ID),
		types.MRP("slug", res.Slug),
		types.MRP("title", title),
	))
}

type DocumentsGetSettings struct {
	ID             int64  `glazed.parameter:"id"`
	Slug           string `glazed.parameter:"slug"`
	IncludeContent bool   `glazed.parameter:"include-content"`
}

type DocumentsGetCommand struct {
	*cmds.CommandDefinition
}

func NewDocumentsGetCommand() (*DocumentsGetCommand, error) {
	apiSection, err := newAPISection()
	if err != nil {
		return nil, err
	}

	section, err := schema.NewSection(
		"documents",
		"Documents",
		schema.WithDescription("Fetch documents"),
		schema.WithFields(
			fields.New("id", fields.TypeInteger,
				fields.WithHelp("Document ID (mutually exclusive with --slug)"),
				fields.WithDefault(0),
			),
			fields.New("slug", fields.TypeString,
				fields.WithHelp("Document slug (mutually exclusive with --id)"),
				fields.WithDefault(""),
			),
			fields.New("include-content", fields.TypeBool,
				fields.WithHelp("Include full content in output"),
				fields.WithDefault(false),
			),
		),
	)
	if err != nil {
		return nil, err
	}

	s := schema.NewSchema(schema.WithSections(apiSection, section))
	desc := cmds.NewCommandDefinition(
		"get",
		cmds.WithShort("Get a document by ID or slug"),
		cmds.WithLong("Fetch a document via REST; use --slug to also include extracted forms."),
		cmds.WithSchema(s),
	)

	return &DocumentsGetCommand{CommandDefinition: desc}, nil
}

var _ cmds.GlazeCommand = &DocumentsGetCommand{}

func (c *DocumentsGetCommand) RunIntoGlazeProcessor(ctx context.Context, parsedLayers *layers.ParsedLayers, gp middlewares.Processor) error {
	api := &APISettings{}
	if err := values.DecodeSectionInto(parsedLayers, "api", api); err != nil {
		return pkgerrors.Wrap(err, "decode api settings")
	}
	settings := &DocumentsGetSettings{}
	if err := values.DecodeSectionInto(parsedLayers, "documents", settings); err != nil {
		return pkgerrors.Wrap(err, "decode documents settings")
	}

	if settings.ID == 0 && strings.TrimSpace(settings.Slug) == "" {
		return pkgerrors.New("one of --id or --slug is required")
	}
	if settings.ID != 0 && strings.TrimSpace(settings.Slug) != "" {
		return pkgerrors.New("--id and --slug are mutually exclusive")
	}

	client, err := restclient.New(restclient.Options{BaseURL: api.BaseURL, Timeout: time.Duration(api.TimeoutSeconds) * time.Second})
	if err != nil {
		return err
	}

	if strings.TrimSpace(settings.Slug) != "" {
		doc, err := client.GetDocumentBySlug(ctx, strings.TrimSpace(settings.Slug))
		if err != nil {
			return err
		}
		row := types.NewRow(
			types.MRP("id", doc.ID),
			types.MRP("title", doc.Title),
			types.MRP("slug", doc.Slug),
			types.MRP("description", derefStringPtr(doc.Description)),
			types.MRP("category", derefStringPtr(doc.Category)),
			types.MRP("isPublished", doc.IsPublished),
			types.MRP("authorId", doc.AuthorID),
			types.MRP("createdAt", doc.CreatedAt),
			types.MRP("updatedAt", doc.UpdatedAt),
			types.MRP("formsCount", len(doc.Forms)),
			types.MRP("forms", doc.Forms),
		)
		if settings.IncludeContent {
			row.Set("content", doc.Content)
		}
		return gp.AddRow(ctx, row)
	}

	doc, err := client.GetDocumentByID(ctx, settings.ID)
	if err != nil {
		return err
	}
	row := types.NewRow(
		types.MRP("id", doc.ID),
		types.MRP("title", doc.Title),
		types.MRP("slug", doc.Slug),
		types.MRP("description", derefStringPtr(doc.Description)),
		types.MRP("category", derefStringPtr(doc.Category)),
		types.MRP("isPublished", doc.IsPublished),
		types.MRP("authorId", doc.AuthorID),
		types.MRP("createdAt", doc.CreatedAt),
		types.MRP("updatedAt", doc.UpdatedAt),
	)
	if settings.IncludeContent {
		row.Set("content", doc.Content)
	}
	return gp.AddRow(ctx, row)
}

type DocumentsUpdateSettings struct {
	ID          int64   `glazed.parameter:"id"`
	Title       *string `glazed.parameter:"title"`
	Content     *string `glazed.parameter:"content"`
	ContentFile string  `glazed.parameter:"content-file"`
	Description *string `glazed.parameter:"description"`
	Category    *string `glazed.parameter:"category"`
	Publish     bool    `glazed.parameter:"publish"`
	Unpublish   bool    `glazed.parameter:"unpublish"`
}

type DocumentsUpdateCommand struct {
	*cmds.CommandDefinition
}

func NewDocumentsUpdateCommand() (*DocumentsUpdateCommand, error) {
	apiSection, err := newAPISection()
	if err != nil {
		return nil, err
	}

	section, err := schema.NewSection(
		"documents",
		"Documents",
		schema.WithDescription("Update documents"),
		schema.WithFields(
			fields.New("id", fields.TypeInteger,
				fields.WithHelp("Document ID"),
				fields.WithDefault(0),
			),
			fields.New("title", fields.TypeString,
				fields.WithHelp("New title (omit to keep unchanged)"),
				fields.WithDefault(""),
			),
			fields.New("content", fields.TypeString,
				fields.WithHelp("New content (omit to keep unchanged; mutually exclusive with --content-file)"),
				fields.WithDefault(""),
			),
			fields.New("content-file", fields.TypeString,
				fields.WithHelp("Path to a file containing the new content (mutually exclusive with --content)"),
				fields.WithDefault(""),
			),
			fields.New("description", fields.TypeString,
				fields.WithHelp("New description (omit to keep unchanged)"),
				fields.WithDefault(""),
			),
			fields.New("category", fields.TypeString,
				fields.WithHelp("New category (omit to keep unchanged)"),
				fields.WithDefault(""),
			),
			fields.New("publish", fields.TypeBool,
				fields.WithHelp("Set isPublished=true (mutually exclusive with --unpublish)"),
				fields.WithDefault(false),
			),
			fields.New("unpublish", fields.TypeBool,
				fields.WithHelp("Set isPublished=false (mutually exclusive with --publish)"),
				fields.WithDefault(false),
			),
		),
	)
	if err != nil {
		return nil, err
	}

	s := schema.NewSchema(schema.WithSections(apiSection, section))
	desc := cmds.NewCommandDefinition(
		"update",
		cmds.WithShort("Update a document (PATCH)"),
		cmds.WithLong("Update a document via REST PATCH /api/documents/:id."),
		cmds.WithSchema(s),
	)
	return &DocumentsUpdateCommand{CommandDefinition: desc}, nil
}

var _ cmds.GlazeCommand = &DocumentsUpdateCommand{}

func (c *DocumentsUpdateCommand) RunIntoGlazeProcessor(ctx context.Context, parsedLayers *layers.ParsedLayers, gp middlewares.Processor) error {
	api := &APISettings{}
	if err := values.DecodeSectionInto(parsedLayers, "api", api); err != nil {
		return pkgerrors.Wrap(err, "decode api settings")
	}
	settings := &DocumentsUpdateSettings{}
	if err := values.DecodeSectionInto(parsedLayers, "documents", settings); err != nil {
		return pkgerrors.Wrap(err, "decode documents settings")
	}

	if settings.ID == 0 {
		return pkgerrors.New("id is required")
	}
	if settings.Publish && settings.Unpublish {
		return pkgerrors.New("publish and unpublish are mutually exclusive")
	}

	content := trimOptionalStringPtr(settings.Content)
	if strings.TrimSpace(settings.ContentFile) != "" {
		if content != nil {
			return pkgerrors.New("content and content-file are mutually exclusive")
		}
		s, err := readFileString(strings.TrimSpace(settings.ContentFile))
		if err != nil {
			return err
		}
		s = strings.TrimSpace(s)
		if s != "" {
			content = &s
		}
	}

	var isPublished *bool
	if settings.Publish {
		v := true
		isPublished = &v
	} else if settings.Unpublish {
		v := false
		isPublished = &v
	}

	req := restclient.UpdateDocumentRequest{
		Title:       trimOptionalStringPtr(settings.Title),
		Content:     content,
		Description: trimOptionalStringPtr(settings.Description),
		Category:    trimOptionalStringPtr(settings.Category),
		IsPublished: isPublished,
	}
	if req.Title == nil && req.Content == nil && req.Description == nil && req.Category == nil && req.IsPublished == nil {
		return pkgerrors.New("no fields to update")
	}

	client, err := restclient.New(restclient.Options{BaseURL: api.BaseURL, Timeout: time.Duration(api.TimeoutSeconds) * time.Second})
	if err != nil {
		return err
	}
	res, err := client.UpdateDocument(ctx, settings.ID, req)
	if err != nil {
		return err
	}

	return gp.AddRow(ctx, types.NewRow(
		types.MRP("id", settings.ID),
		types.MRP("success", res.Success),
	))
}

type DocumentsDeleteSettings struct {
	ID int64 `glazed.parameter:"id"`
}

type DocumentsDeleteCommand struct {
	*cmds.CommandDefinition
}

func NewDocumentsDeleteCommand() (*DocumentsDeleteCommand, error) {
	apiSection, err := newAPISection()
	if err != nil {
		return nil, err
	}

	section, err := schema.NewSection(
		"documents",
		"Documents",
		schema.WithDescription("Delete documents"),
		schema.WithFields(
			fields.New("id", fields.TypeInteger,
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
		"delete",
		cmds.WithShort("Delete a document"),
		cmds.WithLong("Delete a document via REST DELETE /api/documents/:id."),
		cmds.WithSchema(s),
	)
	return &DocumentsDeleteCommand{CommandDefinition: desc}, nil
}

var _ cmds.GlazeCommand = &DocumentsDeleteCommand{}

func (c *DocumentsDeleteCommand) RunIntoGlazeProcessor(ctx context.Context, parsedLayers *layers.ParsedLayers, gp middlewares.Processor) error {
	api := &APISettings{}
	if err := values.DecodeSectionInto(parsedLayers, "api", api); err != nil {
		return pkgerrors.Wrap(err, "decode api settings")
	}
	settings := &DocumentsDeleteSettings{}
	if err := values.DecodeSectionInto(parsedLayers, "documents", settings); err != nil {
		return pkgerrors.Wrap(err, "decode documents settings")
	}

	if settings.ID == 0 {
		return pkgerrors.New("id is required")
	}

	client, err := restclient.New(restclient.Options{BaseURL: api.BaseURL, Timeout: time.Duration(api.TimeoutSeconds) * time.Second})
	if err != nil {
		return err
	}

	res, err := client.DeleteDocument(ctx, settings.ID)
	if err != nil {
		return err
	}

	return gp.AddRow(ctx, types.NewRow(
		types.MRP("id", settings.ID),
		types.MRP("success", res.Success),
	))
}

type DocumentsAnalyticsSettings struct {
	ID int64 `glazed.parameter:"id"`
}

type DocumentsAnalyticsCommand struct {
	*cmds.CommandDefinition
}

func NewDocumentsAnalyticsCommand() (*DocumentsAnalyticsCommand, error) {
	apiSection, err := newAPISection()
	if err != nil {
		return nil, err
	}

	section, err := schema.NewSection(
		"documents",
		"Documents",
		schema.WithDescription("Document analytics"),
		schema.WithFields(
			fields.New("id", fields.TypeInteger,
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
		"analytics",
		cmds.WithShort("Get document analytics"),
		cmds.WithLong("Fetch document quiz analytics via REST GET /api/documents/:id/analytics."),
		cmds.WithSchema(s),
	)
	return &DocumentsAnalyticsCommand{CommandDefinition: desc}, nil
}

var _ cmds.GlazeCommand = &DocumentsAnalyticsCommand{}

func (c *DocumentsAnalyticsCommand) RunIntoGlazeProcessor(ctx context.Context, parsedLayers *layers.ParsedLayers, gp middlewares.Processor) error {
	api := &APISettings{}
	if err := values.DecodeSectionInto(parsedLayers, "api", api); err != nil {
		return pkgerrors.Wrap(err, "decode api settings")
	}
	settings := &DocumentsAnalyticsSettings{}
	if err := values.DecodeSectionInto(parsedLayers, "documents", settings); err != nil {
		return pkgerrors.Wrap(err, "decode documents settings")
	}

	if settings.ID == 0 {
		return pkgerrors.New("id is required")
	}

	client, err := restclient.New(restclient.Options{BaseURL: api.BaseURL, Timeout: time.Duration(api.TimeoutSeconds) * time.Second})
	if err != nil {
		return err
	}

	a, err := client.GetDocumentAnalytics(ctx, settings.ID)
	if err != nil {
		return err
	}

	return gp.AddRow(ctx, types.NewRow(
		types.MRP("id", settings.ID),
		types.MRP("totalSubmissions", a.TotalSubmissions),
		types.MRP("averageScore", a.AverageScore),
		types.MRP("highestScore", a.HighestScore),
		types.MRP("lowestScore", a.LowestScore),
	))
}

type DocumentsSubmissionsSettings struct {
	ID int64 `glazed.parameter:"id"`
}

type DocumentsSubmissionsCommand struct {
	*cmds.CommandDefinition
}

func NewDocumentsSubmissionsCommand() (*DocumentsSubmissionsCommand, error) {
	apiSection, err := newAPISection()
	if err != nil {
		return nil, err
	}

	section, err := schema.NewSection(
		"documents",
		"Documents",
		schema.WithDescription("Document submissions"),
		schema.WithFields(
			fields.New("id", fields.TypeInteger,
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
		"submissions",
		cmds.WithShort("List submissions for a document"),
		cmds.WithLong("List submissions via REST GET /api/documents/:id/submissions."),
		cmds.WithSchema(s),
	)
	return &DocumentsSubmissionsCommand{CommandDefinition: desc}, nil
}

var _ cmds.GlazeCommand = &DocumentsSubmissionsCommand{}

func (c *DocumentsSubmissionsCommand) RunIntoGlazeProcessor(ctx context.Context, parsedLayers *layers.ParsedLayers, gp middlewares.Processor) error {
	api := &APISettings{}
	if err := values.DecodeSectionInto(parsedLayers, "api", api); err != nil {
		return pkgerrors.Wrap(err, "decode api settings")
	}
	settings := &DocumentsSubmissionsSettings{}
	if err := values.DecodeSectionInto(parsedLayers, "documents", settings); err != nil {
		return pkgerrors.Wrap(err, "decode documents settings")
	}

	if settings.ID == 0 {
		return pkgerrors.New("id is required")
	}

	client, err := restclient.New(restclient.Options{BaseURL: api.BaseURL, Timeout: time.Duration(api.TimeoutSeconds) * time.Second})
	if err != nil {
		return err
	}

	items, err := client.GetDocumentSubmissions(ctx, settings.ID)
	if err != nil {
		return err
	}

	for _, it := range items {
		s := it.Submission
		if err := gp.AddRow(ctx, types.NewRow(
			types.MRP("submissionId", s.ID),
			types.MRP("documentId", settings.ID),
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
