package mongoq

import (
	"go.mongodb.org/mongo-driver/bson"
	"go.mongodb.org/mongo-driver/mongo/options"
)

// BuildFilter converts the query's root filter node into a MongoDB filter document (bson.M).
// It returns an empty document (bson.M{}) if the query has no filter, which the
// MongoDB Go driver treats as a match-all query. Passing nil to Collection.Find
// would return ErrNilDocument, so we never return nil here.
func BuildFilter(q *Query) (bson.M, error) {
	if q.filter == nil {
		return bson.M{}, nil
	}
	return q.filter.ToBSON()
}

// BuildOptions constructs a FindOptions object from the query's limit, offset, sort, and projection.
func BuildOptions(q *Query) *options.FindOptions {
	opts := options.Find()

	if q.limit > 0 {
		opts.SetLimit(q.limit)
	}
	if q.offset > 0 {
		opts.SetSkip(q.offset)
	}
	if len(q.sort) > 0 {
		sortDoc := bson.D{}
		for _, s := range q.sort {
			sortDoc = append(sortDoc, bson.E{Key: s.Field, Value: s.Order.ToInt()})
		}
		opts.SetSort(sortDoc)
	}
	if q.projection != nil {
		opts.SetProjection(q.projection)
	}
	return opts
}

// BuildMongoQuery is a convenience function that returns both the filter and options.
// It is equivalent to calling BuildFilter and BuildOptions.
func BuildMongoQuery(q *Query) (bson.M, *options.FindOptions, error) {
	filter, err := BuildFilter(q)
	if err != nil {
		return nil, nil, err
	}
	opts := BuildOptions(q)
	return filter, opts, nil
}
