1. **Collections & schema registry** (define fields, types, search/filter/facet/sort flags)

2. **Document CRUD**

* Index/Upsert single doc
* Bulk index/upsert
* Partial update
* Delete single / bulk
* Get-by-id (optional helper)

3. **Tokenization pipeline**

* Lowercasing + punctuation splitting
* Unicode-safe handling
* Stopwords (optional)
* Field-level weights

4. **Full-text search**

* Multi-field query (`QueryBy`)
* AND/OR semantics
* Phrase match (nice-to-have)
* Prefix search (autocomplete-ish)
* Relevance scoring (TF/IDF or BM25-lite)

5. **Filtering**

* Equality filters (string/bool/enums)
* IN filters (multi-value)
* Numeric/time ranges (`>=, <=, between`)
* Compound logic (AND/OR)

6. **Sorting**

* Relevance (`_score`)
* Numeric/time field sort (asc/desc)
* Secondary sort keys (tie-breakers)

7. **Faceting**

* Tag facets (string/string-array)
* Top-N counts
* Facets after filters (like Typesense)

8. **Highlighting**

* Matched token highlighting per field
* Snippet generation with max length (nice-to-have)

9. **Typo tolerance** (nice-to-have but Typesense-ish)

* Max typos (0–2)
* Candidate generation via n-grams
* Edit-distance ranking

10. **Synonyms** (nice-to-have)

* Per-collection synonym sets
* Query-time expansion

11. **Geo search** (nice-to-have)

* Radius filter
* Sort by distance

12. **Grouping / collapsing** (nice-to-have)

* Group by field (e.g., product_id)
* Top hit per group

13. **Curation rules** (nice-to-have)

* Pin docs for certain queries
* Hide docs for certain queries

14. **Multi-search** (nice-to-have)

* Batch multiple queries in one call (library method)

15. **Schema evolution**

* Add field, remove field (safe mode)
* Reindex requirements detection

16. **Index maintenance**

* Deterministic indexing (stable results)
* Temp-key cleanup (TTL)
* Optional compaction routines

17. **Safety controls**

* Max query length / token count
* Max facets, max facet values
* Max hits/oversampling cap
* Timeouts via context

18. **Observability hooks**

* Timing (`took`)
* Counters (indexed docs, search calls)
* Optional logger interface

19. **Extensibility**

* Pluggable tokenizer/analyzer
* Pluggable scoring strategy
* Optional filter DSL parser (`ParseFilter`)

20. **Testing essentials**

* Golden tests for search results
* Property-style tests: index → search → delete → search
* Redis integration tests (docker in CI)

### “Things to have” (non-negotiables to ship v1 cleanly)

* Stable schema + validation
* Deterministic tokenization
* Upsert + delete correctness (no index leaks)
* Search + filters + facets + relevance
* Temp key TTL + low round-trips (Lua or pipelining)
* Clear limits (MaxPerPage, MaxHits, etc.)
* Integration tests against real Redis
