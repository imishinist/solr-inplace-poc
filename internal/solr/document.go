package solr

import (
	"iter"
	"maps"
	"slices"
	"sort"

	"github.com/imishinist/solr-inplace-poc/internal/myiter"
)

type Field struct {
	Key   string
	Value interface{}
}

type Fields []Field

func (f *Fields) Iter() iter.Seq[Field] {
	sort.Slice(*f, func(i, j int) bool {
		return (*f)[i].Key < (*f)[j].Key
	})
	return func(yield func(Field) bool) {
		for _, field := range *f {
			if !yield(field) {
				return
			}
		}
	}
}

func FieldCompare(f1 Field, f2 Field) int {
	if f1.Key == f2.Key {
		return 0
	}
	if f1.Key < f2.Key {
		return -1
	}
	return 1
}

type Document struct {
	ID     string
	Fields Fields
}

type DocSet map[string]Document

func (d *DocSet) Add(doc Document) {
	(*d)[doc.ID] = doc
}

func (d *DocSet) Iter() iter.Seq[Document] {
	return func(yield func(Document) bool) {
		for _, k := range slices.Sorted(maps.Keys(*d)) {
			if !yield((*d)[k]) {
				return
			}
		}
	}
}

func DocumentCompare(d1, d2 Document) int {
	if d1.ID == d2.ID {
		return 0
	}
	if d1.ID < d2.ID {
		return -1
	}
	return 1
}

type MergedDoc myiter.Merged[Document]

type MergedDocSetIterator struct {
	inner *myiter.MergedIterator[Document]
}

func NewMergedDocSetIterator(left, right iter.Seq[Document]) *MergedDocSetIterator {
	inner := myiter.NewMergedIterator(left, right, DocumentCompare)
	return &MergedDocSetIterator{
		inner: inner,
	}
}

func (m *MergedDocSetIterator) Iter() iter.Seq[myiter.Merged[Document]] {
	return m.inner.Iter()
}
