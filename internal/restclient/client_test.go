package restclient

import (
	"context"
	"encoding/json"
	"net/http"
	"net/http/httptest"
	"net/url"
	"testing"
	"time"

	"github.com/stretchr/testify/require"
)

func TestClient_New_ValidatesBaseURL(t *testing.T) {
	_, err := New(Options{BaseURL: "127.0.0.1:9092"})
	require.Error(t, err)
}

func TestClient_ListDocuments_PathAndQuery(t *testing.T) {
	t.Parallel()

	srv := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		require.Equal(t, http.MethodGet, r.Method)
		require.Equal(t, "/api/documents", r.URL.Path)
		require.Equal(t, "mine", r.URL.Query().Get("scope"))

		_ = json.NewEncoder(w).Encode([]Document{
			{ID: 1, Title: "Doc", Slug: "doc"},
		})
	}))
	defer srv.Close()

	c, err := New(Options{BaseURL: srv.URL, Timeout: 2 * time.Second})
	require.NoError(t, err)

	docs, err := c.ListDocuments(context.Background(), "mine")
	require.NoError(t, err)
	require.Len(t, docs, 1)
	require.Equal(t, int64(1), docs[0].ID)
}

func TestClient_SubmitQuizBatch_Path(t *testing.T) {
	t.Parallel()

	srv := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		require.Equal(t, http.MethodPost, r.Method)
		require.Equal(t, "/api/quiz/submissions/batch", r.URL.Path)

		_ = json.NewEncoder(w).Encode(SubmitQuizBatchResponse{
			Results: []SubmitQuizBatchResult{{FormID: "f1", Score: 1, MaxScore: 2}},
		})
	}))
	defer srv.Close()

	c, err := New(Options{BaseURL: srv.URL, Timeout: 2 * time.Second})
	require.NoError(t, err)

	res, err := c.SubmitQuizBatch(context.Background(), SubmitQuizBatchRequest{
		DocumentID:  1,
		Submissions: []SubmitQuizBatchItem{{FormID: "f1", Responses: map[string]any{"q1": "a"}}},
	})
	require.NoError(t, err)
	require.Len(t, res.Results, 1)
}

func TestClient_ErrorEnvelope(t *testing.T) {
	t.Parallel()

	srv := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, _ *http.Request) {
		w.WriteHeader(http.StatusBadRequest)
		_ = json.NewEncoder(w).Encode(apiErrorEnvelope{
			Error: apiError{
				Code:    "bad_request",
				Message: "title is required",
				Details: map[string]any{"field": "title"},
			},
		})
	}))
	defer srv.Close()

	c, err := New(Options{BaseURL: srv.URL, Timeout: 2 * time.Second})
	require.NoError(t, err)

	_, err = c.CreateDocument(context.Background(), CreateDocumentRequest{
		Title:       "",
		Content:     "x",
		IsPublished: false,
	})
	require.Error(t, err)
	var apiErr *APIError
	require.ErrorAs(t, err, &apiErr)
	require.Equal(t, http.StatusBadRequest, apiErr.Status)
	require.Equal(t, "bad_request", apiErr.Code)
	require.Equal(t, "title is required", apiErr.Msg)
}

func TestClient_BaseURLWithPathPrefix(t *testing.T) {
	t.Parallel()

	srv := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		require.Equal(t, "/prefix/api/documents", r.URL.Path)
		_ = json.NewEncoder(w).Encode([]Document{})
	}))
	defer srv.Close()

	u, err := url.Parse(srv.URL)
	require.NoError(t, err)
	u.Path = "/prefix"

	c, err := New(Options{BaseURL: u.String(), Timeout: 2 * time.Second})
	require.NoError(t, err)

	_, err = c.ListDocuments(context.Background(), "all")
	require.NoError(t, err)
}
