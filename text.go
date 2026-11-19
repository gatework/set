package set

import (
	"encoding"
	"errors"
	"fmt"
	"strconv"
	"strings"
)

// ErrSyntax reports text that is not a valid range list.
var ErrSyntax = errors.New("set: invalid range list")

var (
	_ encoding.TextMarshaler   = (*Set[int])(nil)
	_ encoding.TextUnmarshaler = (*Set[int])(nil)
)

// Parse reads a comma-separated list of integers and inclusive ranges, the
// notation network devices use for VLAN and port lists, for example
// "1-4,7,10-12". Whitespace around items is ignored and items may overlap or
// appear in any order. The empty string is the empty set. Negative numbers
// are written with a leading minus sign; a range separator is the first '-'
// that follows a digit.
func Parse[T Integer](text string) (*Set[T], error) {
	s := &Set[T]{}
	if strings.TrimSpace(text) == "" {
		return s, nil
	}
	ranges := make([]Range[T], 0, strings.Count(text, ",")+1)
	for item := range strings.SplitSeq(text, ",") {
		r, err := parseRange[T](strings.TrimSpace(item))
		if err != nil {
			return nil, err
		}
		ranges = append(ranges, r)
	}
	return New(ranges...), nil
}

// MarshalText formats s in the compact notation read by Parse, for example
// "1-4,7". The empty set is the empty string.
func (s *Set[T]) MarshalText() ([]byte, error) {
	return []byte(s.Compact()), nil
}

// UnmarshalText replaces the contents of s with the parsed range list.
func (s *Set[T]) UnmarshalText(text []byte) error {
	parsed, err := Parse[T](string(text))
	if err != nil {
		return err
	}
	s.ranges = parsed.ranges
	return nil
}

// Compact formats s as a comma-separated range list such as "1-4,7", the
// inverse of Parse.
func (s *Set[T]) Compact() string {
	var b strings.Builder
	for i, r := range s.ranges {
		if i > 0 {
			_ = b.WriteByte(',')
		}
		_, _ = b.WriteString(formatInteger(r.Min))
		if r.Max != r.Min {
			_ = b.WriteByte('-')
			_, _ = b.WriteString(formatInteger(r.Max))
		}
	}
	return b.String()
}

func parseRange[T Integer](item string) (Range[T], error) {
	if item == "" {
		return Range[T]{}, fmt.Errorf("%w: empty item", ErrSyntax)
	}
	split := -1
	for i := 1; i < len(item); i++ {
		if item[i] == '-' && item[i-1] >= '0' && item[i-1] <= '9' {
			split = i
			break
		}
	}
	if split < 0 {
		v, err := parseInteger[T](item)
		if err != nil {
			return Range[T]{}, err
		}
		return Range[T]{Min: v, Max: v}, nil
	}
	lo, err := parseInteger[T](strings.TrimSpace(item[:split]))
	if err != nil {
		return Range[T]{}, err
	}
	hi, err := parseInteger[T](strings.TrimSpace(item[split+1:]))
	if err != nil {
		return Range[T]{}, err
	}
	if lo > hi {
		return Range[T]{}, fmt.Errorf("%w: descending range %q", ErrSyntax, item)
	}
	return Range[T]{Min: lo, Max: hi}, nil
}

func parseInteger[T Integer](text string) (T, error) {
	var zero T
	minusOne := zero - 1
	signed := minusOne < zero
	if signed {
		v, err := strconv.ParseInt(text, 10, 64)
		if err != nil || int64(T(v)) != v {
			return zero, fmt.Errorf("%w: %q", ErrSyntax, text)
		}
		return T(v), nil
	}
	v, err := strconv.ParseUint(text, 10, 64)
	if err != nil || uint64(T(v)) != v {
		return zero, fmt.Errorf("%w: %q", ErrSyntax, text)
	}
	return T(v), nil
}
