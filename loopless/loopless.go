// Package loopless is a faithful Go port of the active-list machinery in
// Knuth's CWEB program SPIDERS (§14-§29): a genuinely loopless generator of a
// spider's order ideals in Gray order. After O(n) initialization it performs a
// bounded number of operations between successive bit changes — not merely
// amortized O(1) like package active.
//
// The two tricks, both from SPIDERS, are:
//
//   - Focus pointers (Ehrlich, Algorithm 7.2.1.1L): finding the rightmost awake
//     node, waking everything to its right, and putting a node to sleep are all
//     O(1) via the focus array, instead of a scan.
//   - Lazy block fixups: when bit[k] flips, whole same-sign sibling "blocks"
//     must enter or leave a doubly linked active list. Only the rightmost block
//     is relinked immediately; a flag is planted so the remaining blocks are
//     fixed just before they are next needed.
//
// The static tables come from spider.Tables().
//
// One correction to the source. The provided spiders.c computes the umaxscope/
// vmaxscope insertion points from a near-set endpoint (vmax[j]/umax[j]). That is
// too shallow when a chain is nested inside a near-set: the largest *forced* node
// can lie deeper, so a block gets linked out of sorted order and the generator
// walks off into non-ideal labelings. Compiling that C source and adding an
// arc-constraint check exhibits it; the minimal case is the 5-vertex spider
// "....++-.+", whose 10 ideals it lists only 8 of. (This is likely not Knuth's
// final SPIDERS.)
//
// preprocess/scopeUnder fixes it by computing each insertion point directly from
// the transition labeling tau_k: umaxscope[k] is the largest node still on the
// active list at the bit[k] 0→1 transition, vmaxscope[k] the same for 1→0. With
// that one change the list stays sorted, so this generator matches package active
// (which is validated against brute force) pattern for pattern on every spider —
// see TestMatchesActive and TestRandomizedAgainstActive. Everything else is a
// faithful port of SPIDERS §13-§29; the per-bit-change work is still O(1) (the
// loopless property), only the one-time preprocessing is O(n^2) instead of O(n).
package loopless

import (
	"iter"
	"slices"

	"github.com/sjnam/spider/spider"
)

// Gen is a loopless generator over one spider.
type Gen struct {
	t   spider.Tables
	n   int
	inf int

	bit   []int // bit[0..n]; bit[0]=0 is the dummy root's bit
	left  []int // doubly linked circular active list; left[0]=rightmost member
	right []int // right[0]=leftmost member
	focus []int // Ehrlich focus pointers encoding wakefulness
	flag  []int // pending lazy fixup at a node (signed: >0 insert, <0 delete)

	bstart               []int // leftmost sibling of k's block
	umaxscope, vmaxscope []int // extreme forced nodes at a bit[k] transition

	done bool // have we hit the all-asleep state once (end of forward listing)?
}

// New builds a loopless generator seeded at the spider's initial configuration.
func New(s *spider.Spider) *Gen {
	t := s.Tables()
	n := t.N
	g := &Gen{
		t: t, n: n, inf: t.Inf,
		bit:       make([]int, n+1),
		left:      make([]int, n+1),
		right:     make([]int, n+1),
		focus:     make([]int, n+1),
		flag:      make([]int, n+1),
		bstart:    make([]int, n+1),
		umaxscope: make([]int, n+1),
		vmaxscope: make([]int, n+1),
	}
	g.preprocess() // §16-17
	g.launch()     // §23
	return g
}

// preprocess establishes the sibling links of every block (SPIDERS §16-17) and
// the umaxscope/vmaxscope insertion points.
//
// CORRECTION to §16: the original umaxscope/vmaxscope recursion is computed from
// vmax[j]/umax[j] (a near-set endpoint), which misses how deep the forced nodes
// reach when a chain is nested inside a near-set, so a block can be linked out of
// sorted order. We instead compute the insertion point directly from the actual
// transition labeling tau_k: umaxscope[k] is the largest node still in the active
// list at the bit[k] 0→1 transition (with bit[k]=0, the U side), and vmaxscope[k]
// the largest at the 1→0 transition (bit[k]=1, the V side). See scopeUnder.
func (g *Gen) preprocess() {
	t := g.t
	for k := g.n; k > 0; k-- {
		if j := t.Lsib[k]; j != 0 {
			g.left[k], g.right[j] = j, k
		} else {
			// Compute bstart for k's family by walking right through siblings.
			l := k
			for jj := k; jj != 0; jj = g.right[jj] {
				rj := g.right[jj]
				simple := (t.Sign[jj] == 0 && t.Vmax[jj] == 0) || (t.Sign[jj] == 1 && t.Umax[jj] == 0)
				if rj != 0 && t.Sign[jj] == t.Sign[rj] && simple {
					continue // jj merges with its right neighbour into one block
				}
				g.bstart[jj] = l
				l = rj
			}
		}
		g.umaxscope[k] = g.scopeUnder(k, 0)
		g.vmaxscope[k] = g.scopeUnder(k, 1)
	}
}

// scopeUnder returns the largest node in (k, scope(k)] that is on the active
// list at the bit[k] transition, when bit[k] = bk and every other node of
// spider k holds its transition value tau_k. (If none are active it returns k,
// so right[k] is used.) A node m is active iff its parent's bit equals 0 when m
// is positive, or 1 when m is negative.
func (g *Gen) scopeUnder(k, bk int) int {
	t := g.t
	tau := t.Tau[k]
	best := k
	for m := k + 1; m <= t.Scope[k]; m++ {
		p := t.Par[m]
		bp := bk
		if p != k {
			bp = tau[p]
		}
		if (t.Sign[m] == 0 && bp == 0) || (t.Sign[m] == 1 && bp == 1) {
			best = m
		}
	}
	return best
}

// setfirst/setlast/setmid compute the initial labeling, recursing over the tree
// exactly as SPIDERS §13. setfirst(k) lays down the first labeling of spider k;
// the "subtle point" is that a negative child whose ueven[k] >= j gets its
// transition bits (setmid) rather than its first bits.
func (g *Gen) setfirst(k int) {
	t := g.t
	g.bit[k] = 0
	for j := t.Rchild[k]; j != 0; j = t.Lsib[j] {
		if t.Sign[j] == 0 {
			if t.Ueven[k] >= j {
				g.setfirst(j)
			} else {
				g.setlast(j)
			}
		} else if t.Ueven[k] >= j {
			g.setmid(j, 0)
		} else {
			g.setfirst(j)
		}
	}
}

func (g *Gen) setlast(k int) {
	t := g.t
	g.bit[k] = 1
	for j := t.Rchild[k]; j != 0; j = t.Lsib[j] {
		if t.Sign[j] == 1 {
			if t.Veven[k] >= j {
				g.setlast(j)
			} else {
				g.setfirst(j)
			}
		} else if t.Veven[k] >= j {
			g.setmid(j, 1)
		} else {
			g.setlast(j)
		}
	}
}

func (g *Gen) setmid(k, b int) {
	t := g.t
	g.bit[k] = b
	for j := t.Rchild[k]; j != 0; j = t.Lsib[j] {
		if t.Sign[j] == 0 {
			g.setlast(j)
		} else {
			g.setfirst(j)
		}
	}
}

// launch computes the initial labeling and builds the active list (SPIDERS §23).
func (g *Gen) launch() {
	g.setfirst(0)
	l := 0
	for k := 0; k <= g.n; k++ {
		g.focus[k] = k
		if k > 0 && g.t.Sign[k] == g.bit[g.t.Par[k]] {
			g.right[l], g.left[k] = k, l
			l = k
		}
	}
	g.right[l], g.left[0] = 0, l
}

// Next advances one bit, returning false (without changing a bit) when a
// complete forward listing has just finished.
func (g *Gen) Next() bool {
	// §26: rightmost nonsleeping node, waking everything to its right.
	j := g.left[0]
	k := g.focus[j]
	g.focus[j] = j
	if k == 0 {
		return false // all asleep: end of the forward listing
	}
	if g.flag[k] != 0 {
		g.fixup(g.flag[k], k)
	}
	if g.bit[k] == 0 {
		g.moveForward(k) // §28
	} else {
		g.moveBackward(k) // §29
	}
	// §27: put k to sleep.
	j = g.left[k]
	g.focus[k] = g.focus[j]
	g.focus[j] = j
	return true
}

// moveForward sets bit[k] = 1, deleting k's positive child block and inserting
// its negative one (SPIDERS §28).
func (g *Gen) moveForward(k int) {
	t := g.t
	g.bit[k] = 1
	j := t.Rchild[k]
	if j == 0 {
		return
	}
	if t.Sign[j] == 0 { // delete j = umax[k]
		if l := t.Vmin[j]; l < g.inf {
			g.fixup(-j, l)
		} else {
			g.fixup(-j, g.right[j]) // j ends a simple block
		}
	} else { // insert j = vmax[k]
		if l := t.Umin[j]; l < g.inf {
			g.fixup(j, l)
		} else {
			g.fixup(j, g.right[g.umaxscope[k]])
		}
	}
}

// moveBackward sets bit[k] = 0, the dual of moveForward (SPIDERS §29).
func (g *Gen) moveBackward(k int) {
	t := g.t
	g.bit[k] = 0
	j := t.Rchild[k]
	if j == 0 {
		return
	}
	if t.Sign[j] == 1 { // delete j = vmax[k]
		if l := t.Umin[j]; l < g.inf {
			g.fixup(-j, l)
		} else {
			g.fixup(-j, g.right[j])
		}
	} else { // insert j = umax[k]
		if l := t.Vmin[j]; l < g.inf {
			g.fixup(j, l)
		} else {
			g.fixup(j, g.right[g.vmaxscope[k]])
		}
	}
}

// fixup is the basic insert/delete mechanism: it relinks one block and plants a
// flag so the previous block is fixed in due time (SPIDERS §19-22).
func (g *Gen) fixup(k, l int) {
	g.flag[l] = 0
	if k > 0 {
		g.insertBlock(k, l)
	} else {
		g.deleteBlock(k, l)
	}
}

// insertBlock inserts block k before l (SPIDERS §20).
func (g *Gen) insertBlock(k, l int) {
	t := g.t
	j := g.bstart[k]
	i := t.Lsib[j]
	g.left[j], g.right[g.left[l]] = g.left[l], j
	g.left[l], g.right[k] = k, l
	if i != 0 {
		if t.Sign[k] == 1 {
			if t.Sign[i] == 0 {
				if t.Vmin[i] < g.inf {
					j = t.Vmin[i]
				}
				i = -i // the next fix will be a deletion
			} else {
				j = t.Umin[i]
			}
		} else {
			if t.Sign[i] == 1 {
				if t.Umin[i] < g.inf {
					j = t.Umin[i]
				}
				i = -i
			} else {
				j = t.Vmin[i]
			}
		}
		g.flag[j] = i
	}
}

// deleteBlock deletes block -k before l (SPIDERS §21).
func (g *Gen) deleteBlock(k, l int) {
	t := g.t
	k = -k
	j := g.bstart[k]
	i := t.Lsib[j]
	if i != 0 && t.Sign[i] != t.Sign[k] {
		if (t.Sign[i] == 0 && t.Vmax[i] == 0) || (t.Sign[i] == 1 && t.Umax[i] == 0) {
			g.replaceBlock(k, i, j, l) // §22
			return
		}
	}
	g.left[l], g.right[g.left[j]] = g.left[j], l
	if i != 0 {
		if t.Sign[k] == 0 {
			if t.Sign[i] == 1 {
				j = t.Umin[i]
			} else {
				j = t.Vmin[i]
				i = -i
			}
		} else {
			if t.Sign[i] == 0 {
				j = t.Vmin[i]
			} else {
				j = t.Umin[i]
				i = -i
			}
		}
		g.flag[j] = i
	}
}

// replaceBlock inserts a simple opposite-sign block i in place of the block k
// being deleted (SPIDERS §22).
func (g *Gen) replaceBlock(k, i, j, l int) {
	t := g.t
	g.left[l], g.right[i] = i, l
	k = g.bstart[i]
	g.left[k], g.right[g.left[j]] = g.left[j], k
	i = t.Lsib[k]
	if i != 0 {
		if t.Sign[k] == 0 {
			if t.Sign[i] == 1 {
				if t.Umin[i] < g.inf {
					k = t.Umin[i]
				}
				i = -i
			} else {
				k = t.Vmin[i]
			}
		} else {
			if t.Sign[i] == 0 {
				if t.Vmin[i] < g.inf {
					k = t.Vmin[i]
				}
				i = -i
			} else {
				k = t.Umin[i]
			}
		}
		g.flag[k] = i
	}
}

// Bits returns the current labeling a[1..n] (a fresh copy).
func (g *Gen) Bits() []int { return slices.Clone(g.bit[1:]) }

// Active returns the active-list members in order with a parallel asleep flag,
// reconstructed from the focus pointers (SPIDERS §31).
func (g *Gen) Active() (nodes []int, asleep []bool) {
	sleep := make([]bool, g.n+1)
	for k := g.left[0]; ; k-- {
		j := k
		k = g.focus[k]
		for ; j > k; j-- {
			sleep[j] = true
		}
		if k == 0 {
			break
		}
		sleep[k] = false
	}
	for k := g.right[0]; k != 0; k = g.right[k] {
		nodes = append(nodes, k)
		asleep = append(asleep, sleep[k])
	}
	return nodes, asleep
}

// Sequence yields one forward listing of the spider's ideals in Gray order.
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
