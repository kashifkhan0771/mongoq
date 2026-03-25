package mongoq

import "errors"

var (
	ErrEmptyKey               = errors.New("mongoq: filter key cannot be empty")
	ErrNilValue               = errors.New("mongoq: filter value cannot be nil")
	ErrInvalidValue           = errors.New("mongoq: invalid filter value type")
	ErrUnknownOperator        = errors.New("mongoq: unknown operator")
	ErrUnknownLogicalOperator = errors.New("mongoq: unknown logical operator")
	ErrEmptyGroup             = errors.New("mongoq: filter group must have at least one child")
	ErrNilChild               = errors.New("mongoq: filter group child cannot be nil")
	ErrInvalidGroup           = errors.New("mongoq: invalid filter group structure")
)
