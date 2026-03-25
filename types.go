package mongoq

// Operator defines comparison and array operators for a leaf filter.
type Operator int

const (
	// Equal matches values exactly.
	Equal Operator = iota
	// NotEqual matches values that are not equal.
	NotEqual
	// GreaterThan matches values greater than the given value.
	GreaterThan
	// GreaterThanOrEqual matches values greater than or equal to the given value.
	GreaterThanOrEqual
	// LessThan matches values less than the given value.
	LessThan
	// LessThanOrEqual matches values less than or equal to the given value.
	LessThanOrEqual
	// In matches values that exist in the provided array.
	In
	// NotIn matches values that do not exist in the provided array.
	NotIn
	// Exists matches documents where the field exists (true) or does not exist (false).
	Exists
	// Regex matches a regular expression pattern.
	Regex
	// Contains matches documents where the field value matches the given regular
	// expression pattern. It is identical to Regex but named to signal intent
	// (e.g. "does this field contain X?"). Use ".*foo.*" for a substring match.
	// Note: the value is a raw regex pattern, NOT a SQL LIKE wildcard string.
	Contains
	// IgnoreCase matches a regular expression pattern case-insensitively
	// (equivalent to Regex with the MongoDB $options:"i" flag).
	IgnoreCase
)

// LogicalOperator defines how groups of conditions are combined.
type LogicalOperator int

const (
	// And combines conditions with a logical AND.
	And LogicalOperator = iota
	// Or combines conditions with a logical OR.
	Or
	// Nor combines conditions with a logical NOR (not OR).
	Nor
	// Not negates a single condition (only valid for groups with exactly one child).
	Not
)

// SortOrder defines the direction of sorting.
type SortOrder int

const (
	// Asc sorts in ascending order (1).
	Asc SortOrder = iota
	// Desc sorts in descending order (-1).
	Desc
)

// ToInt returns the MongoDB sort value (1 for Asc, -1 for Desc).
func (s SortOrder) ToInt() int {
	if s == Desc {
		return -1
	}
	return 1
}
