package solr

import (
	"errors"
	"fmt"
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
	return encode(doc, allowedFields, func(builder *queryBuilder, key, value string, quote bool) {
		// write: `"#{key}": #{value}`
		builder.WriteKVString(key, value, quote)
	})
}

func InPlaceUpdateEncode(doc *Document, allowedFields []string) (string, error) {
	return encode(doc, allowedFields, func(builder *queryBuilder, key, value string, quote bool) {
		// write: `"#{key}":{"set":#{value}}`
		builder.WriteQuoteString(key, true)
		builder.WriteString(`:{`)
		builder.WriteKVString("set", value, quote)
		builder.WriteString(`}`)
	})
}

func encode(doc *Document, allowedFields []string, yield func(builder *queryBuilder, key, value string, quote bool)) (string, error) {
	var builder queryBuilder

	builder.WriteString("{")

	// write ID
	builder.WriteKVString("id", doc.ID, true)
	for _, field := range doc.Fields {
		if allowedFields != nil && !contains(allowedFields, field.Key) {
			continue
		}

		builder.WriteString(",")

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
		yield(&builder, field.Key, value, quote)

	}
	builder.WriteString("}")

	if err := builder.Error(); err != nil {
		return "", err
	}
	return builder.String(), nil
}
