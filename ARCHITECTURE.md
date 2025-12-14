typelite/
  typelite.go            # public facade
  schema/                # schema types, validation, persistence in redis
  doc/                   # doc encoding + validation + primary key handling
  tokenize/              # tokenizer (dependency-free)
  index/                 # write path: build terms + update redis
  query/
    parse/               # filter DSL parser (optional)
    plan/                # builds an executable plan from SearchRequest
  engine                # executes plan on redis, ranking, pagination
  facet/                 # facet computation helpers
  redisx/                # redis interface, lua scripts, pipelining
  dialect/               # optional: scoring models (tf-idf / bm25-lite)
  internal/util/
  scripts/               # lua (go:embed)
