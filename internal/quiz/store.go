package quiz

import (
	"context"
	"database/sql"
	"encoding/json"
	"math"

	"github.com/go-go-golems/XXX/internal/scoring"
	pkgerrors "github.com/pkg/errors"
)

type Store struct {
	db *sql.DB
}

func NewStore(db *sql.DB) *Store {
	return &Store{db: db}
}

type SubmissionInput struct {
	FormID    string
	Responses map[string]any
}

type SubmitMultipleResult struct {
	FormID   string
	Score    int
	MaxScore int
}

func (s *Store) SubmitMultiple(ctx context.Context, userID int64, documentID int64, submissions []SubmissionInput) ([]SubmitMultipleResult, error) {
	results := make([]SubmitMultipleResult, 0, len(submissions))

	for _, sub := range submissions {
		score := 0
		maxScore := 0

		definition, ok, err := s.getFormDefinition(ctx, documentID, sub.FormID)
		if err != nil {
			return nil, err
		}
		if ok {
			r, err := scoring.Calculate(definition, sub.Responses)
			if err != nil {
				return nil, pkgerrors.Wrap(err, "score submission")
			}
			score = r.Score
			maxScore = r.MaxScore
		}

		responsesJSON, err := json.Marshal(sub.Responses)
		if err != nil {
			return nil, pkgerrors.Wrap(err, "marshal responses")
		}

		if _, err := s.db.ExecContext(ctx, `
INSERT INTO quiz_submissions (user_id, document_id, form_id, responses, score, max_score)
VALUES (?, ?, ?, ?, ?, ?)`,
			userID, documentID, sub.FormID, string(responsesJSON), score, maxScore,
		); err != nil {
			return nil, pkgerrors.Wrap(err, "insert quiz submission")
		}

		results = append(results, SubmitMultipleResult{
			FormID:   sub.FormID,
			Score:    score,
			MaxScore: maxScore,
		})
	}

	return results, nil
}

func (s *Store) Submit(ctx context.Context, userID int64, documentID int64, formID string, responses map[string]any) (int64, *int, *int, error) {
	var score sql.NullInt64
	var maxScore sql.NullInt64

	definition, ok, err := s.getFormDefinition(ctx, documentID, formID)
	if err != nil {
		return 0, nil, nil, err
	}
	if ok {
		r, err := scoring.Calculate(definition, responses)
		if err != nil {
			return 0, nil, nil, pkgerrors.Wrap(err, "score submission")
		}
		score = sql.NullInt64{Int64: int64(r.Score), Valid: true}
		maxScore = sql.NullInt64{Int64: int64(r.MaxScore), Valid: true}
	}

	responsesJSON, err := json.Marshal(responses)
	if err != nil {
		return 0, nil, nil, pkgerrors.Wrap(err, "marshal responses")
	}

	res, err := s.db.ExecContext(ctx, `
INSERT INTO quiz_submissions (user_id, document_id, form_id, responses, score, max_score)
VALUES (?, ?, ?, ?, ?, ?)`,
		userID, documentID, formID, string(responsesJSON), score, maxScore,
	)
	if err != nil {
		return 0, nil, nil, pkgerrors.Wrap(err, "insert quiz submission")
	}

	id, err := res.LastInsertId()
	if err != nil {
		return 0, nil, nil, pkgerrors.Wrap(err, "get submission id")
	}

	var scoreOut *int
	var maxOut *int
	if score.Valid {
		v := int(score.Int64)
		scoreOut = &v
	}
	if maxScore.Valid {
		v := int(maxScore.Int64)
		maxOut = &v
	}

	return id, scoreOut, maxOut, nil
}

type Analytics struct {
	TotalSubmissions int
	AverageScore     float64
	HighestScore     float64
	LowestScore      float64
}

func (s *Store) GetQuizAnalytics(ctx context.Context, documentID int64) (Analytics, error) {
	rows, err := s.db.QueryContext(ctx, `
SELECT score, max_score
FROM quiz_submissions
WHERE document_id = ?`, documentID)
	if err != nil {
		return Analytics{}, pkgerrors.Wrap(err, "query analytics submissions")
	}
	defer func() { _ = rows.Close() }()

	total := 0
	var percentages []float64
	for rows.Next() {
		total++
		var score sql.NullInt64
		var maxScore sql.NullInt64
		if err := rows.Scan(&score, &maxScore); err != nil {
			return Analytics{}, pkgerrors.Wrap(err, "scan analytics submission")
		}
		if !score.Valid || !maxScore.Valid {
			continue
		}
		if maxScore.Int64 <= 0 {
			continue
		}
		percentages = append(percentages, (float64(score.Int64)/float64(maxScore.Int64))*100.0)
	}
	if err := rows.Err(); err != nil {
		return Analytics{}, pkgerrors.Wrap(err, "iterate analytics submissions")
	}

	avg := 0.0
	highest := 0.0
	lowest := 0.0
	if len(percentages) > 0 {
		sum := 0.0
		highest = percentages[0]
		lowest = percentages[0]
		for _, p := range percentages {
			sum += p
			if p > highest {
				highest = p
			}
			if p < lowest {
				lowest = p
			}
		}
		avg = sum / float64(len(percentages))
	}

	return Analytics{
		TotalSubmissions: total,
		AverageScore:     round1(avg),
		HighestScore:     round1(highest),
		LowestScore:      round1(lowest),
	}, nil
}

func round1(v float64) float64 {
	return math.Round(v*10.0) / 10.0
}

func (s *Store) getFormDefinition(ctx context.Context, documentID int64, formID string) (any, bool, error) {
	row := s.db.QueryRowContext(ctx, `
SELECT definition
FROM quiz_forms
WHERE document_id = ? AND form_id = ?
LIMIT 1`, documentID, formID)

	var defJSON string
	if err := row.Scan(&defJSON); err != nil {
		if err == sql.ErrNoRows {
			return nil, false, nil
		}
		return nil, false, pkgerrors.Wrap(err, "query quiz form definition")
	}

	var def any
	if err := json.Unmarshal([]byte(defJSON), &def); err != nil {
		return nil, false, pkgerrors.Wrap(err, "unmarshal quiz form definition")
	}

	return def, true, nil
}
