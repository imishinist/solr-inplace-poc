package solr

import (
	"strings"
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
	for doc := range u.Documents.Iter() {
		// write
		if !first {
			builder.WriteString(",")
		}
		first = false

		builder.WriteString(`"add":{"doc":`)
		if err := u.encodeDoc(&builder, doc); err != nil {
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

func (u *UpdateBatchBuilder) encodeDoc(builder *queryBuilder, doc Document) error {
	// encode
	encoded, err := JSONEncode(&doc, u.fields)
	if err != nil {
		return err
	}
	builder.WriteString(encoded)
	return nil
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
