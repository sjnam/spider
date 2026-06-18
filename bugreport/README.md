# Bug report: `umaxscope`/`vmaxscope` in Knuth's SPIDERS

> ✅ **Resolved — acknowledged and fixed by Donald E. Knuth (June 2026).** The
> current [`spiders.w`](https://www-cs-faculty.stanford.edu/~knuth/programs/spiders.w)
> adopts this fix; its Introduction now reads _"My original version of this code,
> written in December 2001, had a serious error, which is corrected on the
> present version (June 2026),"_ and the new section credits it: _"The code here
> is due to Soojin Nam, who kindly pointed out in 2026 that my original
> recurrences were fatally flawed. [See https://github.com/sjnam/spider/tree/main/bugreport.]"_
> The `spiders.w` and `spiders.ch` below are kept as the snapshot of the bug and
> its fix at the time of the report.

A bug in **SPIDERS** (`https://www-cs-faculty.stanford.edu/~knuth/programs/spiders.w`)
made it generate labelings that are **not order ideals**, and on some inputs
loop forever. This directory has the report, the original (pre-fix) program, and
a CWEB change file with the fix.

| file | what |
| --- | --- |
| [spiders.w](spiders.w) | the program, exactly as downloaded from Knuth's site (unchanged) |
| [spiders.ch](spiders.ch) | a CWEB **change file** carrying the fix |

The fix is delivered the CWEB way — as a change file, leaving the master `.w`
untouched:

```sh
ctangle spiders.w           # buggy version
cc -o spiders spiders.c
./spiders "....++-.+" 0      # lists only 8 of 10 ideals (see below)

ctangle spiders.w spiders.ch # fixed version
cc -o spiders spiders.c
./spiders "....++-.+" 0      # lists all 10
cweave  spiders.w spiders.ch # the typeset fixed program
```

## Summary

SPIDERS generates, in Gray-code order, all labelings of a totally acyclic
digraph's vertices with 0s and 1s such that `x → y ⟹ bit[x] ≤ bit[y]`. For some
input digraphs it instead emits labelings that **violate** the constraints (and
omits some valid ones); for others the main loop never terminates.

The fault is in §16, which precomputes `umaxscope[k]` and `vmaxscope[k]` — the
insertion points used by §28/§29 when a block of children enters the active list.

## Minimal reproducer

The smallest failing digraph has **5 vertices**, written `....++-.+` in the
program's Polish notation:

```bash
        1            arcs (all "x → y means bit[x] ≤ bit[y]"):
       / \             1 → 2     (2 is a positive child of 1)
      2   5            3 → 2     (3 is a negative child of 2)
      |                4 → 3     (4 is a negative child of 3)
      3                5 → 1     (5 is a negative child of 1)
      |
      4
```

It has **10** order ideals. The published program lists only **8**, then walks
off into illegal labelings:

```bash
$ ctangle spiders.w ; cc -o spiders spiders.c ; ./spiders "....++-.+" 0
00000
01000
01100
01110
11110
11100
11101
11111
11011        <-- bit[4]=1 but bit[3]=0, i.e. the arc 4 → 3 is violated
11010        <-- likewise
...10 so far; now we generate in reverse:
...
```

The corrected program lists all 10 ideals (`00000 01000 01100 01110 11110 11111
11101 11100 11000 11001`) in valid single-bit Gray order.

## What goes wrong

§16 sets (writing `j = umax[k]`):

```c
umaxscope[k] = (umaxbit[k]==1 ? (vmax[j]? vmax[j]: j) : umaxscope[j]);
```

and the dual for `vmaxscope`. By the prose just above it, `umaxscope[k]` is meant
to be *"the largest node that is forced to be in the active list at a transition
point when `bit[k]=0`."* But when a chain of vertices is nested inside a
near-set, the deepest forced node lies **below** `vmax[j]` (resp. `umax[j]`), and
the recursion stops one level too high.

For `....++-.+`: `umax[1]=2`, `umaxbit[1]=1`, `vmax[2]=3`, so the formula gives
`umaxscope[1]=3`. At the moment `bit[1]` flips, however, spider 2 sits at its
last labeling `bit[2]bit[3]bit[4] = 111`, so node **4** is still active — the
correct `umaxscope[1]` is **4**. Because the value is 3, node 5's block is linked
*before* node 4 instead of after it; the active list is no longer in sorted
order, so `focus[left[0]]` (§26) stops returning the largest awake node and the
generator diverges. On other inputs the corrupted links make the main loop spin
forever.

## The fix

`umaxscope[k]` is the largest node still active when spider `k` holds its
*transition* labeling with `bit[k]=0` — i.e. each positive child at its **last**
labeling, each negative child at its **first** (exactly what `setmid(k,0)`
produces). This is computed with value-returning analogues of `setfirst`,
`setlast`, `setmid`: let `firstdeep[k]`, `lastdeep[k]`, `middeep0[k]`,
`middeep1[k]` be the largest active *descendant* of `k` (or 0 if none) under the
labeling `setfirst(k)`, `setlast(k)`, `setmid(k,0)`, `setmid(k,1)` would write.
Each obeys a one-pass recursion over `k`'s children, and since §16 already runs
`k` from `n` down to `1`, every child is finished before its parent. Then
`umaxscope[k] = middeep0[k]` (or `k`) and `vmaxscope[k] = middeep1[k]` (or `k`).

The fix is local (see [spiders.ch](spiders.ch)): it replaces the six lines above
with one section, adds that section, and adds four `int[maxn]` arrays. It keeps
the program loopless and the preprocessing **O(n)** — the recursion visits each
child once. (A direct recomputation of these values is O(n²); the change file's
recursion gives the same values in O(n).)

## Verification

- The original is correct on the example spider of the writeup (60 ideals) and
  on the poke/bump/nudge analogues; the bug needs a chain nested in a near-set.
- Over all **10,066** connected spiders with ≤ 7 vertices, the published program
  misbehaves on **567**: 375 emit a non-ideal labeling, 23 loop forever, and 169
  produce a correct forward half whose reverse half is not its mirror. The
  corrected program is right on all 10,066 (every listing is the complete set of
  ideals, distinct, in single-bit Gray order, with a clean mirror).
- Over all **64,978** connected spiders with ≤ 8 vertices, the new
  `umaxscope`/`vmaxscope` values agree exactly with an independent O(n²)
  recomputation (find the largest active node under each transition labeling),
  and with a brute-force order-ideal enumerator.

The corrected `spiders.w` passes `ctangle` and `cweave` with no errors.
