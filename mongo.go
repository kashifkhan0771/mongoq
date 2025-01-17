package mongoq

import "go.mongodb.org/mongo-driver/bson"

// BSONFilters converts the filters in the query to a MongoDB-compatible bson.M.
func BSONFilters(q *Query) bson.M {
	query := bson.M{}
	for key, filter := range q.Filters {
		query[key] = filter.Value
	}
	return query
}

// Sort converts the Sortby field into a MongoDB sort array.
func Sort(q *Query) []string {
	return q.Sortby
}

// BuildMongoQuery combines Filters and Sort to build a MongoDB query.
func BuildMongoQuery(q *Query) bson.M {
	return BSONFilters(q)
}
