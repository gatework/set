// Package set provides mutable sets of integers stored as closed ranges.
//
// A Set keeps its ranges sorted, disjoint, and non-adjacent. Operations depend
// on the number of ranges, not on the number of integers they cover: the entire
// uint64 domain can be stored in one range. Signed and unsigned integer types,
// including types defined with those underlying types, are supported.
//
// The zero value of Set is an empty set. New and Add accept endpoints in either
// order. Add and Merge update the receiver; Union, Intersection, and Clone
// return independent sets. Nil *Set receivers and arguments are not supported.
//
// Stored range slices are never modified after publication. Copying a Set value
// therefore gives an independent set, and All returns a reusable snapshot.
// Copying a *Set pointer still refers to the same mutable object. Concurrent
// reads are safe when no goroutine writes to that object; callers must
// synchronize concurrent reads and writes to the same Set.
package set
