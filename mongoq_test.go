package mongoq

import (
	"reflect"
	"testing"

	"go.mongodb.org/mongo-driver/bson"
)

// ── helpers ──────────────────────────────────────────────────────────────────

// bsonEqual compares two bson.M values for semantic equality, regardless of
// key insertion order. It round-trips both values through BSON bytes into
// map[string]interface{} (which bson.Unmarshal fills with Go maps, not ordered
// bson.D slices) and then uses reflect.DeepEqual for a recursive, order-
// independent comparison.
//
// This correctly handles nested documents like {$regex:…, $options:…} where
// key order varies between the want literal and the value actually produced.
func bsonEqual(t *testing.T, want, got bson.M) bool {
	t.Helper()

	normalise := func(label string, m bson.M) map[string]interface{} {
		b, err := bson.Marshal(m)
		if err != nil {
			t.Fatalf("bsonEqual: marshal %s: %v", label, err)
		}
		var out map[string]interface{}
		if err := bson.Unmarshal(b, &out); err != nil {
			t.Fatalf("bsonEqual: unmarshal %s: %v", label, err)
		}
		return out
	}

	w, g := normalise("want", want), normalise("got", got)
	if !reflect.DeepEqual(w, g) {
		t.Errorf("bson mismatch\n  want: %v\n   got: %v", want, got)
		return false
	}
	return true
}

// ── FilterLeaf.Validate ───────────────────────────────────────────────────────

func TestFilterLeaf_Validate_EmptyKey(t *testing.T) {
	leaf := FilterLeaf{Field: "", Operator: Equal, Value: "x"}
	if err := leaf.Validate(); err != ErrEmptyKey {
		t.Errorf("want ErrEmptyKey, got %v", err)
	}
}

func TestFilterLeaf_Validate_NilValueForEqual(t *testing.T) {
	leaf := FilterLeaf{Field: "name", Operator: Equal, Value: nil}
	if err := leaf.Validate(); err != nil {
		t.Errorf("want nil, got %v", err)
	}
}

func TestFilterLeaf_Validate_NilValue(t *testing.T) {
	leaf := FilterLeaf{Field: "name", Operator: GreaterThan, Value: nil}
	if err := leaf.Validate(); err != ErrNilValue {
		t.Errorf("want ErrNilValue, got %v", err)
	}
}

func TestFilterLeaf_Validate_InRequiresSlice(t *testing.T) {
	leaf := FilterLeaf{Field: "age", Operator: In, Value: 42} // int, not a slice
	if err := leaf.Validate(); err == nil {
		t.Error("want error for In with non-slice value, got nil")
	}
}

func TestFilterLeaf_Validate_InAcceptsTypedSlice(t *testing.T) {
	// []string must pass — the old []any assertion would have rejected this.
	leaf := FilterLeaf{Field: "role", Operator: In, Value: []string{"a", "b"}}
	if err := leaf.Validate(); err != nil {
		t.Errorf("want nil for typed []string slice, got %v", err)
	}
}

func TestFilterLeaf_Validate_UnknownOperator(t *testing.T) {
	leaf := FilterLeaf{Field: "age", Operator: Operator(999), Value: 1}
	if err := leaf.Validate(); err == nil {
		t.Error("want error for unknown operator, got nil")
	}
}

func TestFilterGroup_Validate_UnknownLogicalOperator(t *testing.T) {
	group := &FilterGroup{
		Operator: LogicalOperator(999),
		Children: []FilterNode{FilterLeaf{Field: "x", Operator: Equal, Value: 1}},
	}
	if _, err := group.ToBSON(); err == nil {
		t.Error("want error for unknown logical operator, got nil")
	}
}

func TestFilterLeaf_Validate_ExistsRequiresBool(t *testing.T) {
	leaf := FilterLeaf{Field: "age", Operator: Exists, Value: "yes"} // not bool
	if err := leaf.Validate(); err == nil {
		t.Error("want error for Exists with non-bool value, got nil")
	}
}

func TestFilterLeaf_Validate_RegexRequiresString(t *testing.T) {
	leaf := FilterLeaf{Field: "name", Operator: Regex, Value: 123}
	if err := leaf.Validate(); err == nil {
		t.Error("want error for Regex with non-string value, got nil")
	}
}

func TestFilterLeaf_Validate_Valid(t *testing.T) {
	leaf := FilterLeaf{Field: "name", Operator: Equal, Value: "alice"}
	if err := leaf.Validate(); err != nil {
		t.Errorf("want nil, got %v", err)
	}
}

// ── FilterLeaf.ToBSON ─────────────────────────────────────────────────────────

func TestFilterLeaf_ToBSON_Equal(t *testing.T) {
	leaf := FilterLeaf{Field: "age", Operator: Equal, Value: 30}
	got, err := leaf.ToBSON()
	if err != nil {
		t.Fatal(err)
	}
	bsonEqual(t, bson.M{"age": 30}, got)
}

func TestFilterLeaf_ToBSON_NotEqual(t *testing.T) {
	leaf := FilterLeaf{Field: "age", Operator: NotEqual, Value: 30}
	got, err := leaf.ToBSON()
	if err != nil {
		t.Fatal(err)
	}
	bsonEqual(t, bson.M{"age": bson.M{"$ne": 30}}, got)
}

func TestFilterLeaf_ToBSON_GreaterThan(t *testing.T) {
	leaf := FilterLeaf{Field: "score", Operator: GreaterThan, Value: 50}
	got, err := leaf.ToBSON()
	if err != nil {
		t.Fatal(err)
	}
	bsonEqual(t, bson.M{"score": bson.M{"$gt": 50}}, got)
}

func TestFilterLeaf_ToBSON_LessThanOrEqual(t *testing.T) {
	leaf := FilterLeaf{Field: "score", Operator: LessThanOrEqual, Value: 100}
	got, err := leaf.ToBSON()
	if err != nil {
		t.Fatal(err)
	}
	bsonEqual(t, bson.M{"score": bson.M{"$lte": 100}}, got)
}

func TestFilterLeaf_ToBSON_In(t *testing.T) {
	// Use a typed []string (not []any) to verify the reflect-based slice
	// check accepts common typed slices. Go slice invariance means []string
	// fails a []any type assertion even though it is valid for $in.
	vals := []string{"admin", "user"}
	leaf := FilterLeaf{Field: "role", Operator: In, Value: vals}
	got, err := leaf.ToBSON()
	if err != nil {
		t.Fatal(err)
	}
	bsonEqual(t, bson.M{"role": bson.M{"$in": vals}}, got)
}

func TestFilterLeaf_ToBSON_NotIn(t *testing.T) {
	vals := []any{"banned"}
	leaf := FilterLeaf{Field: "role", Operator: NotIn, Value: vals}
	got, err := leaf.ToBSON()
	if err != nil {
		t.Fatal(err)
	}
	bsonEqual(t, bson.M{"role": bson.M{"$nin": vals}}, got)
}

func TestFilterLeaf_ToBSON_Exists(t *testing.T) {
	leaf := FilterLeaf{Field: "email", Operator: Exists, Value: true}
	got, err := leaf.ToBSON()
	if err != nil {
		t.Fatal(err)
	}
	bsonEqual(t, bson.M{"email": bson.M{"$exists": true}}, got)
}

func TestFilterLeaf_ToBSON_Regex(t *testing.T) {
	leaf := FilterLeaf{Field: "name", Operator: Regex, Value: "^alice"}
	got, err := leaf.ToBSON()
	if err != nil {
		t.Fatal(err)
	}
	bsonEqual(t, bson.M{"name": bson.M{"$regex": "^alice"}}, got)
}

func TestFilterLeaf_ToBSON_Like(t *testing.T) {
	// Contains should behave identically to Regex (no options flag)
	leaf := FilterLeaf{Field: "name", Operator: Contains, Value: "alice"}
	got, err := leaf.ToBSON()
	if err != nil {
		t.Fatal(err)
	}
	bsonEqual(t, bson.M{"name": bson.M{"$regex": "alice"}}, got)
}

func TestFilterLeaf_ToBSON_IgnoreCase(t *testing.T) {
	leaf := FilterLeaf{Field: "name", Operator: IgnoreCase, Value: "alice"}
	got, err := leaf.ToBSON()
	if err != nil {
		t.Fatal(err)
	}
	bsonEqual(t, bson.M{"name": bson.M{"$regex": "alice", "$options": "i"}}, got)
}

// ── FilterGroup.ToBSON ────────────────────────────────────────────────────────

func TestFilterGroup_ToBSON_And(t *testing.T) {
	group := &FilterGroup{
		Operator: And,
		Children: []FilterNode{
			FilterLeaf{Field: "age", Operator: GreaterThan, Value: 18},
			FilterLeaf{Field: "active", Operator: Equal, Value: true},
		},
	}
	got, err := group.ToBSON()
	if err != nil {
		t.Fatal(err)
	}
	want := bson.M{
		"$and": []bson.M{
			{"age": bson.M{"$gt": 18}},
			{"active": true},
		},
	}
	bsonEqual(t, want, got)
}

func TestFilterGroup_ToBSON_Or(t *testing.T) {
	group := &FilterGroup{
		Operator: Or,
		Children: []FilterNode{
			FilterLeaf{Field: "role", Operator: Equal, Value: "admin"},
			FilterLeaf{Field: "role", Operator: Equal, Value: "superuser"},
		},
	}
	got, err := group.ToBSON()
	if err != nil {
		t.Fatal(err)
	}
	want := bson.M{
		"$or": []bson.M{
			{"role": "admin"},
			{"role": "superuser"},
		},
	}
	bsonEqual(t, want, got)
}

func TestFilterGroup_ToBSON_Nor(t *testing.T) {
	group := &FilterGroup{
		Operator: Nor,
		Children: []FilterNode{
			FilterLeaf{Field: "status", Operator: Equal, Value: "banned"},
		},
	}
	got, err := group.ToBSON()
	if err != nil {
		t.Fatal(err)
	}
	bsonEqual(t, bson.M{"$nor": []bson.M{{"status": "banned"}}}, got)
}

func TestFilterGroup_ToBSON_Not(t *testing.T) {
	// Not must have exactly one child and produces $nor:[child]
	group := &FilterGroup{
		Operator: Not,
		Children: []FilterNode{
			FilterLeaf{Field: "deleted", Operator: Equal, Value: true},
		},
	}
	got, err := group.ToBSON()
	if err != nil {
		t.Fatal(err)
	}
	bsonEqual(t, bson.M{"$nor": []bson.M{{"deleted": true}}}, got)
}

func TestFilterGroup_ToBSON_Not_TooManyChildren(t *testing.T) {
	// Not with >1 child must return an error
	group := &FilterGroup{
		Operator: Not,
		Children: []FilterNode{
			FilterLeaf{Field: "a", Operator: Equal, Value: 1},
			FilterLeaf{Field: "b", Operator: Equal, Value: 2},
		},
	}
	if _, err := group.ToBSON(); err == nil {
		t.Error("want error for Not group with 2 children, got nil")
	}
}

func TestFilterGroup_ToBSON_EmptyChildren(t *testing.T) {
	group := &FilterGroup{Operator: And, Children: nil}
	if _, err := group.ToBSON(); err != ErrEmptyGroup {
		t.Errorf("want ErrEmptyGroup, got %v", err)
	}
}

// ── Query builder ─────────────────────────────────────────────────────────────

func TestQuery_NoFilter(t *testing.T) {
	q := NewQuery()
	filter, _, err := q.Build()
	if err != nil {
		t.Fatal(err)
	}
	// BuildFilter returns bson.M{} (not nil) so Collection.Find receives a
	// valid match-all document rather than ErrNilDocument.
	if len(filter) != 0 {
		t.Errorf("want empty bson.M{} for empty query, got %v", filter)
	}
}

func TestQuery_SingleFilter(t *testing.T) {
	q := NewQuery().Filter("age", GreaterThan, 18)
	filter, _, err := q.Build()
	if err != nil {
		t.Fatal(err)
	}
	// Single filter is wrapped in an And group
	want := bson.M{"$and": []bson.M{{"age": bson.M{"$gt": 18}}}}
	bsonEqual(t, want, filter)
}

func TestQuery_MultipleFilters_MergedIntoAndGroup(t *testing.T) {
	// BUG REGRESSION: calling Filter() multiple times must accumulate into
	// a single And group, not nest And groups inside each other.
	q := NewQuery().
		Filter("age", GreaterThan, 18).
		Filter("active", Equal, true)

	filter, _, err := q.Build()
	if err != nil {
		t.Fatal(err)
	}

	andArr, ok := filter["$and"].([]bson.M)
	if !ok {
		t.Fatalf("expected $and array, got %T: %v", filter["$and"], filter)
	}
	if len(andArr) != 2 {
		t.Errorf("expected 2 conditions in $and, got %d — possible double-nesting bug", len(andArr))
	}
}

func TestQuery_Where_CustomNode(t *testing.T) {
	node := &FilterGroup{
		Operator: Or,
		Children: []FilterNode{
			FilterLeaf{Field: "role", Operator: Equal, Value: "admin"},
			FilterLeaf{Field: "role", Operator: Equal, Value: "mod"},
		},
	}
	q := NewQuery().Where(node)
	filter, _, err := q.Build()
	if err != nil {
		t.Fatal(err)
	}
	want := bson.M{
		"$or": []bson.M{
			{"role": "admin"},
			{"role": "mod"},
		},
	}
	bsonEqual(t, want, filter)
}

func TestQuery_LimitOffsetSort(t *testing.T) {
	q := NewQuery().
		Filter("active", Equal, true).
		Limit(10).
		Offset(20).
		Sort("name", Asc).
		Sort("age", Desc)

	_, opts, err := q.Build()
	if err != nil {
		t.Fatal(err)
	}
	if opts.Limit == nil || *opts.Limit != 10 {
		t.Errorf("want limit=10, got %v", opts.Limit)
	}
	if opts.Skip == nil || *opts.Skip != 20 {
		t.Errorf("want skip=20, got %v", opts.Skip)
	}

	sortDoc, ok := opts.Sort.(bson.D)
	if !ok {
		t.Fatalf("expected bson.D sort, got %T", opts.Sort)
	}
	if len(sortDoc) != 2 {
		t.Fatalf("expected 2 sort fields, got %d", len(sortDoc))
	}
	if sortDoc[0].Key != "name" || sortDoc[0].Value != 1 {
		t.Errorf("first sort: want name=1, got %v=%v", sortDoc[0].Key, sortDoc[0].Value)
	}
	if sortDoc[1].Key != "age" || sortDoc[1].Value != -1 {
		t.Errorf("second sort: want age=-1, got %v=%v", sortDoc[1].Key, sortDoc[1].Value)
	}
}

func TestQuery_NoLimitNoSkip_NotSetInOptions(t *testing.T) {
	q := NewQuery().Filter("x", Equal, 1)
	_, opts, err := q.Build()
	if err != nil {
		t.Fatal(err)
	}
	if opts.Limit != nil {
		t.Errorf("want nil limit when not set, got %v", *opts.Limit)
	}
	if opts.Skip != nil {
		t.Errorf("want nil skip when not set, got %v", *opts.Skip)
	}
}

func TestQuery_Project_BsonM(t *testing.T) {
	q := NewQuery().Project(bson.M{"name": 1, "email": 1})
	_, opts, err := q.Build()
	if err != nil {
		t.Fatal(err)
	}
	if opts.Projection == nil {
		t.Error("expected projection to be set")
	}
}

func TestQuery_Project_MapStringInt(t *testing.T) {
	q := NewQuery().Project(map[string]int{"name": 1, "email": 1})
	_, opts, err := q.Build()
	if err != nil {
		t.Fatal(err)
	}
	if opts.Projection == nil {
		t.Error("expected projection to be set for map[string]int")
	}
}

func TestQuery_Project_UnsupportedType_ReturnsError(t *testing.T) {
	// An unsupported type must not be silently ignored; Build() must return
	// an error so the caller is not left with a missing projection and
	// unexpectedly full documents returned from MongoDB.
	type customProjection struct{ Name int }
	q := NewQuery().Project(customProjection{Name: 1})
	_, _, err := q.Build()
	if err == nil {
		t.Error("want error for unsupported projection type, got nil")
	}
}

func TestQuery_Project_UnsupportedType_DoesNotPolluteFutureBuilds(t *testing.T) {
	// Even after a bad Project call, Build() should consistently return an
	// error — not silently succeed on a second call.
	type bad struct{}
	q := NewQuery().Filter("x", Equal, 1).Project(bad{})
	for i := 0; i < 2; i++ {
		_, _, err := q.Build()
		if err == nil {
			t.Errorf("call %d: want error, got nil", i+1)
		}
	}
}

func TestQuery_InvalidFilter_PropagatesError(t *testing.T) {
	// Empty key should cause Build() to return an error, not panic
	q := NewQuery().Filter("", Equal, "value")
	_, _, err := q.Build()
	if err == nil {
		t.Error("want error for empty field key, got nil")
	}
}

// ── SortOrder.ToInt ───────────────────────────────────────────────────────────

func TestSortOrder_ToInt(t *testing.T) {
	if Asc.ToInt() != 1 {
		t.Errorf("want Asc=1, got %d", Asc.ToInt())
	}
	if Desc.ToInt() != -1 {
		t.Errorf("want Desc=-1, got %d", Desc.ToInt())
	}
}
