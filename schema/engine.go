package schema

import "time"

type Op string

const (
	OpEq  Op = "=="
	OpNe  Op = "!="
	OpGt  Op = ">"
	OpGte Op = ">="
	OpLt  Op = "<"
	OpLte Op = "<="
	OpIn  Op = "IN"
)

type FilterCondition struct {
	Field string
	Op    Op
	Value any
}

// For v1, we keep it simple: a list of AND conditions + a list of OR conditions.
// (If both are set, semantics are: (AND...) AND (OR...)).
type Filter struct {
	And []FilterCondition
	Or  []FilterCondition
}

type SortOrder int

const (
	Asc SortOrder = iota
	Desc
)

type SortSpec struct {
	// Field can be a schema field name or special fields like "_score".
	Field string
	Order SortOrder
}

type SearchRequest struct {
	Collection string

	// Query is the raw user query string (e.g. "iphone 15").
	Query string

	// QueryBy lists fields to search across. If empty, engine may default
	// to all fields with Search=true.
	QueryBy []string

	// Filter is an optional typed filter expression.
	Filter Filter

	// SortBy controls ordering. Default is relevance if Query is non-empty.
	SortBy []SortSpec

	// FacetBy lists fields to return facets for.
	FacetBy []string

	// Pagination
	PerPage int
	Page    int

	// Options
	PrefixSearch bool // allow prefix expansion for the last token
	MaxHits      int  // engine may oversample candidates before sorting/paging (e.g. 200)
}

type Hit struct {
	ID    string
	Score float64
	Doc   map[string]any
}

type FacetCount struct {
	Value string
	Count int64
}

// facet is like grouping, like if the query passes facet of status
// status can have active, pending, and their number
type FacetResult struct {
	Field  string
	Counts []FacetCount
}

type SearchResponse struct {
	Hits    []Hit
	Total   int64
	Page    int
	PerPage int
	Facets  []FacetResult

	Took  time.Duration
	Query string
}

//
// Indexing API
//

type IndexMode int

const (
	// Upsert replaces existing documents with the same primary key.
	Upsert IndexMode = iota
	// InsertOnly fails if a document with the same primary key exists.
	InsertOnly
)

type IndexOptions struct {
	Mode IndexMode
}

type BulkError struct {
	Index int
	ID    string
	Err   error
}

type BulkResult struct {
	Indexed int
	Errors  []BulkError
}
