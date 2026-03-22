package dto

import "net/http"

type ProblemDetail map[string]any

func NewProblemDetail(status int) ProblemDetail {
	return ProblemDetail{
		"type":   "about:blank",
		"status": status,
		"title":  http.StatusText(status),
	}
}

func (p ProblemDetail) WithStatus(s int) ProblemDetail {
	p["status"] = s
	return p
}

func (p ProblemDetail) WithType(t string) ProblemDetail {
	if t == "" {
		return p
	}
	p["type"] = t
	return p
}

func (p ProblemDetail) WithTitle(t string) ProblemDetail {
	if t == "" {
		return p
	}
	p["title"] = t
	return p
}

func (p ProblemDetail) WithDetail(d string) ProblemDetail {
	if d == "" {
		return p
	}
	p["detail"] = d
	return p
}

func (p ProblemDetail) WithInstance(i string) ProblemDetail {
	if i == "" {
		return p
	}
	p["instance"] = i
	return p
}

func (p ProblemDetail) With(key string, value any) ProblemDetail {
	if key == "" {
		return p
	}
	if value == nil {
		return p
	}
	if s, ok := value.(string); ok && s == "" {
		return p
	}
	p[key] = value
	return p
}

func (p ProblemDetail) Type() string {
	return stringVal(p["type"])
}

func (p ProblemDetail) Title() string {
	return stringVal(p["title"])
}

func (p ProblemDetail) Status() int {
	if v, ok := p["status"].(int); ok {
		return v
	}
	return 0
}

func (p ProblemDetail) Detail() string {
	return stringVal(p["detail"])
}

func (p ProblemDetail) Instance() string {
	return stringVal(p["instance"])
}

func (p ProblemDetail) Get(key string) any {
	return p[key]
}

func stringVal(v any) string {
	if s, ok := v.(string); ok {
		return s
	}
	return ""
}
