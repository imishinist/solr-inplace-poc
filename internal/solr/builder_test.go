package solr_test

import (
	"testing"

	"github.com/imishinist/solr-inplace-poc/internal/myiter"
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
				t.Fatalf("\nexpected:%v\nbut got: %v", c.expected, got)
			}

			builder.Flush()
			got, err = builder.Build()
			if err != nil {
				t.Fatal(err)
			}
			if got != "{}" {
				t.Fatalf("\nexpected:{}\nbut got: %v", got)
			}
		})
	}
}

func TestUpdateBatchBuilder_BuildInPlaceUpdate(t *testing.T) {
	type Input myiter.Merged[solr.Document]
	cases := []struct {
		name string

		add           []Input
		allowedFields []string
		inPlaceFields []string

		expected string
	}{
		{
			name:     "nothing documents",
			add:      []Input{},
			expected: "{}",
		},
		{
			name: "only new documents/only in-place updates",
			add: []Input{
				{
					// new only
					Right: &solr.Document{
						ID: "1",
						Fields: []solr.Field{
							{Key: "int1", Value: 10},
							{Key: "str1", Value: "string"},
						},
					},
				},
			},
			allowedFields: []string{"int1", "str1"},
			inPlaceFields: []string{"int1", "str1"},
			expected:      `{"add":{"doc":{"id":"1","int1":10,"str1":"string"}}}`,
		},
		{
			name: "old and new documents/only in-place updates",
			add: []Input{
				{
					Left: &solr.Document{
						ID: "1",
						Fields: []solr.Field{
							{Key: "int1", Value: 5},
							{Key: "str1", Value: "string"},
						},
					},
					Right: &solr.Document{
						ID: "1",
						Fields: []solr.Field{
							{Key: "int1", Value: 10},
							{Key: "str1", Value: "string"},
						},
					},
				},
			},
			allowedFields: []string{"int1", "str1"},
			inPlaceFields: []string{"int1", "str1"},
			expected:      `{"add":{"doc":{"id":"1","int1":{"set":10}}}}`,
		},
		{
			name: "only new documents/only in-place updates changes",
			add: []Input{
				{
					Right: &solr.Document{
						ID: "1",
						Fields: []solr.Field{
							{Key: "int1", Value: 10},
							{Key: "str1", Value: "string"},
							{Key: "other", Value: "other"},
						},
					},
				},
			},
			allowedFields: []string{"int1", "str1"},
			inPlaceFields: []string{"int1"},
			expected:      `{"add":{"doc":{"id":"1","int1":10,"str1":"string"}}}`,
		},
		{
			name: "old and new documents/only in-place updates changes",
			add: []Input{
				{
					Left: &solr.Document{
						ID: "1",
						Fields: []solr.Field{
							{Key: "int1", Value: 10},
							{Key: "str1", Value: "string"},
							{Key: "other", Value: "other"},
						},
					},
					Right: &solr.Document{
						ID: "1",
						Fields: []solr.Field{
							{Key: "int1", Value: 20}, // changed
							{Key: "str1", Value: "string"},
							{Key: "other", Value: "other"},
						},
					},
				},
			},
			allowedFields: []string{"int1", "str1"},
			inPlaceFields: []string{"int1"},
			expected:      `{"add":{"doc":{"id":"1","int1":{"set":20}}}}`,
		},
		{
			name: "old and new documents/only allowed fields changes",
			add: []Input{
				{
					Left: &solr.Document{
						ID: "1",
						Fields: []solr.Field{
							{Key: "int1", Value: 1},
							{Key: "str1", Value: "string"},
							{Key: "other", Value: "other"},
						},
					},
					Right: &solr.Document{
						ID: "1",
						Fields: []solr.Field{
							{Key: "int1", Value: 1},
							{Key: "str1", Value: "changed"}, // changed
							{Key: "other", Value: "other"},
						},
					},
				},
			},
			allowedFields: []string{"int1", "str1"},
			inPlaceFields: []string{"int1"},
			expected:      `{"add":{"doc":{"id":"1","int1":1,"str1":"changed"}}}`,
		},
		{
			name: "old and new documents/all allowed fields changed",
			add: []Input{
				{
					Left: &solr.Document{
						ID: "1",
						Fields: []solr.Field{
							{Key: "int1", Value: 1},
							{Key: "str1", Value: "string"},
							{Key: "other", Value: "other"},
						},
					},
					Right: &solr.Document{
						ID: "1",
						Fields: []solr.Field{
							{Key: "int1", Value: 10},        // changed
							{Key: "str1", Value: "changed"}, // changed
							{Key: "other", Value: "other"},
						},
					},
				},
			},
			allowedFields: []string{"int1", "str1"},
			inPlaceFields: []string{"int1"},
			expected:      `{"add":{"doc":{"id":"1","int1":10,"str1":"changed"}}}`,
		},
		{
			name: "old and new documents/only not allowed fields change",
			add: []Input{
				{
					Left: &solr.Document{
						ID: "1",
						Fields: []solr.Field{
							{Key: "int1", Value: 1},
							{Key: "str1", Value: "string"},
							{Key: "other", Value: "other"},
						},
					},
					Right: &solr.Document{
						ID: "1",
						Fields: []solr.Field{
							{Key: "int1", Value: 1},
							{Key: "str1", Value: "string"},
							{Key: "other", Value: "changed"},
						},
					},
				},
			},
			allowedFields: []string{"int1", "str1"},
			inPlaceFields: []string{"int1"},
			expected:      `{}`,
		},
		{
			name: "old and new documents/allowed fields removed",
			add: []Input{
				{
					Left: &solr.Document{
						ID: "1",
						Fields: []solr.Field{
							{Key: "int1", Value: 1},
							{Key: "str1", Value: "string"},
							{Key: "other", Value: "other"},
						},
					},
					Right: &solr.Document{
						ID: "1",
						Fields: []solr.Field{
							{Key: "int1", Value: 1},
							// {Key: "str1", Value: "string"}, // removed
							{Key: "other", Value: "other"},
						},
					},
				},
			},
			allowedFields: []string{"int1", "str1"},
			inPlaceFields: []string{"int1"},
			expected:      `{"add":{"doc":{"id":"1","int1":1}}}`,
		},
		{
			name: "old and new documents/inplace fields removed",
			add: []Input{
				{
					Left: &solr.Document{
						ID: "1",
						Fields: []solr.Field{
							{Key: "int1", Value: 1},
							{Key: "str1", Value: "string"},
							{Key: "other", Value: "other"},
						},
					},
					Right: &solr.Document{
						ID: "1",
						Fields: []solr.Field{
							// {Key: "int1", Value: 1}, // removed
							{Key: "str1", Value: "string"},
							{Key: "other", Value: "other"},
						},
					},
				},
			},
			allowedFields: []string{"int1", "str1"},
			inPlaceFields: []string{"int1"},
			expected:      `{"add":{"doc":{"id":"1","str1":"string"}}}`,
		},
	}

	for _, c := range cases {
		t.Run(c.name, func(t *testing.T) {
			builder := solr.NewUpdateBatchBuilder(c.allowedFields, c.inPlaceFields)
			if len(c.add) > 0 {
				for _, m := range c.add {
					if m.Left == nil {
						builder.Add(*m.Right)
					} else {
						builder.Update(*m.Right, *m.Left)
					}
				}
			}

			got, err := builder.Build()
			if err != nil {
				t.Fatal(err)
			}

			if got != c.expected {
				t.Fatalf("\nexpected:%v\n but got:%v", c.expected, got)
			}

			builder.Flush()
			got, err = builder.Build()
			if err != nil {
				t.Fatal(err)
			}
			if got != "{}" {
				t.Fatalf("\nexpected:{}\n but got:%v", got)
			}
		})
	}
}
