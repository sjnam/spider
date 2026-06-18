# SPIDERS

[![Go Reference](https://pkg.go.dev/badge/github.com/sjnam/spider.svg)](https://pkg.go.dev/github.com/sjnam/spider)

> 🏆 **Acknowledged and fixed by Donald E. Knuth.** A bug this project found in
> Knuth's _SPIDERS_ program was confirmed and corrected in the current
> [`spiders.w`](https://www-cs-faculty.stanford.edu/~knuth/programs/spiders.w)
> (June 2026), which adopts this repo's fix and credits it: _"The code here is
> due to Soojin Nam, who kindly pointed out in 2026 that my original recurrences
> were fatally flawed."_ See [`bugreport/`](bugreport/).

A faithful Go implementation of the algorithms in Knuth & Ruskey's paper
**["Efficient Coroutine Generation of Constrained Gray Sequences"](https://www-cs-faculty.stanford.edu/~knuth/papers/p160.ps.gz)**
(dedicated to the memory of Ole-Johan Dahl).

The authors playfully call the problem **"spider squishing."**

![demo](docs/demo.gif)

> The `bump` coroutines, then the active list on the 9-vertex example spider.
> Regenerate with `vhs docs/demo.tape` — see [docs/DEMO.md](docs/DEMO.md).

## The problem

Generate every bit string $a_1 a_2 \dots a_n \in \{0,1\}^n$ satisfying the
constraints of a given **directed graph**: for each arc $j \to k$, $a_j \le a_k$
(i.e. $a_j=1 \Rightarrow a_k=1$). Formally this is **"generating all order ideals
of an acyclic poset."**

On top of that, list them as a **Gray path** — changing exactly one bit per step.

When the graph has no cycles even with arc directions ignored (_totally
acyclic_) it is called a **spider**, and the paper's key result is that a Gray
path always exists and can be generated in **constant time per bit change**.

```text
  3   5          constraints: a1≤a2≤a3, a4≤a3, a2≤a5, ...
   \ / \         flip each vertex 0/1,
    2   6   8    changing one bit at a time,
     \  |  /     and sweep through every valid combination.
        1
```

## Two worlds that check each other

The paper presents the same algorithm two ways; this repo implements **both** and
**cross-validates each against the other.**

| world | packages | what it is |
| --- | --- | --- |
| **cooperating coroutines (trolls)** | `poke` `bump` `nudge` | each troll $T_k$ is one goroutine — a near-line-by-line transcription of the paper's pseudocode. |
| **active list** | `active` | the same coroutine swarm "compiled" into an explicit data structure + an iterative loop (amortized $O(1)$). |

The key trick is the `ret`/`invoke` helpers, which make **a goroutine's program
counter serve as the coroutine's state**. That lets the paper's coroutines be
transcribed with no explicit state machine (`nudge` even enters its cycle
mid-stream with a single `goto`).

The two worlds produce **identical output, pattern for pattern**, on the special
cases — a test-enforced demonstration of the paper's §8 claim that "the three
steps faithfully implement those coroutines."

## Package layout

```text
spider/
├── poke/      §1 unconstrained  — standard reflected Gray code (2^n patterns)
├── bump/      §2 chain          — 0 ≤ a1 ≤ … ≤ an ≤ 1 (n+1 patterns)
├── nudge/     §3 fence          — a1 ≤ a2 ≥ a3 ≤ … (tricky initialization)
├── spider/    spider data model (children/sign/scope, near-sets U_k/V_k, ideal counts n_k)
│              §6 launching (initial α, transition τ, final ω)
│              SPIDERS §7–12 tables, Polish-notation Parse, brute-force enumerator
├── active/    §8 active-list generator — any spider, amortized O(1)
├── loopless/  §13–29 loopless generator (port of Knuth's SPIDERS C program, O(1)/step)
├── bugreport/ a real bug found in Knuth's SPIDERS, with a CWEB change-file fix
└── main.go    demo driver
```

> **We found a real bug in Knuth's published SPIDERS program.** While building
> `loopless`, the provided `spiders.c` turned out to emit _non-ideal_ labelings
> (and on some inputs loop forever) when a chain is nested inside a near-set —
> the minimal case is the 5-vertex spider `....++-.+`. The fault is in §16's
> `umaxscope`/`vmaxscope` recursion. We corrected it (computing the insertion
> point from the transition labeling τ_k), so `loopless` now matches the
> brute-force-validated `active` on every spider, including 500 random ones. The
> report, the unchanged master `spiders.w`, and a CWEB **change file** with the
> fix live in [`bugreport/`](bugreport/).

## Quick start

```bash
go test ./...           # full verification
go test -race ./...     # coroutine concurrency check
```

### Coroutine demo (§1–3)

```bash
go run . -coro poke  -n 3        # standard Gray code
go run . -coro bump  -n 3        # chain
go run . -coro nudge -n 4        # fence
```

```text
$ go run . -coro bump -n 3
bump trolls — n=3  (chain: 0 <= a_1 <= … <= a_n <= 1)

000   initial state
001   bump{1,2,3} = true
011   bump{1,2,3} = true
111   bump{1,2} = true
111   bump{1} = false
---
...
```

The pattern column and the set of trolls each poke wakes match the tables on
pages 5, 7, and 11 of the paper exactly.

### Active-list / loopless demo (§8, §13–29)

```bash
go run . -coro active   -spider example     # the 9-vertex example of §4 (60 ideals)
go run . -coro loopless -spider example     # same, via the loopless generator
go run . -coro active   -spider chain -n 5
go run . -coro active   -spider '....++-.+' # any spider as a Polish string
```

```text
$ go run . -coro active -spider example
active list — spider=example  (60 ideals, generated in Gray order)

000001100   1235679
000001101   1235679     ← asleep nodes are underlined in the terminal
000001001   1235679
...
011011100   124679      ← the P1 → Q1 transition (48th pattern)
111011100   14789          positive children 2,6 replaced by negative child 8
...
111111100   14789       ← final state
```

This trace reproduces the example on page 21 of the paper character for character.

## What is verified

- **`poke`/`bump`/`nudge`** — one bit per step, every valid pattern exactly
  once, and bit-for-bit agreement with the paper's example tables (`-race` clean).
- **`spider`** — the example spider's `U_1={2,6,9}`, `V_1={4,7,8}`, scope, counts
  `n_k` (total **60**), and the §6 launch table (α/τ/ω) all match pages 12, 16,
  and 18; counts cross-checked with a brute-force enumerator.
- **`active`** — exact output match with `poke`/`bump`/`nudge` on
  NoArcs/Chain/Fence, a complete Gray code agreeing with `AllIdeals` on arbitrary
  mixed spiders, and the page-21 trace.
- **`loopless`** — matches `active` pattern for pattern on every spider (named,
  Polish, and 500 random), after correcting the SPIDERS `umaxscope`/`vmaxscope`
  bug; the fix is verified exhaustively over all spiders with ≤ 8 vertices.

## Benchmark — looplessness isn't free

`go test -bench . ./loopless` times one full listing per spider. The amortized
`active` list beats the truly-loopless generator on every spider tried (≈1.3–2.2×):

```text
BenchmarkGenerate/active/noarcs16     569638 ns/op
BenchmarkGenerate/loopless/noarcs16   770591 ns/op
BenchmarkGenerate/active/fence16       22481 ns/op
BenchmarkGenerate/loopless/fence16     37362 ns/op
```

This empirically confirms the paper's own §15 remark: the contortions needed for
looplessness "actually cause the total execution time to be longer than it would
be with a more straightforward algorithm." Looplessness buys a worst-case
guarantee on the work _between_ two outputs, not overall speed.

## Possible next steps

- [x] **§13–29 loopless $O(1)$/step** — the `loopless` package (focus pointers +
  lazy fixups), with the SPIDERS `umaxscope`/`vmaxscope` bug fixed; matches `active`.
- [ ] **general `gen` coroutines (§5)** — a goroutine version unifying
  poke/bump/nudge via the `maxu`/`maxv`/`prev` tables.
- [ ] **TUI visualization** — animate the spider and the active list live.

## References

- D. E. Knuth and F. Ruskey, _Efficient Coroutine Generation of Constrained Gray
  Sequences_. In _From Object-Orientation to Formal Methods: Essays in Memory of
  Ole-Johan Dahl_, LNCS 2635 (2004), 183–204.
- Y. Koda and F. Ruskey, _A Gray code for the ideals of a forest poset_, Journal
  of Algorithms **15** (1993), 324–340.
- D. E. Knuth, _The Art of Computer Programming_, Vol. 4, §7.2.1.1 (Generating
  all $n$-tuples).
- D. E. Knuth, _SPIDERS_, `https://www-cs-faculty.stanford.edu/~knuth/programs/spiders.w`.

## License

The Go code in this repository is released under the [MIT License](LICENSE).

The files under [`bugreport/`](bugreport/) are an exception: `spiders.w` is
Donald E. Knuth's program, included unmodified, and `spiders.ch` is a change
file against it. They are governed by Knuth's own terms, not by the MIT license,
and are included here only to document the bug and its fix.
