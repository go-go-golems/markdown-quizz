package documents

import (
	"context"
	"database/sql"
	"encoding/json"
	stderrors "errors"
	"time"

	"github.com/go-go-golems/XXX/internal/quizdsl"
	pkgerrors "github.com/pkg/errors"
)

type Store struct {
	db  *sql.DB
	now func() time.Time
}

func NewStore(db *sql.DB) *Store {
	return &Store{
		db:  db,
		now: time.Now,
	}
}

type Document struct {
	ID          int64
	Title       string
	Slug        string
	Content     string
	Description *string
	Category    *string
	IsPublished bool
	AuthorID    int64
	CreatedAt   string
	UpdatedAt   string
}

type QuizForm struct {
	ID         int64
	DocumentID int64
	FormID     string
	Definition any
	CreatedAt  string
	UpdatedAt  string
}

type CreateParams struct {
	Title       string
	Content     string
	Description *string
	Category    *string
	IsPublished bool
	AuthorID    int64
}

func (s *Store) Create(ctx context.Context, p CreateParams) (int64, string, error) {
	slug := GenerateSlug(p.Title, s.now())
	isPublished := boolToInt(p.IsPublished)

	res, err := s.db.ExecContext(ctx, `
INSERT INTO documents (title, slug, content, description, category, is_published, author_id)
VALUES (?, ?, ?, ?, ?, ?, ?)`,
		p.Title, slug, p.Content, p.Description, p.Category, isPublished, p.AuthorID,
	)
	if err != nil {
		return 0, "", pkgerrors.Wrap(err, "insert document")
	}

	documentID, err := res.LastInsertId()
	if err != nil {
		return 0, "", pkgerrors.Wrap(err, "get document id")
	}

	if err := s.replaceFormsFromContent(ctx, documentID, p.Content); err != nil {
		return 0, "", err
	}

	return documentID, slug, nil
}

type UpdateParams struct {
	ID          int64
	Title       *string
	Content     *string
	Description *string
	Category    *string
	IsPublished *bool
}

func (s *Store) Update(ctx context.Context, p UpdateParams) error {
	set := map[string]any{}
	if p.Title != nil {
		set["title"] = *p.Title
	}
	if p.Content != nil {
		set["content"] = *p.Content
	}
	if p.Description != nil {
		set["description"] = *p.Description
	}
	if p.Category != nil {
		set["category"] = *p.Category
	}
	if p.IsPublished != nil {
		set["is_published"] = boolToInt(*p.IsPublished)
	}

	if len(set) == 0 {
		return nil
	}

	query := "UPDATE documents SET "
	args := make([]any, 0, len(set)+1)
	i := 0
	for k, v := range set {
		if i > 0 {
			query += ", "
		}
		query += k + " = ?"
		args = append(args, v)
		i++
	}
	query += ", updated_at = CURRENT_TIMESTAMP WHERE id = ?"
	args = append(args, p.ID)

	if _, err := s.db.ExecContext(ctx, query, args...); err != nil {
		return pkgerrors.Wrap(err, "update document")
	}

	if p.Content != nil {
		if err := s.replaceFormsFromContent(ctx, p.ID, *p.Content); err != nil {
			return err
		}
	}

	return nil
}

func (s *Store) Delete(ctx context.Context, id int64) error {
	if _, err := s.db.ExecContext(ctx, "DELETE FROM documents WHERE id = ?", id); err != nil {
		return pkgerrors.Wrap(err, "delete document")
	}
	return nil
}

func (s *Store) GetByID(ctx context.Context, id int64) (*Document, error) {
	row := s.db.QueryRowContext(ctx, `
SELECT id, title, slug, content, description, category, is_published, author_id, created_at, updated_at
FROM documents
WHERE id = ?
LIMIT 1`, id)

	return scanDocument(row)
}

func (s *Store) GetBySlug(ctx context.Context, slug string) (*Document, []QuizForm, error) {
	row := s.db.QueryRowContext(ctx, `
SELECT id, title, slug, content, description, category, is_published, author_id, created_at, updated_at
FROM documents
WHERE slug = ?
LIMIT 1`, slug)

	doc, err := scanDocument(row)
	if err != nil {
		return nil, nil, err
	}
	if doc == nil {
		return nil, nil, nil
	}

	forms, err := s.GetQuizFormsByDocument(ctx, doc.ID)
	if err != nil {
		return nil, nil, err
	}

	return doc, forms, nil
}

func (s *Store) ListDocuments(ctx context.Context, authorID *int64, publishedOnly bool) ([]Document, error) {
	where := "1=1"
	args := []any{}
	if authorID != nil {
		where += " AND author_id = ?"
		args = append(args, *authorID)
	}
	if publishedOnly {
		where += " AND is_published = 1"
	}

	rows, err := s.db.QueryContext(ctx, `
SELECT id, title, slug, content, description, category, is_published, author_id, created_at, updated_at
FROM documents
WHERE `+where+`
ORDER BY updated_at DESC`, args...)
	if err != nil {
		return nil, pkgerrors.Wrap(err, "list documents")
	}
	defer func() { _ = rows.Close() }()

	var out []Document
	for rows.Next() {
		doc, err := scanDocument(rows)
		if err != nil {
			return nil, err
		}
		if doc != nil {
			out = append(out, *doc)
		}
	}
	if err := rows.Err(); err != nil {
		return nil, pkgerrors.Wrap(err, "list documents rows")
	}

	return out, nil
}

func (s *Store) GetQuizFormsByDocument(ctx context.Context, documentID int64) ([]QuizForm, error) {
	rows, err := s.db.QueryContext(ctx, `
SELECT id, document_id, form_id, definition, created_at, updated_at
FROM quiz_forms
WHERE document_id = ?
ORDER BY id ASC`, documentID)
	if err != nil {
		return nil, pkgerrors.Wrap(err, "get quiz forms by document")
	}
	defer func() { _ = rows.Close() }()

	var out []QuizForm
	for rows.Next() {
		var f QuizForm
		var defJSON string
		if err := rows.Scan(&f.ID, &f.DocumentID, &f.FormID, &defJSON, &f.CreatedAt, &f.UpdatedAt); err != nil {
			return nil, pkgerrors.Wrap(err, "scan quiz form")
		}
		var def any
		if err := json.Unmarshal([]byte(defJSON), &def); err != nil {
			return nil, pkgerrors.Wrap(err, "unmarshal quiz form definition")
		}
		f.Definition = def
		out = append(out, f)
	}
	if err := rows.Err(); err != nil {
		return nil, pkgerrors.Wrap(err, "get quiz forms by document rows")
	}

	return out, nil
}

func (s *Store) replaceFormsFromContent(ctx context.Context, documentID int64, content string) error {
	forms, err := quizdsl.ExtractFormsFromContent(content)
	if err != nil {
		return pkgerrors.Wrap(err, "extract forms from content")
	}

	if _, err := s.db.ExecContext(ctx, "DELETE FROM quiz_forms WHERE document_id = ?", documentID); err != nil {
		return pkgerrors.Wrap(err, "delete quiz forms for document")
	}

	for _, f := range forms {
		defJSON, err := quizdsl.MarshalDefinitionJSON(f.Definition)
		if err != nil {
			return pkgerrors.Wrap(err, "marshal form definition")
		}

		_, err = s.db.ExecContext(ctx, `
INSERT INTO quiz_forms (document_id, form_id, definition)
VALUES (?, ?, ?)
ON CONFLICT(document_id, form_id) DO UPDATE SET definition = excluded.definition`,
			documentID, f.FormID, defJSON,
		)
		if err != nil {
			return pkgerrors.Wrapf(err, "upsert quiz form %s", f.FormID)
		}
	}

	return nil
}

func boolToInt(b bool) int {
	if b {
		return 1
	}
	return 0
}

type documentRow interface {
	Scan(dest ...any) error
}

func scanDocument(row documentRow) (*Document, error) {
	var doc Document
	var isPublished int
	var description sql.NullString
	var category sql.NullString
	var createdAt sql.NullString
	var updatedAt sql.NullString

	err := row.Scan(
		&doc.ID,
		&doc.Title,
		&doc.Slug,
		&doc.Content,
		&description,
		&category,
		&isPublished,
		&doc.AuthorID,
		&createdAt,
		&updatedAt,
	)
	if err != nil {
		if stderrors.Is(err, sql.ErrNoRows) {
			return nil, nil
		}
		return nil, pkgerrors.Wrap(err, "scan document")
	}

	if description.Valid {
		doc.Description = &description.String
	}
	if category.Valid {
		doc.Category = &category.String
	}
	doc.IsPublished = isPublished != 0
	if createdAt.Valid {
		doc.CreatedAt = createdAt.String
	}
	if updatedAt.Valid {
		doc.UpdatedAt = updatedAt.String
	}

	return &doc, nil
}
