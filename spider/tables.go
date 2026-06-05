package spider

import "slices"

// This file ports Knuth's SPIDERS program (§7-§12): an implicit, linear-time
// representation of the near-sets U_k/V_k via "progenitors", plus the parity
// tables that the loopless active-list generator needs. The active list only
// ever uses the largest/smallest elements of U_k and V_k and a few parity bits,
// so the full sets (which can total Ω(n^2) elements) are never materialized.
//
// Convention note: SPIDERS uses sign[k] = 0 for a positive child and 1 for a
// negative child — the opposite boolean from our positive[k]. sgn() bridges the
// two; the dummy root 0 is treated as negative (sign 0 ⇒ positive).

// sgn returns SPIDERS' sign[k]: 0 if k is positive, 1 if negative. Vertex 0 is
// regarded as negative.
func (s *Spider) sgn(k int) int {
	if k != 0 && s.positive[k] {
		return 0
	}
	return 1
}

func (s *Spider) computeTables() {
	n := s.n
	s.ppro = make([]int, n+1)
	s.npro = make([]int, n+1)
	s.prv = make([]int, n+1)
	s.umax = make([]int, n+1)
	s.vmax = make([]int, n+1)
	s.umin = make([]int, n+1)
	s.vmin = make([]int, n+1)
	s.ueven = make([]int, n+1)
	s.veven = make([]int, n+1)
	s.umaxbit = make([]int, n+1)
	s.vmaxbit = make([]int, n+1)
	s.buildTreeLinks()
	s.progenitors() // §7
	s.fillMaxLinks() // §8
	s.evenTables()   // §10
	s.maxbits()      // §11
}

// buildTreeLinks fills rchild and lsib (§3), including the dummy root 0 whose
// children are the component roots.
func (s *Spider) buildTreeLinks() {
	n := s.n
	s.rchild = make([]int, n+1)
	s.lsib = make([]int, n+1)
	kids := make([][]int, n+1)
	for k := 1; k <= n; k++ {
		p := s.parent[k] // 0 for a component root
		kids[p] = append(kids[p], k)
	}
	for p := 0; p <= n; p++ {
		slices.Sort(kids[p])
		prev := 0
		for _, c := range kids[p] {
			s.lsib[c] = prev
			prev = c
		}
		if len(kids[p]) > 0 {
			s.rchild[p] = kids[p][len(kids[p])-1]
		}
	}
	s.scope[0] = n // the dummy root spans the whole forest
}

// Tables is a read-only snapshot of the static SPIDERS tables (§3-§12) that the
// loopless active-list generator (package loopless) needs. All slices are
// indexed 0..N, where 0 is the dummy root; Inf marks ∞ in the min tables.
type Tables struct {
	N                int
	Inf              int
	Par, Sign, Scope []int
	Rchild, Lsib     []int
	Umax, Vmax       []int
	Umin, Vmin       []int
	Ueven, Veven     []int
	Umaxbit, Vmaxbit []int
	Bit0             []int // initial labeling, indexed 0..N with Bit0[0]=0
}

// Tables returns the static tables the loopless generator is built on.
func (s *Spider) Tables() Tables {
	n := s.n
	sign := make([]int, n+1)
	for k := 0; k <= n; k++ {
		sign[k] = s.sgn(k)
	}
	bit := make([]int, n+1)
	init := s.Initial()
	copy(bit[1:], init)
	return Tables{
		N: n, Inf: inf,
		Par: slices.Clone(s.parent), Sign: sign, Scope: slices.Clone(s.scope),
		Rchild: slices.Clone(s.rchild), Lsib: slices.Clone(s.lsib),
		Umax: slices.Clone(s.umax), Vmax: slices.Clone(s.vmax),
		Umin: slices.Clone(s.umin), Vmin: slices.Clone(s.vmin),
		Ueven: slices.Clone(s.ueven), Veven: slices.Clone(s.veven),
		Umaxbit: slices.Clone(s.umaxbit), Vmaxbit: slices.Clone(s.vmaxbit),
		Bit0: bit,
	}
}

// progenitors computes ppro, npro, prev, and (as a side effect) seeds umax/vmax
// with the largest element of each progenitorial list — §7.
func (s *Spider) progenitors() {
	for j := 1; j <= s.n; j++ {
		k := s.parent[j]
		if s.sgn(j) == 0 { // positive
			s.ppro[j] = j
			s.npro[j] = s.npro[k]
			if k != 0 {
				s.prv[j] = s.umax[s.ppro[k]]
				s.umax[s.ppro[k]] = j
			} else {
				s.prv[j] = s.lsib[j] // special case: j is a root
			}
		} else { // negative
			s.npro[j] = j
			s.ppro[j] = s.ppro[k]
			s.prv[j] = s.vmax[s.npro[k]]
			s.vmax[s.npro[k]] = j
		}
	}
}

// fillMaxLinks completes umax/vmax for every vertex by a reverse-postorder
// traversal (§8): postorder visits nodes in order of their scopes, so a single
// pointer per progenitorial list suffices.
func (s *Spider) fillMaxLinks() {
	ptr := make([]int, s.n+1)
	s.lsib[0] = -1 // sentinel
	ptr[0] = s.vmax[0]
	s.umax[0] = s.rchild[0]
	for j := s.rchild[0]; ; {
		if s.sgn(j) == 0 { // positive: run a pointer through U_j to find vmax[j]
			ptr[j] = s.umax[j]
			k := s.npro[j]
			l := ptr[k]
			for l > s.scope[j] {
				l = s.prv[l]
			}
			ptr[k] = l
			if l > j {
				s.vmax[j] = l
			}
		} else { // negative: run a pointer through V_j to find umax[j]
			ptr[j] = s.vmax[j]
			k := s.ppro[j]
			l := ptr[k]
			for l > s.scope[j] {
				l = s.prv[l]
			}
			ptr[k] = l
			if l > j {
				s.umax[j] = l
			}
		}
		if s.rchild[j] != 0 {
			j = s.rchild[j]
		} else {
			for s.lsib[j] == 0 {
				j = s.parent[j]
			}
			j = s.lsib[j]
			if j < 0 {
				break
			}
		}
	}
}

// evenTables computes umin/vmin and ueven/veven (§10): the smallest, and
// smallest-with-even-count, elements of U_k and V_k. ueven[0] is forced to ∞ so
// that the digraph's components stay independent.
func (s *Spider) evenTables() {
	for k := 0; k <= s.n; k++ {
		s.ueven[k], s.veven[k], s.umin[k], s.vmin[k] = inf, inf, inf, inf
	}
	for j := s.n; j > 0; j-- {
		k := s.ppro[j]
		if s.umin[k] <= s.scope[j] {
			s.umin[j] = s.umin[k]
		}
		if s.ueven[k] <= s.scope[j] {
			s.ueven[j] = s.ueven[k]
		}
		k = s.npro[j]
		if s.vmin[k] <= s.scope[j] {
			s.vmin[j] = s.vmin[k]
		}
		if s.veven[k] <= s.scope[j] {
			s.veven[j] = s.veven[k]
		}
		// l = n_j mod 2: n_j = ∏U_j + ∏V_j, and ∏ is even iff an even factor
		// exists (ueven/veven < inf), so the parity is just their XOR.
		l := b2i(s.ueven[j] < inf) ^ b2i(s.veven[j] < inf)
		k = s.parent[j]
		if s.sgn(j) == 0 {
			s.umin[s.ppro[k]] = j
			if l == 0 {
				s.ueven[s.ppro[k]] = j
			}
		} else {
			s.vmin[s.npro[k]] = j
			if l == 0 {
				s.veven[s.npro[k]] = j
			}
		}
	}
	s.ueven[0] = inf
}

// maxbits computes umaxbit/vmaxbit (§11): the value of bit[umax[k]] (resp.
// bit[vmax[k]]) at the moment bit[k] flips from 0 to 1.
func (s *Spider) maxbits() {
	for k := s.n; k > 0; k-- {
		l := s.parent[k]
		if k == s.umax[l] {
			s.umaxbit[l] = 1
		} else if j := s.umax[k]; j != 0 && s.umax[l] == j {
			if s.ueven[k] < j { // δ_jk even
				s.umaxbit[l] = s.umaxbit[k]
			} else {
				s.umaxbit[l] = 1 ^ s.umaxbit[k]
			}
		}
		if k == s.vmax[l] {
			s.vmaxbit[l] = 0
		} else if j := s.vmax[k]; j != 0 && s.vmax[l] == j {
			if s.veven[k] < j {
				s.vmaxbit[l] = s.vmaxbit[k]
			} else {
				s.vmaxbit[l] = 1 ^ s.vmaxbit[k]
			}
		}
	}
}

func b2i(b bool) int {
	if b {
		return 1
	}
	return 0
}
