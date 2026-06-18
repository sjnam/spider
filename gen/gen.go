// Package gen implements the general gen[k](l) coroutines of Section 5 — the
// goroutine "troll" that subsumes poke, bump, and nudge and handles an arbitrary
// spider. It is the general form the special-case coroutines of Sections 1–3 are
// instances of, and the form the active list (package active) compiles into.
//
// The paper's coroutine, driven by invoking gen[maxu[0]](0), is:
//
//	Boolean coroutine gen[k](l); integer l;
//	while true do begin
//	awake0:  if maxu[k] != 0 then while gen[maxu[k]](k) do return true;
//	         a[k] := 1; return true;
//	asleep1: if maxv[k] != 0 then while gen[maxv[k]](k) do return true;
//	         if prev[k] > l then return gen[prev[k]](l) else return false;
//	awake1:  if maxv[k] != 0 then while gen[maxv[k]](k) do return true;
//	         a[k] := 0; return true;
//	asleep0: if maxu[k] != 0 then while gen[maxu[k]](k) do return true;
//	         if prev[k] > l then return gen[prev[k]](l) else return false;
//	end.
//
// where maxu[k]/maxv[k] are the largest elements of U_k/V_k (0 if empty) and
// prev[k] chains the elements of a progenitorial list. Two wrinkles beyond
// bump/nudge: gen takes a parameter l that changes between invocations (so each
// resume must re-read it), and it tail-calls gen[prev[k]](l). As before, ret()
// makes the goroutine's program counter the coroutine state — ret now also
// returns the fresh l delivered with the next poke. Everything comes from
// spider.Tables(); the launch label is awake0/awake1 per the initial bit, and
// the driver pokes gen[maxu[0]] with l=0.
package gen

import (
	"iter"
	"runtime"
	"slices"

	"github.com/sjnam/spider/spider"
)

type result struct {
	changed bool   // the paper's Boolean return value
	reached uint64 // bitmask of gen-coroutines touched in this poke (bit k)
}

type troll struct {
	k                int
	maxu, maxv, prev int
	poke             chan int // each invocation delivers the parameter l
	reply            chan result
	all              []*troll
	a                []int // shared labels, 1-indexed
	startOne         bool  // begin at label awake1 rather than awake0
}

// ret returns a value to the caller and suspends until the next invocation,
// whose parameter l it returns. A closed poke channel retires the goroutine.
func (t *troll) ret(changed bool, reached uint64) int {
	t.reply <- result{changed, reached}
	l, ok := <-t.poke
	if !ok {
		runtime.Goexit()
	}
	return l
}

// invoke calls gen[idx](l) and waits for its Boolean.
func (t *troll) invoke(idx, l int) result {
	c := t.all[idx]
	c.poke <- l
	return <-c.reply
}

func (t *troll) run() {
	l, ok := <-t.poke // the first invocation's parameter
	if !ok {
		return
	}
	bit := uint64(1) << uint(t.k)
	var r result
	var reached uint64

	if t.startOne {
		goto awake1
	}

loop:
	// awake0: while gen[maxu](k) do return true; a[k]=1; return true;
	reached = bit
	if t.maxu != 0 {
		for {
			r = t.invoke(t.maxu, t.k)
			reached = r.reached | bit
			if !r.changed {
				break
			}
			l = t.ret(true, reached)
		}
	}
	t.a[t.k] = 1
	l = t.ret(true, reached)

	// asleep1: while gen[maxv](k) do return true;
	//          if prev>l then return gen[prev](l) else return false;
	reached = bit
	if t.maxv != 0 {
		for {
			r = t.invoke(t.maxv, t.k)
			reached = r.reached | bit
			if !r.changed {
				break
			}
			l = t.ret(true, reached)
		}
	}
	if t.prev > l {
		r = t.invoke(t.prev, l)
		l = t.ret(r.changed, r.reached|bit)
	} else {
		l = t.ret(false, reached)
	}

awake1:
	// awake1: while gen[maxv](k) do return true; a[k]=0; return true;
	reached = bit
	if t.maxv != 0 {
		for {
			r = t.invoke(t.maxv, t.k)
			reached = r.reached | bit
			if !r.changed {
				break
			}
			l = t.ret(true, reached)
		}
	}
	t.a[t.k] = 0
	l = t.ret(true, reached)

	// asleep0: while gen[maxu](k) do return true;
	//          if prev>l then return gen[prev](l) else return false;
	reached = bit
	if t.maxu != 0 {
		for {
			r = t.invoke(t.maxu, t.k)
			reached = r.reached | bit
			if !r.changed {
				break
			}
			l = t.ret(true, reached)
		}
	}
	if t.prev > l {
		r = t.invoke(t.prev, l)
		l = t.ret(r.changed, r.reached|bit)
	} else {
		l = t.ret(false, reached)
	}

	goto loop
}

// Trolls is a launched family of gen coroutines for one spider.
type Trolls struct {
	a    []int
	all  []*troll
	root *troll // gen[maxu[0]]
}

// New launches the gen coroutines seeded at the spider's initial configuration.
func New(s *spider.Spider) *Trolls {
	t := s.Tables()
	n := t.N
	a := make([]int, n+1)
	copy(a[1:], t.Bit0[1:])
	all := make([]*troll, n+1)
	for k := 1; k <= n; k++ {
		all[k] = &troll{
			k: k, maxu: t.Umax[k], maxv: t.Vmax[k], prev: t.Prev[k],
			poke:     make(chan int),
			reply:    make(chan result),
			all:      all,
			a:        a,
			startOne: t.Bit0[k] == 1,
		}
	}
	for k := 1; k <= n; k++ {
		go all[k].run()
	}
	return &Trolls{a: a, all: all, root: all[t.Umax[0]]}
}

// Poke pokes the root gen[maxu[0]] once with l=0, advancing one pattern.
func (ts *Trolls) Poke() (changed bool, reached uint64) {
	ts.root.poke <- 0
	r := <-ts.root.reply
	return r.changed, r.reached
}

// Bits returns the current labeling a[1..n] (a fresh copy).
func (ts *Trolls) Bits() []int { return slices.Clone(ts.a[1:]) }

// N reports the number of vertices.
func (ts *Trolls) N() int { return len(ts.all) - 1 }

// Close retires all coroutines. Call it only when the system is idle.
func (ts *Trolls) Close() {
	for k := 1; k < len(ts.all); k++ {
		close(ts.all[k].poke)
	}
}

// Sequence yields one forward listing of the spider's order ideals in Gray
// order, from the initial configuration until the listing completes.
func Sequence(s *spider.Spider) iter.Seq[[]int] {
	return func(yield func([]int) bool) {
		ts := New(s)
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
