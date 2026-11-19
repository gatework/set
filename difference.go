package set

// Difference returns an independent set containing the integers of s that do
// not belong to other. Neither input is modified. It takes O(n+m) time and
// O(n+m) space.
func (s *Set[T]) Difference(other *Set[T]) *Set[T] {
	result := &Set[T]{}
	j := 0
	for _, r := range s.ranges {
		lo := r.Min
		open := true
		for j < len(other.ranges) && other.ranges[j].Max < lo {
			j++
		}
		for k := j; k < len(other.ranges) && other.ranges[k].Min <= r.Max; k++ {
			cut := other.ranges[k]
			if cut.Min > lo {
				result.ranges = append(result.ranges, Range[T]{Min: lo, Max: cut.Min - 1})
			}
			if cut.Max >= r.Max {
				open = false
				break
			}
			lo = cut.Max + 1
		}
		if open {
			result.ranges = append(result.ranges, Range[T]{Min: lo, Max: r.Max})
		}
	}
	return result
}

// Size returns the number of integers in s and true, or false when the count
// does not fit in a uint64 (only the full 64-bit domain overflows).
func (s *Set[T]) Size() (uint64, bool) {
	var total uint64
	for _, r := range s.ranges {
		width := uint64(r.Max) - uint64(r.Min)
		if width == ^uint64(0) {
			return 0, false
		}
		next := total + width + 1
		if next < total {
			return 0, false
		}
		total = next
	}
	return total, true
}
