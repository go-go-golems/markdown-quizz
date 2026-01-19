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

type DocumentsImportSettings struct {
	File        string  `glazed.parameter:"file"`
	Title       string  `glazed.parameter:"title"`
	Description *string `glazed.parameter:"description"`
	Category    *string `glazed.parameter:"category"`
	IsPublished bool    `glazed.parameter:"is-published"`
}

type DocumentsImportCommand struct {
	*cmds.CommandDefinition
}

func NewDocumentsImportCommand() (*DocumentsImportCommand, error) {
	apiSection, err := newAPISection()
	if err != nil {
		return nil, err
	}

	section, err := schema.NewSection(
		"documents",
		"Documents",
		schema.WithDescription("Create documents from Markdown files"),
		schema.WithFields(
			fields.New("file", fields.TypeString,
				fields.WithHelp("Markdown file path"),
			),
			fields.New("title", fields.TypeString,
				fields.WithHelp("Document title (defaults to filename)"),
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
		"import",
		cmds.WithShort("Import a Markdown document (extracts <form> quizzes)"),
		cmds.WithLong("Create a document via REST; any <form id=\"...\"> YAML </form> blocks are extracted into quiz forms by the backend."),
		cmds.WithSchema(s),
	)

	return &DocumentsImportCommand{CommandDefinition: desc}, nil
}

var _ cmds.GlazeCommand = &DocumentsImportCommand{}

func (c *DocumentsImportCommand) RunIntoGlazeProcessor(ctx context.Context, parsedLayers *layers.ParsedLayers, gp middlewares.Processor) error {
	api := &APISettings{}
	if err := values.DecodeSectionInto(parsedLayers, "api", api); err != nil {
		return pkgerrors.Wrap(err, "decode api settings")
	}
	settings := &DocumentsImportSettings{}
	if err := values.DecodeSectionInto(parsedLayers, "documents", settings); err != nil {
		return pkgerrors.Wrap(err, "decode documents settings")
	}
	if strings.TrimSpace(settings.File) == "" {
		return pkgerrors.New("file is required")
	}

	content, err := readFileString(settings.File)
	if err != nil {
		return err
	}

	title := strings.TrimSpace(settings.Title)
	if title == "" {
		title = guessTitleFromPath(settings.File)
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
