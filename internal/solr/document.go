package solr

import (
	"iter"
	"sort"
)

type Field struct {
	Key   string
	Value interface{}
}

type Document struct {
	ID     string
	Fields []Field
}

type DocSet map[string]Document

func (d *DocSet) Add(doc Document) {
	(*d)[doc.ID] = doc
}

func (d *DocSet) Iter() iter.Seq[Document] {
	docs := make([]Document, 0, len(*d))
	for _, doc := range *d {
		docs = append(docs, doc)
	}
	sort.Slice(docs, func(i, j int) bool {
		return docs[i].ID < docs[j].ID
	})

	return func(yield func(Document) bool) {
		for _, doc := range docs {
			if !yield(doc) {
				return
			}
		}
	}
}

type MergedDoc struct {
	Left  *Document
	Right *Document
}

type MergedDocSetIterator struct {
	left  iter.Seq[Document]
	right iter.Seq[Document]
}

func NewMergedDocSetIterator(left, right iter.Seq[Document]) *MergedDocSetIterator {
	return &MergedDocSetIterator{
		left:  left,
		right: right,
	}
}

func (m *MergedDocSetIterator) Iter() iter.Seq[MergedDoc] {
	next1, stop1 := iter.Pull(m.left)
	next2, stop2 := iter.Pull(m.right)

	return func(yield func(MergedDoc) bool) {
		defer stop1()
		defer stop2()

		var cursor1, cursor2 iterHolder[Document]
		for {
			cursor1.next(next1)
			cursor2.next(next2)
			if cursor1.stop && cursor2.stop {
				return
			}

			// only cursor1
			if !cursor1.stop && cursor2.stop {
				doc := *cursor1.value
				cursor1.reset()

				if !yield(MergedDoc{Left: &doc}) {
					return
				}
				continue
			}

			// only cursor2
			if cursor1.stop && !cursor2.stop {
				doc := *cursor2.value
				cursor2.reset()
				if !yield(MergedDoc{Right: &doc}) {
					return
				}
				continue
			}

			doc1 := *cursor1.value
			doc2 := *cursor2.value
			if doc1.ID == doc2.ID {
				cursor1.reset()
				cursor2.reset()
				if !yield(MergedDoc{Left: &doc1, Right: &doc2}) {
					return
				}
				continue
			}
			if doc1.ID < doc2.ID {
				cursor1.reset()
				if !yield(MergedDoc{Left: &doc1}) {
					return
				}
			} else {
				cursor2.reset()
				if !yield(MergedDoc{Right: &doc2}) {
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
