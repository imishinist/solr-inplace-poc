package solr

import (
	"strings"

	"github.com/imishinist/solr-inplace-poc/internal/myiter"
)

type UpdateBatchBuilder struct {
	// configuration
	fields              []string
	inPlaceUpdateFields []string

	// states
	OldDocuments    DocSet // old documents for in-place update
	Documents       DocSet
	DeleteDocuments DocSet
}

func NewUpdateBatchBuilder(fields []string, inPlaceUpdateFields []string) *UpdateBatchBuilder {
	return &UpdateBatchBuilder{
		fields:              fields,
		inPlaceUpdateFields: inPlaceUpdateFields,
		OldDocuments:        make(DocSet),
		Documents:           make(DocSet),
		DeleteDocuments:     make(DocSet),
	}
}

func (u *UpdateBatchBuilder) Add(docs ...Document) {
	for _, doc := range docs {
		u.Documents.Add(doc)
	}
}

func (u *UpdateBatchBuilder) AddOld(docs ...Document) {
	for _, doc := range docs {
		u.OldDocuments.Add(doc)
	}
}

func (u *UpdateBatchBuilder) Update(newDoc, oldDoc Document) {
	u.OldDocuments.Add(oldDoc)
	u.Documents.Add(newDoc)
}

func (u *UpdateBatchBuilder) Delete(docs ...Document) {
	for _, doc := range docs {
		u.DeleteDocuments.Add(doc)
	}
}

func (u *UpdateBatchBuilder) Build() (string, error) {
	var (
		builder queryBuilder
		first   = true
	)
	builder.WriteString("{")

	mergedIter := NewMergedDocSetIterator(u.OldDocuments.Iter(), u.Documents.Iter())
	for merged := range mergedIter.Iter() {
		if !u.hasUpdates(merged.Left, merged.Right) {
			continue
		}

		// write
		if !first {
			builder.WriteString(",")
		}
		first = false

		builder.WriteString(`"add":{"doc":`)
		if err := u.encodeDoc(&builder, merged); err != nil {
			return "", err
		}
		builder.WriteString("}")
	}

	if len(u.DeleteDocuments) > 0 {
		if !first {
			builder.WriteString(",")
		}

		builder.WriteString(`"delete":`)
		if err := u.encodeDelete(&builder); err != nil {
			return "", err
		}
	}
	builder.WriteString("}")

	if err := builder.Error(); err != nil {
		return "", err
	}
	return builder.String(), nil
}

// MergedDoc:
//
//	Left: old document
//	right: new document
func (u *UpdateBatchBuilder) encodeDoc(builder *queryBuilder, merged myiter.Merged[Document]) error {
	// only new document
	if merged.Left == nil || (merged.Left != nil && !u.canInPlaceUpdate(*merged.Left, *merged.Right)) {
		encoded, err := JSONEncode(merged.Right, u.fields)
		if err != nil {
			return err
		}
		builder.WriteString(encoded)
		return nil
	}

	// in-place update
	doc1 := *merged.Left
	doc2 := *merged.Right
	mergedIter := myiter.NewMergedIterator(doc1.Fields.Iter(), doc2.Fields.Iter(), FieldCompare)

	mergedFields := make(Fields, 0)
	for field := range mergedIter.Iter() {
		if (*field.Left).Value == (*field.Right).Value {
			continue
		}
		mergedFields = append(mergedFields, *field.Right)
	}
	encoded, err := InPlaceUpdateEncode(&Document{
		ID:     doc1.ID,
		Fields: mergedFields,
	}, u.inPlaceUpdateFields)
	if err != nil {
		return err
	}
	builder.WriteString(encoded)
	return nil
}

func (u *UpdateBatchBuilder) hasUpdates(old, new *Document) bool {
	if old == nil || new == nil {
		return true
	}
	mi := myiter.NewMergedIterator(old.Fields.Iter(), new.Fields.Iter(), FieldCompare)
	for field := range mi.Iter() {
		if field.Left != nil && field.Right != nil {
			left := *field.Left
			right := *field.Right

			if left.Value != right.Value && contains(u.fields, left.Key) {
				return true
			}
		}
		if field.Left == nil && field.Right != nil {
			right := *field.Right
			if contains(u.fields, right.Key) {
				return true
			}
		}
		if field.Left != nil && field.Right == nil {
			left := *field.Left
			if contains(u.fields, left.Key) {
				return true
			}
		}
	}
	return false
}

func (u *UpdateBatchBuilder) canInPlaceUpdate(old, new Document) bool {
	mi := myiter.NewMergedIterator(old.Fields.Iter(), new.Fields.Iter(), FieldCompare)
	for field := range mi.Iter() {
		left := field.Left
		right := field.Right

		if left != nil && right != nil {
			if (*left).Value != (*right).Value {
				if !contains(u.inPlaceUpdateFields, (*left).Key) {
					return false
				}
			}
		}
		if left == nil && right != nil {
			if !contains(u.inPlaceUpdateFields, (*right).Key) {
				return false
			}
		}
		// removed field
		if left != nil && right == nil {
			return false
		}
	}
	return true
}

func (u *UpdateBatchBuilder) encodeDelete(builder *queryBuilder) error {
	first := true
	builder.WriteString("[")
	for doc := range u.DeleteDocuments.Iter() {
		if !first {
			builder.WriteString(",")
		}
		first = false
		builder.WriteQuoteString(doc.ID, true)
	}
	builder.WriteString("]")
	return nil
}

func (u *UpdateBatchBuilder) Flush() {
	u.OldDocuments = make(DocSet)
	u.Documents = make(DocSet)
	u.DeleteDocuments = make(DocSet)
}

type queryBuilder struct {
	builder strings.Builder
	err     error
}

func (q *queryBuilder) WriteString(x string) {
	if q.err != nil {
		return
	}
	if _, err := q.builder.WriteString(x); err != nil {
		q.err = err
		return
	}
}

func (q *queryBuilder) WriteQuoteString(x string, quote bool) {
	if q.err != nil {
		return
	}

	if quote {
		q.WriteString("\"")
	}
	q.WriteString(x)
	if quote {
		q.WriteString("\"")
	}
}

func (q *queryBuilder) WriteKVString(key string, value string, quote bool) {
	if q.err != nil {
		return
	}

	q.WriteQuoteString(key, true)
	q.WriteString(":")
	q.WriteQuoteString(value, quote)
}

func (q *queryBuilder) String() string {
	if q.err != nil {
		return ""
	}
	return q.builder.String()
}

func (q *queryBuilder) Error() error {
	return q.err
}
