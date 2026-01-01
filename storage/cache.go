package storage

import (
	"context"
	"errors"

	"github.com/chuksgpfr/typelite/schema"
)

var (
	ErrCollectionName   = errors.New("failed to set name")
	ErrCollectionHeader = errors.New("failed to set name")
	ErrCollection       = errors.New("failed to set collection")
	ErrKeyAlreadyExist  = errors.New("key already exist, you can upsert")
	ErrStorageFailed    = errors.New("failed to query storage")
)

type Storage interface {
	CreateCollection(ctx context.Context, data *schema.CreateCollection) error
	IndexDocument(ctx context.Context, dockKey string, data *schema.IndexDocumentPayload) error
	CheckDocumentKeyExist(ctx context.Context, dockKey string) (bool, error)
	// IndexDocumentFields(ctx context.Context, dockKey string, data *schema.IndexDocumentPayload) error
	IndexTextField(ctx context.Context, term, collectionName, documentIndexKey, invertedIndexKey, dictionaryKey, primaryKey string, field *schema.Field) error
}
