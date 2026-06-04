package spider

import (
	"slices"
	"testing"
)

// bits turns a string like "11011100" into []int, with '*' becoming -1 (the
// flipping bit used in transition patterns).
func bitsOf(s string) []int {
	out := make([]int, len(s))
	for i, c := range s {
		switch c {
		case '0':
			out[i] = 0
		case '1':
			out[i] = 1
		case '*':
			out[i] = -1
		}
	}
	return out
}

// TestLaunchMatchesPaperTable checks alpha_k, tau_k, omega_k for every vertex of
// the example spider against the page-18 table.
func TestLaunchMatchesPaperTable(t *testing.T) {
	s := Example()
	cases := []struct {
		k                 int
		alpha, tau, omega string
	}{
		{9, "0", "*", "1"},
		{8, "00", "*1", "11"},
		{7, "0", "*", "1"},
		{6, "00", "*0", "11"},
		{5, "0", "*", "1"},
		{4, "0", "*", "1"},
		{3, "00", "*0", "11"},
		{2, "0000", "*111", "1101"},
		{1, "000001100", "*11011100", "111111100"},
	}
	for _, c := range cases {
		if got, want := s.Alpha(c.k), bitsOf(c.alpha); !slices.Equal(got, want) {
			t.Errorf("Alpha(%d) = %v, want %v (%s)", c.k, got, want, c.alpha)
		}
		if got, want := s.Tau(c.k), bitsOf(c.tau); !slices.Equal(got, want) {
			t.Errorf("Tau(%d) = %v, want %v (%s)", c.k, got, want, c.tau)
		}
		if got, want := s.Omega(c.k), bitsOf(c.omega); !slices.Equal(got, want) {
			t.Errorf("Omega(%d) = %v, want %v (%s)", c.k, got, want, c.omega)
		}
	}
}

// TestInitialIsValidIdeal checks the starting configuration is itself an ideal,
// for every spider we can build.
func TestInitialIsValidIdeal(t *testing.T) {
	spiders := map[string]*Spider{
		"example": Example(),
		"chain6":  Chain(6),
		"fence6":  Fence(6),
		"fence7":  Fence(7),
		"noarcs5": NoArcs(5),
	}
	for name, s := range spiders {
		if a := s.Initial(); !s.IsIdeal(a) {
			t.Errorf("%s: Initial() = %v is not an ideal", name, a)
		}
	}
}

// TestInitialMatchesSpecialCases ties the launch seed to Sections 1–3: poke and
// bump start at all-zeros, the example starts at 000001100, and the fence starts
// at the first n bits of 000111000111…
func TestInitialMatchesSpecialCases(t *testing.T) {
	if got := Example().Initial(); !slices.Equal(got, bitsOf("000001100")) {
		t.Errorf("Example().Initial() = %v, want 000001100", got)
	}
	for n := 1; n <= 8; n++ {
		if got := NoArcs(n).Initial(); !slices.Equal(got, make([]int, n)) {
			t.Errorf("NoArcs(%d).Initial() = %v, want all zeros", n, got)
		}
		if got := Chain(n).Initial(); !slices.Equal(got, make([]int, n)) {
			t.Errorf("Chain(%d).Initial() = %v, want all zeros", n, got)
		}
	}
	// Fence: 000111000111… (the nudge starting configuration).
	want := bitsOf("00011100") // first 8 bits
	if got := Fence(8).Initial(); !slices.Equal(got, want) {
		t.Errorf("Fence(8).Initial() = %v, want %v", got, want)
	}
}
