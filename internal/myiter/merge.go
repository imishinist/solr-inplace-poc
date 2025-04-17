package myiter

import (
	"iter"
)

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
	return func(yield func(Merged[T]) bool) {
		next, stop := iter.Pull(m.right)
		defer stop()

		v2, ok2 := next()
	Outer:
		for v1 := range m.left {
			for ok2 {
				ordering := m.compare(v1, v2)
				if ordering < 0 {
					break
				}

				v2_ := v2
				if ordering > 0 {
					if !yield(Merged[T]{Right: &v2_}) {
						return
					}
					v2, ok2 = next()
					continue
				}

				// v1 == v2
				if !yield(Merged[T]{Left: &v1, Right: &v2_}) {
					return
				}
				v2, ok2 = next()
				continue Outer
			}
			if !yield(Merged[T]{Left: &v1}) {
				return
			}
		}
		for ok2 {
			v2_ := v2
			if !yield(Merged[T]{Right: &v2_}) {
				return
			}
			v2, ok2 = next()
		}
	}
}
