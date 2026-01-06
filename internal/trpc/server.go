package trpc

import (
	"context"
	"encoding/json"
	stderrors "errors"
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

func (s *Server) ServeHTTP(w http.ResponseWriter, r *http.Request) {
	w.Header().Set("Content-Type", "application/json; charset=utf-8")

	paths, err := parseTRPCPaths(r.URL)
	if err != nil {
		_ = json.NewEncoder(w).Encode(errorEnvelope("BAD_REQUEST", err.Error(), ""))
		return
	}

	isBatch := r.URL.Query().Get("batch") == "1" || len(paths) > 1

	rawInput, err := readTRPCInput(r, isBatch)
	if err != nil {
		_ = json.NewEncoder(w).Encode(errorEnvelope("PARSE_ERROR", err.Error(), strings.Join(paths, ",")))
		return
	}

	if !isBatch {
		out := s.handleCall(r.Context(), paths[0], decodeSuperJSON(rawInput))
		_ = json.NewEncoder(w).Encode(out)
		return
	}

	inputs := map[string]any{}
	if rawInput != nil {
		if m, ok := rawInput.(map[string]any); ok {
			inputs = m
		}
	}

	out := make([]any, 0, len(paths))
	for i, path := range paths {
		raw := inputs[strconv.Itoa(i)]
		out = append(out, s.handleCall(r.Context(), path, decodeSuperJSON(raw)))
	}

	_ = json.NewEncoder(w).Encode(out)
}

func parseTRPCPaths(u *url.URL) ([]string, error) {
	raw := u.Path
	if raw == "/api/trpc" || raw == "/api/trpc/" {
		return nil, stderrors.New("missing procedure path")
	}

	raw = strings.TrimPrefix(raw, "/api/trpc")
	raw = strings.TrimPrefix(raw, "/")
	if raw == "" {
		return nil, stderrors.New("missing procedure path")
	}

	parts := strings.Split(raw, ",")
	out := make([]string, 0, len(parts))
	for _, p := range parts {
		p = strings.TrimSpace(p)
		if p == "" {
			continue
		}
		out = append(out, p)
	}
	if len(out) == 0 {
		return nil, stderrors.New("missing procedure path")
	}
	return out, nil
}

func readTRPCInput(r *http.Request, isBatch bool) (any, error) {
	if r.Method == http.MethodGet {
		raw := r.URL.Query().Get("input")
		if raw == "" {
			return nil, nil
		}
		var v any
		if err := json.Unmarshal([]byte(raw), &v); err != nil {
			return nil, pkgerrors.Wrap(err, "decode input query param")
		}
		return v, nil
	}

	body, err := io.ReadAll(io.LimitReader(r.Body, 10<<20))
	if err != nil {
		return nil, pkgerrors.Wrap(err, "read request body")
	}
	if len(body) == 0 {
		if isBatch {
			return map[string]any{}, nil
		}
		return nil, nil
	}
	defer func() { _ = r.Body.Close() }()

	var v any
	if err := json.Unmarshal(body, &v); err != nil {
		return nil, pkgerrors.Wrap(err, "decode input body")
	}
	return v, nil
}

func decodeSuperJSON(v any) any {
	if v == nil {
		return nil
	}
	m, ok := v.(map[string]any)
	if !ok {
		return v
	}
	if len(m) == 0 {
		return nil
	}
	if jsonVal, ok := m["json"]; ok {
		return jsonVal
	}
	return v
}

func (s *Server) handleCall(ctx context.Context, path string, input any) any {
	switch path {
	case "system.health":
		return resultEnvelope(map[string]any{"ok": true})
	case "system.notifyOwner":
		return resultEnvelope(map[string]any{"success": false})
	case "auth.me":
		return resultEnvelope(map[string]any{"id": s.UserID, "role": "admin"})
	case "auth.logout":
		return resultEnvelope(map[string]any{"success": true})
	case "documents.list":
		docs, err := s.Documents.ListDocuments(ctx, nil, false)
		if err != nil {
			return errorEnvelope("INTERNAL_SERVER_ERROR", "failed to list documents", path)
		}
		out := make([]any, 0, len(docs))
		for _, d := range docs {
			out = append(out, documentToMap(d))
		}
		return resultEnvelope(out)
	case "documents.myDocuments":
		authorID := s.UserID
		docs, err := s.Documents.ListDocuments(ctx, &authorID, false)
		if err != nil {
			return errorEnvelope("INTERNAL_SERVER_ERROR", "failed to list documents", path)
		}
		out := make([]any, 0, len(docs))
		for _, d := range docs {
			out = append(out, documentToMap(d))
		}
		return resultEnvelope(out)
	case "documents.getBySlug":
		in, ok := input.(map[string]any)
		if !ok {
			return errorEnvelope("BAD_REQUEST", "invalid input", path)
		}
		slug, _ := in["slug"].(string)
		if slug == "" {
			return errorEnvelope("BAD_REQUEST", "slug is required", path)
		}
		doc, forms, err := s.Documents.GetBySlug(ctx, slug)
		if err != nil {
			return errorEnvelope("INTERNAL_SERVER_ERROR", "failed to load document", path)
		}
		if doc == nil {
			return errorEnvelope("NOT_FOUND", "Document not found", path)
		}

		out := documentToMap(*doc)
		out["forms"] = quizFormsToAny(forms)
		return resultEnvelope(out)
	case "documents.getById":
		in, ok := input.(map[string]any)
		if !ok {
			return errorEnvelope("BAD_REQUEST", "invalid input", path)
		}
		id, ok := asInt64(in["id"])
		if !ok {
			return errorEnvelope("BAD_REQUEST", "id is required", path)
		}
		doc, err := s.Documents.GetByID(ctx, id)
		if err != nil {
			return errorEnvelope("INTERNAL_SERVER_ERROR", "failed to load document", path)
		}
		if doc == nil {
			return errorEnvelope("NOT_FOUND", "Document not found", path)
		}
		return resultEnvelope(documentToMap(*doc))
	case "documents.create":
		in, ok := input.(map[string]any)
		if !ok {
			return errorEnvelope("BAD_REQUEST", "invalid input", path)
		}
		title, _ := in["title"].(string)
		content, _ := in["content"].(string)
		if title == "" {
			return errorEnvelope("BAD_REQUEST", "title is required", path)
		}
		isPublished := false
		if v, ok := in["isPublished"].(bool); ok {
			isPublished = v
		}
		var description *string
		if v, ok := in["description"].(string); ok {
			description = &v
		}
		var category *string
		if v, ok := in["category"].(string); ok {
			category = &v
		}

		id, slug, err := s.Documents.Create(ctx, documents.CreateParams{
			Title:       title,
			Content:     content,
			Description: description,
			Category:    category,
			IsPublished: isPublished,
			AuthorID:    s.UserID,
		})
		if err != nil {
			return errorEnvelope("INTERNAL_SERVER_ERROR", "failed to create document", path)
		}
		return resultEnvelope(map[string]any{"id": id, "slug": slug})
	case "documents.update":
		in, ok := input.(map[string]any)
		if !ok {
			return errorEnvelope("BAD_REQUEST", "invalid input", path)
		}
		id, ok := asInt64(in["id"])
		if !ok {
			return errorEnvelope("BAD_REQUEST", "id is required", path)
		}
		var title *string
		if v, ok := in["title"].(string); ok {
			title = &v
		}
		var content *string
		if v, ok := in["content"].(string); ok {
			content = &v
		}
		var description *string
		if v, ok := in["description"].(string); ok {
			description = &v
		}
		var category *string
		if v, ok := in["category"].(string); ok {
			category = &v
		}
		var isPublished *bool
		if v, ok := in["isPublished"].(bool); ok {
			isPublished = &v
		}

		if err := s.Documents.Update(ctx, documents.UpdateParams{
			ID:          id,
			Title:       title,
			Content:     content,
			Description: description,
			Category:    category,
			IsPublished: isPublished,
		}); err != nil {
			return errorEnvelope("INTERNAL_SERVER_ERROR", "failed to update document", path)
		}
		return resultEnvelope(map[string]any{"success": true})
	case "documents.delete":
		in, ok := input.(map[string]any)
		if !ok {
			return errorEnvelope("BAD_REQUEST", "invalid input", path)
		}
		id, ok := asInt64(in["id"])
		if !ok {
			return errorEnvelope("BAD_REQUEST", "id is required", path)
		}
		if err := s.Documents.Delete(ctx, id); err != nil {
			return errorEnvelope("INTERNAL_SERVER_ERROR", "failed to delete document", path)
		}
		return resultEnvelope(map[string]any{"success": true})
	case "documents.analytics":
		in, ok := input.(map[string]any)
		if !ok {
			return errorEnvelope("BAD_REQUEST", "invalid input", path)
		}
		id, ok := asInt64(in["id"])
		if !ok {
			return errorEnvelope("BAD_REQUEST", "id is required", path)
		}
		a, err := s.Quiz.GetQuizAnalytics(ctx, id)
		if err != nil {
			return errorEnvelope("INTERNAL_SERVER_ERROR", "failed to compute analytics", path)
		}
		return resultEnvelope(map[string]any{
			"totalSubmissions": a.TotalSubmissions,
			"averageScore":     a.AverageScore,
			"highestScore":     a.HighestScore,
			"lowestScore":      a.LowestScore,
		})
	case "documents.submissions":
		in, ok := input.(map[string]any)
		if !ok {
			return errorEnvelope("BAD_REQUEST", "invalid input", path)
		}
		id, ok := asInt64(in["id"])
		if !ok {
			return errorEnvelope("BAD_REQUEST", "id is required", path)
		}
		subs, err := s.Quiz.GetSubmissionsByDocument(ctx, id)
		if err != nil {
			return errorEnvelope("INTERNAL_SERVER_ERROR", "failed to list submissions", path)
		}
		out := make([]any, 0, len(subs))
		for _, it := range subs {
			out = append(out, submissionWithUserToAny(it))
		}
		return resultEnvelope(out)
	case "quiz.submitMultiple":
		in, ok := input.(map[string]any)
		if !ok {
			return errorEnvelope("BAD_REQUEST", "invalid input", path)
		}
		documentID, ok := asInt64(in["documentId"])
		if !ok {
			return errorEnvelope("BAD_REQUEST", "documentId is required", path)
		}
		rawSubs, ok := in["submissions"].([]any)
		if !ok {
			return errorEnvelope("BAD_REQUEST", "submissions is required", path)
		}
		subs := make([]quiz.SubmissionInput, 0, len(rawSubs))
		for _, it := range rawSubs {
			m, ok := it.(map[string]any)
			if !ok {
				continue
			}
			formID, _ := m["formId"].(string)
			resp, _ := m["responses"].(map[string]any)
			subs = append(subs, quiz.SubmissionInput{FormID: formID, Responses: resp})
		}
		results, err := s.Quiz.SubmitMultiple(ctx, s.UserID, documentID, subs)
		if err != nil {
			return errorEnvelope("INTERNAL_SERVER_ERROR", "failed to submit quizzes", path)
		}
		out := make([]any, 0, len(results))
		for _, r := range results {
			out = append(out, map[string]any{"formId": r.FormID, "score": r.Score, "maxScore": r.MaxScore})
		}
		return resultEnvelope(map[string]any{"results": out})
	case "quiz.submit":
		in, ok := input.(map[string]any)
		if !ok {
			return errorEnvelope("BAD_REQUEST", "invalid input", path)
		}
		documentID, ok := asInt64(in["documentId"])
		if !ok {
			return errorEnvelope("BAD_REQUEST", "documentId is required", path)
		}
		formID, _ := in["formId"].(string)
		if formID == "" {
			return errorEnvelope("BAD_REQUEST", "formId is required", path)
		}
		resp, _ := in["responses"].(map[string]any)

		id, score, max, err := s.Quiz.Submit(ctx, s.UserID, documentID, formID, resp)
		if err != nil {
			return errorEnvelope("INTERNAL_SERVER_ERROR", "failed to submit quiz", path)
		}
		out := map[string]any{"id": id, "score": score, "maxScore": max}
		return resultEnvelope(out)
	case "quiz.mySubmissions":
		subs, err := s.Quiz.GetSubmissionsByUser(ctx, s.UserID)
		if err != nil {
			return errorEnvelope("INTERNAL_SERVER_ERROR", "failed to list submissions", path)
		}
		out := make([]any, 0, len(subs))
		for _, it := range subs {
			out = append(out, submissionWithDocToAny(it))
		}
		return resultEnvelope(out)
	case "quiz.getSubmission":
		in, ok := input.(map[string]any)
		if !ok {
			return errorEnvelope("BAD_REQUEST", "invalid input", path)
		}
		id, ok := asInt64(in["id"])
		if !ok {
			return errorEnvelope("BAD_REQUEST", "id is required", path)
		}
		detail, err := s.Quiz.GetSubmissionByID(ctx, id)
		if err != nil {
			return errorEnvelope("INTERNAL_SERVER_ERROR", "failed to load submission", path)
		}
		if detail == nil {
			return errorEnvelope("NOT_FOUND", "Submission not found", path)
		}
		return resultEnvelope(map[string]any{
			"submission":     submissionToAny(detail.Submission),
			"documentTitle":  detail.DocumentTitle,
			"documentSlug":   detail.DocumentSlug,
			"formDefinition": detail.FormDefinition,
		})
	default:
		return errorEnvelope("NOT_FOUND", "Procedure not found", path)
	}
}

func resultEnvelope(data any) any {
	return map[string]any{
		"result": map[string]any{
			"data": map[string]any{
				"json": data,
			},
		},
	}
}

func errorEnvelope(codeKey string, message string, path string) any {
	codeNumber, httpStatus := trpcCode(codeKey)
	shape := map[string]any{
		"code":    codeNumber,
		"message": message,
		"data": map[string]any{
			"code":       codeKey,
			"httpStatus": httpStatus,
			"path":       path,
		},
	}

	return map[string]any{
		"error": map[string]any{
			"json": shape,
		},
	}
}

func trpcCode(codeKey string) (int, int) {
	switch codeKey {
	case "PARSE_ERROR":
		return -32700, 400
	case "BAD_REQUEST":
		return -32600, 400
	case "INTERNAL_SERVER_ERROR":
		return -32603, 500
	case "UNAUTHORIZED":
		return -32001, 401
	case "FORBIDDEN":
		return -32003, 403
	case "NOT_FOUND":
		return -32004, 404
	case "METHOD_NOT_SUPPORTED":
		return -32005, 405
	default:
		return -32603, 500
	}
}

func asInt64(v any) (int64, bool) {
	switch t := v.(type) {
	case float64:
		return int64(t), true
	case float32:
		return int64(t), true
	case int:
		return int64(t), true
	case int64:
		return t, true
	case json.Number:
		i, err := t.Int64()
		if err != nil {
			return 0, false
		}
		return i, true
	default:
		return 0, false
	}
}

func documentToMap(d documents.Document) map[string]any {
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

func quizFormsToAny(forms []documents.QuizForm) []any {
	out := make([]any, 0, len(forms))
	for _, f := range forms {
		out = append(out, map[string]any{
			"id":         f.ID,
			"documentId": f.DocumentID,
			"formId":     f.FormID,
			"definition": f.Definition,
			"createdAt":  f.CreatedAt,
			"updatedAt":  f.UpdatedAt,
		})
	}
	return out
}

func submissionToAny(s quiz.Submission) map[string]any {
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

func submissionWithDocToAny(s quiz.SubmissionWithDocument) map[string]any {
	return map[string]any{
		"submission":    submissionToAny(s.Submission),
		"documentTitle": s.DocumentTitle,
		"documentSlug":  s.DocumentSlug,
	}
}

func submissionWithUserToAny(s quiz.SubmissionWithUser) map[string]any {
	return map[string]any{
		"submission": submissionToAny(s.Submission),
		"userName":   s.UserName,
	}
}
