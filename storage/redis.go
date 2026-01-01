package storage

import (
	"context"
	"encoding/json"
	"fmt"
	"strconv"
	"strings"

	"github.com/chuksgpfr/typelite/schema"
	"github.com/redis/go-redis/v9"
)

type RedisClient struct {
	rdb       *redis.Client
	namespace func(name string) string
}

func NewRedisClient(rdb *redis.Client, namespace string) *RedisClient {
	return &RedisClient{
		rdb: rdb,
		namespace: func(name string) string {
			return fmt.Sprintf("%s:%s", namespace, name)
		},
	}
}

func (r *RedisClient) CreateCollection(ctx context.Context, data *schema.CreateCollection) error {
	if strings.EqualFold(data.Name, "") {
		return fmt.Errorf("%w: pass collection name", ErrCollectionName)
	}

	if len(data.Headers) == 0 {
		return fmt.Errorf("%w: pass a header", ErrCollectionHeader)
	}

	_, err := r.do(ctx, "SADD", r.namespace("collections"), data.Name)
	if err != nil {
		return fmt.Errorf("%w: please fix error %v", ErrCollection, err)
	}

	if _, err := r.do(ctx, "ZADD", r.namespace("collections:by_name"), 0, data.Name); err != nil {
		return err
	}

	fieldsJSON, err := json.Marshal(data.Headers)
	if err != nil {
		return err
	}

	metaKey := r.namespace("collection:" + data.Name)
	if _, err := r.do(ctx,
		"HSET", metaKey,
		"name", data.Name,
		"fields", string(fieldsJSON),
	); err != nil {
		return err
	}

	for _, h := range data.Headers {
		fieldKey := r.namespace("collections:field:" + h)
		// SADD tl:collections:field:id "products"
		if _, err := r.do(ctx, "SADD", fieldKey, data.Name); err != nil {
			return err
		}
	}

	return nil
}

func (r *RedisClient) IndexDocument(ctx context.Context, dockKey string, data *schema.IndexDocumentPayload) error {
	flatArgs := []any{}
	flatArgs = append(flatArgs, "HSET", dockKey)

	for key, value := range *data {
		flatArgs = append(flatArgs, []any{key, value}...)
	}

	_, err := r.do(ctx, flatArgs...)

	return err
}

func (r *RedisClient) CheckDocumentKeyExist(ctx context.Context, dockKey string) (bool, error) {
	resp, err := r.do(ctx, "EXISTS", dockKey)
	if err != nil {
		return false, ErrStorageFailed
	}

	var exists int64
	switch v := resp.(type) {
	case int64:
		exists = v
	case int:
		exists = int64(v)
	case string:
		iv, err := strconv.ParseInt(v, 10, 64)
		if err != nil {
			return false, ErrStorageFailed
		}
		exists = iv
	case []byte:
		iv, err := strconv.ParseInt(string(v), 10, 64)
		if err != nil {
			return false, ErrStorageFailed
		}
		exists = iv
	default:
		iv, err := strconv.ParseInt(fmt.Sprint(v), 10, 64)
		if err != nil {
			return false, ErrStorageFailed
		}
		exists = iv
	}

	if exists != 1 {
		return false, nil
	}

	return true, nil
}

func (r *RedisClient) IndexTextField(ctx context.Context, term, collectionName, documentIndexKey, invertedIndexKey, dictionaryKey, primaryKey string, field *schema.Field) error {
	// this is for prefix search
	_, err := r.do(ctx, "ZADD", dictionaryKey, 0, term)
	if err != nil {
		return err
	}

	// score here could be tf * weight; keeping it simple = weight
	score := field.Weight
	if score == 0 {
		score = 1.0
	}
	_, err = r.do(ctx, "ZADD", invertedIndexKey, score, primaryKey)
	if err != nil {
		return err
	}

	_, err = r.do(ctx, "SADD", documentIndexKey, invertedIndexKey)
	if err != nil {
		return err
	}

	return nil
}

func (r *RedisClient) do(ctx context.Context, args ...any) (interface{}, error) {
	cmd := r.rdb.Do(ctx, args...)
	return cmd.Result()
}
