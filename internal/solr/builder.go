package solr

import "strings"

type UpdateBatchBuilder struct {
	// configuration
	fields              []string
	inPlaceUpdateFields []string

	// states
	Documents       []Document
	DeleteDocuments []Document
}

func NewUpdateBatchBuilder(fields []string, inPlaceUpdateFields []string) *UpdateBatchBuilder {
	return &UpdateBatchBuilder{
		fields:              fields,
		inPlaceUpdateFields: inPlaceUpdateFields,
		Documents:           make([]Document, 0),
		DeleteDocuments:     make([]Document, 0),
	}
}

func (u *UpdateBatchBuilder) Add(doc ...Document) {
	u.Documents = append(u.Documents, doc...)
}

func (u *UpdateBatchBuilder) Delete(doc ...Document) {
	u.DeleteDocuments = append(u.DeleteDocuments, doc...)
}

func (u *UpdateBatchBuilder) Build() (string, error) {
	var (
		builder strings.Builder
		first   = true
	)
	if _, err := builder.WriteString("{"); err != nil {
		return "", err
	}
	for _, doc := range u.Documents {
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
	for i, doc := range u.DeleteDocuments {
		if i != 0 {
			if _, err := builder.WriteString(","); err != nil {
				return err
			}
		}
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
	u.Documents = u.Documents[:0]
	u.DeleteDocuments = u.DeleteDocuments[:0]
}
