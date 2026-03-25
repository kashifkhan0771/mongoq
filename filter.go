package mongoq

import (
	"fmt"
	"reflect"

	"go.mongodb.org/mongo-driver/bson"
)

// FilterLeaf represents a single field condition.
type FilterLeaf struct {
	Field    string
	Operator Operator
	Value    any
}

// validate checks if the leaf filter is valid. It is the single validation
// entry point and satisfies the FilterNode interface.
func (f FilterLeaf) validate() error {
	if f.Field == "" {
		return ErrEmptyKey
	}

	// $eq/$ne do not require a value, and can be nil
	if f.Operator != Equal && f.Operator != NotEqual {
		if f.Value == nil {
			return ErrNilValue
		}
	}

	// Reject unknown operators early, before BSON rendering.
	switch f.Operator {
	case Equal, NotEqual, GreaterThan, GreaterThanOrEqual, LessThan, LessThanOrEqual,
		In, NotIn, Exists, Regex, Contains, IgnoreCase:
		// known — proceed to type checks below
	default:
		return fmt.Errorf("%w: %v", ErrUnknownOperator, f.Operator)
	}
	switch f.Operator {
	case In, NotIn:
		// Accept any slice or array kind, not just []any, since Go slices are
		// invariant ([]string does not satisfy a []any type assertion).
		kind := reflect.TypeOf(f.Value).Kind()
		if kind != reflect.Slice && kind != reflect.Array {
			return fmt.Errorf("%w: expected slice/array for operator %v", ErrInvalidValue, f.Operator)
		}
	case Exists:
		// Value must be bool
		if _, ok := f.Value.(bool); !ok {
			return fmt.Errorf("%w: expected bool for operator %v", ErrInvalidValue, f.Operator)
		}
	case Regex, Contains, IgnoreCase:
		// Value must be string
		if _, ok := f.Value.(string); !ok {
			return fmt.Errorf("%w: expected string for operator %v", ErrInvalidValue, f.Operator)
		}
	}
	return nil
}

// Validate is the exported counterpart of validate, for callers outside the package.
func (f FilterLeaf) Validate() error {
	return f.validate()
}

// ToBSON converts the leaf to a MongoDB filter document (bson.M).
func (f FilterLeaf) ToBSON() (bson.M, error) {
	if err := f.validate(); err != nil {
		return nil, err
	}

	op := f.Operator
	val := f.Value

	// Special case: IgnoreCase
	if op == IgnoreCase {
		s, ok := val.(string)
		if !ok {
			return nil, fmt.Errorf("%w: expected string for operator IgnoreCase", ErrInvalidValue)
		}
		return bson.M{f.Field: bson.M{"$regex": s, "$options": "i"}}, nil
	}

	// Contains is just Regex without options
	if op == Contains {
		op = Regex
	}

	var expr any
	switch op {
	case Equal:
		expr = val
	case NotEqual:
		expr = bson.M{"$ne": val}
	case GreaterThan:
		expr = bson.M{"$gt": val}
	case GreaterThanOrEqual:
		expr = bson.M{"$gte": val}
	case LessThan:
		expr = bson.M{"$lt": val}
	case LessThanOrEqual:
		expr = bson.M{"$lte": val}
	case In:
		expr = bson.M{"$in": val}
	case NotIn:
		expr = bson.M{"$nin": val}
	case Exists:
		expr = bson.M{"$exists": val}
	case Regex:
		if s, ok := val.(string); ok {
			expr = bson.M{"$regex": s}
		} else {
			expr = bson.M{"$regex": val}
		}
	default:
		return nil, fmt.Errorf("%w: %v", ErrUnknownOperator, op)
	}
	return bson.M{f.Field: expr}, nil
}

// FilterGroup represents a logical combination of conditions.
// It is used as a pointer (*FilterGroup) when stored as the root of a Query,
// so that Filter() can append children without silent value copies.
type FilterGroup struct {
	Operator LogicalOperator
	Children []FilterNode // each child implements FilterNode
}

// FilterNode is the common interface for FilterLeaf and FilterGroup.
type FilterNode interface {
	// ToBSON returns the MongoDB filter representation (bson.M).
	ToBSON() (bson.M, error)
	// validate recursively checks the node.
	validate() error
}

// validate ensures the group is well-formed. It satisfies the FilterNode interface.
func (g *FilterGroup) validate() error {
	if g == nil {
		return ErrEmptyGroup
	}

	if len(g.Children) == 0 {
		return ErrEmptyGroup
	}
	// Reject unknown logical operators early, before BSON rendering.
	switch g.Operator {
	case And, Or, Nor, Not:
		// known — continue
	default:
		return fmt.Errorf("%w: %v", ErrUnknownLogicalOperator, g.Operator)
	}
	if g.Operator == Not && len(g.Children) != 1 {
		return fmt.Errorf("%w: Not group must have exactly one child", ErrInvalidGroup)
	}
	for _, child := range g.Children {
		if child == nil {
			return ErrNilChild
		}

		if err := child.validate(); err != nil {
			return err
		}
	}
	return nil
}

// ToBSON converts the group to a MongoDB filter document.
//
// Note on Not: MongoDB's $not is a field-level operator and cannot negate an
// arbitrary sub-document. The idiomatic way to negate a whole condition at the
// document level is $nor: [condition], which matches documents that do NOT
// satisfy the condition. That is exactly what this implementation produces for
// a Not group, so Not behaves as a document-level logical NOT.
func (g *FilterGroup) ToBSON() (bson.M, error) {
	if err := g.validate(); err != nil {
		return nil, err
	}

	childrenBSON := make([]bson.M, 0, len(g.Children))
	for _, child := range g.Children {
		childBSON, err := child.ToBSON()
		if err != nil {
			return nil, err
		}
		childrenBSON = append(childrenBSON, childBSON)
	}

	switch g.Operator {
	case And:
		return bson.M{"$and": childrenBSON}, nil
	case Or:
		return bson.M{"$or": childrenBSON}, nil
	case Nor:
		return bson.M{"$nor": childrenBSON}, nil
	case Not:
		// MongoDB has no document-level $not; $nor:[child] is equivalent to
		// "NOT (child)" and is the correct way to negate a whole condition.
		return bson.M{"$nor": []bson.M{childrenBSON[0]}}, nil
	default:
		return nil, fmt.Errorf("%w: %v", ErrUnknownLogicalOperator, g.Operator)
	}
}
