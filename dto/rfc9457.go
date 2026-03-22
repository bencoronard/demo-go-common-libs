package dto

import (
	"maps"

	"github.com/bencoronard/demo-go-common-libs/utility"
)

type ProblemDetail map[string]any

func NewProblemDetail(status int) ProblemDetail {
	return ProblemDetail{
		"type":   "about:blank",
		"status": status,
	}
}

func (p ProblemDetail) WithStatus(s int) ProblemDetail {
	pd := p.clone()
	pd["status"] = s
	return pd
}

func (p ProblemDetail) WithType(t string) ProblemDetail {
	if t == "" {
		return p
	}
	pd := p.clone()
	pd["type"] = t
	return pd
}

func (p ProblemDetail) WithTitle(t string) ProblemDetail {
	if t == "" {
		return p
	}
	pd := p.clone()
	pd["title"] = t
	return pd
}

func (p ProblemDetail) WithDetail(d string) ProblemDetail {
	if d == "" {
		return p
	}
	pd := p.clone()
	pd["detail"] = d
	return pd
}

func (p ProblemDetail) WithInstance(i string) ProblemDetail {
	if i == "" {
		return p
	}
	pd := p.clone()
	pd["instance"] = i
	return pd
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
	pd := p.clone()
	pd[key] = value
	return pd
}

func (p ProblemDetail) Type() string {
	return utility.CastToTypeOrZero[string](p["type"])
}

func (p ProblemDetail) Title() string {
	return utility.CastToTypeOrZero[string](p["title"])
}

func (p ProblemDetail) Status() int {
	return utility.CastToTypeOrZero[int](p["status"])
}

func (p ProblemDetail) Detail() string {
	return utility.CastToTypeOrZero[string](p["detail"])
}

func (p ProblemDetail) Instance() string {
	return utility.CastToTypeOrZero[string](p["instance"])
}

func (p ProblemDetail) Get(key string) any {
	return p[key]
}

func (p ProblemDetail) clone() ProblemDetail {
	n := make(ProblemDetail, len(p))
	maps.Copy(n, p)
	return n
}
