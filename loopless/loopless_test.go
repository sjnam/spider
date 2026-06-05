package loopless

import (
	"fmt"
	"iter"
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
	// Shapes the provided spiders.c handles correctly (the poke/bump/nudge
	// analogues plus the running example). Some other shapes are mishandled by
	// that source — see the package comment — so they are intentionally absent.
	m := map[string]*spider.Spider{
		"example": spider.Example(),
		"noarcs6": spider.NoArcs(6),
		"chain7":  spider.Chain(7),
		"fence8":  spider.Fence(8),
	}
	polish := []string{
		"...+-",             // 1 -> 2 <- 3 (Section 2 example)
		"....+-.--..+-..-+", // the running example again
		".....----",         // a positive chain of 5
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

// TestSameIdealSetAsActive checks the loopless generator enumerates exactly the
// same set of ideals as the (brute-force-validated) package active. The ORDER
// may differ — SPIDERS' focus[left[0]] rule isn't always the largest awake node
// — so this compares the visited sets, not the sequences. (Order is pinned to
// the real SPIDERS program in golden_test.go.)
func TestSameIdealSetAsActive(t *testing.T) {
	for name, s := range testSpiders(t) {
		loop := map[string]bool{}
		for p := range Sequence(s) {
			loop[fmt.Sprint(p)] = true
		}
		act := map[string]bool{}
		for p := range active.Sequence(s) {
			act[fmt.Sprint(p)] = true
		}
		if len(loop) != len(act) {
			t.Errorf("%s: loopless visited %d ideals, active %d", name, len(loop), len(act))
		}
		for k := range act {
			if !loop[k] {
				t.Errorf("%s: loopless missing ideal %s", name, k)
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
