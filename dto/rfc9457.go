package dto

import (
	"encoding/json"
	"net/http"
)

type ProblemDetail struct {
	Type       string         `json:"type,omitempty"`
	Title      string         `json:"title,omitempty"`
	Status     int            `json:"status"`
	Detail     string         `json:"detail,omitempty"`
	Instance   string         `json:"instance,omitempty"`
	Properties map[string]any `json:"-"`
}

func (p *ProblemDetail) SetType(t string) *ProblemDetail {
	p.Type = t
	return p
}

func (p *ProblemDetail) SetTitle(title string) *ProblemDetail {
	p.Title = title
	return p
}

func (p *ProblemDetail) SetStatus(status int) *ProblemDetail {
	p.Status = status
	return p
}

func (p *ProblemDetail) SetDetail(detail string) *ProblemDetail {
	p.Detail = detail
	return p
}

func (p *ProblemDetail) SetInstance(intance string) *ProblemDetail {
	p.Instance = intance
	return p
}

func (p *ProblemDetail) SetProperty(key string, value any) *ProblemDetail {
	if p.Properties == nil {
		p.Properties = make(map[string]any)
	}
	p.Properties[key] = value
	return p
}

func ForStatusAndDetail(status int, detail string) *ProblemDetail {
	return &ProblemDetail{
		Type:   "about:blank",
		Title:  http.StatusText(status),
		Status: status,
		Detail: detail,
	}
}

func (p ProblemDetail) MarshalJSON() ([]byte, error) {
	type Alias ProblemDetail

	base := make(map[string]any)

	b, err := json.Marshal(Alias(p))
	if err != nil {
		return nil, err
	}

	if err := json.Unmarshal(b, &base); err != nil {
		return nil, err
	}

	for k, v := range p.Properties {
		if _, exists := base[k]; !exists {
			base[k] = v
		}
	}

	return json.Marshal(base)
}

func (p *ProblemDetail) UnmarshalJSON(data []byte) error {
	type Alias ProblemDetail

	var raw map[string]any
	if err := json.Unmarshal(data, &raw); err != nil {
		return err
	}

	var known Alias
	if err := json.Unmarshal(data, &known); err != nil {
		return err
	}

	*p = ProblemDetail(known)

	if p.Properties == nil {
		p.Properties = make(map[string]any)
	}

	for k, v := range raw {
		switch k {
		case "type", "title", "status", "detail", "instance":
		default:
			p.Properties[k] = v
		}
	}

	return nil
}
