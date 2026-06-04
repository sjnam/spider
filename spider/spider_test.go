package spider

import (
	"fmt"
	"slices"
	"testing"
)

// TestExampleStructure checks the example spider against the tables on pages
// 12, 16 and 18 of the paper.
func TestExampleStructure(t *testing.T) {
	s := Example()

	if got, want := s.U(1), []int{2, 6, 9}; !slices.Equal(got, want) {
		t.Errorf("U_1 = %v, want %v", got, want)
	}
	if got, want := s.V(1), []int{4, 7, 8}; !slices.Equal(got, want) {
		t.Errorf("V_1 = %v, want %v", got, want)
	}
	if got, want := s.U(8), []int{9}; !slices.Equal(got, want) {
		t.Errorf("U_8 = %v, want %v", got, want)
	}
	if got, want := s.U(2), []int{3, 5}; !slices.Equal(got, want) {
		t.Errorf("U_2 = %v, want %v", got, want)
	}
	if got, want := s.V(2), []int{4}; !slices.Equal(got, want) {
		t.Errorf("V_2 = %v, want %v", got, want)
	}

	// scope(k) from the page-16 table.
	wantScope := map[int]int{1: 9, 2: 5, 3: 4, 4: 4, 5: 5, 6: 7, 7: 7, 8: 9, 9: 9}
	for k, want := range wantScope {
		if got := s.Scope(k); got != want {
			t.Errorf("scope(%d) = %d, want %d", k, got, want)
		}
	}

	// n_k from the page-18 table.
	wantN := map[int]int{1: 60, 2: 8, 3: 3, 4: 2, 5: 2, 6: 3, 7: 2, 8: 3, 9: 2}
	for k, want := range wantN {
		if got := s.Count(k); got != want {
			t.Errorf("n_%d = %d, want %d", k, got, want)
		}
	}

	if got := s.Total(); got != 60 {
		t.Errorf("Total() = %d, want 60", got)
	}
}

// TestCountsMatchBruteForce cross-checks the recursive ideal count against the
// brute-force enumerator, and confirms every enumerated pattern is an ideal.
func TestCountsMatchBruteForce(t *testing.T) {
	spiders := map[string]*Spider{
		"example":  Example(),
		"noarcs5":  NoArcs(5),
		"chain6":   Chain(6),
		"fence6":   Fence(6),
		"fence7":   Fence(7),
		"noarcs10": NoArcs(10),
	}
	for name, s := range spiders {
		ideals := s.AllIdeals()
		if got, want := len(ideals), s.Total(); got != want {
			t.Errorf("%s: brute force found %d ideals, recursive Total() = %d", name, got, want)
		}
		seen := make(map[string]bool)
		for _, a := range ideals {
			if !s.IsIdeal(a) {
				t.Errorf("%s: enumerated non-ideal %v", name, a)
			}
			key := fmt.Sprint(a)
			if seen[key] {
				t.Errorf("%s: duplicate ideal %v", name, a)
			}
			seen[key] = true
		}
	}
}

// TestSpecialCaseCounts ties the structural counts to the Section 1–3 results:
// 2^n for no arcs, n+1 for the chain, and the fence sizes the paper names.
func TestSpecialCaseCounts(t *testing.T) {
	for n := 1; n <= 12; n++ {
		if got, want := NoArcs(n).Total(), 1<<n; got != want {
			t.Errorf("NoArcs(%d).Total() = %d, want %d", n, got, want)
		}
		if got, want := Chain(n).Total(), n+1; got != want {
			t.Errorf("Chain(%d).Total() = %d, want %d", n, got, want)
		}
	}
	// Fence sizes the paper mentions: 8 for n=4, 21 for n=6.
	if got := Fence(4).Total(); got != 8 {
		t.Errorf("Fence(4).Total() = %d, want 8", got)
	}
	if got := Fence(6).Total(); got != 21 {
		t.Errorf("Fence(6).Total() = %d, want 21", got)
	}
}
