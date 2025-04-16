package solr_test

import (
	"iter"
	"testing"

	"github.com/google/go-cmp/cmp"

	"github.com/imishinist/solr-inplace-poc/internal/myiter"
	"github.com/imishinist/solr-inplace-poc/internal/solr"
)

func collect[T any](iter iter.Seq[T]) []T {
	ret := make([]T, 0)
	for v := range iter {
		ret = append(ret, v)
	}
	return ret
}

func TestFields_Iter(t *testing.T) {
	cases := []struct {
		name     string
		fields   solr.Fields
		expected []solr.Field
	}{
		{
			name:     "empty",
			fields:   make(solr.Fields, 0),
			expected: make([]solr.Field, 0),
		},
		{
			name: "sorted",
			fields: solr.Fields{
				{Key: "a", Value: 1},
				{Key: "b", Value: 1},
				{Key: "c", Value: 1},
			},
			expected: []solr.Field{
				{Key: "a", Value: 1},
				{Key: "b", Value: 1},
				{Key: "c", Value: 1},
			},
		},
		{
			name: "not sorted",
			fields: solr.Fields{
				{Key: "a", Value: 1},
				{Key: "c", Value: 1},
				{Key: "b", Value: 1},
			},
			expected: []solr.Field{
				{Key: "a", Value: 1},
				{Key: "b", Value: 1},
				{Key: "c", Value: 1},
			},
		},
	}

	for _, c := range cases {
		t.Run(c.name, func(t *testing.T) {
			actual := collect(c.fields.Iter())
			if diff := cmp.Diff(c.expected, actual); diff != "" {
				t.Fatalf("diff: %s", diff)
			}
		})
	}
}

func docsToDocSet(docs []solr.Document) solr.DocSet {
	ds := make(solr.DocSet)
	for _, doc := range docs {
		ds.Add(doc)
	}
	return ds
}

func genDoc(id string, val1 int, val2 string) *solr.Document {
	return &solr.Document{
		ID: id,
		Fields: []solr.Field{
			{Key: "int1", Value: val1},
			{Key: "str1", Value: val2},
		},
	}
}

func TestMergedDocSetIterator(t *testing.T) {
	type Expected = myiter.Merged[solr.Document]
	cases := []struct {
		name     string
		ds1      solr.DocSet
		ds2      solr.DocSet
		expected []Expected
	}{
		{
			name:     "empty",
			ds1:      make(solr.DocSet),
			ds2:      make(solr.DocSet),
			expected: make([]Expected, 0),
		},
		{
			name: "only ds1",
			ds1: docsToDocSet([]solr.Document{
				*genDoc("2", 2, "string"),
				*genDoc("1", 1, "string"),
			}),
			ds2: make(solr.DocSet),
			expected: []Expected{
				{Left: genDoc("1", 1, "string")},
				{Left: genDoc("2", 2, "string")},
			},
		},
		{
			name: "only ds2",
			ds1:  make(solr.DocSet),
			ds2: docsToDocSet([]solr.Document{
				*genDoc("2", 2, "string"),
				*genDoc("1", 1, "string"),
			}),
			expected: []Expected{
				{Right: genDoc("1", 1, "string")},
				{Right: genDoc("2", 2, "string")},
			},
		},
		{
			name: "both",
			ds1: docsToDocSet([]solr.Document{
				*genDoc("9", 9, "str9 ds1"),
				*genDoc("1", 1, "str1 ds1"),
				*genDoc("4", 4, "str4 ds1"),
				*genDoc("6", 6, "str6 ds1"),
				*genDoc("7", 7, "str7 ds1"),
			}),
			ds2: docsToDocSet([]solr.Document{
				*genDoc("1", 10, "str1 ds2"),
				*genDoc("2", 20, "str2 ds2"),
				*genDoc("3", 30, "str3 ds2"),
				*genDoc("5", 50, "str5 ds2"),
				*genDoc("8", 80, "str8 ds2"),
			}),
			expected: []Expected{
				{
					Left:  genDoc("1", 1, "str1 ds1"),
					Right: genDoc("1", 10, "str1 ds2"),
				},
				{Right: genDoc("2", 20, "str2 ds2")},
				{Right: genDoc("3", 30, "str3 ds2")},
				{Left: genDoc("4", 4, "str4 ds1")},
				{Right: genDoc("5", 50, "str5 ds2")},
				{Left: genDoc("6", 6, "str6 ds1")},
				{Left: genDoc("7", 7, "str7 ds1")},
				{Right: genDoc("8", 80, "str8 ds2")},
				{Left: genDoc("9", 9, "str9 ds1")},
			},
		},
	}
	for _, c := range cases {
		t.Run(c.name, func(t *testing.T) {
			iter := solr.NewMergedDocSetIterator(c.ds1.Iter(), c.ds2.Iter())
			actual := collect(iter.Iter())

			if diff := cmp.Diff(c.expected, actual); diff != "" {
				t.Fatalf("diff: %s\n", diff)
			}
		})
	}
}
