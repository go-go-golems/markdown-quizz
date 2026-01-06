package trpc

import (
	"context"
	"encoding/json"
	"net/http"
	"net/http/httptest"
	"net/url"
	"os"
	"strings"
	"testing"

	"github.com/go-go-golems/XXX/internal/db"
	"github.com/go-go-golems/XXX/internal/documents"
	"github.com/go-go-golems/XXX/internal/quiz"
	"github.com/stretchr/testify/require"
)

func TestTRPCServer_SystemHealth(t *testing.T) {
	s, _ := newTestServer(t)

	input := `{"json":{"timestamp":123}}`
	req := httptest.NewRequest(http.MethodGet, "/api/trpc/system.health?input="+url.QueryEscape(input), nil)
	rec := httptest.NewRecorder()
	s.ServeHTTP(rec, req)

	require.Equal(t, http.StatusOK, rec.Code)
	var resp map[string]any
	require.NoError(t, json.Unmarshal(rec.Body.Bytes(), &resp))
	require.Equal(t, true, resp["result"].(map[string]any)["data"].(map[string]any)["json"].(map[string]any)["ok"])
}

func TestTRPCServer_BatchQuery(t *testing.T) {
	s, _ := newTestServer(t)

	input := `{"0":{},"1":{}}`
	req := httptest.NewRequest(http.MethodGet, "/api/trpc/documents.list,auth.me?batch=1&input="+url.QueryEscape(input), nil)
	rec := httptest.NewRecorder()
	s.ServeHTTP(rec, req)

	require.Equal(t, http.StatusOK, rec.Code)
	var resp []any
	require.NoError(t, json.Unmarshal(rec.Body.Bytes(), &resp))
	require.Len(t, resp, 2)
}

func TestTRPCServer_DocumentsCreateAndGetBySlug(t *testing.T) {
	s, stores := newTestServer(t)

	createBody := `{"json":{"title":"Doc","content":"hello","isPublished":false}}`
	req := httptest.NewRequest(http.MethodPost, "/api/trpc/documents.create", strings.NewReader(createBody))
	rec := httptest.NewRecorder()
	s.ServeHTTP(rec, req)

	require.Equal(t, http.StatusOK, rec.Code)
	var createResp map[string]any
	require.NoError(t, json.Unmarshal(rec.Body.Bytes(), &createResp))
	data := createResp["result"].(map[string]any)["data"].(map[string]any)["json"].(map[string]any)
	slug, _ := data["slug"].(string)
	require.NotEmpty(t, slug)

	doc, forms, err := stores.docStore.GetBySlug(context.Background(), slug)
	require.NoError(t, err)
	require.NotNil(t, doc)
	require.Len(t, forms, 0)

	getInput := `{"json":{"slug":"` + slug + `"}}`
	getReq := httptest.NewRequest(http.MethodGet, "/api/trpc/documents.getBySlug?input="+url.QueryEscape(getInput), nil)
	getRec := httptest.NewRecorder()
	s.ServeHTTP(getRec, getReq)

	require.Equal(t, http.StatusOK, getRec.Code)
}

type testStores struct {
	docStore *documents.Store
}

func newTestServer(t *testing.T) (*Server, testStores) {
	t.Helper()

	ctx := context.Background()
	f, err := os.CreateTemp("", "markdown-quizz-trpc-*.sqlite")
	require.NoError(t, err)
	_ = f.Close()
	t.Cleanup(func() { _ = os.Remove(f.Name()) })

	sqliteDB, err := db.OpenSQLite(ctx, db.SQLiteOptions{Path: f.Name()})
	require.NoError(t, err)
	t.Cleanup(func() { _ = sqliteDB.Close() })

	docStore := documents.NewStore(sqliteDB)
	quizStore := quiz.NewStore(sqliteDB)

	s := NewServer(Server{
		Documents: docStore,
		Quiz:      quizStore,
		UserID:    1,
	})

	return s, testStores{docStore: docStore}
}
