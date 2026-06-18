package active_test

import (
	"fmt"

	"github.com/sjnam/spider/active"
	"github.com/sjnam/spider/spider"
)

// Sequence lists every order ideal of a spider in Gray-code order. For the
// chain 1 -> 2 -> 3 these are the monotone patterns 0 <= a1 <= a2 <= a3.
func ExampleSequence() {
	for bits := range active.Sequence(spider.Chain(3)) {
		fmt.Println(bits)
	}
	// Output:
	// [0 0 0]
	// [0 0 1]
	// [0 1 1]
	// [1 1 1]
}
