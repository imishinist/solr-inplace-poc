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
