package restclient

import (
	"context"
	"fmt"
	"net/http"
	"net/url"
)

func (c *Client) SubmitQuiz(ctx context.Context, req SubmitQuizRequest) (*SubmitQuizResponse, error) {
	var out SubmitQuizResponse
	if err := c.doJSON(ctx, http.MethodPost, "api/quiz/submissions", nil, req, &out); err != nil {
		return nil, err
	}
	return &out, nil
}

func (c *Client) SubmitQuizBatch(ctx context.Context, req SubmitQuizBatchRequest) (*SubmitQuizBatchResponse, error) {
	var out SubmitQuizBatchResponse
	if err := c.doJSON(ctx, http.MethodPost, "api/quiz/submissions/batch", nil, req, &out); err != nil {
		return nil, err
	}
	return &out, nil
}

func (c *Client) ListMySubmissions(ctx context.Context) ([]SubmissionWithDocument, error) {
	var out []SubmissionWithDocument
	q := url.Values{}
	q.Set("scope", "mine")
	if err := c.doJSON(ctx, http.MethodGet, "api/quiz/submissions", q, nil, &out); err != nil {
		return nil, err
	}
	return out, nil
}

func (c *Client) GetSubmissionByID(ctx context.Context, id int64) (*SubmissionDetail, error) {
	var out SubmissionDetail
	if err := c.doJSON(ctx, http.MethodGet, fmt.Sprintf("api/quiz/submissions/%d", id), nil, nil, &out); err != nil {
		return nil, err
	}
	return &out, nil
}
