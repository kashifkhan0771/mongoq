package mongoq

// Query defines a MongoDB query with filters, sorting, limit, etc.
type Query struct {
	Filters Filters
	Limit   int64
	Offset  int64
	Sortby  []string
}

// NewQuery creates a new query instance.
func NewQuery() *Query {
	return &Query{
		Filters: make(Filters),
	}
}

// AddFilter adds a filter to the query.
func (q *Query) AddFilter(key string, value interface{}, op Operator) error {
	filter, err := newFilter(key, value, op)
	if err != nil {
		return err
	}

	q.Filters[key] = filter

	return nil
}
