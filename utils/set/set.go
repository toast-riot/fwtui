package set

type Set[T comparable] map[T]struct{}

func NewSet[T comparable]() Set[T] {
	return make(Set[T])
}

func (s Set[T]) Add(value T) Set[T] {
	s[value] = struct{}{}
	return s
}

func (s Set[T]) Remove(value T) Set[T] {
	delete(s, value)
	return s
}

func (s Set[T]) Has(value T) bool {
	_, ok := s[value]
	return ok
}

func (s Set[T]) Toggle(value T) Set[T] {
	if s.Has(value) {
		s.Remove(value)
	} else {
		s.Add(value)
	}
	return s
}

func (s Set[T]) ToSlice() []T {
	result := make([]T, 0, len(s))
	for v := range s {
		result = append(result, v)
	}
	return result
}

func (s Set[T]) IsEmpty() bool {
	return len(s) == 0
}
