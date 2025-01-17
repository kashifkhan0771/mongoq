# mongoq
Build MongoDB queries ⚙️


## Example usage:
```go
package main

import (
	"context"
	"fmt"
	"log"

	"github.com/kashifkhan0771/mongoq"
	"go.mongodb.org/mongo-driver/mongo"
	"go.mongodb.org/mongo-driver/mongo/options"
)

func main() {
	// Create a new query
	query := mongoq.NewQuery()
	query.AddFilter("age", 30, mongoq.GTE) // Add a filter: age >= 30
	query.Sortby = []string{"name"}       // Sort by name in ascending order
	query.Limit = 10                      // Limit to 10 results
	query.Offset = 0                      // Skip 0 records

	// Connect to MongoDB
	client, err := mongo.Connect(context.Background(), options.Client().ApplyURI("mongodb://localhost:27017"))
	if err != nil {
		log.Fatal(err)
	}
	defer client.Disconnect(context.Background())

	// Access the collection
	collection := client.Database("testdb").Collection("users")

	// Execute the query directly using the mongoq query
	cursor, err := collection.Find(
		context.Background(),
		mongoq.BuildMongoQuery(query),
	)
	defer cursor.Close(context.Background())

	// Print the results
	for cursor.Next(context.Background()) {
		var doc map[string]interface{}
		if err := cursor.Decode(&doc); err != nil {
			log.Fatal(err)
		}
		fmt.Println(doc)
	}

	if err := cursor.Err(); err != nil {
		log.Fatal(err)
	}
}
```