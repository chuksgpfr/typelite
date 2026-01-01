package schema

type CreateCollection struct {
	Name    string
	Headers []string
}

type FieldType string

const (
	Text   FieldType = "text"
	String FieldType = "string" // tags/facets
	Int    FieldType = "int"
	Float  FieldType = "float"
	Bool   FieldType = "bool"
	Time   FieldType = "time"
	Geo    FieldType = "geo"
)

type Field struct {
	Name       string
	Type       FieldType
	Search     bool
	Filter     bool // for String fields
	Sortable   bool // for numeric/time (and maybe string enum)
	Optional   bool
	PrimaryKey bool
	Weight     int
}

type Collection struct {
	Name   string
	Fields []Field
}

// PrimaryKeyField returns the PK field or false if not found.
func (c *Collection) PrimaryKeyField() (*Field, bool) {
	for _, f := range c.Fields {
		if f.PrimaryKey {
			return &f, true
		}
	}
	return &Field{}, false
}

type indexDocParam map[string]any
type IndexDocumentPayload indexDocParam
