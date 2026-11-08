package set

import (
	"cmp"
	"fmt"
	"iter"
	"slices"
	"sort"
	"strconv"
	"strings"
)

// Integer permits every built-in integer type and types derived from them.
type Integer interface {
	~int | ~int8 | ~int16 | ~int32 | ~int64 |
		~uint | ~uint8 | ~uint16 | ~uint32 | ~uint64 | ~uintptr
}

// Range describes the closed interval [Min, Max]. New normalizes reversed
// endpoints. The zero value is the singleton containing zero, not an empty range.
type Range[T Integer] struct {
	Min T
	Max T
}

// Relation describes set inclusion, not a total ordering of sets.
type Relation uint8

const (
	// RelationUnknown is the zero value; Compare never returns it.
	RelationUnknown Relation = iota
	// RelationEqual means both sets contain exactly the same integers.
	RelationEqual
	// RelationSuperset means the receiver strictly contains the other set.
	RelationSuperset
	// RelationSubset means the receiver is a strict subset of the other set.
	RelationSubset
	// RelationOther means neither set contains the other, whether they overlap or not.
	RelationOther
)

// Set stores integers as sorted, disjoint, non-adjacent closed ranges.
// Its zero value is an empty set ready for use. Value copies are independent;
// pointer copies refer to the same mutable object. See the package documentation
// for the concurrency and snapshot contracts.
type Set[T Integer] struct {
	// A nil slice represents the zero-value empty set. Once published to a
	// caller, a backing array is read-only; mutators build a replacement slice.
	ranges []Range[T]
}

var _ fmt.Stringer = (*Set[int])(nil)

// New copies and normalizes ranges, merging overlaps and adjacent intervals.
// Endpoints may be reversed, and ranges may be provided in any order.
// With no arguments it returns an empty set. It takes O(n log n) time.
func New[T Integer](ranges ...Range[T]) *Set[T] {
	owned := slices.Clone(ranges)
	for i, r := range owned {
		owned[i] = orderedRange(r.Min, r.Max)
	}
	slices.SortFunc(owned, func(a, b Range[T]) int {
		if order := cmp.Compare(a.Min, b.Min); order != 0 {
			return order
		}
		return cmp.Compare(a.Max, b.Max)
	})

	normalized := owned[:0]
	for _, r := range owned {
		normalized = appendRange(normalized, r)
	}
	return &Set[T]{ranges: normalized}
}

// Add includes the closed interval [lo, hi], swapping reversed endpoints.
// It merges overlapping and adjacent ranges and takes O(n) time and space.
func (s *Set[T]) Add(lo, hi T) {
	s.ranges = unionRanges(s.ranges, []Range[T]{orderedRange(lo, hi)})
}

// Merge includes every integer in other, leaving other unchanged unless it is
// the receiver itself. Merging a set with itself is safe and idempotent.
// It takes O(n+m) time and space.
func (s *Set[T]) Merge(other *Set[T]) {
	s.ranges = unionRanges(s.ranges, other.ranges)
}

// Clone returns a set with its own copy of the stored ranges.
func (s *Set[T]) Clone() *Set[T] {
	return &Set[T]{ranges: slices.Clone(s.ranges)}
}

// Union returns an independent set containing every integer in either input.
// Neither input is modified. It takes O(n+m) time and space.
func (s *Set[T]) Union(other *Set[T]) *Set[T] {
	return &Set[T]{ranges: unionRanges(s.ranges, other.ranges)}
}

// Intersection returns an independent set containing only integers present in
// both inputs. Neither input is modified. It takes O(n+m) time and O(k) space,
// where k is the number of output ranges.
func (s *Set[T]) Intersection(other *Set[T]) *Set[T] {
	result := &Set[T]{}
	i, j := 0, 0
	for i < len(s.ranges) && j < len(other.ranges) {
		left, right := s.ranges[i], other.ranges[j]
		if r, ok := intersectRange(left, right); ok {
			result.ranges = appendRange(result.ranges, r)
		}
		if left.Max <= right.Max {
			i++
		}
		if right.Max <= left.Max {
			j++
		}
	}
	return result
}

// Contains reports whether value belongs to s, using O(log n) time.
func (s *Set[T]) Contains(value T) bool {
	i := sort.Search(len(s.ranges), func(i int) bool {
		return s.ranges[i].Max >= value
	})
	return i < len(s.ranges) && s.ranges[i].Min <= value
}

// Equal reports whether both sets contain exactly the same integers.
func (s *Set[T]) Equal(other *Set[T]) bool {
	return slices.Equal(s.ranges, other.ranges)
}

// IsSubset reports whether every integer in s also belongs to other.
// Equality counts as inclusion; the empty set is a subset of every set.
// It takes O(n+m) time without allocating.
func (s *Set[T]) IsSubset(other *Set[T]) bool {
	return other.IsSuperset(s)
}

// IsSuperset reports whether every integer in other also belongs to s.
// Equality counts as inclusion; every set contains the empty set.
// It takes O(n+m) time without allocating.
func (s *Set[T]) IsSuperset(other *Set[T]) bool {
	i := 0
	for _, want := range other.ranges {
		for i < len(s.ranges) && s.ranges[i].Max < want.Min {
			i++
		}
		if i == len(s.ranges) {
			return false
		}
		have := s.ranges[i]
		if have.Min > want.Min || have.Max < want.Max {
			return false
		}
	}
	return true
}

// IsProperSubset reports whether s is a subset of other and is not equal to it.
func (s *Set[T]) IsProperSubset(other *Set[T]) bool {
	return !s.Equal(other) && s.IsSubset(other)
}

// IsProperSuperset reports whether s contains other and is not equal to it.
func (s *Set[T]) IsProperSuperset(other *Set[T]) bool {
	return !s.Equal(other) && s.IsSuperset(other)
}

// Compare classifies s relative to other as equal, a strict superset, a strict
// subset, or neither. RelationOther includes both disjoint and overlapping sets.
func (s *Set[T]) Compare(other *Set[T]) Relation {
	switch {
	case s.Equal(other):
		return RelationEqual
	case s.IsSuperset(other):
		return RelationSuperset
	case s.IsSubset(other):
		return RelationSubset
	default:
		return RelationOther
	}
}

// Len returns the number of normalized ranges, not the number of integers.
func (s *Set[T]) Len() int {
	return len(s.ranges)
}

// IsEmpty reports whether s contains no integers.
func (s *Set[T]) IsEmpty() bool {
	return len(s.ranges) == 0
}

// Min returns the smallest integer and true, or the zero value and false if empty.
func (s *Set[T]) Min() (T, bool) {
	if s.IsEmpty() {
		var zero T
		return zero, false
	}
	return s.ranges[0].Min, true
}

// Max returns the largest integer and true, or the zero value and false if empty.
func (s *Set[T]) Max() (T, bool) {
	if s.IsEmpty() {
		var zero T
		return zero, false
	}
	return s.ranges[len(s.ranges)-1].Max, true
}

// Ranges returns a copy of the normalized ranges, sorted by their lower bounds.
// The result may be nil for an empty set. Editing it does not modify s.
func (s *Set[T]) Ranges() []Range[T] {
	return slices.Clone(s.ranges)
}

// All returns a reusable iterator over the ranges present when All is called.
// Later mutations of s do not change this snapshot, including mutations during
// iteration. The iterator stops as soon as yield returns false.
func (s *Set[T]) All() iter.Seq[Range[T]] {
	return slices.Values(s.ranges)
}

// String formats s as ascending closed ranges, for example "{[1,4], [7,10]}".
// The empty set is "{}". Endpoints use decimal notation, including for types
// that define their own formatting methods.
func (s *Set[T]) String() string {
	var b strings.Builder
	// Writes to strings.Builder cannot fail.
	_ = b.WriteByte('{')
	for i, r := range s.ranges {
		if i > 0 {
			_, _ = b.WriteString(", ")
		}
		_ = b.WriteByte('[')
		_, _ = b.WriteString(formatInteger(r.Min))
		_ = b.WriteByte(',')
		_, _ = b.WriteString(formatInteger(r.Max))
		_ = b.WriteByte(']')
	}
	_ = b.WriteByte('}')
	return b.String()
}

func orderedRange[T Integer](lo, hi T) Range[T] {
	if lo > hi {
		lo, hi = hi, lo
	}
	return Range[T]{Min: lo, Max: hi}
}

// appendRange merges next into an owned output slice. Its caller supplies
// normalized ranges in nondecreasing order of their lower bounds.
func appendRange[T Integer](ranges []Range[T], next Range[T]) []Range[T] {
	if len(ranges) == 0 {
		return append(ranges, next)
	}
	last := &ranges[len(ranges)-1]
	// Subtract only after proving next.Min exceeds a representable integer,
	// so this cannot underflow for signed or unsigned types.
	if next.Min > last.Max && next.Min-1 > last.Max {
		return append(ranges, next)
	}
	last.Max = max(last.Max, next.Max)
	return ranges
}

func unionRanges[T Integer](left, right []Range[T]) []Range[T] {
	result := make([]Range[T], 0, max(len(left), len(right)))
	i, j := 0, 0
	for i < len(left) && j < len(right) {
		if left[i].Min <= right[j].Min {
			result = appendRange(result, left[i])
			i++
			continue
		}
		result = appendRange(result, right[j])
		j++
	}
	for _, r := range left[i:] {
		result = appendRange(result, r)
	}
	for _, r := range right[j:] {
		result = appendRange(result, r)
	}
	return result
}

func intersectRange[T Integer](left, right Range[T]) (Range[T], bool) {
	lo, hi := max(left.Min, right.Min), min(left.Max, right.Max)
	if lo > hi {
		return Range[T]{}, false
	}
	return Range[T]{Min: lo, Max: hi}, true
}

func formatInteger[T Integer](value T) string {
	if value < 0 {
		return strconv.FormatInt(int64(value), 10)
	}
	return strconv.FormatUint(uint64(value), 10)
}
