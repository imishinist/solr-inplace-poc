package solr

type Field struct {
	Key   string
	Value interface{}
}

type Document struct {
	ID     string
	Fields []Field
}
