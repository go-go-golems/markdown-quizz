package quiz

import (
	"context"
	"os"
	"testing"

	"github.com/go-go-golems/XXX/internal/db"
	"github.com/go-go-golems/XXX/internal/documents"
	"github.com/stretchr/testify/require"
)

func TestQuizStore_SubmitAndAnalytics(t *testing.T) {
	ctx := context.Background()

	f, err := os.CreateTemp("", "markdown-quizz-quiz-*.sqlite")
	require.NoError(t, err)
	_ = f.Close()
	defer func() { _ = os.Remove(f.Name()) }()

	sqliteDB, err := db.OpenSQLite(ctx, db.SQLiteOptions{Path: f.Name()})
	require.NoError(t, err)
	defer func() { _ = sqliteDB.Close() }()

	docStore := documents.NewStore(sqliteDB)

	content := `
<form id="f1">
fields:
  - name: q1
    correct: "a"
  - name: q2
    correct: ["x","y"]
</form>
`
	docID, _, err := docStore.Create(ctx, documents.CreateParams{
		Title:    "Doc",
		Content:  content,
		AuthorID: 1,
	})
	require.NoError(t, err)

	quizStore := NewStore(sqliteDB)

	results, err := quizStore.SubmitMultiple(ctx, 1, docID, []SubmissionInput{
		{FormID: "f1", Responses: map[string]any{"q1": "a", "q2": []any{"y", "x"}}},
		{FormID: "missing", Responses: map[string]any{"q": "x"}},
	})
	require.NoError(t, err)
	require.Len(t, results, 2)
	require.Equal(t, 2, results[0].Score)
	require.Equal(t, 2, results[0].MaxScore)
	require.Equal(t, 0, results[1].Score)
	require.Equal(t, 0, results[1].MaxScore)

	id, score, maxScore, err := quizStore.Submit(ctx, 1, docID, "missing", map[string]any{"q": "x"})
	require.NoError(t, err)
	require.NotZero(t, id)
	require.Nil(t, score)
	require.Nil(t, maxScore)

	a, err := quizStore.GetQuizAnalytics(ctx, docID)
	require.NoError(t, err)
	require.Equal(t, 3, a.TotalSubmissions)
	require.Equal(t, 100.0, a.AverageScore)
	require.Equal(t, 100.0, a.HighestScore)
	require.Equal(t, 100.0, a.LowestScore)
}
