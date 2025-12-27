package dto

import "math"

type Pageable struct {
	Page int
	Size int
	Sort []Sort
}

type Sort struct {
	Property  string
	Direction Direction
}

type Direction string

const (
	ASC  Direction = "ASC"
	DESC Direction = "DESC"
)

func NewPageable(page, size int) *Pageable {
	if page < 0 {
		page = 0
	}

	if size <= 0 {
		size = 20
	}

	return &Pageable{
		Page: page,
		Size: size,
		Sort: []Sort{},
	}
}

func (p *Pageable) WithSort(prop string, dir Direction) *Pageable {
	p.Sort = append(p.Sort, Sort{
		Property:  prop,
		Direction: dir,
	})
	return p
}

func (p *Pageable) GetOffset() int {
	return p.Page * p.Size
}

func (p *Pageable) GetLimit() int {
	return p.Size
}

type Slice[T any] struct {
	Content          []T
	Pageable         *Pageable
	Size             int
	NumberOfElements int
	HasNext          bool
	HasPrev          bool
	IsFirst          bool
	IsLast           bool
}

func NewSlice[T any](content []T, pageable *Pageable, totalFetched int) *Slice[T] {
	hasNext := totalFetched > pageable.Size

	actualContent := content
	if hasNext && len(content) > pageable.Size {
		actualContent = content[:pageable.Size]
	}

	return &Slice[T]{
		Content:          actualContent,
		Pageable:         pageable,
		Size:             pageable.Size,
		NumberOfElements: len(actualContent),
		HasNext:          hasNext,
		HasPrev:          pageable.Page > 0,
		IsFirst:          pageable.Page == 0,
		IsLast:           !hasNext,
	}
}

func (s *Slice[T]) NextPageable() *Pageable {
	if !s.HasNext {
		return nil
	}
	return NewPageable(s.Pageable.Page+1, s.Pageable.Size)
}

func (s *Slice[T]) PreviousPageable() *Pageable {
	if !s.HasPrev {
		return nil
	}
	return NewPageable(s.Pageable.Page-1, s.Pageable.Size)
}

func (s *Slice[T]) Map(fn func(T) T) *Slice[T] {
	mapped := make([]T, len(s.Content))
	for i, item := range s.Content {
		mapped[i] = fn(item)
	}

	return &Slice[T]{
		Content:          mapped,
		Pageable:         s.Pageable,
		Size:             s.Size,
		NumberOfElements: s.NumberOfElements,
		HasNext:          s.HasNext,
		HasPrev:          s.HasPrev,
		IsFirst:          s.IsFirst,
		IsLast:           s.IsLast,
	}
}

type Page[T any] struct {
	*Slice[T]
	TotalElements int
	TotalPages    int
}

func NewPage[T any](content []T, pageable *Pageable, totalElements int) *Page[T] {
	totalPages := int(math.Ceil(float64(totalElements) / float64(pageable.Size)))
	hasNext := pageable.Page < totalPages-1

	slice := &Slice[T]{
		Content:          content,
		Pageable:         pageable,
		Size:             pageable.Size,
		NumberOfElements: len(content),
		HasNext:          hasNext,
		HasPrev:          pageable.Page > 0,
		IsFirst:          pageable.Page == 0,
		IsLast:           pageable.Page >= totalPages-1,
	}

	return &Page[T]{
		Slice:         slice,
		TotalElements: totalElements,
		TotalPages:    totalPages,
	}
}
