package httpaction

import (
	"bytes"
	"encoding/json"
	"fmt"
	"io"
	"net/http"
	"strings"
)

type Params struct {
	Method         string            `json:"method,omitempty"`
	URL            string            `json:"url"`
	Headers        map[string]string `json:"headers,omitempty"`
	Body           json.RawMessage   `json:"body,omitempty"`
	TimeoutSeconds *int              `json:"timeoutSeconds,omitempty"`
	ResultVariable string            `json:"resultVariable"`
}

// DecodeParams parses params with DisallowUnknownFields and rejects trailing JSON.
func DecodeParams(data []byte) (Params, error) {
	dec := json.NewDecoder(bytes.NewReader(data))
	dec.DisallowUnknownFields()
	var p Params
	if err := dec.Decode(&p); err != nil {
		return Params{}, err
	}
	var tail json.RawMessage
	if err := dec.Decode(&tail); err != nil {
		if err == io.EOF {
			return normalize(p)
		}
		return Params{}, err
	}
	return Params{}, fmt.Errorf("trailing JSON after params object")
}

func normalize(p Params) (Params, error) {
	if strings.TrimSpace(p.URL) == "" {
		return Params{}, fmt.Errorf("url is required")
	}
	if strings.TrimSpace(p.ResultVariable) == "" {
		return Params{}, fmt.Errorf("resultVariable is required")
	}
	if p.Method == "" {
		p.Method = http.MethodGet
	}
	p.Method = strings.ToUpper(strings.TrimSpace(p.Method))

	if p.TimeoutSeconds != nil {
		if *p.TimeoutSeconds <= 0 {
			return Params{}, fmt.Errorf("timeoutSeconds must be positive")
		}
		if *p.TimeoutSeconds > 300 {
			return Params{}, fmt.Errorf("timeoutSeconds exceeds max (300)")
		}
	}
	return p, nil
}
