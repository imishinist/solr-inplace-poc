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
		builder strings.Builder
		first   = true
	)
	if _, err := builder.WriteString("{"); err != nil {
		return "", err
	}
	for doc := range u.Documents.Iter() {
		var err error
		// write
		if !first {
			if _, err = builder.WriteString(","); err != nil {
				return "", err
			}
		}
		first = false

		if err = writeString(&builder, "add", true); err != nil {
			return "", err
		}
		if _, err = builder.WriteString(`:{"doc":`); err != nil {
			return "", err
		}
		if err = u.encodeDoc(&builder, doc); err != nil {
			return "", err
		}
		if _, err = builder.WriteString("}"); err != nil {
			return "", err
		}
	}

	if len(u.DeleteDocuments) > 0 {
		if !first {
			if _, err := builder.WriteString(","); err != nil {
				return "", err
			}
		}

		var delete strings.Builder
		if err := u.encodeDelete(&delete); err != nil {
			return "", err
		}
		if err := writeField(&builder, "delete", delete.String(), false); err != nil {
			return "", err
		}
	}

	if _, err := builder.WriteString("}"); err != nil {
		return "", err
	}

	return builder.String(), nil
}

func (u *UpdateBatchBuilder) encodeDoc(builder *strings.Builder, doc Document) error {
	// encode
	encoded, err := JSONEncode(&doc, u.fields)
	if err != nil {
		return err
	}

	if err := writeString(builder, encoded, false); err != nil {
		return err
	}

	return nil
}

func (u *UpdateBatchBuilder) encodeDelete(builder *strings.Builder) error {
	if _, err := builder.WriteString("["); err != nil {
		return err
	}
	first := true
	for doc := range u.DeleteDocuments.Iter() {
		if !first {
			if _, err := builder.WriteString(","); err != nil {
				return err
			}
		}
		first = false
		if err := writeString(builder, doc.ID, true); err != nil {
			return err
		}
	}
	if _, err := builder.WriteString("]"); err != nil {
		return err
	}
	return nil
}

func (u *UpdateBatchBuilder) Flush() {
	u.OldDocuments = make(DocSet)
	u.Documents = make(DocSet)
	u.DeleteDocuments = make(DocSet)
}
