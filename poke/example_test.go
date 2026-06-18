package poke_test

import (
	"fmt"

	"github.com/sjnam/spider/poke"
)

// Sequence yields standard reflected Gray code: every n-bit pattern once,
// changing one bit at a time.
func ExampleSequence() {
	for bits := range poke.Sequence(2) {
		fmt.Println(bits)
	}
	// Output:
	// [0 0]
	// [0 1]
	// [1 1]
	// [1 0]
}
