package typelite

import (
	"context"
	"errors"
	"fmt"

	"github.com/chuksgpfr/typelite/schema"
	"github.com/chuksgpfr/typelite/storage"
	"github.com/redis/go-redis/v9"
)

type TypeLite struct {
	engine *Engine
}

/*
*	Redis url is the url of your redis server
* Redis password is the password of your redis server, you can pass as empty
* Namespace is the folder/namespace you want typelite to use for your redis, if you pass empty, it defaults to "tl"
 */
func NewTypeLite(redisUrl, redisPassword, namespace string) *TypeLite {
	rdb := redis.NewClient(&redis.Options{
		Addr:     redisUrl,
		Password: redisPassword,
		DB:       0,
	})

	redisClient := storage.NewRedisClient(rdb, namespace)

	engineConfig := &EngineConfig{
		Storage:   redisClient,
		Namespace: namespace,
	}

	engine := NewEngine(engineConfig)

	return &TypeLite{
		engine: engine,
	}
}

func (t *TypeLite) CreateCollection(ctx context.Context, s *schema.Collection) error {
	if s == nil {
		return errors.New("please pass a non nil collection")
	}

	return t.engine.RegisterCollection(ctx, s)
}

func (t *TypeLite) IndexDocument(ctx context.Context, collectionName string, body *schema.IndexDocumentPayload, opts *schema.IndexOptions) error {

	collection := t.engine.collections[collectionName]
	primaryKey, ok := collection.PrimaryKeyField()

	if !ok {
		return ErrPrimaryKeyMissing
	}

	documentKey := fmt.Sprintf("%s:%s:doc:%s", t.engine.cfg.Namespace, collectionName, primaryKey.Name)
	documentIndexKey := fmt.Sprintf("%s:%s:docix:%s", t.engine.cfg.Namespace, collectionName, primaryKey.Name)

	err := t.engine.IndexDocument(ctx, collectionName, documentKey, documentIndexKey, body, opts)

	return err
}

// func (c *TypeLite) DropCollection(ctx context.Context, name string) error
// func (c *TypeLite) GetSchema(ctx context.Context, name string) (schema.Schema, error)
