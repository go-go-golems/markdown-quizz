package restclient

import (
	"bytes"
	"context"
	"encoding/json"
	"fmt"
	"io"
	"net/http"
	"net/url"
	"path"
	"strings"
	"time"

	pkgerrors "github.com/pkg/errors"
)

type Client struct {
	baseURL *url.URL
	http    *http.Client
}

type Options struct {
	BaseURL string
	Timeout time.Duration
}

func New(opts Options) (*Client, error) {
	if strings.TrimSpace(opts.BaseURL) == "" {
		return nil, pkgerrors.New("base URL is required")
	}
	u, err := url.Parse(opts.BaseURL)
	if err != nil {
		return nil, pkgerrors.Wrap(err, "parse base URL")
	}
	if u.Scheme == "" || u.Host == "" {
		return nil, pkgerrors.New("base URL must include scheme and host")
	}
	if opts.Timeout <= 0 {
		opts.Timeout = 10 * time.Second
	}

	u.Path = strings.TrimSuffix(u.Path, "/")

	return &Client{
		baseURL: u,
		http: &http.Client{
			Timeout: opts.Timeout,
		},
	}, nil
}

type apiErrorEnvelope struct {
	Error apiError `json:"error"`
}

type apiError struct {
	Code    string `json:"code"`
	Message string `json:"message"`
	Details any    `json:"details,omitempty"`
}

type APIError struct {
	Status int
	Code   string
	Msg    string
	Body   []byte
}

func (e *APIError) Error() string {
	if e.Code != "" {
		return fmt.Sprintf("api error: http %d: %s: %s", e.Status, e.Code, e.Msg)
	}
	return fmt.Sprintf("api error: http %d: %s", e.Status, e.Msg)
}

func (c *Client) doJSON(ctx context.Context, method, relPath string, query url.Values, in any, out any) error {
	u := *c.baseURL
	u.Path = path.Join(c.baseURL.Path, relPath)
	if !strings.HasPrefix(u.Path, "/") {
		u.Path = "/" + u.Path
	}
	if len(query) > 0 {
		u.RawQuery = query.Encode()
	}

	var body io.Reader
	if in != nil {
		b, err := json.Marshal(in)
		if err != nil {
			return pkgerrors.Wrap(err, "marshal request JSON")
		}
		body = bytes.NewReader(b)
	}

	req, err := http.NewRequestWithContext(ctx, method, u.String(), body)
	if err != nil {
		return pkgerrors.Wrap(err, "new request")
	}
	if in != nil {
		req.Header.Set("Content-Type", "application/json")
	}

	resp, err := c.http.Do(req)
	if err != nil {
		return pkgerrors.Wrap(err, "http request failed")
	}
	defer func() { _ = resp.Body.Close() }()

	respBody, err := io.ReadAll(io.LimitReader(resp.Body, 10<<20))
	if err != nil {
		return pkgerrors.Wrap(err, "read response body")
	}

	if resp.StatusCode >= 400 {
		var env apiErrorEnvelope
		if err := json.Unmarshal(respBody, &env); err == nil && env.Error.Message != "" {
			return &APIError{
				Status: resp.StatusCode,
				Code:   env.Error.Code,
				Msg:    env.Error.Message,
				Body:   respBody,
			}
		}
		return &APIError{
			Status: resp.StatusCode,
			Msg:    strings.TrimSpace(string(respBody)),
			Body:   respBody,
		}
	}

	if out == nil {
		return nil
	}
	if len(respBody) == 0 {
		return pkgerrors.New("empty response body")
	}
	if err := json.Unmarshal(respBody, out); err != nil {
		return pkgerrors.Wrap(err, "unmarshal response JSON")
	}

	return nil
}
