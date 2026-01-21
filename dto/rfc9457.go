package dto

import (
	"encoding/json"
	"maps"
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

func (p ProblemDetail) WithType(t string) ProblemDetail {
	p.Type = t
	return p
}

func (p ProblemDetail) WithTitle(title string) ProblemDetail {
	p.Title = title
	return p
}

func (p ProblemDetail) WithStatus(status int) ProblemDetail {
	p.Status = status
	return p
}

func (p ProblemDetail) WithDetail(detail string) ProblemDetail {
	p.Detail = detail
	return p
}

func (p ProblemDetail) WithInstance(intance string) ProblemDetail {
	p.Instance = intance
	return p
}

func (p ProblemDetail) WithProperty(key string, value any) ProblemDetail {
	if p.Properties == nil {
		p.Properties = make(map[string]any, 1)
	} else {
		newProps := make(map[string]any, len(p.Properties)+1)
		maps.Copy(newProps, p.Properties)
		p.Properties = newProps
	}

	p.Properties[key] = value

	return p
}

func ForStatusAndDetail(status int, detail string) ProblemDetail {
	return ProblemDetail{
		Type:   "about:blank",
		Title:  http.StatusText(status),
		Status: status,
		Detail: detail,
	}
}

func (p ProblemDetail) MarshalJSON() ([]byte, error) {
	type Alias ProblemDetail

	if len(p.Properties) == 0 {
		return json.Marshal(Alias(p))
	}

	base := make(map[string]any, 5+len(p.Properties))

	baseBytes, err := json.Marshal(Alias(p))
	if err != nil {
		return nil, err
	}

	if err := json.Unmarshal(baseBytes, &base); err != nil {
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

	customProps := make(map[string]any)
	for k, v := range raw {
		switch k {
		case "type", "title", "status", "detail", "instance":
		default:
			customProps[k] = v
		}
	}

	if len(customProps) > 0 {
		p.Properties = customProps
	}

	return nil
}
