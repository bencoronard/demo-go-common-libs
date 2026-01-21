package dto

import "math"

type Sort struct {
	Property  string
	Direction Direction
}

type Direction string

const (
	ASC  Direction = "ASC"
	DESC Direction = "DESC"
)

type Pageable struct {
	Page int
	Size int
	Sort []Sort
}

func NewPageable(page, size int) Pageable {
	if page < 0 {
		page = 0
	}

	if size <= 0 {
		size = 20
	}

	return Pageable{
		Page: page,
		Size: size,
		Sort: nil,
	}
}

func (p Pageable) WithSort(prop string, dir Direction) Pageable {
	var newSort []Sort
	if len(p.Sort) == 0 {
		newSort = []Sort{{Property: prop, Direction: dir}}
	} else {
		newSort = make([]Sort, len(p.Sort)+1)
		copy(newSort, p.Sort)
		newSort[len(p.Sort)] = Sort{Property: prop, Direction: dir}
	}

	return Pageable{
		Page: p.Page,
		Size: p.Size,
		Sort: newSort,
	}
}

func (p Pageable) Offset() int {
	return p.Page * p.Size
}

func (p Pageable) Limit() int {
	return p.Size
}

type Slice[T any] struct {
	Content          []T
	Pageable         Pageable
	NumberOfElements int
	HasNext          bool
	HasPrev          bool
	IsFirst          bool
	IsLast           bool
}

func NewSlice[T any](content []T, pageable Pageable, totalFetched int) Slice[T] {
	hasNext := totalFetched > pageable.Size

	actualContent := content
	if hasNext && len(content) > pageable.Size {
		actualContent = content[:pageable.Size]
	}

	return Slice[T]{
		Content:          actualContent,
		Pageable:         pageable,
		NumberOfElements: len(actualContent),
		HasNext:          hasNext,
		HasPrev:          pageable.Page > 0,
		IsFirst:          pageable.Page == 0,
		IsLast:           !hasNext,
	}
}

func (s Slice[T]) NextPageable() (Pageable, bool) {
	if !s.HasNext {
		return Pageable{}, false
	}
	return NewPageable(s.Pageable.Page+1, s.Pageable.Size), true
}

func (s Slice[T]) PreviousPageable() (Pageable, bool) {
	if !s.HasPrev {
		return Pageable{}, false
	}
	return NewPageable(s.Pageable.Page-1, s.Pageable.Size), true
}

func Map[T any, U any](s Slice[T], fn func(T) U) Slice[U] {
	if len(s.Content) == 0 {
		return Slice[U]{
			Content:          nil,
			Pageable:         s.Pageable,
			NumberOfElements: 0,
			HasNext:          s.HasNext,
			HasPrev:          s.HasPrev,
			IsFirst:          s.IsFirst,
			IsLast:           s.IsLast,
		}
	}

	mapped := make([]U, len(s.Content))
	for i, item := range s.Content {
		mapped[i] = fn(item)
	}

	return Slice[U]{
		Content:          mapped,
		Pageable:         s.Pageable,
		NumberOfElements: len(mapped),
		HasNext:          s.HasNext,
		HasPrev:          s.HasPrev,
		IsFirst:          s.IsFirst,
		IsLast:           s.IsLast,
	}
}

type Page[T any] struct {
	Content          []T
	Pageable         Pageable
	NumberOfElements int
	HasNext          bool
	HasPrev          bool
	IsFirst          bool
	IsLast           bool
	TotalElements    int
	TotalPages       int
}

func NewPage[T any](content []T, pageable Pageable, totalElements int) Page[T] {
	totalPages := 0

	if pageable.Size > 0 {
		totalPages = int(math.Ceil(float64(totalElements) / float64(pageable.Size)))
	}

	return Page[T]{
		Content:          content,
		Pageable:         pageable,
		NumberOfElements: len(content),
		HasNext:          pageable.Page < totalPages-1,
		HasPrev:          pageable.Page > 0,
		IsFirst:          pageable.Page == 0,
		IsLast:           totalPages == 0 || pageable.Page >= totalPages-1,
		TotalElements:    totalElements,
		TotalPages:       totalPages,
	}
}

func (p Page[T]) NextPageable() (Pageable, bool) {
	if !p.HasNext {
		return Pageable{}, false
	}
	return NewPageable(p.Pageable.Page+1, p.Pageable.Size), true
}

func (p Page[T]) PreviousPageable() (Pageable, bool) {
	if !p.HasPrev {
		return Pageable{}, false
	}
	return NewPageable(p.Pageable.Page-1, p.Pageable.Size), true
}

func MapPage[T any, U any](p Page[T], fn func(T) U) Page[U] {
	if len(p.Content) == 0 {
		return Page[U]{
			Content:          nil,
			Pageable:         p.Pageable,
			NumberOfElements: 0,
			HasNext:          p.HasNext,
			HasPrev:          p.HasPrev,
			IsFirst:          p.IsFirst,
			IsLast:           p.IsLast,
			TotalElements:    p.TotalElements,
			TotalPages:       p.TotalPages,
		}
	}

	mapped := make([]U, len(p.Content))
	for i, item := range p.Content {
		mapped[i] = fn(item)
	}

	return Page[U]{
		Content:          mapped,
		Pageable:         p.Pageable,
		NumberOfElements: len(mapped),
		HasNext:          p.HasNext,
		HasPrev:          p.HasPrev,
		IsFirst:          p.IsFirst,
		IsLast:           p.IsLast,
		TotalElements:    p.TotalElements,
		TotalPages:       p.TotalPages,
	}
}
