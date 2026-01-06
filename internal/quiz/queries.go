package quiz

import (
	"context"
	"database/sql"
	"encoding/json"

	pkgerrors "github.com/pkg/errors"
)

type Submission struct {
	ID          int64
	UserID      int64
	DocumentID  int64
	FormID      string
	Responses   map[string]any
	Score       *int
	MaxScore    *int
	SubmittedAt string
}

type SubmissionWithDocument struct {
	Submission    Submission
	DocumentTitle string
	DocumentSlug  string
}

type SubmissionWithUser struct {
	Submission Submission
	UserName   *string
}

type SubmissionDetail struct {
	SubmissionWithDocument
	FormDefinition any
}

func (s *Store) GetSubmissionsByUser(ctx context.Context, userID int64) ([]SubmissionWithDocument, error) {
	rows, err := s.db.QueryContext(ctx, `
SELECT
  qs.id, qs.user_id, qs.document_id, qs.form_id, qs.responses, qs.score, qs.max_score, qs.submitted_at,
  d.title, d.slug
FROM quiz_submissions qs
LEFT JOIN documents d ON qs.document_id = d.id
WHERE qs.user_id = ?
ORDER BY qs.submitted_at DESC`, userID)
	if err != nil {
		return nil, pkgerrors.Wrap(err, "get submissions by user")
	}
	defer func() { _ = rows.Close() }()

	var out []SubmissionWithDocument
	for rows.Next() {
		item, err := scanSubmissionWithDoc(rows)
		if err != nil {
			return nil, err
		}
		out = append(out, item)
	}
	if err := rows.Err(); err != nil {
		return nil, pkgerrors.Wrap(err, "get submissions by user rows")
	}
	return out, nil
}

func (s *Store) GetSubmissionsByDocument(ctx context.Context, documentID int64) ([]SubmissionWithUser, error) {
	rows, err := s.db.QueryContext(ctx, `
SELECT
  qs.id, qs.user_id, qs.document_id, qs.form_id, qs.responses, qs.score, qs.max_score, qs.submitted_at,
  u.name
FROM quiz_submissions qs
LEFT JOIN users u ON qs.user_id = u.id
WHERE qs.document_id = ?
ORDER BY qs.submitted_at DESC`, documentID)
	if err != nil {
		return nil, pkgerrors.Wrap(err, "get submissions by document")
	}
	defer func() { _ = rows.Close() }()

	var out []SubmissionWithUser
	for rows.Next() {
		sub, userName, err := scanSubmissionWithUser(rows)
		if err != nil {
			return nil, err
		}
		out = append(out, SubmissionWithUser{
			Submission: sub,
			UserName:   userName,
		})
	}
	if err := rows.Err(); err != nil {
		return nil, pkgerrors.Wrap(err, "get submissions by document rows")
	}
	return out, nil
}

func (s *Store) GetSubmissionByID(ctx context.Context, id int64) (*SubmissionDetail, error) {
	row := s.db.QueryRowContext(ctx, `
SELECT
  qs.id, qs.user_id, qs.document_id, qs.form_id, qs.responses, qs.score, qs.max_score, qs.submitted_at,
  d.title, d.slug
FROM quiz_submissions qs
LEFT JOIN documents d ON qs.document_id = d.id
WHERE qs.id = ?
LIMIT 1`, id)

	item, err := scanSubmissionWithDocRow(row)
	if err != nil {
		return nil, err
	}
	if item == nil {
		return nil, nil
	}

	definition, _, err := s.getFormDefinition(ctx, item.Submission.DocumentID, item.Submission.FormID)
	if err != nil {
		return nil, err
	}

	return &SubmissionDetail{
		SubmissionWithDocument: *item,
		FormDefinition:         definition,
	}, nil
}

type submissionRow interface {
	Scan(dest ...any) error
}

func scanSubmissionWithDoc(rows submissionRow) (SubmissionWithDocument, error) {
	item, err := scanSubmissionWithDocRow(rows)
	if err != nil {
		return SubmissionWithDocument{}, err
	}
	if item == nil {
		return SubmissionWithDocument{}, pkgerrors.New("unexpected nil submission")
	}
	return *item, nil
}

func scanSubmissionWithDocRow(row submissionRow) (*SubmissionWithDocument, error) {
	var s SubmissionWithDocument
	var responsesJSON string
	var score sql.NullInt64
	var maxScore sql.NullInt64
	var submittedAt sql.NullString
	var title sql.NullString
	var slug sql.NullString

	err := row.Scan(
		&s.Submission.ID,
		&s.Submission.UserID,
		&s.Submission.DocumentID,
		&s.Submission.FormID,
		&responsesJSON,
		&score,
		&maxScore,
		&submittedAt,
		&title,
		&slug,
	)
	if err != nil {
		if err == sql.ErrNoRows {
			return nil, nil
		}
		return nil, pkgerrors.Wrap(err, "scan submission")
	}

	if err := json.Unmarshal([]byte(responsesJSON), &s.Submission.Responses); err != nil {
		return nil, pkgerrors.Wrap(err, "unmarshal responses")
	}
	if score.Valid {
		v := int(score.Int64)
		s.Submission.Score = &v
	}
	if maxScore.Valid {
		v := int(maxScore.Int64)
		s.Submission.MaxScore = &v
	}
	if submittedAt.Valid {
		s.Submission.SubmittedAt = submittedAt.String
	}
	if title.Valid {
		s.DocumentTitle = title.String
	}
	if slug.Valid {
		s.DocumentSlug = slug.String
	}

	return &s, nil
}

func scanSubmissionWithUser(row submissionRow) (Submission, *string, error) {
	var s Submission
	var responsesJSON string
	var score sql.NullInt64
	var maxScore sql.NullInt64
	var submittedAt sql.NullString
	var userName sql.NullString

	err := row.Scan(
		&s.ID,
		&s.UserID,
		&s.DocumentID,
		&s.FormID,
		&responsesJSON,
		&score,
		&maxScore,
		&submittedAt,
		&userName,
	)
	if err != nil {
		return Submission{}, nil, pkgerrors.Wrap(err, "scan submission with user")
	}

	if err := json.Unmarshal([]byte(responsesJSON), &s.Responses); err != nil {
		return Submission{}, nil, pkgerrors.Wrap(err, "unmarshal responses")
	}
	if score.Valid {
		v := int(score.Int64)
		s.Score = &v
	}
	if maxScore.Valid {
		v := int(maxScore.Int64)
		s.MaxScore = &v
	}
	if submittedAt.Valid {
		s.SubmittedAt = submittedAt.String
	}

	var userNameOut *string
	if userName.Valid {
		userNameOut = &userName.String
	}
	return s, userNameOut, nil
}
