package loopless

import (
	"math/rand"
	"slices"
	"strings"
	"testing"

	"github.com/sjnam/spider/active"
	"github.com/sjnam/spider/spider"
)

// randomPolish builds a random valid Polish description of a connected spider on
// n vertices: it emits dots and (sign-random) combine operators, never letting
// the stack drop below two before an operator, and drains to a single root.
func randomPolish(n int, rng *rand.Rand) string {
	var sb strings.Builder
	stack, dots := 0, 0
	for dots < n || stack > 1 {
		canDot := dots < n
		canOp := stack >= 2
		if canDot && (!canOp || rng.Intn(2) == 0) {
			sb.WriteByte('.')
			stack++
			dots++
		} else { // canOp
			if rng.Intn(2) == 0 {
				sb.WriteByte('+')
			} else {
				sb.WriteByte('-')
			}
			stack--
		}
	}
	return sb.String()
}

// TestRandomizedAgainstActive stresses the block/fixup logic on hundreds of
// random spider shapes, requiring the corrected loopless generator to match
// package active (which is validated against brute force) pattern for pattern.
// This is what previously failed on chain-in-near-set shapes and is now fixed by
// computing the umaxscope/vmaxscope insertion points from the transition labeling
// (see preprocess/scopeUnder).
func TestRandomizedAgainstActive(t *testing.T) {
	rng := rand.New(rand.NewSource(20260605))
	for trial := 0; trial < 500; trial++ {
		n := 1 + rng.Intn(10)
		p := randomPolish(n, rng)
		s, err := spider.Parse(p)
		if err != nil {
			t.Fatalf("randomPolish produced invalid %q: %v", p, err)
		}

		got := collect(Sequence(s))
		want := collect(active.Sequence(s))
		if len(got) != len(want) {
			t.Fatalf("spider %q: loopless %d patterns, active %d", p, len(got), len(want))
		}
		for i := range got {
			if !slices.Equal(got[i], want[i]) {
				t.Fatalf("spider %q: step %d loopless %v, active %v", p, i, got[i], want[i])
			}
		}
	}
}
