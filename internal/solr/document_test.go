package solr_test

import (
	"testing"

	"github.com/imishinist/solr-inplace-poc/internal/solr"
)

func TestJSONEncode(t *testing.T) {
	cases := []struct {
		name     string
		doc      solr.Document
		expected string
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
			expected: `{"id":"1","int_1":10,"str_1":"foo","float_1":10.5}`,
		},
	}

	for _, c := range cases {
		t.Run(c.name, func(t *testing.T) {
			got, err := solr.JSONEncode(&c.doc)
			if err != nil {
				t.Fatal(err)
			}
			if got != c.expected {
				t.Fatalf("expected: %v, but got: %v", c.expected, got)
			}
		})
	}
}
