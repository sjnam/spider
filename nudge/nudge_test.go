package nudge

import (
	"fmt"
	"testing"
)

func diff(x, y []int) int {
	d := 0
	for i := range x {
		if x[i] != y[i] {
			d++
		}
	}
	return d
}

// isFence reports whether a_1 <= a_2 >= a_3 <= a_4 >= … (the up-down constraint).
func isFence(bits []int) bool {
	for i := 1; i < len(bits); i++ {
		if i%2 == 1 { // a_i <= a_{i+1}: positions (i-1) < i in 0-indexed are a_i,a_{i+1}
			if bits[i-1] > bits[i] {
				return false
			}
		} else {
			if bits[i-1] < bits[i] {
				return false
			}
		}
	}
	return true
}

// TestSequenceIsFenceGray checks one forward listing: all patterns distinct,
// every one satisfies the fence constraint, and consecutive patterns differ in
// exactly one bit.
func TestSequenceIsFenceGray(t *testing.T) {
	for n := 1; n <= 12; n++ {
		var prev []int
		seen := make(map[string]bool)
		count := 0
		for bits := range Sequence(n) {
			if !isFence(bits) {
				t.Fatalf("n=%d: %v violates a_1<=a_2>=a_3<=…", n, bits)
			}
			key := fmt.Sprint(bits)
			if seen[key] {
				t.Fatalf("n=%d: pattern %v repeated", n, bits)
			}
			seen[key] = true
			if prev != nil {
				if d := diff(prev, bits); d != 1 {
					t.Fatalf("n=%d: %v -> %v differ in %d bits, want 1", n, prev, bits, d)
				}
			}
			prev = bits
			count++
		}
		if count == 0 {
			t.Fatalf("n=%d: produced no patterns", n)
		}
	}
}

// TestCountsMatchPaper checks the listing sizes the paper mentions: 8 fence
// patterns for n=4 (page 11) and 21 (= 11 odd + 10 even) for n=6 (page 11).
func TestCountsMatchPaper(t *testing.T) {
	cases := map[int]int{4: 8, 6: 21}
	for n, want := range cases {
		count := 0
		for range Sequence(n) {
			count++
		}
		if count != want {
			t.Fatalf("n=%d: produced %d fence patterns, want %d", n, count, want)
		}
	}
}

// TestMatchesPaperN4 reproduces the bit-pattern column of the n=4 table on
// page 11 of the paper, forward half then reverse.
func TestMatchesPaperN4(t *testing.T) {
	want := []string{
		"0001", // initial configuration
		"0000", "0100", "0101", "0111", "1111", "1101", "1100", // forward
		"1100", "1101", "1111", "0111", "0101", "0100", "0000", "0001", // reverse
	}
	ts := New(4)
	defer ts.Close()

	render := func() string {
		s := ""
		for _, v := range ts.Bits() {
			s += fmt.Sprint(v)
		}
		return s
	}
	var got []string
	got = append(got, render())
	for len(got) < len(want) {
		ts.Poke()
		got = append(got, render())
	}
	for i := range want {
		if got[i] != want[i] {
			t.Fatalf("step %d: got %s, want %s (full: %v)", i, got[i], want[i], got)
		}
	}
}

// TestStartConfig confirms the initial pattern is the first n bits of
// 000111000111…
func TestStartConfig(t *testing.T) {
	ts := New(8)
	defer ts.Close()
	want := []int{0, 0, 0, 1, 1, 1, 0, 0}
	got := ts.Bits()
	for i := range want {
		if got[i] != want[i] {
			t.Fatalf("start config %v, want %v", got, want)
		}
	}
}
