package spider_test

import (
	"fmt"

	"github.com/sjnam/spider/spider"
)

// Parse reads a spider from the right-Polish notation of Knuth's SPIDERS.
func ExampleParse() {
	s, _ := spider.Parse("...+-") // the digraph 1 -> 2 <- 3
	fmt.Println(s.Total())
	// Output: 5
}

// The near-sets and ideal count of the paper's 9-vertex running example.
func ExampleSpider_U() {
	s := spider.Example()
	fmt.Println(s.U(1), s.V(1), s.Total())
	// Output: [2 6 9] [4 7 8] 60
}
