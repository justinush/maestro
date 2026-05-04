package stub

import (
	"bytes"
	"encoding/json"
	"fmt"
	"io"
)

type Params struct {
	Set map[string]json.RawMessage `json:"set,omitempty"`
}

// DecodeParams parses params with DisallowUnknownFields and rejects trailing JSON after the object.
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
			return p, nil
		}
		return Params{}, err
	}
	return Params{}, fmt.Errorf("trailing JSON after params object")
}
