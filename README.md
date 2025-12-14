# typelite

**typelite** is a lean, dependency-light search layer for Go, built on top of **Redis**.

- No HTTP server.
- No separate search daemon.
- You import it as a **Go package** and call it directly from your API/service.
- It gives you **Typesense-style** features (collections, full-text search, filters, facets, sorting) backed by Redis keys.

> Think: “mini Typesense / Algolia, but as a Go library on top of Redis”.

---

## Features

- **Collections & Schema**
  - Define collections with fields and types (`string`, `text`, `number`, `bool`, `time`, `string[]`, `geo`).
  - Flags per field: `PrimaryKey`, `Search`, `Filter`, `Facet`, `Sort`.
  - Schema validation and registration at startup.

- **Document Indexing**
  - Index/Upsert single documents.
  - Bulk indexing (with per-document error reporting).
  - Delete single or multiple documents.

- **Full-text Search**
  - Multi-field search (`QueryBy`).
  - Field weights.
  - Prefix search toggle (for autocomplete-ish behavior).
  - Relevance scoring (TF/IDF-style).

- **Filtering & Facets**
  - Equality filters (`==`, `IN`) on string/bool/enums.
  - Numeric/time range filters (`>`, `>=`, `<`, `<=`).
  - Facets with counts per value (like “status: active (120), draft (5)”).

- **Sorting**
  - Sort by relevance (`_score`).
  - Sort by numeric/time fields (asc/desc).
  - Secondary sort fields.

- **Redis-first Design**
  - All index structures live in Redis:
    - Document hashes
    - Inverted indexes (term → docIDs)
    - Tag/Facet sets
    - Numeric sorted sets
  - Uses a minimal `RedisClient` interface (`Do(ctx, args...)`).

---

## Status

This README describes the **intended public API** and design.

- Engine / schema types are stable.
- Index & search internals are designed to be pluggable (Redis-based).

---

## Installation

```bash
go get github.com/chuksgpfr/typelite
````

---

## Quick Start

### 1. Create a collection

`typelite` depends on a minimal Redis interface:


```go
redisUrl := "localhost:6379"
redisPassword := ""
namespace := ""
tle := typelite.NewTypeLite(redisUrl, redisPassword, namespace)

ctx := context.Background()
userCollection := &schema.Collection{
  Name: "users",
  Fields: []schema.Field{
    {
      Name: "id", Type: schema.String, Search: true, PrimaryKey: true,
    },
    {
      Name: "first_name", Type: schema.String, Search: true,
    },
    {
      Name: "last_name", Type: schema.String, Search: true,
    },
  },
}

err := tle.CreateCollection(ctx, userCollection)

fmt.Println(err)
```


### 2. Index a document

```go
doc := map[string]any{
    "id":          "p1",
    "name":        "iPhone 15 Pro",
    "description": "New Apple phone with titanium frame",
    "price":       1299,
    "status":      "active",
    "tags":        []string{"phone", "apple"},
}

if err := engine.IndexDoc(ctx, "products", doc); err != nil {
    log.Fatalf("index doc: %v", err)
}
```

Bulk indexing:

```go
docs := []map[string]any{
    {
        "id":     "p1",
        "name":   "iPhone 15 Pro",
        "price":  1299,
        "status": "active",
        "tags":   []string{"phone", "apple"},
    },
    {
        "id":     "p2",
        "name":   "Galaxy S24",
        "price":  1099,
        "status": "active",
        "tags":   []string{"phone", "samsung"},
    },
}

res, err := engine.IndexDocs(ctx, "products", docs)
if err != nil {
    log.Fatalf("bulk index: %v", err)
}
log.Printf("indexed=%d errors=%d", res.Indexed, len(res.Errors))
```

### 3. Run a search

```go
searchReq := typelite.SearchRequest{
    Collection: "products",
    Query:      "iphone",
    QueryBy:    []string{"name", "description"},

    // Filters: status == "active" AND price between 500 and 2000
    Filter: typelite.Filter{
        And: []typelite.FilterCondition{
            {Field: "status", Op: typelite.OpEq, Value: "active"},
            {Field: "price",  Op: typelite.OpGte, Value: 500},
            {Field: "price",  Op: typelite.OpLte, Value: 2000},
        },
    },

    FacetBy: []string{"status", "tags"},

    SortBy: []typelite.SortSpec{
        {Field: "_score", Order: typelite.Desc}, // relevance
    },

    PerPage: 20,
    Page:    1,
}

resp, err := engine.Search(ctx, searchReq)
if err != nil {
    log.Fatalf("search: %v", err)
}

log.Printf("total=%d hits=%d took=%s", resp.Total, len(resp.Hits), resp.Took)
for _, h := range resp.Hits {
    log.Printf("id=%s score=%.2f name=%v", h.ID, h.Score, h.Doc["name"])
}

for _, f := range resp.Facets {
    log.Printf("Facet %s:", f.Field)
    for _, c := range f.Counts {
        log.Printf("  %s (%d)", c.Value, c.Count)
    }
}
```

---

## Core Concepts

### Engine

`Engine` is the main entry point:

It is **safe for concurrent use** and holds:

* Config
* Registered collections
* Internals for indexing/search.

### Collections & Fields

```go
type Collection struct {
    Name   string
    Fields []Field
}

type FieldType int

const (
    FieldString FieldType = iota
    FieldText
    FieldNumber
    FieldBool
    FieldTime
    FieldStringArray
)

type Field struct {
    Name       string
    Type       FieldType
    PrimaryKey bool

    Search bool
    Weight float64

    Filter bool
    Facet  bool
    Sort   bool
}
```


Validation rules:

* Collection name must be non-empty.
* At least one field.
* Field names unique.
* Exactly **one** `PrimaryKey` field.

### Search

```go
type SortOrder int

const (
    Asc SortOrder = iota
    Desc
)

type SortSpec struct {
    Field string
    Order SortOrder
}

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

type Filter struct {
    And []FilterCondition
    Or  []FilterCondition
}

type SearchRequest struct {
    Collection string
    Query      string
    QueryBy    []string
    Filter     Filter
    SortBy     []SortSpec
    FacetBy    []string

    PerPage     int
    Page        int
    PrefixSearch bool
    MaxHits      int
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
```



## Roadmap / Ideas

* Typed filter DSL parser (`ParseFilter("status:='active' && price:>=1000")`).
* Typo tolerance (n-gram + edit distance).
* Synonyms per collection.
* Geo search (lat/lon + radius).
* Pluggable analyzers/tokenizers.
* More advanced scoring functions (BM25, custom scoring hooks).

---

## License
