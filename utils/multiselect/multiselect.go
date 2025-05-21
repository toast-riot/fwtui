package multiselect

import (
	"fwtui/utils/set"
)

type MultiSelectableList[T any] struct {
	Items    []T
	Focused  int          // Index of focused item
	Selected set.Set[int] // Set of selected item indices
}

func FromList[T any](items []T) MultiSelectableList[T] {
	return MultiSelectableList[T]{
		Items:    items,
		Focused:  0,
		Selected: set.NewSet[int](),
	}
}

func (s *MultiSelectableList[T]) Next() {
	if len(s.Items) == 0 {
		return
	}
	s.Focused = (s.Focused + 1) % len(s.Items)
}

func (s *MultiSelectableList[T]) Prev() {
	if len(s.Items) == 0 {
		return
	}
	s.Focused = (s.Focused - 1 + len(s.Items)) % len(s.Items)
}

func (s *MultiSelectableList[T]) Toggle() {
	s.Selected.Toggle(s.Focused)
}

func (s *MultiSelectableList[T]) ClearSelection() {
	s.Selected = set.NewSet[int]()
}

func (s *MultiSelectableList[T]) IsSelected(i int) bool {
	return s.Selected.Has(i)
}

func (s *MultiSelectableList[T]) NoneSelected() bool {
	return s.Selected.IsEmpty()
}

func (s *MultiSelectableList[T]) FocusedItem() T {
	return s.Items[s.Focused]
}

func (s *MultiSelectableList[T]) FocusedIndex() int {
	return s.Focused
}
func (s *MultiSelectableList[T]) SetItems(items []T) {
	if s.Focused >= len(items) {
		s.Focused = len(items) - 1
	}
	s.Items = items
	s.Selected = set.NewSet[int]()
}

func (s *MultiSelectableList[T]) GetSelectedItems() []T {
	selectedItems := make([]T, 0, len(s.Selected))
	s.ForEach(func(item T, index int, isFocused, isSelected bool) {
		if isSelected {
			selectedItems = append(selectedItems, item)
		}
	})
	return selectedItems
}

func (s *MultiSelectableList[T]) GetSelectedIndexes() []int {
	return s.Selected.ToSlice()
}

func (s *MultiSelectableList[T]) FocusFirst() {
	s.Focused = 0
}

func (s *MultiSelectableList[T]) ForEach(f func(item T, index int, isFocused, isSelected bool)) {
	for i, item := range s.Items {
		f(item, i, i == s.Focused, s.Selected.Has(i))
	}
}
