package bump

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

// isMonotone reports whether a_1 <= a_2 <= … <= a_n (the chain constraint).
func isMonotone(bits []int) bool {
	for i := 1; i < len(bits); i++ {
		if bits[i-1] > bits[i] {
			return false
		}
	}
	return true
}

// TestSequenceIsConstrainedGray checks one forward listing: exactly n+1
// patterns, all distinct, every one monotone, and consecutive patterns one bit
// apart.
func TestSequenceIsConstrainedGray(t *testing.T) {
	for n := 1; n <= 10; n++ {
		var prev []int
		seen := make(map[string]bool)
		count := 0
		for bits := range Sequence(n) {
			if !isMonotone(bits) {
				t.Fatalf("n=%d: %v violates a_1<=…<=a_n", n, bits)
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
		if want := n + 1; count != want {
			t.Fatalf("n=%d: produced %d patterns, want %d", n, count, want)
		}
	}
}

// TestMatchesPaperN3 reproduces the bit-pattern column of the n=3 table on
// page 7 of the paper (forward half, then reverse).
func TestMatchesPaperN3(t *testing.T) {
	want := []string{
		"000",               // initial
		"001", "011", "111", // forward
		"111", "011", "001", "000", // reverse (after the midpoint false)
	}
	ts := New(3)
	defer ts.Close()

	var got []string
	render := func() string {
		s := ""
		for _, v := range ts.Bits() {
			s += fmt.Sprint(v)
		}
		return s
	}
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

// TestFullPeriodReturnsHome drives a full period (2(n+1) pokes) and checks the
// chain returns to all-off.
func TestFullPeriodReturnsHome(t *testing.T) {
	const n = 6
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
	if want := 2 * (n + 1); pokes != want {
		t.Fatalf("period took %d pokes, want %d", pokes, want)
	}
	for i, v := range ts.Bits() {
		if v != 0 {
			t.Fatalf("after full period, bit %d = %d, want 0", i, v)
		}
	}
}
