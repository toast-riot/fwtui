package focusablelist

type SelectableList[T comparable] struct {
	Items   []T
	Current int // index of selected item
}

// FromList creates a new selectable list, defaulting to the first item.
func FromList[T comparable](items []T) *SelectableList[T] {
	return &SelectableList[T]{
		Items:   items,
		Current: 0,
	}
}

func (s *SelectableList[T]) Focus(item T) *SelectableList[T] {
	for i, v := range s.Items {
		if v == item {
			s.Current = i
			return s
		}
	}
	return s
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

// Focused returns the currently selected item.
func (s *SelectableList[T]) Focused() T {
	return s.Items[s.Current]
}

func (s *SelectableList[T]) ForEach(f func(item T, index int, isSelected bool)) {
	for i, item := range s.Items {
		f(item, i, i == s.Current)
	}
}

func (s *SelectableList[T]) FocusFirst() {
	s.Current = 0
}
func (s *SelectableList[T]) GetItems() []T {
	return s.Items
}
func (s *SelectableList[T]) SetItems(items []T) {
	if s.Current >= len(items) {
		s.Current = len(items) - 1
	}
	s.Items = items
}
