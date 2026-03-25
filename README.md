# mongoq

Build MongoDB queries in Go with a clean, fluent API. ⚙️

## Installation

```bash
go get github.com/kashifkhan0771/mongoq
```

## Quick start

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
	// Build a query: active users older than 18, sorted by name, page 1 of 10
	filter, opts, err := mongoq.NewQuery().
		Filter("age", mongoq.GreaterThan, 18).
		Filter("active", mongoq.Equal, true).
		Sort("name", mongoq.Asc).
		Limit(10).
		Offset(0).
		Build()
	if err != nil {
		log.Fatal(err)
	}

	client, err := mongo.Connect(context.Background(), options.Client().ApplyURI("mongodb://localhost:27017"))
	if err != nil {
		log.Fatal(err)
	}
	defer client.Disconnect(context.Background())

	cursor, err := client.Database("testdb").Collection("users").Find(context.Background(), filter, opts)
	if err != nil {
		log.Fatal(err)
	}
	defer cursor.Close(context.Background())

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

## API reference

### Operators

| Constant             | MongoDB equivalent        | Notes                                                               |
|----------------------|---------------------------|---------------------------------------------------------------------|
| `Equal`              | `field: value`            |                                                                     |
| `NotEqual`           | `$ne`                     |                                                                     |
| `GreaterThan`        | `$gt`                     |                                                                     |
| `GreaterThanOrEqual` | `$gte`                    |                                                                     |
| `LessThan`           | `$lt`                     |                                                                     |
| `LessThanOrEqual`    | `$lte`                    |                                                                     |
| `In`                 | `$in`                     | Value must be `[]any`                                               |
| `NotIn`              | `$nin`                    | Value must be `[]any`                                               |
| `Exists`             | `$exists`                 | Value must be `bool`                                                |
| `Regex`              | `$regex`                  | Value must be a string regex pattern                                |
| `Contains`               | `$regex`                  | Raw regex pattern; use `".*foo.*"` for substring match. **Not** SQL `%` wildcards. |
| `IgnoreCase`         | `$regex` + `$options:"i"` | Case-insensitive regex                                              |

### Logical operators (for `Where`)

| Constant | MongoDB equivalent  | Notes                                                                                                      |
|----------|---------------------|------------------------------------------------------------------------------------------------------------|
| `And`    | `$and`              | Default when using chained `Filter()` calls                                                                |
| `Or`     | `$or`               |                                                                                                            |
| `Nor`    | `$nor`              |                                                                                                            |
| `Not`    | `$nor: [child]`     | Document-level NOT; must have exactly one child. Uses `$nor` because MongoDB's `$not` is field-level only. |

### Sort orders

| Constant | MongoDB equivalent | Notes              |
|----------|--------------------|--------------------|
| `Asc`    | `1`                | Ascending order    |
| `Desc`   | `-1`               | Descending order   |

### Query methods

```go
q := mongoq.NewQuery()

// Add a field condition (AND-ed together by default)
q.Filter(field string, op mongoq.Operator, value any) *Query

// Replace the root filter with a custom condition tree
q.Where(node mongoq.FilterNode) *Query

// Pagination
q.Limit(n int64) *Query
q.Offset(n int64) *Query

// Sorting (call multiple times for multi-field sort)
q.Sort(field string, order mongoq.SortOrder) *Query   // mongoq.Asc | mongoq.Desc

// Projection — accepts bson.D, bson.M, or map[string]int
q.Project(projection interface{}) *Query

// Produce the final filter document and FindOptions
filter, opts, err := q.Build()
```

### Filter tree structure

The query builder represents filters as a tree of nodes:

- **`FilterNode`** — interface implemented by all filter nodes
- **`FilterLeaf`** — a single field condition (e.g. `age > 18`)
- **`FilterGroup`** — a logical group with an `Operator` (`And`, `Or`, `Nor`, `Not`) and a slice of child `FilterNode` values

When you call `Filter()`, it creates `FilterLeaf` nodes and accumulates them into a root `And` group automatically. Use `Where()` with a `*FilterGroup` when you need explicit control over logical operators or nested groups.

### Custom condition trees with `Where`

Use `Where` when you need `$or`, `$nor`, or `Not` at the top level, or when
you need to nest logical groups:

```go
// Match documents where role is "admin" OR "moderator"
node := mongoq.FilterGroup{
	Operator: mongoq.Or,
	Children: []mongoq.FilterNode{
		mongoq.FilterLeaf{Field: "role", Operator: mongoq.Equal, Value: "admin"},
		mongoq.FilterLeaf{Field: "role", Operator: mongoq.Equal, Value: "moderator"},
	},
}
filter, opts, err := mongoq.NewQuery().Where(&node).Limit(50).Build()
```

> **Tip:** Pass `FilterGroup` as a pointer (`&node`) when used with `Where`.
