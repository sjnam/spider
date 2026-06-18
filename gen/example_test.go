package gen_test

import (
	"fmt"

	"github.com/sjnam/spider/gen"
	"github.com/sjnam/spider/spider"
)

// Sequence drives the general §5 coroutines over any spider. On the chain
// 1 -> 2 -> 3 they reduce to bump: 0 <= a1 <= a2 <= a3 in Gray order.
func ExampleSequence() {
	for bits := range gen.Sequence(spider.Chain(3)) {
		fmt.Println(bits)
	}
	// Output:
	// [0 0 0]
	// [0 0 1]
	// [0 1 1]
	// [1 1 1]
}
