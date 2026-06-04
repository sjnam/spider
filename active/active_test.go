package active

import (
	"fmt"
	"iter"
	"slices"
	"testing"

	"github.com/sjnam/spider/bump"
	"github.com/sjnam/spider/nudge"
	"github.com/sjnam/spider/poke"
	"github.com/sjnam/spider/spider"
)

func collect[T any](seq iter.Seq[T]) []T {
	var out []T
	for v := range seq {
		out = append(out, v)
	}
	return out
}

func equalSeqs(a, b [][]int) bool {
	if len(a) != len(b) {
		return false
	}
	for i := range a {
		if !slices.Equal(a[i], b[i]) {
			return false
		}
	}
	return true
}

// TestMatchesGoroutineCoroutines is the payoff: the iterative active-list
// generator reproduces, pattern for pattern, the goroutine coroutines of
// Sections 1–3 on their respective spiders.
func TestMatchesGoroutineCoroutines(t *testing.T) {
	for n := 1; n <= 9; n++ {
		if got, want := collect(Sequence(spider.NoArcs(n))), collect(poke.Sequence(n)); !equalSeqs(got, want) {
			t.Errorf("NoArcs(%d) != poke: got %v want %v", n, got, want)
		}
		if got, want := collect(Sequence(spider.Chain(n))), collect(bump.Sequence(n)); !equalSeqs(got, want) {
			t.Errorf("Chain(%d) != bump: got %v want %v", n, got, want)
		}
		if got, want := collect(Sequence(spider.Fence(n))), collect(nudge.Sequence(n)); !equalSeqs(got, want) {
			t.Errorf("Fence(%d) != nudge: got %v want %v", n, got, want)
		}
	}
}

// isGrayListing checks a one-bit-per-step listing of distinct ideals that
// covers exactly the spider's ideal set.
func isGrayListing(t *testing.T, name string, s *spider.Spider, seq [][]int) {
	t.Helper()
	if got, want := len(seq), s.Total(); got != want {
		t.Errorf("%s: produced %d patterns, want %d", name, got, want)
	}
	seen := make(map[string]bool)
	var prev []int
	for _, p := range seq {
		if !s.IsIdeal(p) {
			t.Errorf("%s: non-ideal %v", name, p)
		}
		key := fmt.Sprint(p)
		if seen[key] {
			t.Errorf("%s: repeated pattern %v", name, p)
		}
		seen[key] = true
		if prev != nil {
			d := 0
			for i := range prev {
				if prev[i] != p[i] {
					d++
				}
			}
			if d != 1 {
				t.Errorf("%s: %v -> %v differ in %d bits, want 1", name, prev, p, d)
			}
		}
		prev = p
	}
	// Completeness: the produced set equals the brute-force ideal set.
	for _, p := range s.AllIdeals() {
		if !seen[fmt.Sprint(p)] {
			t.Errorf("%s: missing ideal %v", name, p)
		}
	}
}

// TestValidGrayCode checks the active-list output is a complete single-bit Gray
// listing of the ideals for a variety of spiders, including the example (60
// ideals) and an arbitrary mixed spider.
func TestValidGrayCode(t *testing.T) {
	spiders := map[string]*spider.Spider{
		"example": spider.Example(),
		"noarcs6": spider.NoArcs(6),
		"chain7":  spider.Chain(7),
		"fence8":  spider.Fence(8),
		"mixed":   mixedSpider(),
	}
	for name, s := range spiders {
		isGrayListing(t, name, s, collect(Sequence(s)))
	}
}

// TestExampleCount confirms the example spider yields exactly 60 patterns.
func TestExampleCount(t *testing.T) {
	if got := len(collect(Sequence(spider.Example()))); got != 60 {
		t.Errorf("Example produced %d patterns, want 60", got)
	}
}

// mixedSpider is an arbitrary totally-acyclic digraph with both positive and
// negative children, numbered in preorder:
//
//	1
//	├─2 (pos) ── 3 (neg)
//	├─4 (neg) ── 5 (pos)
//	└─6 (pos)
func mixedSpider() *spider.Spider {
	//             1   2   3   4   5   6
	parent := []int{0, 0, 1, 2, 1, 4, 1}
	positive := []bool{false, true, true, false, false, true, true}
	return spider.New(6, parent, positive)
}
