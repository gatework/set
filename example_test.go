package set_test

import (
	"fmt"

	"github.com/gatework/set"
)

func Example() {
	var ports set.Set[uint16]
	ports.Add(80, 80)
	ports.Add(443, 443)
	ports.Add(8080, 8082)
	ports.Add(8083, 8085)

	allowed := set.New(set.Range[uint16]{Min: 440, Max: 450})
	fmt.Println(&ports)
	fmt.Println(ports.Contains(443))
	fmt.Println(ports.Intersection(allowed))
	fmt.Println(ports.Len())
	// Output:
	// {[80,80], [443,443], [8080,8085]}
	// true
	// {[443,443]}
	// 3
}

func ExampleSet_All() {
	numbers := set.New(set.Range[int]{Min: -2, Max: 0}, set.Range[int]{Min: 4, Max: 5})
	snapshot := numbers.All()
	numbers.Add(1, 3)
	for r := range snapshot {
		fmt.Println(r.Min, r.Max)
	}
	fmt.Println(numbers)
	// Output:
	// -2 0
	// 4 5
	// {[-2,5]}
}

func ExampleSet_Clone() {
	original := set.New(set.Range[int]{Min: 1, Max: 4})
	cloned := original.Clone()
	cloned.Merge(set.New(set.Range[int]{Min: 5, Max: 8}))
	fmt.Println(original)
	fmt.Println(cloned)
	// Output:
	// {[1,4]}
	// {[1,8]}
}
