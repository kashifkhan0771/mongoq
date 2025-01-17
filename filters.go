package mongoq

import "go.mongodb.org/mongo-driver/bson"

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

// newFilter creates a new filter with validation and operator mapping.
func newFilter(key string, value interface{}, op ...Operator) (Filter, error) {
	if key == "" {
		return Filter{}, errors.New("key cannot be empty")
	}

	if value == nil {
		return Filter{}, errors.New("value cannot be nil")
	}

	f := Filter{Key: key, Value: value, Op: Equal}
	if len(op) > 0 {
		if !op[0].IsValid() {
			return Filter{}, fmt.Errorf("invalid operator: %v", op[0])
		}
		f.Op = op[0]
	}

	// Map operators to MongoDB query syntax
	switch f.Op {
	case Equal:
		f.Value = value
	case NotEqual:
		f.Value = bson.M{"$ne": value}
	case GTE:
		f.Value = bson.M{"$gte": value}
	case LTE:
		f.Value = bson.M{"$lte": value}
	case Like, IgnoreCase:
		pattern, ok := value.(string)
		if !ok {
			return Filter{}, errors.New("regex pattern must be a string")
		}
		if pattern == "" {
			return Filter{}, errors.New("regex pattern cannot be empty")
		}
		if f.Op == Like {
			f.Value = bson.M{"$regex": pattern}
		} else {
			f.Value = bson.M{"$regex": pattern, "$options": "i"}
		}
	}

	return f, nil
}

// Filters is a map of filters for building MongoDB queries.
type Filters map[string]Filter
