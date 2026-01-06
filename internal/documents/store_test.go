package documents

import (
	"context"
	"os"
	"testing"
	"time"

	"github.com/go-go-golems/XXX/internal/db"
	"github.com/stretchr/testify/require"
)

func TestGenerateSlug_MatchesLegacyShape(t *testing.T) {
	now := time.UnixMilli(1700000000123)
	slug := GenerateSlug("Hello, World!", now)
	require.Equal(t, "hello-world-"+strconvFormatInt36(now.UnixMilli()), slug)
}

func TestStore_CreateAndUpdate_ReextractForms(t *testing.T) {
	ctx := context.Background()

	f, err := os.CreateTemp("", "markdown-quizz-*.sqlite")
	require.NoError(t, err)
	_ = f.Close()
	defer func() { _ = os.Remove(f.Name()) }()

	sqliteDB, err := db.OpenSQLite(ctx, db.SQLiteOptions{Path: f.Name()})
	require.NoError(t, err)
	defer func() { _ = sqliteDB.Close() }()

	store := NewStore(sqliteDB)
	store.now = func() time.Time { return time.UnixMilli(1700000000123) }

	content1 := `
<form id="f1">
fields:
  - name: q1
    correct: "a"
</form>
<form id="f2">
fields:
  - name: q2
    correct: ["x","y"]
</form>
`

	docID, slug, err := store.Create(ctx, CreateParams{
		Title:       "Doc",
		Content:     content1,
		IsPublished: false,
		AuthorID:    1,
	})
	require.NoError(t, err)
	require.NotZero(t, docID)
	require.NotEmpty(t, slug)

	forms, err := store.GetQuizFormsByDocument(ctx, docID)
	require.NoError(t, err)
	require.Len(t, forms, 2)

	content2 := `
<form id="f1">
fields:
  - name: q1
    correct: "b"
</form>
`

	require.NoError(t, store.Update(ctx, UpdateParams{
		ID:      docID,
		Content: &content2,
	}))

	forms, err = store.GetQuizFormsByDocument(ctx, docID)
	require.NoError(t, err)
	require.Len(t, forms, 1)
	require.Equal(t, "f1", forms[0].FormID)
}
