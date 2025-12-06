package dto

type Pageable struct {
	Limit  int
	Offset int
	SortBy string
}

type Slice[T any] struct {
	Content    []T
	TotalCount int
	HasNext    bool
}
