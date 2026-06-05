package spider

import "testing"

// sameShape reports whether two spiders have identical parent and sign data.
func sameShape(a, b *Spider) bool {
	if a.N() != b.N() {
		return false
	}
	for k := 1; k <= a.N(); k++ {
		if a.Parent(k) != b.Parent(k) || a.IsPositive(k) != b.IsPositive(k) {
			return false
		}
	}
	return true
}

// TestParseExample confirms the Polish string from SPIDERS §6 reproduces the
// running example spider exactly — this also pins the sign convention ('+' is a
// negative child, '-' a positive child).
func TestParseExample(t *testing.T) {
	s, err := Parse("....+-.--..+-..-+")
	if err != nil {
		t.Fatalf("Parse error: %v", err)
	}
	if !sameShape(s, Example()) {
		t.Fatalf("Parse(example) does not match Example()")
	}
	if s.Total() != 60 {
		t.Errorf("Total() = %d, want 60", s.Total())
	}
}

// TestParseSpecialCases checks Polish encodings of the standard spiders and the
// three-vertex example x -> y <- z from §2 (five labelings).
func TestParseSpecialCases(t *testing.T) {
	cases := []struct {
		polish string
		want   *Spider
	}{
		{"....", NoArcs(4)},
		{"...--", Chain(3)},   // 1 -> 2 -> 3
		{"....---", Chain(4)}, // 1 -> 2 -> 3 -> 4
	}
	for _, c := range cases {
		s, err := Parse(c.polish)
		if err != nil {
			t.Fatalf("Parse(%q) error: %v", c.polish, err)
		}
		if !sameShape(s, c.want) {
			t.Errorf("Parse(%q) shape differs from expected", c.polish)
		}
	}

	// The §2 example 1 -> 2 <- 3 has five ideals: 000,010,011,110,111.
	s, err := Parse("...+-")
	if err != nil {
		t.Fatalf("Parse error: %v", err)
	}
	if s.Total() != 5 {
		t.Errorf("Parse(\"...+-\").Total() = %d, want 5", s.Total())
	}
	if !s.IsIdeal([]int{0, 1, 0}) || s.IsIdeal([]int{1, 0, 0}) {
		t.Errorf("constraints wrong: want 1->2<-3 (a1<=a2, a3<=a2)")
	}
}

// TestParseErrors checks malformed descriptions are rejected.
func TestParseErrors(t *testing.T) {
	for _, bad := range []string{"", "+", "..+x", "...+++"} {
		if _, err := Parse(bad); err == nil {
			t.Errorf("Parse(%q) = nil error, want error", bad)
		}
	}
}
