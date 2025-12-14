package main

import (
	"context"
	"fmt"

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

	fmt.Println(err)
}
