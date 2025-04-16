package myiter

import "iter"

type Merged[T any] struct {
	Left  *T
	Right *T
}

type MergedIterator[T any] struct {
	left  iter.Seq[T]
	right iter.Seq[T]

	// compare compares left and right
	// left == right => 0
	// left < right => -1
	// left > right => 1
	compare func(T, T) int
}

func NewMergedIterator[T any](left, right iter.Seq[T], compare func(T, T) int) *MergedIterator[T] {
	return &MergedIterator[T]{
		left:    left,
		right:   right,
		compare: compare,
	}
}

func (m *MergedIterator[T]) Iter() iter.Seq[Merged[T]] {
	next1, stop1 := iter.Pull(m.left)
	next2, stop2 := iter.Pull(m.right)

	return func(yield func(Merged[T]) bool) {
		defer stop1()
		defer stop2()

		var cursor1, cursor2 iterHolder[T]
		for {
			cursor1.next(next1)
			cursor2.next(next2)
			if cursor1.stop && cursor2.stop {
				return
			}

			// only cursor1
			if !cursor1.stop && cursor2.stop {
				val := *cursor1.value
				cursor1.reset()

				if !yield(Merged[T]{Left: &val}) {
					return
				}
				continue
			}

			// only cursor2
			if cursor1.stop && !cursor2.stop {
				val := *cursor2.value
				cursor2.reset()
				if !yield(Merged[T]{Right: &val}) {
					return
				}
				continue
			}

			val1 := *cursor1.value
			val2 := *cursor2.value
			cmp := m.compare(val1, val2)
			if cmp == 0 {
				cursor1.reset()
				cursor2.reset()
				if !yield(Merged[T]{Left: &val1, Right: &val2}) {
					return
				}
				continue
			}
			if cmp < 0 {
				cursor1.reset()
				if !yield(Merged[T]{Left: &val1}) {
					return
				}
			} else {
				cursor2.reset()
				if !yield(Merged[T]{Right: &val2}) {
					return
				}
			}
		}
	}
}

type iterHolder[T any] struct {
	value *T
	stop  bool
}

func (i *iterHolder[T]) next(yield func() (T, bool)) {
	if i.stop {
		return
	}
	if i.value != nil {
		return
	}

	value, ok := yield()
	i.stop = !ok
	if ok {
		i.value = &value
	}
}

func (i *iterHolder[T]) reset() {
	i.value = nil
}
