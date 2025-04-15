package solr_test

import (
	"testing"

	"github.com/imishinist/solr-inplace-poc/internal/solr"
)

func TestUpdateBatchBuilder_Build(t *testing.T) {
	cases := []struct {
		name string

		add           []solr.Document
		delete        []solr.Document
		allowedFields []string

		expected string
	}{
		{
			name:     "nothing documents",
			add:      []solr.Document{},
			delete:   []solr.Document{},
			expected: "{}",
		},
		{
			name: "1 documents",
			add: []solr.Document{
				{
					ID: "1",
					Fields: []solr.Field{
						{Key: "int1", Value: 10},
						{Key: "str1", Value: "string"},
					},
				},
			},
			delete:   []solr.Document{},
			expected: `{"add":{"doc":{"id":"1","int1":10,"str1":"string"}}}`,
		},
		{
			name: "2 documents and deletes",
			add: []solr.Document{
				{
					ID: "1",
					Fields: []solr.Field{
						{Key: "int1", Value: 10},
						{Key: "str1", Value: "string1"},
					},
				},
				{
					ID: "2",
					Fields: []solr.Field{
						{Key: "int1", Value: 20},
						{Key: "str1", Value: "string2"},
					},
				},
			},
			delete: []solr.Document{
				{ID: "11"},
				{ID: "12"},
			},
			expected: `{` +
				`"add":{"doc":{"id":"1","int1":10,"str1":"string1"}}` +
				`,"add":{"doc":{"id":"2","int1":20,"str1":"string2"}}` +
				`,"delete":["11","12"]` +
				`}`,
		},
		{
			name: "documents with allowed fields",
			add: []solr.Document{
				{
					ID: "1",
					Fields: []solr.Field{
						{Key: "int1", Value: 10},
						{Key: "str1", Value: "string1"},
						{Key: "int2", Value: 10},
						{Key: "str2", Value: "string1"},
					},
				},
				{
					ID: "2",
					Fields: []solr.Field{
						{Key: "int1", Value: 20},
						{Key: "str1", Value: "string2"},
						{Key: "int2", Value: 10},
						{Key: "str2", Value: "string1"},
					},
				},
			},
			delete: []solr.Document{
				{ID: "11"},
				{ID: "12"},
			},
			allowedFields: []string{"int1", "str1"},
			expected: `{` +
				`"add":{"doc":{"id":"1","int1":10,"str1":"string1"}}` +
				`,"add":{"doc":{"id":"2","int1":20,"str1":"string2"}}` +
				`,"delete":["11","12"]` +
				`}`,
		},
	}

	for _, c := range cases {
		t.Run(c.name, func(t *testing.T) {
			builder := solr.NewUpdateBatchBuilder(c.allowedFields, nil)
			if len(c.add) > 0 {
				builder.Add(c.add...)
			}
			if len(c.delete) > 0 {
				builder.Delete(c.delete...)
			}

			got, err := builder.Build()
			if err != nil {
				t.Fatal(err)
			}

			if got != c.expected {
				t.Fatalf("expected %v, but got: %v", c.expected, got)
			}

			builder.Flush()
			got, err = builder.Build()
			if err != nil {
				t.Fatal(err)
			}
			if got != "{}" {
				t.Fatalf("expected {}, but got: %v", got)
			}
		})
	}
}
