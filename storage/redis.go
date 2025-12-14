package storage

import (
	"context"
	"encoding/json"
	"errors"
	"fmt"
	"strings"

	"github.com/chuksgpfr/typelite/schema"
	"github.com/redis/go-redis/v9"
)

var (
	ErrRedisCollectionName   = errors.New("failed to set name")
	ErrRedisCollectionHeader = errors.New("failed to set name")
	ErrRedisCollection       = errors.New("failed to set collection")
)

type RedisClient struct {
	rdb *redis.Client
}

func NewRedisClient(rdb *redis.Client) *RedisClient {
	return &RedisClient{
		rdb: rdb,
	}
}

func (r *RedisClient) CreateCollection(ctx context.Context, data *schema.CreateCollection) error {
	if strings.EqualFold(data.Name, "") {
		return fmt.Errorf("%w: pass collection name", ErrRedisCollectionName)
	}

	if len(data.Headers) == 0 {
		return fmt.Errorf("%w: pass a header", ErrRedisCollectionHeader)
	}

	res, err := r.do(ctx, "SADD", "tl:collections", data.Name)
	if err != nil {
		return fmt.Errorf("%w: please fix error %v", ErrRedisCollection, err)
	}

	fmt.Println("WER ", res)

	if _, err := r.do(ctx, "ZADD", "tl:collections:by_name", 0, data.Name); err != nil {
		return err
	}

	fieldsJSON, err := json.Marshal(data.Headers)
	if err != nil {
		return err
	}

	metaKey := "tl:collection:" + data.Name
	if _, err := r.do(ctx,
		"HSET", metaKey,
		"name", data.Name,
		"fields", string(fieldsJSON),
	); err != nil {
		return err
	}

	for _, h := range data.Headers {
		fieldKey := "tl:collections:field:" + h
		// SADD tl:collections:field:id "products"
		if _, err := r.do(ctx, "SADD", fieldKey, data.Name); err != nil {
			return err
		}
	}

	return nil
}

func (r *RedisClient) do(ctx context.Context, args ...any) (interface{}, error) {
	cmd := r.rdb.Do(ctx, args...)
	return cmd.Result()
}
