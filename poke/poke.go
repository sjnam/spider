// Package poke implements the "troll" coroutines from Section 1 of
// Knuth & Ruskey, "Efficient Coroutine Generation of Constrained Gray
// Sequences" — the unconstrained case, which generates standard reflected
// Gray binary code.
//
// The paper's coroutine is:
//
//	Boolean coroutine poke[k];
//	while true do begin
//	awake:  a[k] := 1 - a[k]; return true;
//	asleep: if k > 1 then return poke[k-1] else return false;
//	end.
//
// Here each troll T_k is its own goroutine, and "invoking" a coroutine is
// modelled by sending on its poke channel and waiting for a reply — the
// closest faithful rendering of the paper's symmetric coroutine semantics
// that Go's primitives allow. The whole poke chain runs synchronously (only
// one troll acts at a time), so the shared lamp array needs no locking: every
// access is ordered by a channel send/receive.
package poke

import (
	"iter"
	"slices"
)

// result is what an invocation of a troll reports back, mirroring the paper's
// Boolean return value plus a little extra bookkeeping for tracing.
type result struct {
	changed bool   // true if some lamp toggled (paper's `return true`)
	reached uint64 // bitmask of trolls that woke during this poke chain (bit k)
}

// troll is one cooperating coroutine T_k.
type troll struct {
	k     int
	awake bool
	poke  chan struct{} // an invocation of poke[k]
	reply chan result   // the value poke[k] returns
	left  *troll        // T_{k-1}, or nil for the leftmost troll T_1
	a     []int         // shared lamps, 1-indexed: a[k] is troll k's lamp
}

func (t *troll) run() {
	// `for range t.poke` is the immortal `while true` loop; closing the
	// channel is how we retire the troll once everything is idle.
	for range t.poke {
		if t.awake {
			// awake: flip the lamp, then nod off.
			t.a[t.k] = 1 - t.a[t.k]
			t.awake = false
			t.reply <- result{changed: true, reached: 1 << uint(t.k)}
		} else {
			// asleep: wake up and pass the poke leftward, adding ourself to
			// the chain (return poke[k-1]).
			t.awake = true
			if t.left != nil {
				t.left.poke <- struct{}{}
				r := <-t.left.reply
				t.reply <- result{changed: r.changed, reached: r.reached | 1<<uint(t.k)}
			} else {
				t.reply <- result{changed: false, reached: 1 << uint(t.k)} // return false
			}
		}
	}
}

// Trolls is a launched array of poke coroutines, driven from the right end.
type Trolls struct {
	a    []int    // a[0] unused; a[1..n] are the lamps
	all  []*troll // all[1..n]; all[0] is nil
	root *troll   // T_n, the coroutine the driver pokes
}

// New launches n trolls T_1..T_n, all awake with their lamps off, exactly the
// paper's initial configuration.
func New(n int) *Trolls {
	a := make([]int, n+1)
	all := make([]*troll, n+1)
	var left *troll
	for k := 1; k <= n; k++ {
		t := &troll{
			k:     k,
			awake: true,
			poke:  make(chan struct{}),
			reply: make(chan result),
			left:  left,
			a:     a,
		}
		all[k] = t
		go t.run()
		left = t
	}
	return &Trolls{a: a, all: all, root: all[n]}
}

// Poke pokes the root troll T_n once — the external driving force D from the
// paper — advancing to the next pattern. It reports whether a lamp changed
// (false marks a complete listing) and how far left the poke propagated.
func (ts *Trolls) Poke() (changed bool, reached uint64) {
	ts.root.poke <- struct{}{}
	r := <-ts.root.reply
	return r.changed, r.reached
}

// Bits returns the current lamp pattern a[1..n] (a fresh copy).
func (ts *Trolls) Bits() []int { return slices.Clone(ts.a[1:]) }

// N reports the number of trolls.
func (ts *Trolls) N() int { return len(ts.all) - 1 }

// Close retires all trolls. Only call it when the system is idle (right after
// a Poke returns), so every troll is blocked waiting to be poked.
func (ts *Trolls) Close() {
	for k := 1; k < len(ts.all); k++ {
		close(ts.all[k].poke)
	}
}

// Sequence yields the 2^n bit patterns of one forward listing in Gray order,
// starting from 00…0. Consecutive patterns differ in exactly one bit, and the
// 2^n patterns are exactly the reflected Gray binary code. Each yielded slice
// is freshly allocated and owned by the caller.
func Sequence(n int) iter.Seq[[]int] {
	return func(yield func([]int) bool) {
		ts := New(n)
		defer ts.Close()

		if !yield(ts.Bits()) {
			return
		}
		for {
			changed, _ := ts.Poke()
			if !changed {
				// poke[n] returned false: the forward listing is done.
				return
			}
			if !yield(ts.Bits()) {
				return
			}
		}
	}
}
