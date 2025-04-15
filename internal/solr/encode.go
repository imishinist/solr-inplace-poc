package solr

import (
	"errors"
	"fmt"
	"strings"
)

func contains(slice []string, str string) bool {
	for _, s := range slice {
		if s == str {
			return true
		}
	}
	return false
}

// JSONEncode encodes document with json format.
// allowedFields is a fields slice that allowed encoding
func JSONEncode(doc *Document, allowedFields []string) (string, error) {
	var builder strings.Builder
	var err error

	_, err = builder.WriteString("{")
	if err != nil {
		return "", err
	}

	// write ID
	if err = writeField(&builder, "id", doc.ID, true); err != nil {
		return "", err
	}
	for _, field := range doc.Fields {
		if allowedFields != nil && !contains(allowedFields, field.Key) {
			continue
		}

		_, err = builder.WriteString(",")
		if err != nil {
			return "", err
		}
		var value string
		var quote bool
		switch v := field.Value.(type) {
		case int, int8, int16, int32, int64:
			value = fmt.Sprintf("%d", v)
			quote = false
		case uint, uint8, uint16, uint32, uint64:
			value = fmt.Sprintf("%ud", v)
			quote = false
		case float32, float64:
			value = fmt.Sprintf("%g", v)
			quote = false
		case string:
			value = v
			quote = true
		default:
			return "", errors.New("unsupported field type")
		}

		if err = writeField(&builder, field.Key, value, quote); err != nil {
			return "", err
		}
	}

	builder.WriteString("}")
	return builder.String(), nil
}

func writeString(builder *strings.Builder, value string, quote bool) error {
	var err error
	if quote {
		if _, err = builder.WriteString("\""); err != nil {
			return err
		}
	}

	if _, err = builder.WriteString(value); err != nil {
		return err
	}

	if quote {
		if _, err = builder.WriteString("\""); err != nil {
			return err
		}
	}
	return nil
}

func writeField(builder *strings.Builder, key, value string, quote bool) error {
	var err error
	if err = writeString(builder, key, true); err != nil {
		return err
	}

	if _, err = builder.WriteString(":"); err != nil {
		return err
	}

	if err = writeString(builder, value, quote); err != nil {
		return err
	}
	return nil
}
