package selectable_list

type SelectableList[T any] struct {
	Items   []T
	Current int // index of selected item
}

// NewSelectableList creates a new selectable list, defaulting to the first item.
func NewSelectableList[T any](items []T) *SelectableList[T] {
	return &SelectableList[T]{
		Items:   items,
		Current: 0,
	}
}

// Next moves selection forward (wraps around).
func (s *SelectableList[T]) Next() {
	if len(s.Items) == 0 {
		return
	}
	s.Current = (s.Current + 1) % len(s.Items)
}

// Prev moves selection backward (wraps around).
func (s *SelectableList[T]) Prev() {
	if len(s.Items) == 0 {
		return
	}
	s.Current = (s.Current - 1 + len(s.Items)) % len(s.Items)
}

// Selected returns the currently selected item.
func (s *SelectableList[T]) Selected() T {
	return s.Items[s.Current]
}
