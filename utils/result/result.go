package result

type Result[T any] struct {
	ok  *T
	err error
}

func Ok[T any](value T) Result[T] {
	return Result[T]{ok: &value}
}

func Err[T any](err error) Result[T] {
	return Result[T]{err: err}
}

func (r Result[T]) IsOk() bool {
	return r.err == nil
}

func (r Result[T]) IsErr() bool {
	return r.err != nil
}

func (r Result[T]) Value() T {
	if r.IsErr() {
		panic("called Value on an Err value")
	}
	return *r.ok
}

func (r Result[T]) WithDefault(defaultVal T) T {
	if r.IsErr() {
		return defaultVal
	}
	return *r.ok
}

func (r Result[T]) Err() error {
	return r.err
}
