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
	"github.com/sjnam/spider/nudge"
	"github.com/sjnam/spider/poke"
	"github.com/sjnam/spider/spider"
)

// trolls is the small surface the three coroutine families share.
type trolls interface {
	Poke() (changed bool, reached uint64)
	Bits() []int
	N() int
	Close()
}

func main() {
	coro := flag.String("coro", "poke", "poke, bump, nudge, or active")
	n := flag.Int("n", 3, "number of trolls/bits (poke, bump, nudge)")
	periods := flag.Int("periods", 1, "how many full periods to print (poke, bump, nudge)")
	which := flag.String("spider", "example", "spider for -coro active: example, noarcs, chain, fence")
	steps := flag.Int("steps", 0, "steps to print for -coro active (0 = one full listing)")
	flag.Parse()

	if *coro == "active" {
		runActive(*which, *n, *steps)
		return
	}
	runTrolls(*coro, *n, *periods)
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
		fmt.Printf("unknown -coro %q (want poke, bump, nudge, or active)\n", coro)
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

func runActive(which string, n, steps int) {
	var s *spider.Spider
	switch which {
	case "example":
		s = spider.Example()
	case "noarcs":
		s = spider.NoArcs(n)
	case "chain":
		s = spider.Chain(n)
	case "fence":
		s = spider.Fence(n)
	default:
		fmt.Printf("unknown -spider %q (want example, noarcs, chain, fence)\n", which)
		return
	}
	fmt.Printf("active list — spider=%s  (%d ideals, generated in Gray order)\n\n", which, s.Total())

	g := active.New(s)
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
func listStr(g *active.Gen) string {
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
func bitsAsleep(g *active.Gen) string {
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
