package dto

import (
	"bytes"
	"encoding/json"
)

type ProblemDetail struct {
	Type       string         `json:"type,omitempty"`
	Title      string         `json:"title,omitempty"`
	Status     int            `json:"status"`
	Detail     string         `json:"detail,omitempty"`
	Instance   string         `json:"instance,omitempty"`
	Properties map[string]any `json:"-"`
}

func (p *ProblemDetail) MarshalJSON() ([]byte, error) {
	type Alias ProblemDetail
	alias := (*Alias)(p)

	if len(p.Properties) == 0 {
		return json.Marshal(alias)
	}

	baseBytes, err := json.Marshal(alias)
	if err != nil {
		return nil, err
	}

	baseBytes = baseBytes[:len(baseBytes)-1]

	var buf bytes.Buffer
	buf.Grow(len(baseBytes) + 64)
	buf.Write(baseBytes)

	for k, v := range p.Properties {
		buf.WriteByte(',')

		keyBytes, err := json.Marshal(k)
		if err != nil {
			return nil, err
		}
		valBytes, err := json.Marshal(v)
		if err != nil {
			return nil, err
		}

		buf.Write(keyBytes)
		buf.WriteByte(':')
		buf.Write(valBytes)
	}

	buf.WriteByte('}')

	return buf.Bytes(), nil
}

func (p *ProblemDetail) UnmarshalJSON(data []byte) error {
	type Alias ProblemDetail

	var raw map[string]json.RawMessage
	if err := json.Unmarshal(data, &raw); err != nil {
		return err
	}

	var alias Alias
	if err := json.Unmarshal(data, &alias); err != nil {
		return err
	}

	*p = ProblemDetail(alias)

	for _, field := range []string{"type", "title", "status", "detail", "instance"} {
		delete(raw, field)
	}

	if len(raw) > 0 {
		p.Properties = make(map[string]any, len(raw))
		for k, v := range raw {
			var decoded any
			if err := json.Unmarshal(v, &decoded); err != nil {
				return err
			}
			p.Properties[k] = decoded
		}
	}

	return nil
}
