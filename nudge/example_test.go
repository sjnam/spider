package nudge_test

import (
	"fmt"

	"github.com/sjnam/spider/nudge"
)

// Sequence lists the fence patterns a1 <= a2 >= a3 <= … in Gray order. Note it
// begins at 0001, not all-zeros: the fence needs the 000111… start (Section 3).
func ExampleSequence() {
	for bits := range nudge.Sequence(4) {
		fmt.Println(bits)
	}
	// Output:
	// [0 0 0 1]
	// [0 0 0 0]
	// [0 1 0 0]
	// [0 1 0 1]
	// [0 1 1 1]
	// [1 1 1 1]
	// [1 1 0 1]
	// [1 1 0 0]
}
