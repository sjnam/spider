// Package bump implements the chain coroutines from Section 2 of Knuth &
// Ruskey's spider-squishing paper. The constraint digraph is the oriented
// path 1 -> 2 -> … -> n, so the patterns generated are all n-tuples with
//
//	0 <= a_1 <= a_2 <= … <= a_n <= 1,
//
// in Gray order (one bit changes per step). There are n+1 such tuples.
//
// The paper's coroutine, driven by repeatedly invoking bump[1], is:
//
//	Boolean coroutine bump[k];
//	while true do begin
//	awake0:  if k < n then while bump[k+1] do return true;
//	         a[k] := 1; return true;
//	asleep1: return false;            comment a_k…a_n = 1…1;
//	awake1:  a[k] := 0; return true;
//	asleep0: if k < n then while bump[k+1] do return true;
//	         return false;            comment a_k…a_n = 0…0;
//	end.
//
// Unlike poke (Section 1), a bump troll has four distinct resume points and
// internal while-loops. We don't track that state by hand: each troll is a
// goroutine, and ret() — "send a Boolean to the caller, then block until poked
// again" — lets the goroutine's own program counter be the coroutine state.
// The body below is then an almost verbatim transcription of the paper.
package bump

import (
	"iter"
	"runtime"
	"slices"
)

type result struct {
	changed bool   // the paper's Boolean return value
	reached uint64 // bitmask of trolls touched in this poke chain (bit k)
}

type troll struct {
	k     int
	poke  chan struct{}
	reply chan result
	right *troll // T_{k+1}, or nil for the rightmost troll T_n
	a     []int  // shared lamps, 1-indexed: a[k] is troll k's lamp
}

// ret returns a value to our caller and suspends until the next invocation —
// the coroutine "return" of the paper. A closed poke channel means the system
// is shutting down, so we exit the goroutine cleanly.
func (t *troll) ret(changed bool, reached uint64) {
	t.reply <- result{changed, reached}
	if _, ok := <-t.poke; !ok {
		runtime.Goexit()
	}
}

// invoke calls bump[k+1] and waits for its Boolean, i.e. evaluates the
// `bump[k+1]` in the paper's while-conditions.
func (t *troll) invoke() result {
	t.right.poke <- struct{}{}
	return <-t.right.reply
}

func (t *troll) run() {
	if _, ok := <-t.poke; !ok { // await the first invocation
		return
	}
	for {
		bit := uint64(1) << uint(t.k)

		// awake0: if k < n then while bump[k+1] do return true;
		//         a[k] := 1; return true;
		reached := bit
		if t.right != nil {
			for {
				r := t.invoke()
				reached = r.reached | bit
				if !r.changed {
					break
				}
				t.ret(true, reached)
			}
		}
		t.a[t.k] = 1
		t.ret(true, reached)

		// asleep1: return false;
		t.ret(false, bit)

		// awake1: a[k] := 0; return true;
		t.a[t.k] = 0
		t.ret(true, bit)

		// asleep0: if k < n then while bump[k+1] do return true;
		//          return false;
		reached = bit
		if t.right != nil {
			for {
				r := t.invoke()
				reached = r.reached | bit
				if !r.changed {
					break
				}
				t.ret(true, reached)
			}
		}
		t.ret(false, reached)
	}
}

// Trolls is a launched chain of bump coroutines, driven from the left end.
type Trolls struct {
	a    []int
	all  []*troll
	root *troll // bump[1]
}

// New launches n trolls with their lamps off, all starting at label awake0 —
// the paper's initial configuration for the chain.
func New(n int) *Trolls {
	a := make([]int, n+1)
	all := make([]*troll, n+1)
	var right *troll
	for k := n; k >= 1; k-- { // right-to-left so each troll knows its right neighbour
		t := &troll{
			k:     k,
			poke:  make(chan struct{}),
			reply: make(chan result),
			right: right,
			a:     a,
		}
		all[k] = t
		go t.run()
		right = t
	}
	return &Trolls{a: a, all: all, root: all[1]}
}

// Poke pokes the root troll bump[1] once, advancing to the next pattern. It
// reports whether a lamp changed (false marks a complete listing) and how far
// right the poke propagated.
func (ts *Trolls) Poke() (changed bool, reached uint64) {
	ts.root.poke <- struct{}{}
	r := <-ts.root.reply
	return r.changed, r.reached
}

// Bits returns the current lamp pattern a[1..n] (a fresh copy).
func (ts *Trolls) Bits() []int { return slices.Clone(ts.a[1:]) }

// N reports the number of trolls.
func (ts *Trolls) N() int { return len(ts.all) - 1 }

// Root reports the index of the troll the driver pokes (always 1 for bump).
func (ts *Trolls) Root() int { return 1 }

// Close retires all trolls. Call it only when the system is idle.
func (ts *Trolls) Close() {
	for k := 1; k < len(ts.all); k++ {
		close(ts.all[k].poke)
	}
}

// Sequence yields the n+1 patterns of one forward listing, from 00…0 up to
// 11…1, each satisfying 0 <= a_1 <= … <= a_n <= 1 and differing from its
// predecessor in exactly one bit. Each yielded slice is owned by the caller.
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
				return
			}
			if !yield(ts.Bits()) {
				return
			}
		}
	}
}
