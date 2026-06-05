package spider

import "testing"

// TestTablesMatchSpidersProgram checks the progenitor and parity tables against
// the values printed by Knuth's SPIDERS program for the example spider: the §9
// table (ppro, npro, prev, umax, vmax) and the §12 table (umin, ueven, umaxbit,
// vmin, veven, vmaxbit). Node-0 entries are dummy-root bookkeeping and are not
// asserted; vertices 1..9 are unambiguous in the program's output.
func TestTablesMatchSpidersProgram(t *testing.T) {
	s := Example()

	// §9: k -> {ppro, npro, prev, umax, vmax}
	nine := map[int][5]int{
		1: {1, 0, 0, 9, 8},
		2: {2, 0, 0, 5, 4},
		3: {3, 0, 0, 0, 4},
		4: {3, 4, 0, 0, 0},
		5: {5, 0, 3, 0, 0},
		6: {6, 0, 2, 0, 7},
		7: {6, 7, 4, 0, 0},
		8: {1, 8, 7, 9, 0},
		9: {9, 8, 6, 0, 0},
	}
	for k, w := range nine {
		got := [5]int{s.ppro[k], s.npro[k], s.prv[k], s.umax[k], s.vmax[k]}
		if got != w {
			t.Errorf("k=%d: {ppro,npro,prev,umax,vmax}=%v, want %v", k, got, w)
		}
	}
	// §9 also pins the dummy root's max links.
	if s.umax[0] != 1 || s.vmax[0] != 8 {
		t.Errorf("umax[0]=%d vmax[0]=%d, want 1 and 8", s.umax[0], s.vmax[0])
	}

	// §12: k -> {umin, ueven, umaxbit, vmin, veven, vmaxbit}, with inf for ∞.
	twelve := map[int][6]int{
		1: {2, 2, 0, 4, 4, 0},
		2: {3, 5, 1, 4, 4, 1},
		3: {inf, inf, 0, 4, 4, 0},
		4: {inf, inf, 0, inf, inf, 0},
		5: {inf, inf, 0, inf, inf, 0},
		6: {inf, inf, 0, 7, 7, 0},
		7: {inf, inf, 0, inf, inf, 0},
		8: {9, 9, 1, inf, inf, 0},
		9: {inf, inf, 0, inf, inf, 0},
	}
	for k, w := range twelve {
		got := [6]int{s.umin[k], s.ueven[k], s.umaxbit[k], s.vmin[k], s.veven[k], s.vmaxbit[k]}
		if got != w {
			t.Errorf("k=%d: {umin,ueven,umaxbit,vmin,veven,vmaxbit}=%v, want %v", k, got, w)
		}
	}
}

// TestUMaxIsLargestOfU cross-checks umax/vmax/umin/vmin against the explicit
// near-sets U_k/V_k for several spiders, so the implicit representation agrees
// with the recursive one we already trust.
func TestUMaxIsLargestOfU(t *testing.T) {
	spiders := []*Spider{Example(), Chain(7), Fence(8), NoArcs(6)}
	for _, s := range spiders {
		for k := 1; k <= s.N(); k++ {
			u := s.U(k)
			wantUmax, wantUmin := 0, inf
			if len(u) > 0 {
				wantUmax, wantUmin = u[len(u)-1], u[0]
			}
			if s.umax[k] != wantUmax {
				t.Errorf("umax[%d]=%d, want %d (U=%v)", k, s.umax[k], wantUmax, u)
			}
			if s.umin[k] != wantUmin {
				t.Errorf("umin[%d]=%d, want %d (U=%v)", k, s.umin[k], wantUmin, u)
			}
			v := s.V(k)
			wantVmax, wantVmin := 0, inf
			if len(v) > 0 {
				wantVmax, wantVmin = v[len(v)-1], v[0]
			}
			if s.vmax[k] != wantVmax {
				t.Errorf("vmax[%d]=%d, want %d (V=%v)", k, s.vmax[k], wantVmax, v)
			}
			if s.vmin[k] != wantVmin {
				t.Errorf("vmin[%d]=%d, want %d (V=%v)", k, s.vmin[k], wantVmin, v)
			}
		}
	}
}
