package loopless

import (
	"fmt"
	"iter"
	"slices"
	"testing"

	"github.com/sjnam/spider/active"
	"github.com/sjnam/spider/spider"
)

func collect[T any](seq iter.Seq[T]) []T {
	var out []T
	for v := range seq {
		out = append(out, v)
	}
	return out
}

// testSpiders returns a varied set of spiders, several given as Polish strings.
func testSpiders(t *testing.T) map[string]*spider.Spider {
	t.Helper()
	m := map[string]*spider.Spider{
		"example": spider.Example(),
		"noarcs6": spider.NoArcs(6),
		"chain7":  spider.Chain(7),
		"fence8":  spider.Fence(8),
	}
	polish := []string{
		"...+-",               // 1 -> 2 <- 3 (Section 2 example)
		"....+-.--..+-..-+",    // the running example again
		".....----",           // a positive chain of 5
		"....++-.+",           // chain nested in a near-set (the §16-bug minimal case)
		".....---..+++.+..++",  // a larger chain-in-near-set shape
		"..+....+--.+..-+-.+",  // another mixed shape
	}
	for _, p := range polish {
		s, err := spider.Parse(p)
		if err != nil {
			t.Fatalf("Parse(%q): %v", p, err)
		}
		m["polish:"+p] = s
	}
	return m
}

// TestMatchesActive checks the loopless generator produces exactly the same
// Gray listing as the (brute-force-validated) package active, pattern for
// pattern. With the corrected umaxscope/vmaxscope (see preprocess) the list
// stays sorted, so loopless picks the same largest-awake node as active.
func TestMatchesActive(t *testing.T) {
	for name, s := range testSpiders(t) {
		got := collect(Sequence(s))
		want := collect(active.Sequence(s))
		if len(got) != len(want) {
			t.Errorf("%s: loopless %d patterns, active %d", name, len(got), len(want))
			continue
		}
		for i := range got {
			if !slices.Equal(got[i], want[i]) {
				t.Errorf("%s: step %d loopless %v, active %v", name, i, got[i], want[i])
				break
			}
		}
	}
}

// TestValidGrayCode independently checks the loopless output is a complete
// single-bit Gray listing of the ideals.
func TestValidGrayCode(t *testing.T) {
	for name, s := range testSpiders(t) {
		seq := collect(Sequence(s))
		if got, want := len(seq), s.Total(); got != want {
			t.Errorf("%s: %d patterns, want %d", name, got, want)
		}
		seen := map[string]bool{}
		var prev []int
		for _, p := range seq {
			if !s.IsIdeal(p) {
				t.Errorf("%s: non-ideal %v", name, p)
			}
			key := fmt.Sprint(p)
			if seen[key] {
				t.Errorf("%s: repeat %v", name, p)
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
					t.Errorf("%s: %v->%v differ in %d bits", name, prev, p, d)
				}
			}
			prev = p
		}
	}
}
