package rest

import (
	"bytes"
	"encoding/json"
	"io"
	"net/http"
	"net/url"
	"strconv"
	"strings"

	"github.com/go-go-golems/XXX/internal/documents"
	"github.com/go-go-golems/XXX/internal/quiz"
	pkgerrors "github.com/pkg/errors"
)

type Server struct {
	Documents *documents.Store
	Quiz      *quiz.Store

	UserID int64
}

func NewServer(opts Server) *Server {
	s := opts
	if s.UserID == 0 {
		s.UserID = 1
	}
	return &s
}

type apiErrorEnvelope struct {
	Error apiError `json:"error"`
}

type apiError struct {
	Code    string `json:"code"`
	Message string `json:"message"`
	Details any    `json:"details,omitempty"`
}

func (s *Server) ServeHTTP(w http.ResponseWriter, r *http.Request) {
	w.Header().Set("Content-Type", "application/json; charset=utf-8")

	if !strings.HasPrefix(r.URL.Path, "/api/") {
		writeError(w, http.StatusNotFound, "not_found", "not found", nil)
		return
	}

	parts, err := splitPath(strings.TrimPrefix(r.URL.Path, "/api/"))
	if err != nil {
		writeError(w, http.StatusBadRequest, "bad_request", err.Error(), nil)
		return
	}
	if len(parts) == 0 {
		writeError(w, http.StatusNotFound, "not_found", "not found", nil)
		return
	}

	switch parts[0] {
	case "documents":
		s.handleDocuments(w, r, parts[1:])
		return
	case "quiz":
		s.handleQuiz(w, r, parts[1:])
		return
	default:
		writeError(w, http.StatusNotFound, "not_found", "not found", nil)
		return
	}
}

func (s *Server) handleDocuments(w http.ResponseWriter, r *http.Request, parts []string) {
	if len(parts) == 0 {
		switch r.Method {
		case http.MethodGet:
			scope := strings.TrimSpace(r.URL.Query().Get("scope"))
			if scope == "" {
				scope = "all"
			}
			switch scope {
			case "all":
				docs, err := s.Documents.ListDocuments(r.Context(), nil, false)
				if err != nil {
					writeError(w, http.StatusInternalServerError, "internal_server_error", "failed to list documents", nil)
					return
				}
				writeJSON(w, http.StatusOK, documentsToAPI(docs))
				return
			case "mine":
				authorID := s.UserID
				docs, err := s.Documents.ListDocuments(r.Context(), &authorID, false)
				if err != nil {
					writeError(w, http.StatusInternalServerError, "internal_server_error", "failed to list documents", nil)
					return
				}
				writeJSON(w, http.StatusOK, documentsToAPI(docs))
				return
			default:
				writeError(w, http.StatusBadRequest, "bad_request", "invalid scope", map[string]any{"scope": scope})
				return
			}
		case http.MethodPost:
			var in createDocumentRequest
			if err := decodeJSON(r, &in); err != nil {
				writeError(w, http.StatusBadRequest, "bad_request", "invalid JSON body", map[string]any{"error": err.Error()})
				return
			}
			if strings.TrimSpace(in.Title) == "" {
				writeError(w, http.StatusBadRequest, "bad_request", "title is required", map[string]any{"field": "title"})
				return
			}
			if strings.TrimSpace(in.Content) == "" {
				writeError(w, http.StatusBadRequest, "bad_request", "content is required", map[string]any{"field": "content"})
				return
			}

			id, slug, err := s.Documents.Create(r.Context(), documents.CreateParams{
				Title:       in.Title,
				Content:     in.Content,
				Description: in.Description,
				Category:    in.Category,
				IsPublished: in.IsPublished,
				AuthorID:    s.UserID,
			})
			if err != nil {
				writeError(w, http.StatusInternalServerError, "internal_server_error", "failed to create document", nil)
				return
			}

			writeJSON(w, http.StatusCreated, map[string]any{"id": id, "slug": slug})
			return
		default:
			writeError(w, http.StatusMethodNotAllowed, "method_not_allowed", "method not allowed", nil)
			return
		}
	}

	if len(parts) >= 2 && parts[0] == "by-slug" {
		if r.Method != http.MethodGet {
			writeError(w, http.StatusMethodNotAllowed, "method_not_allowed", "method not allowed", nil)
			return
		}
		slug := parts[1]
		if slug == "" {
			writeError(w, http.StatusBadRequest, "bad_request", "slug is required", map[string]any{"field": "slug"})
			return
		}
		doc, forms, err := s.Documents.GetBySlug(r.Context(), slug)
		if err != nil {
			writeError(w, http.StatusInternalServerError, "internal_server_error", "failed to load document", nil)
			return
		}
		if doc == nil {
			writeError(w, http.StatusNotFound, "not_found", "document not found", nil)
			return
		}
		out := documentToAPI(*doc)
		out["forms"] = quizFormsToAPI(forms)
		writeJSON(w, http.StatusOK, out)
		return
	}

	id, ok := parseInt64(parts[0])
	if !ok {
		writeError(w, http.StatusBadRequest, "bad_request", "invalid document id", map[string]any{"id": parts[0]})
		return
	}

	if len(parts) == 1 {
		switch r.Method {
		case http.MethodGet:
			doc, err := s.Documents.GetByID(r.Context(), id)
			if err != nil {
				writeError(w, http.StatusInternalServerError, "internal_server_error", "failed to load document", nil)
				return
			}
			if doc == nil {
				writeError(w, http.StatusNotFound, "not_found", "document not found", nil)
				return
			}
			writeJSON(w, http.StatusOK, documentToAPI(*doc))
			return
		case http.MethodPatch:
			var in updateDocumentRequest
			if err := decodeJSON(r, &in); err != nil {
				writeError(w, http.StatusBadRequest, "bad_request", "invalid JSON body", map[string]any{"error": err.Error()})
				return
			}
			if in.Title == nil && in.Content == nil && in.Description == nil && in.Category == nil && in.IsPublished == nil {
				writeError(w, http.StatusBadRequest, "bad_request", "no fields to update", nil)
				return
			}
			if err := s.Documents.Update(r.Context(), documents.UpdateParams{
				ID:          id,
				Title:       in.Title,
				Content:     in.Content,
				Description: in.Description,
				Category:    in.Category,
				IsPublished: in.IsPublished,
			}); err != nil {
				writeError(w, http.StatusInternalServerError, "internal_server_error", "failed to update document", nil)
				return
			}
			writeJSON(w, http.StatusOK, map[string]any{"success": true})
			return
		case http.MethodDelete:
			if err := s.Documents.Delete(r.Context(), id); err != nil {
				writeError(w, http.StatusInternalServerError, "internal_server_error", "failed to delete document", nil)
				return
			}
			writeJSON(w, http.StatusOK, map[string]any{"success": true})
			return
		default:
			writeError(w, http.StatusMethodNotAllowed, "method_not_allowed", "method not allowed", nil)
			return
		}
	}

	if len(parts) == 2 && parts[1] == "analytics" {
		if r.Method != http.MethodGet {
			writeError(w, http.StatusMethodNotAllowed, "method_not_allowed", "method not allowed", nil)
			return
		}
		a, err := s.Quiz.GetQuizAnalytics(r.Context(), id)
		if err != nil {
			writeError(w, http.StatusInternalServerError, "internal_server_error", "failed to compute analytics", nil)
			return
		}
		writeJSON(w, http.StatusOK, map[string]any{
			"totalSubmissions": a.TotalSubmissions,
			"averageScore":     a.AverageScore,
			"highestScore":     a.HighestScore,
			"lowestScore":      a.LowestScore,
		})
		return
	}

	if len(parts) == 2 && parts[1] == "submissions" {
		if r.Method != http.MethodGet {
			writeError(w, http.StatusMethodNotAllowed, "method_not_allowed", "method not allowed", nil)
			return
		}
		subs, err := s.Quiz.GetSubmissionsByDocument(r.Context(), id)
		if err != nil {
			writeError(w, http.StatusInternalServerError, "internal_server_error", "failed to list submissions", nil)
			return
		}
		out := make([]any, 0, len(subs))
		for _, it := range subs {
			out = append(out, submissionWithUserToAPI(it))
		}
		writeJSON(w, http.StatusOK, out)
		return
	}

	writeError(w, http.StatusNotFound, "not_found", "not found", nil)
}

func (s *Server) handleQuiz(w http.ResponseWriter, r *http.Request, parts []string) {
	if len(parts) == 0 {
		writeError(w, http.StatusNotFound, "not_found", "not found", nil)
		return
	}

	if parts[0] != "submissions" {
		writeError(w, http.StatusNotFound, "not_found", "not found", nil)
		return
	}

	if len(parts) == 1 {
		switch r.Method {
		case http.MethodGet:
			scope := strings.TrimSpace(r.URL.Query().Get("scope"))
			if scope == "" {
				scope = "mine"
			}
			if scope != "mine" {
				writeError(w, http.StatusBadRequest, "bad_request", "invalid scope", map[string]any{"scope": scope})
				return
			}
			subs, err := s.Quiz.GetSubmissionsByUser(r.Context(), s.UserID)
			if err != nil {
				writeError(w, http.StatusInternalServerError, "internal_server_error", "failed to list submissions", nil)
				return
			}
			out := make([]any, 0, len(subs))
			for _, it := range subs {
				out = append(out, submissionWithDocumentToAPI(it))
			}
			writeJSON(w, http.StatusOK, out)
			return
		case http.MethodPost:
			var in submitQuizRequest
			if err := decodeJSON(r, &in); err != nil {
				writeError(w, http.StatusBadRequest, "bad_request", "invalid JSON body", map[string]any{"error": err.Error()})
				return
			}
			if in.DocumentID == 0 {
				writeError(w, http.StatusBadRequest, "bad_request", "documentId is required", map[string]any{"field": "documentId"})
				return
			}
			if strings.TrimSpace(in.FormID) == "" {
				writeError(w, http.StatusBadRequest, "bad_request", "formId is required", map[string]any{"field": "formId"})
				return
			}
			if in.Responses == nil {
				in.Responses = map[string]any{}
			}
			id, score, maxScore, err := s.Quiz.Submit(r.Context(), s.UserID, in.DocumentID, in.FormID, in.Responses)
			if err != nil {
				writeError(w, http.StatusInternalServerError, "internal_server_error", "failed to submit quiz", nil)
				return
			}
			writeJSON(w, http.StatusCreated, map[string]any{"id": id, "score": score, "maxScore": maxScore})
			return
		default:
			writeError(w, http.StatusMethodNotAllowed, "method_not_allowed", "method not allowed", nil)
			return
		}
	}

	if len(parts) == 2 && parts[1] == "batch" {
		if r.Method != http.MethodPost {
			writeError(w, http.StatusMethodNotAllowed, "method_not_allowed", "method not allowed", nil)
			return
		}
		var in submitQuizBatchRequest
		if err := decodeJSON(r, &in); err != nil {
			writeError(w, http.StatusBadRequest, "bad_request", "invalid JSON body", map[string]any{"error": err.Error()})
			return
		}
		if in.DocumentID == 0 {
			writeError(w, http.StatusBadRequest, "bad_request", "documentId is required", map[string]any{"field": "documentId"})
			return
		}
		if len(in.Submissions) == 0 {
			writeError(w, http.StatusBadRequest, "bad_request", "submissions is required", map[string]any{"field": "submissions"})
			return
		}
		subs := make([]quiz.SubmissionInput, 0, len(in.Submissions))
		for _, it := range in.Submissions {
			resp := it.Responses
			if resp == nil {
				resp = map[string]any{}
			}
			subs = append(subs, quiz.SubmissionInput{FormID: it.FormID, Responses: resp})
		}
		results, err := s.Quiz.SubmitMultiple(r.Context(), s.UserID, in.DocumentID, subs)
		if err != nil {
			writeError(w, http.StatusInternalServerError, "internal_server_error", "failed to submit quizzes", nil)
			return
		}
		out := make([]any, 0, len(results))
		for _, r := range results {
			out = append(out, map[string]any{"formId": r.FormID, "score": r.Score, "maxScore": r.MaxScore})
		}
		writeJSON(w, http.StatusOK, map[string]any{"results": out})
		return
	}

	id, ok := parseInt64(parts[1])
	if !ok || len(parts) != 2 {
		writeError(w, http.StatusNotFound, "not_found", "not found", nil)
		return
	}
	if r.Method != http.MethodGet {
		writeError(w, http.StatusMethodNotAllowed, "method_not_allowed", "method not allowed", nil)
		return
	}
	detail, err := s.Quiz.GetSubmissionByID(r.Context(), id)
	if err != nil {
		writeError(w, http.StatusInternalServerError, "internal_server_error", "failed to load submission", nil)
		return
	}
	if detail == nil {
		writeError(w, http.StatusNotFound, "not_found", "submission not found", nil)
		return
	}

	writeJSON(w, http.StatusOK, map[string]any{
		"submission":     submissionToAPI(detail.SubmissionWithDocument.Submission),
		"documentTitle":  detail.DocumentTitle,
		"documentSlug":   detail.DocumentSlug,
		"formDefinition": detail.FormDefinition,
	})
}

type createDocumentRequest struct {
	Title       string  `json:"title"`
	Content     string  `json:"content"`
	Description *string `json:"description,omitempty"`
	Category    *string `json:"category,omitempty"`
	IsPublished bool    `json:"isPublished"`
}

type updateDocumentRequest struct {
	Title       *string `json:"title,omitempty"`
	Content     *string `json:"content,omitempty"`
	Description *string `json:"description,omitempty"`
	Category    *string `json:"category,omitempty"`
	IsPublished *bool   `json:"isPublished,omitempty"`
}

type submitQuizRequest struct {
	DocumentID int64          `json:"documentId"`
	FormID     string         `json:"formId"`
	Responses  map[string]any `json:"responses"`
}

type submitQuizBatchRequest struct {
	DocumentID  int64                   `json:"documentId"`
	Submissions []submitQuizBatchItemIn `json:"submissions"`
}

type submitQuizBatchItemIn struct {
	FormID     string         `json:"formId"`
	Responses  map[string]any `json:"responses"`
}

func splitPath(p string) ([]string, error) {
	p = strings.Trim(p, "/")
	if p == "" {
		return nil, nil
	}
	raw := strings.Split(p, "/")
	out := make([]string, 0, len(raw))
	for _, it := range raw {
		if it == "" {
			continue
		}
		v, err := url.PathUnescape(it)
		if err != nil {
			return nil, pkgerrors.Wrap(err, "unescape path segment")
		}
		out = append(out, v)
	}
	return out, nil
}

func parseInt64(s string) (int64, bool) {
	v, err := strconv.ParseInt(s, 10, 64)
	if err != nil {
		return 0, false
	}
	return v, true
}

func decodeJSON(r *http.Request, dst any) error {
	body, err := io.ReadAll(io.LimitReader(r.Body, 2<<20))
	if err != nil {
		return pkgerrors.Wrap(err, "read body")
	}
	defer func() { _ = r.Body.Close() }()
	if len(body) == 0 {
		return pkgerrors.New("empty body")
	}
	dec := json.NewDecoder(bytes.NewReader(body))
	dec.DisallowUnknownFields()
	if err := dec.Decode(dst); err != nil {
		return pkgerrors.Wrap(err, "decode json")
	}
	if err := dec.Decode(&struct{}{}); err != io.EOF {
		return pkgerrors.New("unexpected trailing JSON")
	}
	return nil
}

func writeJSON(w http.ResponseWriter, status int, v any) {
	w.WriteHeader(status)
	_ = json.NewEncoder(w).Encode(v)
}

func writeError(w http.ResponseWriter, status int, code, message string, details any) {
	writeJSON(w, status, apiErrorEnvelope{
		Error: apiError{
			Code:    code,
			Message: message,
			Details: details,
		},
	})
}

func documentsToAPI(docs []documents.Document) []any {
	out := make([]any, 0, len(docs))
	for _, d := range docs {
		out = append(out, documentToAPI(d))
	}
	return out
}

func documentToAPI(d documents.Document) map[string]any {
	return map[string]any{
		"id":          d.ID,
		"title":       d.Title,
		"slug":        d.Slug,
		"content":     d.Content,
		"description": d.Description,
		"category":    d.Category,
		"isPublished": d.IsPublished,
		"authorId":    d.AuthorID,
		"createdAt":   d.CreatedAt,
		"updatedAt":   d.UpdatedAt,
	}
}

func quizFormsToAPI(forms []documents.QuizForm) []any {
	out := make([]any, 0, len(forms))
	for _, f := range forms {
		out = append(out, map[string]any{
			"formId":     f.FormID,
			"definition": f.Definition,
		})
	}
	return out
}

func submissionToAPI(s quiz.Submission) map[string]any {
	return map[string]any{
		"id":          s.ID,
		"userId":      s.UserID,
		"documentId":  s.DocumentID,
		"formId":      s.FormID,
		"responses":   s.Responses,
		"score":       s.Score,
		"maxScore":    s.MaxScore,
		"submittedAt": s.SubmittedAt,
	}
}

func submissionWithDocumentToAPI(s quiz.SubmissionWithDocument) map[string]any {
	return map[string]any{
		"submission":     submissionToAPI(s.Submission),
		"documentTitle":  s.DocumentTitle,
		"documentSlug":   s.DocumentSlug,
	}
}

func submissionWithUserToAPI(s quiz.SubmissionWithUser) map[string]any {
	return map[string]any{
		"submission": submissionToAPI(s.Submission),
		"userName":   s.UserName,
	}
}

var _ http.Handler = (*Server)(nil)
