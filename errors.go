package typelite

import "errors"

var (
	ErrNotImplemented           = errors.New("typelite: not implemented")
	ErrCollectionNotFound       = errors.New("typelite: collection not found")
	ErrInvalidSchema            = errors.New("typelite: invalid schema")
	ErrInvalidDocument          = errors.New("typelite: invalid document")
	ErrPrimaryKeyMissing        = errors.New("typelite: primary key missing")
	ErrNoCollectionName         = errors.New("typelite: no collection name")
	ErrNoHeader                 = errors.New("typelite: no header")
	ErrFailedToCreateCollection = errors.New("typelite: failed to create collection")
)
