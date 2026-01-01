package main

import (
	"context"

	"github.com/chuksgpfr/typelite"
	"github.com/chuksgpfr/typelite/schema"
)

func main() {
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
	if err != nil {
		panic(err)
	}

	doc := &schema.IndexDocumentPayload{
		"id":         "p1",
		"first_name": "Lord",
		"last_name":  "Khagan",
	}

	err = tle.IndexDocument(ctx, userCollection.Name, doc, &schema.IndexOptions{Mode: schema.InsertOnly})
	if err != nil {
		panic(err)
	}

}
