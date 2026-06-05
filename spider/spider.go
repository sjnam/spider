// Package spider models a "spider" — a totally acyclic digraph from Knuth &
// Ruskey's spider-squishing paper — and the static quantities the generation
// algorithms are built on: the children, signs, and scope of each vertex, the
// near-sets U_k and V_k, and the ideal counts n_k.
//
// A spider's vertices are numbered 1..n in preorder (when the digraph is
// regarded as a forest), so the descendants of k occupy the contiguous range
// [k, scope(k)]. A virtual vertex 0 is the parent of every component root, with
// an arc 0 -> root, making the roots positive children of 0.
//
// Sign convention. For a non-root vertex k with parent p, the tree edge is one
// of the digraph arcs. If the arc is p -> k (constraint a_p <= a_k) then k is a
// positive child of p; if the arc is k -> p (constraint a_k <= a_p) then k is a
// negative child. Component roots are treated as positive children of vertex 0.
package spider

import "slices"

// Spider is an immutable spider with its derived quantities precomputed.
type Spider struct {
	n        int
	parent   []int   // parent[k], 0 for component roots; index 0 unused
	positive []bool  // positive[k]: k is a positive child of its parent
	children [][]int // children[k], in increasing (preorder) order
	roots    []int   // component roots (children of virtual vertex 0)
	scope    []int   // scope[k]: largest vertex in the subtree rooted at k
	u, v     [][]int // u[k]=U_k, v[k]=V_k (sorted)
	count    []int   // count[k]=n_k, the number of ideals of subspider k

	// Section-6 launch data, each indexed [k][vertex], meaningful on the range
	// [k, scope(k)]. alpha/omega are the first/last patterns of the Gray path
	// G_k; tau is the transition (with tau[k] = -1 marking the bit that flips).
	alpha, omega, tau [][]int

	// Implicit U_k/V_k representation and parity tables, ported from Knuth's
	// SPIDERS program (§7-§12). Index 0 is the dummy root. See tables.go.
	rchild, lsib     []int // rightmost child, left sibling (0 = none)
	ppro, npro       []int // positive / negative progenitors
	prv              []int // previous element in the same progenitorial list
	umax, vmax       []int // largest elements of U_k, V_k (0 = empty)
	umin, vmin       []int // smallest elements of U_k, V_k (inf = empty)
	ueven, veven     []int // smallest u in U_k / v in V_k with n_u / n_v even
	umaxbit, vmaxbit []int // bit[umax_k] / bit[vmax_k] at the bit[k] transition
}

// inf represents ∞ in the umin/vmin/ueven/veven tables (SPIDERS uses maxn).
const inf = 1 << 30

// New builds a spider on vertices 1..n. parent and positive are 1-indexed
// slices of length n+1 (index 0 is ignored); parent[k]==0 marks a component
// root. The numbering must be a valid preorder: each vertex's descendants form
// a contiguous block that immediately follows it.
func New(n int, parent []int, positive []bool) *Spider {
	s := &Spider{
		n:        n,
		parent:   slices.Clone(parent),
		positive: slices.Clone(positive),
		children: make([][]int, n+1),
		scope:    make([]int, n+1),
		u:        make([][]int, n+1),
		v:        make([][]int, n+1),
		count:    make([]int, n+1),
		alpha:    make([][]int, n+1),
		omega:    make([][]int, n+1),
		tau:      make([][]int, n+1),
	}
	for k := 1; k <= n; k++ {
		if p := parent[k]; p == 0 {
			s.roots = append(s.roots, k)
		} else {
			s.children[p] = append(s.children[p], k)
		}
	}
	for k := 1; k <= n; k++ {
		slices.Sort(s.children[k])
	}
	// Postorder (descending vertex number works, since children > parent and a
	// subtree is a contiguous suffix-block) to fill scope, U, V, and counts.
	for k := n; k >= 1; k-- {
		s.scope[k] = k
		for _, c := range s.children[k] {
			if s.scope[c] > s.scope[k] {
				s.scope[k] = s.scope[c]
			}
		}
		s.u[k] = s.nearU(k)
		s.v[k] = s.nearV(k)
		nP, nV := 1, 1
		for _, x := range s.u[k] {
			nP *= s.count[x]
		}
		for _, x := range s.v[k] {
			nV *= s.count[x]
		}
		s.count[k] = nP + nV
		s.computeLaunch(k)
	}
	s.computeTables()
	return s
}

// nearU computes U_k = {positive children} ∪ ⋃_{negative child c} U_c.
func (s *Spider) nearU(k int) []int {
	var out []int
	for _, c := range s.children[k] {
		if s.positive[c] {
			out = append(out, c)
		} else {
			out = append(out, s.u[c]...)
		}
	}
	slices.Sort(out)
	return out
}

// nearV computes V_k = {negative children} ∪ ⋃_{positive child c} V_c.
func (s *Spider) nearV(k int) []int {
	var out []int
	for _, c := range s.children[k] {
		if !s.positive[c] {
			out = append(out, c)
		} else {
			out = append(out, s.v[c]...)
		}
	}
	slices.Sort(out)
	return out
}

// N reports the number of vertices.
func (s *Spider) N() int { return s.n }

// Roots returns the component roots (children of the virtual vertex 0).
func (s *Spider) Roots() []int { return slices.Clone(s.roots) }

// Parent returns the parent of k (0 for a component root).
func (s *Spider) Parent(k int) int { return s.parent[k] }

// IsPositive reports whether k is a positive child of its parent.
func (s *Spider) IsPositive(k int) bool { return s.positive[k] }

// Children returns the children of k in increasing order.
func (s *Spider) Children(k int) []int { return slices.Clone(s.children[k]) }

// Scope returns scope(k), the largest vertex in subspider k.
func (s *Spider) Scope(k int) int { return s.scope[k] }

// U returns U_k, the positive vertices near k.
func (s *Spider) U(k int) []int { return slices.Clone(s.u[k]) }

// V returns V_k, the negative vertices near k.
func (s *Spider) V(k int) []int { return slices.Clone(s.v[k]) }

// Count returns n_k, the number of order ideals of subspider k.
func (s *Spider) Count(k int) int { return s.count[k] }

// Total returns the number of order ideals of the whole spider, the product of
// n_r over the component roots r.
func (s *Spider) Total() int {
	total := 1
	for _, r := range s.roots {
		total *= s.count[r]
	}
	return total
}

// constraints returns the list of pairs (j, k) meaning a_j <= a_k, one per arc.
func (s *Spider) constraints() [][2]int {
	var cs [][2]int
	for k := 1; k <= s.n; k++ {
		p := s.parent[k]
		if p == 0 {
			continue
		}
		if s.positive[k] { // arc p -> k: a_p <= a_k
			cs = append(cs, [2]int{p, k})
		} else { // arc k -> p: a_k <= a_p
			cs = append(cs, [2]int{k, p})
		}
	}
	return cs
}

// IsIdeal reports whether the pattern a (length n, with a[k-1] = a_k) satisfies
// every arc constraint a_j <= a_k.
func (s *Spider) IsIdeal(a []int) bool {
	for _, c := range s.constraints() {
		if a[c[0]-1] > a[c[1]-1] {
			return false
		}
	}
	return true
}

// AllIdeals returns every order ideal, by brute force over all 2^n patterns,
// in lexicographic order. It is the simple reference enumerator used to
// validate the Gray-code generators; intended for small n.
func (s *Spider) AllIdeals() [][]int {
	cs := s.constraints()
	var out [][]int
	a := make([]int, s.n)
	var rec func(i int)
	rec = func(i int) {
		if i == s.n {
			ok := true
			for _, c := range cs {
				if a[c[0]-1] > a[c[1]-1] {
					ok = false
					break
				}
			}
			if ok {
				out = append(out, slices.Clone(a))
			}
			return
		}
		for b := 0; b <= 1; b++ {
			a[i] = b
			rec(i + 1)
		}
	}
	rec(0)
	return out
}
