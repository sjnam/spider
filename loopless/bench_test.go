package loopless

import (
	"testing"

	"github.com/sjnam/spider/active"
	"github.com/sjnam/spider/spider"
)

// BenchmarkGenerate compares the amortized-O(1) active list against the
// loopless generator across a few spiders, timing one full forward listing
// (New + Next until the listing completes) per iteration.
func BenchmarkGenerate(b *testing.B) {
	cases := []struct {
		name string
		s    *spider.Spider
	}{
		{"example", spider.Example()},
		{"noarcs16", spider.NoArcs(16)},
		{"fence16", spider.Fence(16)},
		{"chain16", spider.Chain(16)},
	}
	for _, c := range cases {
		b.Run("active/"+c.name, func(b *testing.B) {
			for i := 0; i < b.N; i++ {
				g := active.New(c.s)
				for g.Next() {
				}
			}
		})
		b.Run("loopless/"+c.name, func(b *testing.B) {
			for i := 0; i < b.N; i++ {
				g := New(c.s)
				for g.Next() {
				}
			}
		})
	}
}
