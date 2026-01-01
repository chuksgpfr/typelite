package typelite

import (
	"context"
	"fmt"
	"strings"
	"sync"

	"github.com/chuksgpfr/typelite/logger"
	"github.com/chuksgpfr/typelite/schema"
	"github.com/chuksgpfr/typelite/storage"
	"github.com/chuksgpfr/typelite/tokenizer"
)

type EngineConfig struct {
	Storage storage.Storage

	// Namespace prefixes all keys, allowing multiple engines to share a Redis.
	// Default: "tl"
	Namespace string

	// DefaultPerPage is used when *schema.SearchRequest.PerPage is 0.
	// Default: 20
	DefaultPerPage int

	// MaxPerPage caps *schema.SearchRequest.PerPage.
	// Default: 250
	MaxPerPage int

	// Logger is optional.
	Logger logger.Logger
}

// normalize fills defaults.
func (c *EngineConfig) normalize() *EngineConfig {
	if c.Namespace == "" {
		c.Namespace = "tl"
	}
	if c.DefaultPerPage <= 0 {
		c.DefaultPerPage = 20
	}
	if c.MaxPerPage <= 0 {
		c.MaxPerPage = 250
	}
	return c
}

// MustCollection is a convenience for startup-time schema definitions.
func MustCollection(c *schema.Collection, err error) *schema.Collection {
	if err != nil {
		panic(err)
	}
	return c
}

// ValidateCollection validates a schema definition.
// This should be called before RegisterCollection; RegisterCollection will call it too.
func ValidateCollection(c *schema.Collection) error {
	if c.Name == "" {
		return fmt.Errorf("%w: collection name required", ErrInvalidSchema)
	}
	if len(c.Fields) == 0 {
		return fmt.Errorf("%w: at least one field required", ErrInvalidSchema)
	}

	primaryKeys := 0

	fieldNames := map[string]struct{}{}
	for _, f := range c.Fields {
		if f.Name == "" {
			return fmt.Errorf("%w: field name required", ErrInvalidSchema)
		}
		if _, ok := fieldNames[f.Name]; ok {
			return fmt.Errorf("%w: duplicate field %q", ErrInvalidSchema, f.Name)
		}
		fieldNames[f.Name] = struct{}{}

		if f.PrimaryKey {
			primaryKeys++
		}

		if primaryKeys > 1 {
			return fmt.Errorf("%w: you have more than one primary key, it is not allowed", ErrInvalidSchema)
		}
	}

	if primaryKeys == 0 {
		return fmt.Errorf("%w: you did not set a primary key", ErrInvalidSchema)
	}
	return nil
}

func ParseFilter(_ string) (schema.Filter, error) { return schema.Filter{}, ErrNotImplemented }

// It is safe for concurrent use.
type Engine struct {
	cfg *EngineConfig

	mu          sync.RWMutex
	collections map[string]*schema.Collection
	getDocKey   func(collectionName, primaryKey string) string
}

func NewEngine(cfg *EngineConfig) *Engine {
	cfg = cfg.normalize()
	// collection := &Collection{}
	return &Engine{
		cfg:         cfg,
		collections: make(map[string]*schema.Collection),
		getDocKey: func(collectionName, primaryKey string) string {
			return fmt.Sprintf("%s:%s:docx:%s", cfg.Namespace, collectionName, primaryKey)
		},
	}
}

// RegisterCollection validates and registers a collection schema.
// Implementations may persist schema metadata to Redis
func (e *Engine) RegisterCollection(ctx context.Context, c *schema.Collection) error {
	if e.cfg.Storage == nil {
		return fmt.Errorf("%w: storage client is required", ErrInvalidSchema)
	}
	if err := ValidateCollection(c); err != nil {
		return err
	}

	headers := []string{}
	for _, f := range c.Fields {
		headers = append(headers, f.Name)
	}

	data := &schema.CreateCollection{
		Name:    c.Name,
		Headers: headers,
	}

	var err error

	e.mu.Lock()

	err = e.cfg.Storage.CreateCollection(ctx, data)
	e.collections[data.Name] = c
	e.mu.Unlock()

	// TODO: persist schema to Redis (optional)
	return err
}

// Collection returns the registered collection schema, if present.
func (e *Engine) Collection(ctx context.Context, name string) (*schema.Collection, bool) {
	_ = ctx
	e.mu.RLock()
	defer e.mu.RUnlock()
	c, ok := e.collections[name]
	return c, ok
}

// IndexDoc indexes or upserts a single document into a collection.
func (e *Engine) IndexDocument(ctx context.Context, collectionName, documentKey, documentIndexKey string, doc *schema.IndexDocumentPayload, opts *schema.IndexOptions) error {
	e.mu.RLock()
	collection, ok := e.collections[collectionName]
	e.mu.RUnlock()

	if !ok {
		return ErrCollectionNotFound
	}

	// dockKey := e.getDocKey(c.Name, primaryKey.Name)

	if opts.Mode == schema.InsertOnly {
		keyExist, err := e.cfg.Storage.CheckDocumentKeyExist(ctx, documentKey)
		if err != nil {
			return err
		}

		if keyExist {
			// ErrDocumentKeyAlreadyExist
			return nil
		}
	}

	err := e.cfg.Storage.IndexDocument(ctx, documentKey, doc)

	if err != nil {
		return err
	}

	err = e.indexDocFields(ctx, documentIndexKey, collection, *doc)

	if err != nil {
		return err
	}
	return nil
}

// IndexDocs indexes multiple documents. It should continue indexing even if some docs fail.
// func (e *Engine) IndexDocs(ctx context.Context, collection, documentKey string, docs *schema.IndexDocumentPayload, opts *schema.IndexOptions) (*schema.BulkResult, error) {
// 	_ = ctx
// 	_ = collection
// 	_ = docs
// 	_ = opts

// 	if opts.Mode == schema.InsertOnly {
// 		exist, err := e.cfg.Storage.CheckDocumentKeyExist(ctx, documentKey)
// 		if err != nil {
// 			return nil, ErrFailedToIndexDocument
// 		}

// 		if !exist {
// 			return nil, ErrCollectionNotFound
// 		}
// 	}

// 	// store a flat document
// 	err := e.cfg.Storage.IndexDocument(ctx, documentKey, docs)
// 	if err != nil {
// 		return nil, ErrFailedToIndexDocument
// 	}

// 	err = e.indexDocFields(ctx, collection, *docs)
// 	// TODO: implement with pipelining / batching
// 	return &schema.BulkResult{}, ErrNotImplemented
// }

func (e *Engine) indexDocFields(ctx context.Context, documentIndexKey string, collection *schema.Collection, doc schema.IndexDocumentPayload) error {

	for _, f := range collection.Fields {
		filedValue, ok := doc[f.Name]
		if !ok {
			continue
		}

		switch f.Type {
		case schema.Text:
			if f.Search {
				s, _ := filedValue.(string)
				err := e.indexTextFields(ctx, collection, &f, documentIndexKey, s)
				if err != nil {
					return ErrFailedToIndexTextFields
				}
			}
		}
	}

	return nil
}

// DeleteDoc deletes a document and removes it from all indexes.
func (e *Engine) DeleteDoc(ctx context.Context, collection, id string) error {
	_ = ctx
	_ = collection
	_ = id
	// TODO: implement delete path
	return ErrNotImplemented
}

// DeleteDocs deletes multiple documents.
func (e *Engine) DeleteDocs(ctx context.Context, collection string, ids []string) (*schema.BulkResult, error) {
	_ = ctx
	_ = collection
	_ = ids
	// TODO: implement delete batching
	return &schema.BulkResult{}, ErrNotImplemented
}

// Search performs a query over a collection.
func (e *Engine) Search(ctx context.Context, req *schema.SearchRequest) (*schema.SearchResponse, error) {
	_ = ctx
	_ = req
	// TODO: implement internal/search executor (Lua + temp keys + TTL)
	return nil, ErrNotImplemented
}

// Close releases any resources held by the engine.
// (Many Redis clients do not require this; keep for symmetry.)
func (e *Engine) Close() error { return nil }

func (e *Engine) indexTextFields(ctx context.Context, c *schema.Collection, f *schema.Field, documentIndexKey, text string) error {
	if strings.EqualFold(text, "") {
		return nil
	}

	ns := e.cfg.Namespace

	dictionaryKey := fmt.Sprintf("%s:%s:terms", ns, c.Name)

	tokenizedTexts := tokenizer.Tokenize(text)

	pk, ok := c.PrimaryKeyField()
	if !ok {
		return ErrPrimaryKeyMissing
	}

	for _, term := range tokenizedTexts {
		invertedIndexKey := fmt.Sprintf("%s:%s:ix:%s:t:%s", ns, c.Name, f.Name, term)
		err := e.cfg.Storage.IndexTextField(ctx, term, c.Name, documentIndexKey, invertedIndexKey, dictionaryKey, pk.Name, f)
		if err != nil {
			return err
		}
	}

	return nil
}
