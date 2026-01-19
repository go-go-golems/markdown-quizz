package restclient

import (
	"context"
	"fmt"
	"net/http"
	"net/url"
)

func (c *Client) ListDocuments(ctx context.Context, scope string) ([]Document, error) {
	var out []Document
	q := url.Values{}
	if scope != "" {
		q.Set("scope", scope)
	}
	if err := c.doJSON(ctx, http.MethodGet, "api/documents", q, nil, &out); err != nil {
		return nil, err
	}
	return out, nil
}

func (c *Client) GetDocumentByID(ctx context.Context, id int64) (*Document, error) {
	var out Document
	if err := c.doJSON(ctx, http.MethodGet, fmt.Sprintf("api/documents/%d", id), nil, nil, &out); err != nil {
		return nil, err
	}
	return &out, nil
}

func (c *Client) GetDocumentBySlug(ctx context.Context, slug string) (*DocumentWithForms, error) {
	var out DocumentWithForms
	if err := c.doJSON(ctx, http.MethodGet, "api/documents/by-slug/"+url.PathEscape(slug), nil, nil, &out); err != nil {
		return nil, err
	}
	return &out, nil
}

func (c *Client) CreateDocument(ctx context.Context, req CreateDocumentRequest) (*CreateDocumentResponse, error) {
	var out CreateDocumentResponse
	if err := c.doJSON(ctx, http.MethodPost, "api/documents", nil, req, &out); err != nil {
		return nil, err
	}
	return &out, nil
}

func (c *Client) UpdateDocument(ctx context.Context, id int64, req UpdateDocumentRequest) (*SuccessResponse, error) {
	var out SuccessResponse
	if err := c.doJSON(ctx, http.MethodPatch, fmt.Sprintf("api/documents/%d", id), nil, req, &out); err != nil {
		return nil, err
	}
	return &out, nil
}

func (c *Client) DeleteDocument(ctx context.Context, id int64) (*SuccessResponse, error) {
	var out SuccessResponse
	if err := c.doJSON(ctx, http.MethodDelete, fmt.Sprintf("api/documents/%d", id), nil, nil, &out); err != nil {
		return nil, err
	}
	return &out, nil
}

func (c *Client) GetDocumentAnalytics(ctx context.Context, documentID int64) (*DocumentAnalytics, error) {
	var out DocumentAnalytics
	if err := c.doJSON(ctx, http.MethodGet, fmt.Sprintf("api/documents/%d/analytics", documentID), nil, nil, &out); err != nil {
		return nil, err
	}
	return &out, nil
}

func (c *Client) GetDocumentSubmissions(ctx context.Context, documentID int64) ([]SubmissionWithUser, error) {
	var out []SubmissionWithUser
	if err := c.doJSON(ctx, http.MethodGet, fmt.Sprintf("api/documents/%d/submissions", documentID), nil, nil, &out); err != nil {
		return nil, err
	}
	return out, nil
}
