// Command spider demonstrates Knuth & Ruskey's spider-squishing paper. With
// -coro poke|bump|nudge it runs the goroutine "troll" coroutines of Sections
// 1–3; with -coro active it runs the Section-8 active-list generator over an
// arbitrary spider (-spider), printing the pattern and the active list (asleep
// nodes underlined) to mirror the tables in the paper.
package main

import (
	"flag"
	"fmt"
	"math/bits"
	"strings"

	"github.com/sjnam/spider/active"
	"github.com/sjnam/spider/bump"
	"github.com/sjnam/spider/gen"
	"github.com/sjnam/spider/loopless"
	"github.com/sjnam/spider/nudge"
	"github.com/sjnam/spider/poke"
	"github.com/sjnam/spider/spider"
	"github.com/sjnam/spider/tui"
)

// trolls is the small surface the three coroutine families share.
type trolls interface {
	Poke() (changed bool, reached uint64)
	Bits() []int
	N() int
	Close()
}

// listGen is the surface the active-list generators (active, loopless) share.
type listGen interface {
	Next() bool
	Bits() []int
	Active() (nodes []int, asleep []bool)
}

func main() {
	coro := flag.String("coro", "poke", "poke, bump, nudge, gen, active, or loopless")
	n := flag.Int("n", 3, "number of trolls/bits (poke, bump, nudge)")
	periods := flag.Int("periods", 1, "how many full periods to print (poke, bump, nudge, gen)")
	which := flag.String("spider", "example", "spider for -coro gen/active/loopless: example, noarcs, chain, fence, or a Polish string like \"...+-\"")
	steps := flag.Int("steps", 0, "steps to print for -coro active/loopless (0 = one full listing)")
	flag.Parse()

	switch *coro {
	case "active", "loopless":
		runActive(*coro, *which, *n, *steps)
	case "gen":
		runGen(*which, *n, *periods)
	case "tui":
		if s := buildSpider(*which, *n); s != nil {
			if err := tui.Run(s, *which); err != nil {
				fmt.Println(err)
			}
		}
	default:
		runTrolls(*coro, *n, *periods)
	}
}

// buildSpider resolves the -spider flag to a Spider (a name, or a Polish string).
func buildSpider(which string, n int) *spider.Spider {
	switch which {
	case "example":
		return spider.Example()
	case "noarcs":
		return spider.NoArcs(n)
	case "chain":
		return spider.Chain(n)
	case "fence":
		return spider.Fence(n)
	default:
		s, err := spider.Parse(which) // treat anything else as a Polish description
		if err != nil {
			fmt.Printf("unknown -spider %q: not a name (example, noarcs, chain, fence) and %v\n", which, err)
			return nil
		}
		return s
	}
}

// runGen drives the general gen coroutines over a spider, printing each pattern
// and the set of gen-coroutines the poke woke (cf. the poke/bump/nudge demos).
func runGen(which string, n, periods int) {
	s := buildSpider(which, n)
	if s == nil {
		return
	}
	ts := gen.New(s)
	defer ts.Close()
	fmt.Printf("gen coroutines — spider=%s  (%d ideals)\n\n", which, s.Total())

	bitStr := func() string {
		var b strings.Builder
		for _, v := range ts.Bits() {
			fmt.Fprintf(&b, "%d", v)
		}
		return b.String()
	}

	fmt.Printf("%s   initial state\n", bitStr())
	falses, want := 0, periods*2
	for falses < want {
		changed, reached := ts.Poke()
		fmt.Printf("%s   gen%s = %t\n", bitStr(), setStr(reached), changed)
		if !changed {
			falses++
			fmt.Println(strings.Repeat("-", s.N()))
		}
	}
}

func runTrolls(coro string, n, periods int) {
	var ts trolls
	switch coro {
	case "poke":
		ts = poke.New(n)
		fmt.Printf("poke trolls — n=%d  (unconstrained: standard reflected Gray code)\n\n", n)
	case "bump":
		ts = bump.New(n)
		fmt.Printf("bump trolls — n=%d  (chain: 0 <= a_1 <= … <= a_n <= 1)\n\n", n)
	case "nudge":
		ts = nudge.New(n)
		fmt.Printf("nudge trolls — n=%d  (fence: a_1 <= a_2 >= a_3 <= …)\n\n", n)
	default:
		fmt.Printf("unknown -coro %q (want poke, bump, nudge, gen, active, loopless, or tui)\n", coro)
		return
	}
	defer ts.Close()

	bitStr := func() string {
		var b strings.Builder
		for _, v := range ts.Bits() {
			fmt.Fprintf(&b, "%d", v)
		}
		return b.String()
	}

	fmt.Printf("%s   initial state\n", bitStr())
	falses, want := 0, periods*2 // each period returns false twice (forward + reverse)
	for falses < want {
		changed, reached := ts.Poke()
		fmt.Printf("%s   %s%s = %t\n", bitStr(), coro, setStr(reached), changed)
		if !changed {
			falses++
			fmt.Println(strings.Repeat("-", n))
		}
	}
}

func runActive(engine, which string, n, steps int) {
	s := buildSpider(which, n)
	if s == nil {
		return
	}
	fmt.Printf("%s — spider=%s  (%d ideals, generated in Gray order)\n\n", engine, which, s.Total())

	var g listGen
	if engine == "loopless" {
		g = loopless.New(s)
	} else {
		g = active.New(s)
	}
	limit := steps
	if limit == 0 {
		limit = s.Total() - 1 // one full forward listing
	}
	fmt.Printf("%s   %s\n", bitsAsleep(g), listStr(g))
	for i := 0; i < limit; i++ {
		if !g.Next() {
			fmt.Println(strings.Repeat("-", s.N()))
			continue
		}
		fmt.Printf("%s   %s\n", bitsAsleep(g), listStr(g))
	}
}

const (
	ulOn  = "\x1b[4m"
	ulOff = "\x1b[0m"
)

// listStr renders the active list, underlining the asleep nodes.
func listStr(g listGen) string {
	nodes, asleep := g.Active()
	var b strings.Builder
	for i, k := range nodes {
		if asleep[i] {
			fmt.Fprintf(&b, "%s%d%s", ulOn, k, ulOff)
		} else {
			fmt.Fprintf(&b, "%d", k)
		}
	}
	return b.String()
}

// bitsAsleep renders the pattern, underlining bit a_j when node j is asleep and
// on the active list — exactly the paper's left-column convention.
func bitsAsleep(g listGen) string {
	bits := g.Bits()
	nodes, asleep := g.Active()
	down := make(map[int]bool)
	for i, k := range nodes {
		if asleep[i] {
			down[k] = true
		}
	}
	var b strings.Builder
	for i, v := range bits {
		if down[i+1] {
			fmt.Fprintf(&b, "%s%d%s", ulOn, v, ulOff)
		} else {
			fmt.Fprintf(&b, "%d", v)
		}
	}
	return b.String()
}

// setStr renders a troll bitmask as the sorted set notation the paper uses,
// e.g. mask with bits 1,2,4 set -> "{1,2,4}".
func setStr(mask uint64) string {
	var b strings.Builder
	b.WriteByte('{')
	first := true
	for mask != 0 {
		k := bits.TrailingZeros64(mask)
		mask &= mask - 1
		if !first {
			b.WriteByte(',')
		}
		fmt.Fprintf(&b, "%d", k)
		first = false
	}
	b.WriteByte('}')
	return b.String()
}
