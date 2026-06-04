package poke

import (
	"fmt"
	"testing"
)

// diff returns the number of bit positions in which x and y differ.
func diff(x, y []int) int {
	d := 0
	for i := range x {
		if x[i] != y[i] {
			d++
		}
	}
	return d
}

// TestSequenceIsGrayCode checks the three defining properties of one forward
// listing: it has 2^n patterns, they are all distinct (so it covers every
// n-tuple exactly once), and consecutive patterns differ in exactly one bit.
func TestSequenceIsGrayCode(t *testing.T) {
	for n := 1; n <= 10; n++ {
		var prev []int
		seen := make(map[string]bool)
		count := 0
		for bits := range Sequence(n) {
			if len(bits) != n {
				t.Fatalf("n=%d: pattern has length %d, want %d", n, len(bits), n)
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
		if want := 1 << n; count != want {
			t.Fatalf("n=%d: produced %d patterns, want %d", n, count, want)
		}
	}
}

// TestStartsAtZero confirms the listing begins at the all-off pattern.
func TestStartsAtZero(t *testing.T) {
	for bits := range Sequence(4) {
		for i, v := range bits {
			if v != 0 {
				t.Fatalf("first pattern %v has a 1 at index %d, want all zeros", bits, i)
			}
		}
		return
	}
}

// TestFullPeriodReturnsHome drives a full period (2^(n+1) pokes) and checks the
// system comes back to the initial all-off, all-awake state — i.e. false is
// returned exactly twice and the last pattern is all zeros.
func TestFullPeriodReturnsHome(t *testing.T) {
	const n = 5
	ts := New(n)
	defer ts.Close()

	falses, pokes := 0, 0
	for falses < 2 {
		changed, _ := ts.Poke()
		pokes++
		if !changed {
			falses++
		}
	}
	if want := 1 << (n + 1); pokes != want {
		t.Fatalf("period took %d pokes, want %d", pokes, want)
	}
	for i, v := range ts.Bits() {
		if v != 0 {
			t.Fatalf("after full period, bit %d = %d, want 0", i, v)
		}
	}
}
