package solr_test

import (
	"testing"

	"github.com/imishinist/solr-inplace-poc/internal/solr"
)

func TestJSONEncode(t *testing.T) {
	cases := []struct {
		name          string
		doc           solr.Document
		allowedFields []string
		expected      string
	}{
		{
			name: "simple document",
			doc: solr.Document{
				ID: "1",
				Fields: []solr.Field{
					{Key: "int_1", Value: 10},
					{Key: "str_1", Value: "foo"},
					{Key: "float_1", Value: 10.5},
				},
			},
			allowedFields: nil,
			expected:      `{"id":"1","int_1":10,"str_1":"foo","float_1":10.5}`,
		},
		{
			name: "allowed fields",
			doc: solr.Document{
				ID: "1",
				Fields: []solr.Field{
					{Key: "int_1", Value: 10},
					{Key: "str_1", Value: "foo"},
					{Key: "float_1", Value: 10.5},
					{Key: "int_2", Value: 20},
					{Key: "str_2", Value: "bar"},
					{Key: "float_2", Value: 10.1},
				},
			},
			allowedFields: []string{
				"int_1",
				"str_1",
				"float_1",
			},
			expected: `{"id":"1","int_1":10,"str_1":"foo","float_1":10.5}`,
		},
	}

	for _, c := range cases {
		t.Run(c.name, func(t *testing.T) {
			got, err := solr.JSONEncode(&c.doc, c.allowedFields)
			if err != nil {
				t.Fatal(err)
			}
			if got != c.expected {
				t.Fatalf("\nexpected: %v\n but got: %v", c.expected, got)
			}
		})
	}
}

func TestInPlaceUpdateEncode(t *testing.T) {
	cases := []struct {
		name          string
		doc           solr.Document
		allowedFields []string
		expected      string
	}{
		{
			name: "simple document",
			doc: solr.Document{
				ID: "1",
				Fields: []solr.Field{
					{Key: "int_1", Value: 10},
					{Key: "str_1", Value: "foo"},
					{Key: "float_1", Value: 10.5},
				},
			},
			allowedFields: nil,
			expected:      `{"id":"1","int_1":{"set":10},"str_1":{"set":"foo"},"float_1":{"set":10.5}}`,
		},
		{
			name: "allowed fields",
			doc: solr.Document{
				ID: "1",
				Fields: []solr.Field{
					{Key: "int_1", Value: 10},
					{Key: "str_1", Value: "foo"},
					{Key: "float_1", Value: 10.5},
					{Key: "int_2", Value: 20},
					{Key: "str_2", Value: "bar"},
					{Key: "float_2", Value: 10.1},
				},
			},
			allowedFields: []string{
				"int_1",
				"str_1",
				"float_1",
			},
			expected: `{"id":"1","int_1":{"set":10},"str_1":{"set":"foo"},"float_1":{"set":10.5}}`,
		},
	}

	for _, c := range cases {
		t.Run(c.name, func(t *testing.T) {
			got, err := solr.InPlaceUpdateEncode(&c.doc, c.allowedFields)
			if err != nil {
				t.Fatal(err)
			}
			if got != c.expected {
				t.Fatalf("\nexpected: %v\n but got: %v", c.expected, got)
			}
		})
	}
}
