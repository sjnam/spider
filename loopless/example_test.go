package loopless_test

import (
	"fmt"

	"github.com/sjnam/spider/loopless"
	"github.com/sjnam/spider/spider"
)

// Sequence lists every order ideal of an arbitrary spider in Gray order. The
// three-vertex spider 1 -> 2 <- 3 ("...+-" in Polish notation) has the five
// ideals Knuth gives in the SPIDERS introduction.
func ExampleSequence() {
	s, _ := spider.Parse("...+-")
	for bits := range loopless.Sequence(s) {
		fmt.Println(bits)
	}
	// Output:
	// [0 0 0]
	// [0 1 0]
	// [0 1 1]
	// [1 1 1]
	// [1 1 0]
}
