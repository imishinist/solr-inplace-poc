package solr_test

import (
	"testing"

	"github.com/google/go-cmp/cmp"

	"github.com/imishinist/solr-inplace-poc/internal/solr"
)

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
	cases := []struct {
		name     string
		ds1      solr.DocSet
		ds2      solr.DocSet
		expected []solr.MergedDoc
	}{
		{
			name:     "empty",
			ds1:      make(solr.DocSet),
			ds2:      make(solr.DocSet),
			expected: make([]solr.MergedDoc, 0),
		},
		{
			name: "only ds1",
			ds1: docsToDocSet([]solr.Document{
				*genDoc("2", 2, "string"),
				*genDoc("1", 1, "string"),
			}),
			ds2: make(solr.DocSet),
			expected: []solr.MergedDoc{
				solr.MergedDoc{Left: genDoc("1", 1, "string")},
				solr.MergedDoc{Left: genDoc("2", 2, "string")},
			},
		},
		{
			name: "only ds2",
			ds1:  make(solr.DocSet),
			ds2: docsToDocSet([]solr.Document{
				*genDoc("2", 2, "string"),
				*genDoc("1", 1, "string"),
			}),
			expected: []solr.MergedDoc{
				solr.MergedDoc{Right: genDoc("1", 1, "string")},
				solr.MergedDoc{Right: genDoc("2", 2, "string")},
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
			expected: []solr.MergedDoc{
				solr.MergedDoc{
					Left:  genDoc("1", 1, "str1 ds1"),
					Right: genDoc("1", 10, "str1 ds2"),
				},
				solr.MergedDoc{Right: genDoc("2", 20, "str2 ds2")},
				solr.MergedDoc{Right: genDoc("3", 30, "str3 ds2")},
				solr.MergedDoc{Left: genDoc("4", 4, "str4 ds1")},
				solr.MergedDoc{Right: genDoc("5", 50, "str5 ds2")},
				solr.MergedDoc{Left: genDoc("6", 6, "str6 ds1")},
				solr.MergedDoc{Left: genDoc("7", 7, "str7 ds1")},
				solr.MergedDoc{Right: genDoc("8", 80, "str8 ds2")},
				solr.MergedDoc{Left: genDoc("9", 9, "str9 ds1")},
			},
		},
	}
	for _, c := range cases {
		t.Run(c.name, func(t *testing.T) {
			iter := solr.NewMergedDocSetIterator(c.ds1.Iter(), c.ds2.Iter())
			actual := make([]solr.MergedDoc, 0)
			for doc := range iter.Iter() {
				actual = append(actual, doc)
			}

			if diff := cmp.Diff(c.expected, actual); diff != "" {
				t.Fatalf("diff: %s\n", diff)
			}
		})
	}
}
