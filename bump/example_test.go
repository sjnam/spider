package bump_test

import (
	"fmt"

	"github.com/sjnam/spider/bump"
)

// Sequence lists the chain patterns 0 <= a1 <= … <= an <= 1 in Gray order.
func ExampleSequence() {
	for bits := range bump.Sequence(3) {
		fmt.Println(bits)
	}
	// Output:
	// [0 0 0]
	// [0 0 1]
	// [0 1 1]
	// [1 1 1]
}
