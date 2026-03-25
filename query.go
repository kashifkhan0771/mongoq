package mongoq

import (
	"fmt"
	"sort"

	"go.mongodb.org/mongo-driver/bson"
	"go.mongodb.org/mongo-driver/mongo/options"
)

// SortField defines a field and direction for sorting.
type SortField struct {
	Field string
	Order SortOrder
}

// Query is the main builder for MongoDB queries.
type Query struct {
	filter        FilterNode // root filter (nil means no filter)
	limit         int64      // 0 means no limit
	offset        int64      // 0 means no offset
	sort          []SortField
	projection    bson.D // using bson.D to preserve field order
	projectionErr error  // set if Project is called with an unsupported type
	validationErr error
}

// NewQuery creates a new empty query.
func NewQuery() *Query {
	return &Query{
		filter: nil,
		limit:  0,
		offset: 0,
		sort:   nil,
	}
}

// Filter adds a leaf condition to the root And group.
// If the root filter is nil, it initialises a new And group with this leaf.
// If the root filter is already an And group, the leaf is appended to it.
// Otherwise (e.g. the root was set via Where), a new And group is created
// that wraps the existing root together with the new leaf.
func (q *Query) Filter(field string, op Operator, value any) *Query {
	leaf := FilterLeaf{
		Field:    field,
		Operator: op,
		Value:    value,
	}
	if q.filter == nil {
		q.filter = &FilterGroup{
			Operator: And,
			Children: []FilterNode{leaf},
		}
	} else if group, ok := q.filter.(*FilterGroup); ok && group.Operator == And {
		// Mutate through the pointer — no silent copy/re-assign needed.
		group.Children = append(group.Children, leaf)
	} else {
		q.filter = &FilterGroup{
			Operator: And,
			Children: []FilterNode{q.filter, leaf},
		}
	}
	return q
}

// Where replaces the root filter with a custom condition tree.
// It panics if node is nil (including typed-nil pointers wrapped in the interface)
// because passing a nil FilterNode to Build would cause a nil-pointer dereference
// when ToBSON is called. Use NewQuery() without Where to match all documents.
func (q *Query) Where(node FilterNode) *Query {
	if node == nil {
		panic("mongoq: Where called with a nil FilterNode")
	}
	q.filter = node
	return q
}

// Limit sets the maximum number of documents to return.
func (q *Query) Limit(limit int64) *Query {
	if limit < 0 {
		q.validationErr = fmt.Errorf("mongoq: limit must be >= 0")
		return q
	}

	q.limit = limit
	return q
}

// Offset sets the number of documents to skip.
func (q *Query) Offset(offset int64) *Query {
	if offset < 0 {
		q.validationErr = fmt.Errorf("mongoq: offset must be >= 0")
		return q
	}

	q.offset = offset
	return q
}

// Sort adds a sort field (order can be Asc or Desc).
func (q *Query) Sort(field string, order SortOrder) *Query {
	q.sort = append(q.sort, SortField{Field: field, Order: order})
	return q
}

// Project sets the projection (fields to include or exclude).
// Accepted types: bson.D, bson.M, or map[string]int.
// Use bson.M{"field": 1} for inclusion, or bson.M{"field": 0} for exclusion.
// For order-sensitive projections, use bson.D directly.
//
// If an unsupported type is passed, the error is stored and returned by Build()
// so that the fluent call chain is not broken. This is the fail-fast pattern:
// the projection is never silently ignored.
func (q *Query) Project(projection any) *Query {
	switch p := projection.(type) {
	case bson.D:
		q.projection = p
	case bson.M:
		// Convert to bson.D with sorted keys for deterministic output.
		d := make(bson.D, 0, len(p))
		for k, v := range p {
			d = append(d, bson.E{Key: k, Value: v})
		}
		sort.Slice(d, func(i, j int) bool { return d[i].Key < d[j].Key })
		q.projection = d
	case map[string]int:
		// Handle the common case where callers pass a plain map[string]int.
		d := make(bson.D, 0, len(p))
		for k, v := range p {
			d = append(d, bson.E{Key: k, Value: v})
		}
		sort.Slice(d, func(i, j int) bool { return d[i].Key < d[j].Key })
		q.projection = d
	default:
		// Store the error and surface it at Build() time so the fluent chain
		// is not broken and the caller is not silently left without a projection.
		q.projectionErr = fmt.Errorf(
			"mongoq: unsupported projection type %T: use bson.D, bson.M, or map[string]int",
			projection,
		)
	}
	return q
}

// Build validates the query and returns the MongoDB filter document and FindOptions.
// It returns an error if any method in the chain was called with invalid arguments
// (e.g. an unsupported projection type passed to Project).
func (q *Query) Build() (bson.M, *options.FindOptions, error) {
	if q.projectionErr != nil {
		return nil, nil, q.projectionErr
	}

	if q.validationErr != nil {
		return nil, nil, q.validationErr
	}

	return BuildMongoQuery(q)
}
