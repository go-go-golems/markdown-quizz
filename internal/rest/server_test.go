package rest

import (
	"bytes"
	"context"
	"encoding/json"
	"net/http"
	"net/http/httptest"
	"os"
	"strconv"
	"testing"

	"github.com/go-go-golems/XXX/internal/db"
	"github.com/go-go-golems/XXX/internal/documents"
	"github.com/go-go-golems/XXX/internal/quiz"
	"github.com/stretchr/testify/require"
)

func TestREST_DocumentsAndQuiz_HappyPath(t *testing.T) {
	ctx := context.Background()

	f, err := os.CreateTemp("", "markdown-quizz-rest-*.sqlite")
	require.NoError(t, err)
	_ = f.Close()
	defer func() { _ = os.Remove(f.Name()) }()

	sqliteDB, err := db.OpenSQLite(ctx, db.SQLiteOptions{Path: f.Name()})
	require.NoError(t, err)
	defer func() { _ = sqliteDB.Close() }()

	srv := NewServer(Server{
		Documents: documents.NewStore(sqliteDB),
		Quiz:      quiz.NewStore(sqliteDB),
		UserID:    1,
	})

	content := `
<form id="f1">
fields:
  - name: q1
    correct: "a"
  - name: q2
    correct: ["x","y"]
</form>
`

	// Create document
	createBody := map[string]any{
		"title":       "Doc",
		"content":     content,
		"description": "desc",
		"category":    "cat",
		"isPublished": false,
	}
	createRes := doJSON(t, srv, http.MethodPost, "/api/documents", createBody)
	require.Equal(t, http.StatusCreated, createRes.Status)

	var created struct {
		ID   int64  `json:"id"`
		Slug string `json:"slug"`
	}
	require.NoError(t, json.Unmarshal(createRes.Body, &created))
	require.NotZero(t, created.ID)
	require.NotEmpty(t, created.Slug)

	// List documents
	listRes := doJSON(t, srv, http.MethodGet, "/api/documents?scope=all", nil)
	require.Equal(t, http.StatusOK, listRes.Status)
	var docs []map[string]any
	require.NoError(t, json.Unmarshal(listRes.Body, &docs))
	require.Len(t, docs, 1)

	// Get by ID
	getRes := doJSON(t, srv, http.MethodGet, "/api/documents/"+itoa(created.ID), nil)
	require.Equal(t, http.StatusOK, getRes.Status)
	var doc map[string]any
	require.NoError(t, json.Unmarshal(getRes.Body, &doc))
	require.Equal(t, "Doc", doc["title"])

	// Get by slug (includes forms)
	getSlugRes := doJSON(t, srv, http.MethodGet, "/api/documents/by-slug/"+created.Slug, nil)
	require.Equal(t, http.StatusOK, getSlugRes.Status)
	var docSlug map[string]any
	require.NoError(t, json.Unmarshal(getSlugRes.Body, &docSlug))
	forms, ok := docSlug["forms"].([]any)
	require.True(t, ok)
	require.NotEmpty(t, forms)

	// Update document
	updateBody := map[string]any{
		"title": "Doc2",
	}
	upRes := doJSON(t, srv, http.MethodPatch, "/api/documents/"+itoa(created.ID), updateBody)
	require.Equal(t, http.StatusOK, upRes.Status)

	// Submit batch quizzes
	subBody := map[string]any{
		"documentId": created.ID,
		"submissions": []any{
			map[string]any{"formId": "f1", "responses": map[string]any{"q1": "a", "q2": []any{"y", "x"}}},
			map[string]any{"formId": "missing", "responses": map[string]any{"q": "x"}},
		},
	}
	subRes := doJSON(t, srv, http.MethodPost, "/api/quiz/submissions/batch", subBody)
	require.Equal(t, http.StatusOK, subRes.Status)
	var subOut map[string]any
	require.NoError(t, json.Unmarshal(subRes.Body, &subOut))
	results, ok := subOut["results"].([]any)
	require.True(t, ok)
	require.Len(t, results, 2)

	// Analytics reflects 2 submissions
	analyticsRes := doJSON(t, srv, http.MethodGet, "/api/documents/"+itoa(created.ID)+"/analytics", nil)
	require.Equal(t, http.StatusOK, analyticsRes.Status)
	var analytics map[string]any
	require.NoError(t, json.Unmarshal(analyticsRes.Body, &analytics))
	require.Equal(t, float64(2), analytics["totalSubmissions"])

	// My submissions
	mineRes := doJSON(t, srv, http.MethodGet, "/api/quiz/submissions?scope=mine", nil)
	require.Equal(t, http.StatusOK, mineRes.Status)
	var mine []map[string]any
	require.NoError(t, json.Unmarshal(mineRes.Body, &mine))
	require.Len(t, mine, 2)

	// Delete document
	delRes := doJSON(t, srv, http.MethodDelete, "/api/documents/"+itoa(created.ID), nil)
	require.Equal(t, http.StatusOK, delRes.Status)
}

type jsonResponse struct {
	Status int
	Body   []byte
}

func doJSON(t *testing.T, h http.Handler, method, path string, body any) jsonResponse {
	t.Helper()

	var buf *bytes.Reader
	if body != nil {
		b, err := json.Marshal(body)
		require.NoError(t, err)
		buf = bytes.NewReader(b)
	} else {
		buf = bytes.NewReader(nil)
	}

	req := httptest.NewRequest(method, path, buf)
	if body != nil {
		req.Header.Set("Content-Type", "application/json")
	}
	w := httptest.NewRecorder()
	h.ServeHTTP(w, req)
	return jsonResponse{
		Status: w.Result().StatusCode,
		Body:   w.Body.Bytes(),
	}
}

func itoa(v int64) string {
	return strconv.FormatInt(v, 10)
}
