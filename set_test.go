package set_test

import (
	"fmt"
	"math"
	"slices"
	"testing"

	"github.com/gatework/set"
)

func TestNew(t *testing.T) {
	t.Parallel()
	tests := []struct {
		name  string
		input []set.Range[int]
		want  []set.Range[int]
	}{
		{name: "empty"},
		{
			name:  "zero singleton",
			input: []set.Range[int]{{}},
			want:  []set.Range[int]{{Min: 0, Max: 0}},
		},
		{
			name:  "reversed endpoints",
			input: []set.Range[int]{{Min: 4, Max: 1}},
			want:  []set.Range[int]{{Min: 1, Max: 4}},
		},
		{
			name:  "unordered disjoint ranges",
			input: []set.Range[int]{{Min: 7, Max: 10}, {Min: 4, Max: 1}},
			want:  []set.Range[int]{{Min: 1, Max: 4}, {Min: 7, Max: 10}},
		},
		{
			name:  "overlapping ranges",
			input: []set.Range[int]{{Min: 1, Max: 4}, {Min: 3, Max: 10}},
			want:  []set.Range[int]{{Min: 1, Max: 10}},
		},
		{
			name:  "touching endpoints",
			input: []set.Range[int]{{Min: 4, Max: 9}, {Min: 1, Max: 4}},
			want:  []set.Range[int]{{Min: 1, Max: 9}},
		},
		{
			name:  "adjacent ranges",
			input: []set.Range[int]{{Min: 3, Max: 4}, {Min: 1, Max: 2}},
			want:  []set.Range[int]{{Min: 1, Max: 4}},
		},
		{
			name: "nested and repeated ranges",
			input: []set.Range[int]{
				{Min: 3, Max: 7}, {Min: 1, Max: 10}, {Min: 1, Max: 4}, {Min: 1, Max: 10},
			},
			want: []set.Range[int]{{Min: 1, Max: 10}},
		},
		{
			name:  "negative through zero",
			input: []set.Range[int]{{Min: -1, Max: -4}, {Min: 0, Max: 0}, {Min: 1, Max: 2}},
			want:  []set.Range[int]{{Min: -4, Max: 2}},
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			t.Parallel()
			requireRanges(t, set.New(tt.input...), tt.want)
		})
	}
}

func TestAdd(t *testing.T) {
	t.Parallel()
	tests := []struct {
		name string
		lo   int
		hi   int
		want []set.Range[int]
	}{
		{
			name: "before", lo: -4, hi: -2,
			want: []set.Range[int]{{Min: -4, Max: -2}, {Min: 1, Max: 4}, {Min: 7, Max: 10}},
		},
		{
			name: "after", lo: 15, hi: 12,
			want: []set.Range[int]{{Min: 1, Max: 4}, {Min: 7, Max: 10}, {Min: 12, Max: 15}},
		},
		{
			name: "contained", lo: 3, hi: 2,
			want: []set.Range[int]{{Min: 1, Max: 4}, {Min: 7, Max: 10}},
		},
		{
			name: "bridge", lo: 5, hi: 6,
			want: []set.Range[int]{{Min: 1, Max: 10}},
		},
		{
			name: "overlap both", lo: 3, hi: 9,
			want: []set.Range[int]{{Min: 1, Max: 10}},
		},
		{
			name: "cover all", lo: -1, hi: 20,
			want: []set.Range[int]{{Min: -1, Max: 20}},
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			t.Parallel()
			s := set.New(set.Range[int]{Min: 1, Max: 4}, set.Range[int]{Min: 7, Max: 10})
			s.Add(tt.lo, tt.hi)
			requireRanges(t, s, tt.want)
		})
	}

	var zero set.Set[int]
	requireRanges(t, &zero, nil)
	zero.Add(0, 0)
	requireRanges(t, &zero, []set.Range[int]{{Min: 0, Max: 0}})
}

func TestSetAlgebra(t *testing.T) {
	t.Parallel()
	// These ranges come from the Perl interSet regression case.
	leftRanges := []set.Range[int]{{Min: 1, Max: 2}, {Min: 4, Max: 10}, {Min: 12, Max: 15}}
	rightRanges := []set.Range[int]{{Min: 3, Max: 7}, {Min: 9, Max: 16}}
	left, right := set.New(leftRanges...), set.New(rightRanges...)
	wantUnion := []set.Range[int]{{Min: 1, Max: 16}}
	wantIntersection := []set.Range[int]{{Min: 4, Max: 7}, {Min: 9, Max: 10}, {Min: 12, Max: 15}}
	requireRanges(t, left.Union(right), wantUnion)
	requireRanges(t, left.Intersection(right), wantIntersection)
	merged := left.Clone()
	merged.Merge(right)
	requireRanges(t, merged, wantUnion)
	requireRanges(t, left, leftRanges)
	requireRanges(t, right, rightRanges)
	left.Merge(left)
	requireRanges(t, left, leftRanges)

	var empty set.Set[int]
	requireRanges(t, left.Union(&empty), leftRanges)
	requireRanges(t, empty.Union(left), leftRanges)
	requireRanges(t, left.Intersection(&empty), nil)
	requireRanges(t, empty.Intersection(left), nil)
	empty.Merge(left)
	requireRanges(t, &empty, leftRanges)
}

func TestContains(t *testing.T) {
	t.Parallel()
	s := set.New(set.Range[int]{Min: -4, Max: -2}, set.Range[int]{Min: 0, Max: 3})
	for point := -6; point <= 5; point++ {
		want := (point >= -4 && point <= -2) || (point >= 0 && point <= 3)
		if got := s.Contains(point); got != want {
			t.Fatalf("Contains(%d) = %v, want %v", point, got, want)
		}
	}
}

func TestSetRelations(t *testing.T) {
	t.Parallel()
	base := set.New(set.Range[int]{Min: 1, Max: 4}, set.Range[int]{Min: 7, Max: 10})
	tests := []struct {
		name  string
		other []set.Range[int]
		want  set.Relation
	}{
		{
			name:  "equal despite segmentation",
			other: []set.Range[int]{{Min: 7, Max: 10}, {Min: 1, Max: 2}, {Min: 3, Max: 4}},
			want:  set.RelationEqual,
		},
		{
			name:  "strict superset",
			other: []set.Range[int]{{Min: 1, Max: 4}, {Min: 8, Max: 9}},
			want:  set.RelationSuperset,
		},
		{
			name:  "strict subset",
			other: []set.Range[int]{{Min: 1, Max: 4}, {Min: 7, Max: 11}},
			want:  set.RelationSubset,
		},
		{
			name:  "partial overlap",
			other: []set.Range[int]{{Min: 1, Max: 5}, {Min: 8, Max: 9}},
			want:  set.RelationOther,
		},
		{
			name:  "disjoint",
			other: []set.Range[int]{{Min: 5, Max: 6}},
			want:  set.RelationOther,
		},
		{name: "contains empty", want: set.RelationSuperset},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			t.Parallel()
			requireRelation(t, base, set.New(tt.other...), tt.want)
		})
	}
	requireRelation(t, set.New[int](), set.New[int](), set.RelationEqual)
	requireRelation(t, set.New[int](), base, set.RelationSubset)
}

func TestSetIsolation(t *testing.T) {
	t.Parallel()
	input := make([]set.Range[int], 3, 8)
	copy(input, []set.Range[int]{{Min: 8, Max: 7}, {Min: 1, Max: 2}, {Min: 3, Max: 4}})
	savedInput := slices.Clone(input)
	original := set.New(input...)
	if !slices.Equal(input, savedInput) {
		t.Fatal("New modified its caller's input slice")
	}
	input[0] = set.Range[int]{Min: -100, Max: 100}
	returned := original.Ranges()
	returned[0] = set.Range[int]{Min: -100, Max: 100}
	want := []set.Range[int]{{Min: 1, Max: 4}, {Min: 7, Max: 8}}
	requireRanges(t, original, want)

	valueCopy := *original
	empty := set.New[int]()
	copies := []*set.Set[int]{
		original.Clone(),
		&valueCopy,
		original.Union(empty),
		empty.Union(original),
		original.Intersection(original),
	}
	merged := set.New[int]()
	merged.Merge(original)
	copies = append(copies, merged)
	original.Add(5, 6)
	original.Merge(set.New(set.Range[int]{Min: 9, Max: 10}))
	for _, independent := range copies {
		requireRanges(t, independent, want)
		independent.Add(-10, 0)
		independent.Merge(set.New(set.Range[int]{Min: 100, Max: 200}))
		requireRanges(t, original, []set.Range[int]{{Min: 1, Max: 10}})
	}

	for _, result := range []*set.Set[int]{
		original.Intersection(empty), empty.Intersection(original), empty.Union(empty), empty.Clone(),
	} {
		result.Add(99, 99)
		requireRanges(t, empty, nil)
		requireRanges(t, original, []set.Range[int]{{Min: 1, Max: 10}})
	}
}

func TestAll(t *testing.T) {
	t.Parallel()
	want := []set.Range[int]{{Min: 1, Max: 2}, {Min: 7, Max: 8}}
	s := set.New(want...)
	snapshot := s.All()
	s.Add(3, 6)
	s.Merge(set.New(set.Range[int]{Min: 9, Max: 10}))
	for range 2 {
		if got := slices.Collect(snapshot); !slices.Equal(got, want) {
			t.Fatalf("snapshot = %v, want %v", got, want)
		}
	}
	var calls int
	snapshot(func(r set.Range[int]) bool {
		calls++
		r.Min = -100 // The yielded value belongs to the caller.
		s.Add(r.Min, r.Max)
		return false
	})
	if calls != 1 {
		t.Fatalf("early exit yielded %d ranges", calls)
	}
	if got := slices.Collect(snapshot); !slices.Equal(got, want) {
		t.Fatalf("mutation during iteration changed snapshot: %v", got)
	}
	var empty set.Set[int]
	emptySnapshot := empty.All()
	empty.Add(0, 0)
	if got := slices.Collect(emptySnapshot); len(got) != 0 {
		t.Fatalf("empty snapshot changed: %v", got)
	}
}

type port uint16
type signedInteger int64

func (signedInteger) String() string { return "custom" }
func (signedInteger) Format(state fmt.State, _ rune) {
	_, _ = state.Write([]byte("custom"))
}

func TestString(t *testing.T) {
	t.Parallel()
	tests := []struct {
		name string
		set  fmt.Stringer
		want string
	}{
		{name: "empty", set: set.New[int](), want: "{}"},
		{
			name: "ranges",
			set:  set.New(set.Range[int]{Min: 7, Max: 10}, set.Range[int]{Min: -4, Max: 1}),
			want: "{[-4,1], [7,10]}",
		},
		{
			name: "full signed domain",
			set:  set.New(set.Range[int64]{Min: math.MinInt64, Max: math.MaxInt64}),
			want: "{[-9223372036854775808,9223372036854775807]}",
		},
		{
			name: "full unsigned domain",
			set:  set.New(set.Range[uint64]{Min: 0, Max: math.MaxUint64}),
			want: "{[0,18446744073709551615]}",
		},
		{
			name: "custom formatting does not affect endpoints",
			set:  set.New(set.Range[signedInteger]{Min: -2, Max: 3}),
			want: "{[-2,3]}",
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			t.Parallel()
			if got := tt.set.String(); got != tt.want {
				t.Fatalf("String() = %q, want %q", got, tt.want)
			}
		})
	}
}

func TestIntegerTypes(t *testing.T) {
	t.Parallel()
	t.Run("int", func(t *testing.T) {
		t.Parallel()
		testIntegerBounds[int](t, math.MinInt, math.MaxInt)
	})
	t.Run("int8", func(t *testing.T) {
		t.Parallel()
		testIntegerBounds[int8](t, math.MinInt8, math.MaxInt8)
	})
	t.Run("int16", func(t *testing.T) {
		t.Parallel()
		testIntegerBounds[int16](t, math.MinInt16, math.MaxInt16)
	})
	t.Run("int32", func(t *testing.T) {
		t.Parallel()
		testIntegerBounds[int32](t, math.MinInt32, math.MaxInt32)
	})
	t.Run("int64", func(t *testing.T) {
		t.Parallel()
		testIntegerBounds[int64](t, math.MinInt64, math.MaxInt64)
	})
	t.Run("uint", func(t *testing.T) {
		t.Parallel()
		testIntegerBounds[uint](t, 0, math.MaxUint)
	})
	t.Run("uint8", func(t *testing.T) {
		t.Parallel()
		testIntegerBounds[uint8](t, 0, math.MaxUint8)
	})
	t.Run("uint16", func(t *testing.T) {
		t.Parallel()
		testIntegerBounds[uint16](t, 0, math.MaxUint16)
	})
	t.Run("uint32", func(t *testing.T) {
		t.Parallel()
		testIntegerBounds[uint32](t, 0, math.MaxUint32)
	})
	t.Run("uint64", func(t *testing.T) {
		t.Parallel()
		testIntegerBounds[uint64](t, 0, math.MaxUint64)
	})
	t.Run("uintptr", func(t *testing.T) {
		t.Parallel()
		testIntegerBounds[uintptr](t, 0, ^uintptr(0))
	})
	t.Run("named unsigned", func(t *testing.T) {
		t.Parallel()
		testIntegerBounds[port](t, 0, math.MaxUint16)
	})
	t.Run("named signed", func(t *testing.T) {
		t.Parallel()
		testIntegerBounds[signedInteger](t, math.MinInt64, math.MaxInt64)
	})
}

func testIntegerBounds[T set.Integer](t *testing.T, lo, hi T) {
	t.Helper()
	fullRange := []set.Range[T]{{Min: lo, Max: hi}}
	full := set.New(set.Range[T]{Min: hi, Max: lo})
	requireRanges(t, full, fullRange)
	if !full.Contains(lo) || !full.Contains(hi) {
		t.Fatal("full domain does not contain its endpoints")
	}
	endpoints := []set.Range[T]{{Min: lo, Max: lo}, {Min: hi, Max: hi}}
	edges := set.New(endpoints...)
	requireRanges(t, edges, endpoints)
	if edges.Contains(lo+1) || edges.Contains(hi-1) {
		t.Fatal("separated endpoints cover a gap")
	}
	requireRanges(t, edges.Intersection(full), endpoints)
	requireRanges(t, full.Union(edges), fullRange)
	requireRelation(t, edges, full, set.RelationSubset)
	edges.Add(lo+1, lo+1)
	edges.Add(hi-1, hi-1)
	requireRanges(t, edges, []set.Range[T]{{Min: lo, Max: lo + 1}, {Min: hi - 1, Max: hi}})
	edges.Add(hi-2, lo+2)
	requireRanges(t, edges, fullRange)
	edges.Merge(edges)
	requireRanges(t, edges, fullRange)
	requireRanges(t, set.New(endpoints[0]).Intersection(set.New(endpoints[1])), nil)
}

func TestExhaustiveSmallDomain(t *testing.T) {
	t.Parallel()
	const count = 1 << 6
	sets := make([]*set.Set[int], count)
	models := make([]pointModel, count)
	for mask := range count {
		s := set.New[int]()
		for bit := 5; bit >= 0; bit-- {
			if mask&(1<<bit) != 0 {
				point := bit - 2
				s.Add(point, point)
				models[mask][point+128] = true
			}
		}
		sets[mask] = s
		requireModel(t, s, models[mask])
		s.Merge(s)
		requireModel(t, s, models[mask])
	}
	for left := range count {
		for right := range count {
			a, b := sets[left], sets[right]
			requireModel(t, a.Union(b), models[left|right])
			requireModel(t, a.Intersection(b), models[left&right])
			merged := a.Clone()
			merged.Merge(b)
			requireModel(t, merged, models[left|right])
			requireRelation(t, a, b, modelRelation(models[left], models[right]))
		}
	}
	for i, s := range sets {
		requireModel(t, s, models[i])
	}
}

// pointModel is an independent oracle for the finite domain [-128, 127].
// It performs pointwise operations without using any interval algorithm.
type pointModel [256]bool

func (m *pointModel) add(lo, hi int8) {
	first, last := int(lo)+128, int(hi)+128
	if first > last {
		first, last = last, first
	}
	for i := first; i <= last; i++ {
		m[i] = true
	}
}

func modelRelation(left, right pointModel) set.Relation {
	var leftOnly, rightOnly bool
	for i := range left {
		leftOnly = leftOnly || (left[i] && !right[i])
		rightOnly = rightOnly || (right[i] && !left[i])
	}
	switch {
	case leftOnly && rightOnly:
		return set.RelationOther
	case leftOnly:
		return set.RelationSuperset
	case rightOnly:
		return set.RelationSubset
	default:
		return set.RelationEqual
	}
}

func requireRanges[T set.Integer](t *testing.T, s *set.Set[T], want []set.Range[T]) {
	t.Helper()
	got := s.Ranges()
	if !slices.Equal(got, want) {
		t.Fatalf("Ranges() = %v, want %v", got, want)
	}
	if s.Len() != len(want) || s.IsEmpty() != (len(want) == 0) {
		t.Fatalf("Len/IsEmpty disagree with expected ranges: %v", want)
	}
	var wantMin, wantMax T
	if len(want) != 0 {
		wantMin, wantMax = want[0].Min, want[len(want)-1].Max
	}
	if value, ok := s.Min(); value != wantMin || ok != (len(want) != 0) {
		t.Fatalf("Min() = (%v, %v) for expected ranges %v", value, ok, want)
	}
	if value, ok := s.Max(); value != wantMax || ok != (len(want) != 0) {
		t.Fatalf("Max() = (%v, %v) for expected ranges %v", value, ok, want)
	}
}

func requireRelation[T set.Integer](t *testing.T, left, right *set.Set[T], want set.Relation) {
	t.Helper()
	if got := left.Compare(right); got != want {
		t.Fatalf("Compare(%v, %v) = %v, want %v", left, right, got, want)
	}
	isEqual := want == set.RelationEqual
	isSubset := want == set.RelationSubset
	isSuperset := want == set.RelationSuperset
	checks := []struct {
		name string
		got  bool
		want bool
	}{
		{name: "equal", got: left.Equal(right), want: isEqual},
		{name: "subset", got: left.IsSubset(right), want: isEqual || isSubset},
		{name: "superset", got: left.IsSuperset(right), want: isEqual || isSuperset},
		{name: "proper subset", got: left.IsProperSubset(right), want: isSubset},
		{name: "proper superset", got: left.IsProperSuperset(right), want: isSuperset},
	}
	for _, check := range checks {
		if check.got != check.want {
			t.Fatalf("%s(%v, %v) = %v, want %v", check.name, left, right, check.got, check.want)
		}
	}
}

// requireModel inspects every covered point and the representation, then checks
// membership separately so Contains cannot conceal a malformed stored range.
func requireModel[T set.Integer](t *testing.T, s *set.Set[T], want pointModel) {
	t.Helper()
	var got pointModel
	ranges := s.Ranges()
	for i, r := range ranges {
		lo, hi := int(r.Min), int(r.Max)
		if lo < -128 || hi > 127 || lo > hi {
			t.Fatalf("invalid range for bounded model: %v", r)
		}
		if i > 0 && int(ranges[i-1].Max)+1 >= lo {
			t.Fatalf("ranges are not normalized: %v", ranges)
		}
		for value := lo; value <= hi; value++ {
			got[value+128] = true
		}
	}
	if got != want {
		t.Fatalf("ranges %v disagree with independent point model", ranges)
	}
	var wantMin, wantMax T
	var hasPoints bool
	for i, included := range want {
		value := T(i - 128)
		if s.Contains(value) != included {
			t.Fatalf("Contains(%v) disagrees with point model for %v", value, s)
		}
		if included {
			if !hasPoints {
				wantMin = value
				hasPoints = true
			}
			wantMax = value
		}
	}
	if s.Len() != len(ranges) || s.IsEmpty() == hasPoints {
		t.Fatal("Len/IsEmpty disagree with point model")
	}
	if value, ok := s.Min(); value != wantMin || ok != hasPoints {
		t.Fatal("Min disagrees with point model")
	}
	if value, ok := s.Max(); value != wantMax || ok != hasPoints {
		t.Fatal("Max disagrees with point model")
	}
}

func FuzzSet(f *testing.F) {
	for _, seed := range [][]byte{
		{},
		{0, 0, 0},
		{1, 3, 1},
		{0, 128, 127},
		{0, 128, 128, 1, 127, 127},
		{0, 1, 2, 0, 3, 4, 1, 2, 3},
		{0, 1, 4, 0, 7, 10, 1, 8, 9, 1, 3, 5},
	} {
		f.Add(seed)
	}
	f.Fuzz(func(t *testing.T, data []byte) {
		// Bound work per case while retaining up to 128 arbitrary ranges.
		if len(data) > 384 {
			data = data[:384]
		}
		left, right := set.New[int8](), set.New[int8]()
		var leftModel, rightModel pointModel
		leftRanges := []set.Range[int8]{}
		rightRanges := []set.Range[int8]{}
		for i := 0; i+2 < len(data); i += 3 {
			lo, hi := int8(data[i+1]), int8(data[i+2])
			r := set.Range[int8]{Min: lo, Max: hi}
			if data[i]&1 == 0 {
				left.Add(lo, hi)
				leftModel.add(lo, hi)
				leftRanges = append(leftRanges, r)
				continue
			}
			right.Add(lo, hi)
			rightModel.add(lo, hi)
			rightRanges = append(rightRanges, r)
		}
		requireModel(t, left, leftModel)
		requireModel(t, right, rightModel)
		requireModel(t, set.New(leftRanges...), leftModel)
		requireModel(t, set.New(rightRanges...), rightModel)
		requireRelation(t, left, right, modelRelation(leftModel, rightModel))
		var unionModel, intersectionModel pointModel
		for i := range unionModel {
			unionModel[i] = leftModel[i] || rightModel[i]
			intersectionModel[i] = leftModel[i] && rightModel[i]
		}
		union, intersection := left.Union(right), left.Intersection(right)
		requireModel(t, union, unionModel)
		requireModel(t, intersection, intersectionModel)
		clone := left.Clone()
		valueCopy := *left
		snapshot := slices.Collect(left.All())
		iterator := left.All()
		left.Merge(right)
		requireModel(t, left, unionModel)
		left.Merge(left)
		requireModel(t, left, unionModel)
		left.Add(math.MinInt8, math.MaxInt8)
		right.Add(math.MinInt8, math.MaxInt8)
		requireModel(t, clone, leftModel)
		requireModel(t, &valueCopy, leftModel)
		requireModel(t, union, unionModel)
		requireModel(t, intersection, intersectionModel)
		if got := slices.Collect(iterator); !slices.Equal(got, snapshot) {
			t.Fatalf("snapshot changed after mutation: %v", got)
		}
	})
}
