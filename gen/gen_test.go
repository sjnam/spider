package gen_test

import (
	"iter"
	"math/rand"
	"slices"
	"strings"
	"testing"

	"github.com/sjnam/spider/active"
	"github.com/sjnam/spider/bump"
	"github.com/sjnam/spider/gen"
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

func randomPolish(n int, rng *rand.Rand) string {
	var sb strings.Builder
	stack, dots := 0, 0
	for dots < n || stack > 1 {
		if dots < n && (stack < 2 || rng.Intn(2) == 0) {
			sb.WriteByte('.')
			stack++
			dots++
		} else {
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

// TestMatchesActive is the payoff: the general gen coroutines produce the same
// Gray listing as the validated active list, on named spiders, Polish-described
// ones, and 400 random shapes.
func TestMatchesActive(t *testing.T) {
	named := map[string]*spider.Spider{
		"example": spider.Example(),
		"noarcs6": spider.NoArcs(6),
		"chain7":  spider.Chain(7),
		"fence8":  spider.Fence(8),
		"bug":     mustParse(t, "....++-.+"),
	}
	for name, s := range named {
		if got, want := collect(gen.Sequence(s)), collect(active.Sequence(s)); !equalSeqs(got, want) {
			t.Errorf("%s: gen != active\n got=%v\nwant=%v", name, got, want)
		}
	}

	rng := rand.New(rand.NewSource(20260618))
	for trial := 0; trial < 400; trial++ {
		n := 1 + rng.Intn(9)
		p := randomPolish(n, rng)
		s, err := spider.Parse(p)
		if err != nil {
			t.Fatalf("randomPolish %q: %v", p, err)
		}
		if got, want := collect(gen.Sequence(s)), collect(active.Sequence(s)); !equalSeqs(got, want) {
			t.Fatalf("spider %q: gen != active\n got=%v\nwant=%v", p, got, want)
		}
	}
}

// TestReducesToSpecialCases confirms the paper's claim that gen reduces to poke
// on the empty digraph, bump on the chain, and nudge on the fence.
func TestReducesToSpecialCases(t *testing.T) {
	for n := 1; n <= 9; n++ {
		if got, want := collect(gen.Sequence(spider.NoArcs(n))), collect(poke.Sequence(n)); !equalSeqs(got, want) {
			t.Errorf("NoArcs(%d): gen != poke", n)
		}
		if got, want := collect(gen.Sequence(spider.Chain(n))), collect(bump.Sequence(n)); !equalSeqs(got, want) {
			t.Errorf("Chain(%d): gen != bump", n)
		}
		if got, want := collect(gen.Sequence(spider.Fence(n))), collect(nudge.Sequence(n)); !equalSeqs(got, want) {
			t.Errorf("Fence(%d): gen != nudge", n)
		}
	}
}

func mustParse(t *testing.T, p string) *spider.Spider {
	t.Helper()
	s, err := spider.Parse(p)
	if err != nil {
		t.Fatalf("Parse(%q): %v", p, err)
	}
	return s
}
