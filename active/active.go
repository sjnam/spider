// Package active implements the active-list algorithm of Section 8 — the
// iterative "compilation" of the gen coroutines into an O(1)-amortized loop
// over an arbitrary spider.
//
// The active list is the set of nodes satisfying the membership rule: a
// positive child j of k is in the list iff k is the virtual root or a_k = 0; a
// negative child j of k is in the list iff a_k = 1. Each listed node is awake
// or asleep. Starting from spider.Initial() with everything awake, each call to
// Next advances one bit using three rules:
//
//  1. Let k be the largest non-sleeping node on the list; wake every larger
//     node. If all are asleep they all wake and no bit changes (Next reports
//     false — the gen coroutines returning false at a complete listing).
//  2. Flip a_k. If a_k became 1, delete k's positive children and insert its
//     negative children; otherwise insert its positive children and delete its
//     negative ones. Newly inserted nodes are awake.
//  3. Put k to sleep.
//
// The list is kept in increasing node order as a doubly linked list, so rules
// (1) and (3) and the deletions are O(1); insertion currently scans for the
// sorted position. (Section 9's focus pointers and lazy family updates make the
// whole thing loopless; that optimization is left as future work.)
package active

import (
	"iter"
	"slices"

	"github.com/sjnam/spider/spider"
)

// Gen drives the active-list generation of a spider's order ideals.
type Gen struct {
	s *spider.Spider
	n int
	a []int // a[1..n]; a[0] unused

	// Doubly linked list over 0..n+1, with 0 the head sentinel and n+1 the
	// tail sentinel; listed nodes appear in increasing order.
	next, prev    []int
	inList, awake []bool

	posKids, negKids [][]int
}

// New builds a generator seeded at the spider's initial configuration with
// every listed node awake.
func New(s *spider.Spider) *Gen {
	n := s.N()
	g := &Gen{
		s:       s,
		n:       n,
		a:       make([]int, n+1),
		next:    make([]int, n+2),
		prev:    make([]int, n+2),
		inList:  make([]bool, n+2),
		awake:   make([]bool, n+2),
		posKids: make([][]int, n+1),
		negKids: make([][]int, n+1),
	}
	init := s.Initial()
	for k := 1; k <= n; k++ {
		g.a[k] = init[k-1]
	}
	for k := 1; k <= n; k++ {
		for _, c := range s.Children(k) {
			if s.IsPositive(c) {
				g.posKids[k] = append(g.posKids[k], c)
			} else {
				g.negKids[k] = append(g.negKids[k], c)
			}
		}
	}
	// Empty list, then append the members in increasing order (all awake).
	g.next[0], g.prev[n+1] = n+1, 0
	for k := 1; k <= n; k++ {
		if g.member(k) {
			last := g.prev[n+1]
			g.next[last], g.prev[k] = k, last
			g.next[k], g.prev[n+1] = n+1, k
			g.inList[k], g.awake[k] = true, true
		}
	}
	return g
}

// member reports whether k satisfies the active-list membership rule under the
// current pattern.
func (g *Gen) member(k int) bool {
	p := g.s.Parent(k)
	if p == 0 {
		return true // a component root: positive child of the virtual vertex 0
	}
	if g.s.IsPositive(k) {
		return g.a[p] == 0
	}
	return g.a[p] == 1
}

// Next advances to the next pattern, returning false (without changing a bit)
// exactly when a complete listing wraps around.
func (g *Gen) Next() bool {
	n := g.n
	// Rule 1: largest awake node, waking the larger (asleep) ones we pass.
	node := g.prev[n+1]
	for node != 0 && !g.awake[node] {
		g.awake[node] = true
		node = g.prev[node]
	}
	if node == 0 {
		return false // all asleep (now all awake): no bit change
	}
	k := node

	// Rule 2: flip a_k and update the membership of k's children.
	if g.a[k] == 0 {
		g.a[k] = 1
		for _, j := range g.posKids[k] {
			g.del(j)
		}
		for _, j := range g.negKids[k] {
			g.ins(j)
		}
	} else {
		g.a[k] = 0
		for _, j := range g.posKids[k] {
			g.ins(j)
		}
		for _, j := range g.negKids[k] {
			g.del(j)
		}
	}

	// Rule 3: k nods off.
	g.awake[k] = false
	return true
}

// ins inserts j into the sorted list (awake), if it is not already present.
func (g *Gen) ins(j int) {
	if g.inList[j] {
		return
	}
	p := 0
	for g.next[p] < j {
		p = g.next[p]
	}
	nx := g.next[p]
	g.next[p], g.prev[j] = j, p
	g.next[j], g.prev[nx] = nx, j
	g.inList[j], g.awake[j] = true, true
}

// del removes j from the list, if present.
func (g *Gen) del(j int) {
	if !g.inList[j] {
		return
	}
	g.next[g.prev[j]] = g.next[j]
	g.prev[g.next[j]] = g.prev[j]
	g.inList[j] = false
}

// Bits returns the current pattern a[1..n] (a fresh copy).
func (g *Gen) Bits() []int { return slices.Clone(g.a[1:]) }

// Active returns the active-list nodes in order, with a parallel slice marking
// which are asleep — the data behind the right-hand column of the paper's
// tables.
func (g *Gen) Active() (nodes []int, asleep []bool) {
	for k := g.next[0]; k != g.n+1; k = g.next[k] {
		nodes = append(nodes, k)
		asleep = append(asleep, !g.awake[k])
	}
	return nodes, asleep
}

// Sequence yields one forward listing of the spider's order ideals in Gray
// order, from the initial configuration until the listing completes. Each
// yielded slice is owned by the caller.
func Sequence(s *spider.Spider) iter.Seq[[]int] {
	return func(yield func([]int) bool) {
		g := New(s)
		if !yield(g.Bits()) {
			return
		}
		for g.Next() {
			if !yield(g.Bits()) {
				return
			}
		}
	}
}
