package engine

import (
	"bytes"
	"context"
	"fmt"
	"io"
	"net/http"
	"time"

	"github.com/justinush/maestro/internal/httpaction"
)

// Default HTTP response body cap per action (bytes).
const maxHTTPResponseBodyBytes = 1 << 20 // 1 MiB

type httpRunner struct {
	client *http.Client
}

// NewHTTPRunner returns an ActionRunner for workflow actions with type "http".
func NewHTTPRunner(client *http.Client) ActionRunner {
	if client == nil {
		client = http.DefaultClient
	}
	return &httpRunner{client: client}
}

func (r *httpRunner) Run(ctx ActionContext) error {
	if len(ctx.Action.Params) == 0 {
		return fmt.Errorf("http action requires params")
	}
	p, err := httpaction.DecodeParams(ctx.Action.Params)
	if err != nil {
		return err
	}

	var body io.Reader
	if len(p.Body) > 0 {
		body = bytes.NewReader(p.Body)
	}

	timeout := 30 * time.Second
	if p.TimeoutSeconds != nil {
		timeout = time.Duration(*p.TimeoutSeconds) * time.Second
	}
	cctx, cancel := context.WithTimeout(context.Background(), timeout)
	defer cancel()

	req, err := http.NewRequestWithContext(cctx, p.Method, p.URL, body)
	if err != nil {
		return err
	}
	for k, v := range p.Headers {
		req.Header.Set(k, v)
	}
	if len(p.Body) > 0 && req.Header.Get("Content-Type") == "" {
		req.Header.Set("Content-Type", "application/json")
	}

	resp, err := r.client.Do(req)
	if err != nil {
		return err
	}
	defer func() { _ = resp.Body.Close() }()

	raw, err := io.ReadAll(io.LimitReader(resp.Body, maxHTTPResponseBodyBytes+1))
	if err != nil {
		return err
	}
	if len(raw) > maxHTTPResponseBodyBytes {
		return fmt.Errorf("http response body exceeds limit (%d bytes)", maxHTTPResponseBodyBytes)
	}

	hdr := make(map[string]string, len(resp.Header))
	for k, vs := range resp.Header {
		if len(vs) > 0 {
			hdr[k] = vs[0]
		}
	}

	ctx.Variables[p.ResultVariable] = map[string]any{
		"statusCode": resp.StatusCode,
		"headers":    hdr,
		"body":       string(raw),
	}
	return nil
}
