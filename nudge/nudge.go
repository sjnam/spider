// Package nudge implements the "fence" coroutines from Section 3 of Knuth &
// Ruskey's spider-squishing paper. The constraint digraph zig-zags,
//
//	1 -> 2 <- 3 -> 4 <- … ,
//
// so the patterns generated are all n-tuples with the up-down constraints
//
//	a_1 <= a_2 >= a_3 <= a_4 >= … ,
//
// in Gray order. The paper's coroutine, driven by repeatedly invoking
// nudge[1], is:
//
//	Boolean coroutine nudge[k];
//	while true do begin
//	awake0:  if k' <= n then while nudge[k'] do return true;
//	         a[k] := 1; return true;
//	asleep1: if k'' <= n then while nudge[k''] do return true;
//	         return false;
//	awake1:  if k'' <= n then while nudge[k''] do return true;
//	         a[k] := 0; return true;
//	asleep0: if k' <= n then while nudge[k'] do return true;
//	         return false;
//	end.
//
// where (k', k'') = (k+1, k+2) when k is odd, (k+2, k+1) when k is even.
//
// Two things make nudge more interesting than poke/bump:
//
//   - It cannot start from 00…0 with everyone at awake0. The valid starting
//     state sets a_1…a_n to the first n bits of 000111000111…, and each troll
//     begins at awake1 (rather than awake0) exactly when its lamp starts on.
//   - A troll invokes its neighbours k' and k'', which jump by 1 or 2, so a
//     poke chain visits a non-contiguous set of trolls (e.g. {1,2,4}).
//
// As with bump, ret() makes the goroutine's program counter the coroutine
// state; here a single `goto awake1` is all it takes to launch a troll mid-cycle.
package nudge

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
	k        int
	kp, kpp  int   // k' and k''
	n        int   // number of trolls
	poke     chan struct{}
	reply    chan result
	all      []*troll // all[1..n], so a troll can invoke k' and k''
	a        []int    // shared lamps, 1-indexed
	startOne bool     // begin at label awake1 rather than awake0
}

func (t *troll) ret(changed bool, reached uint64) {
	t.reply <- result{changed, reached}
	if _, ok := <-t.poke; !ok {
		runtime.Goexit()
	}
}

// invoke calls nudge[idx] and waits for its Boolean.
func (t *troll) invoke(idx int) result {
	c := t.all[idx]
	c.poke <- struct{}{}
	return <-c.reply
}

func (t *troll) run() {
	if _, ok := <-t.poke; !ok { // await the first invocation
		return
	}
	bit := uint64(1) << uint(t.k)

	// Declared up front so the goto below crosses no declarations.
	var r result
	var reached uint64

	if t.startOne {
		goto awake1
	}

loop:
	// awake0: if k' <= n then while nudge[k'] do return true;
	//         a[k] := 1; return true;
	reached = bit
	if t.kp <= t.n {
		for {
			r = t.invoke(t.kp)
			reached = r.reached | bit
			if !r.changed {
				break
			}
			t.ret(true, reached)
		}
	}
	t.a[t.k] = 1
	t.ret(true, reached)

	// asleep1: if k'' <= n then while nudge[k''] do return true;
	//          return false;
	reached = bit
	if t.kpp <= t.n {
		for {
			r = t.invoke(t.kpp)
			reached = r.reached | bit
			if !r.changed {
				break
			}
			t.ret(true, reached)
		}
	}
	t.ret(false, reached)

awake1:
	// awake1: if k'' <= n then while nudge[k''] do return true;
	//         a[k] := 0; return true;
	reached = bit
	if t.kpp <= t.n {
		for {
			r = t.invoke(t.kpp)
			reached = r.reached | bit
			if !r.changed {
				break
			}
			t.ret(true, reached)
		}
	}
	t.a[t.k] = 0
	t.ret(true, reached)

	// asleep0: if k' <= n then while nudge[k'] do return true;
	//          return false;
	reached = bit
	if t.kp <= t.n {
		for {
			r = t.invoke(t.kp)
			reached = r.reached | bit
			if !r.changed {
				break
			}
			t.ret(true, reached)
		}
	}
	t.ret(false, reached)

	goto loop
}

// startBit returns the kth bit (1-indexed) of the string 000111000111…, which
// is the paper's proper starting pattern for the fence.
func startBit(k int) int {
	if (k-1)%6 < 3 {
		return 0
	}
	return 1
}

// Trolls is a launched fence of nudge coroutines, driven from the left end.
type Trolls struct {
	a    []int
	all  []*troll
	root *troll // nudge[1]
}

// New launches n trolls in the paper's proper starting configuration: lamps
// set to the first n bits of 000111000111…, each troll beginning at awake0 or
// awake1 according to its lamp.
func New(n int) *Trolls {
	a := make([]int, n+1)
	all := make([]*troll, n+1)
	for k := 1; k <= n; k++ {
		a[k] = startBit(k)
		var kp, kpp int
		if k%2 == 1 { // odd
			kp, kpp = k+1, k+2
		} else { // even
			kp, kpp = k+2, k+1
		}
		all[k] = &troll{
			k: k, kp: kp, kpp: kpp, n: n,
			poke:     make(chan struct{}),
			reply:    make(chan result),
			all:      all,
			a:        a,
			startOne: a[k] == 1,
		}
	}
	for k := 1; k <= n; k++ {
		go all[k].run()
	}
	return &Trolls{a: a, all: all, root: all[1]}
}

// Poke pokes the root troll nudge[1] once, advancing to the next pattern.
func (ts *Trolls) Poke() (changed bool, reached uint64) {
	ts.root.poke <- struct{}{}
	r := <-ts.root.reply
	return r.changed, r.reached
}

// Bits returns the current lamp pattern a[1..n] (a fresh copy).
func (ts *Trolls) Bits() []int { return slices.Clone(ts.a[1:]) }

// N reports the number of trolls.
func (ts *Trolls) N() int { return len(ts.all) - 1 }

// Close retires all trolls. Call it only when the system is idle.
func (ts *Trolls) Close() {
	for k := 1; k < len(ts.all); k++ {
		close(ts.all[k].poke)
	}
}

// Sequence yields one forward listing of the fence patterns, starting from the
// proper initial configuration, until the first complete listing (false) is
// reached. Consecutive patterns differ in exactly one bit and each satisfies
// a_1 <= a_2 >= a_3 <= …. Each yielded slice is owned by the caller.
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
