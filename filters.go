package mongoq

// Operator represents filter operations.
type Operator int

const (
	Equal Operator = iota
	NotEqual
	GTE
	LTE
	Like
	IgnoreCase
)

// Filter defines a search filter for MongoDB queries.
type Filter struct {
	Key   string
	Value interface{}
	Op    Operator
}

func newFilter(key string, value interface{}, op ...Operator) Filter {
	f := Filter{Key: key, Value: value, Op: Equal}

	if len(op) > 0 {
		f.Op = op[0]
	}

	return f
}

// Filters is a map of filters for building MongoDB queries.
type Filters map[string]Filter
