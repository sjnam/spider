# Regenerating the demo GIF

The animations in the READMEs are produced with
[VHS](https://github.com/charmbracelet/vhs) from declarative `.tape` scripts, so
the GIFs are fully reproducible (no live screen recording).

| tape | GIF | shown in |
| --- | --- | --- |
| [`demo.tape`](demo.tape) | `demo.gif` | top of the root README |
| [`demo-coroutines.tape`](demo-coroutines.tape) | `demo-coroutines.gif` | README coroutine demo (poke/bump/nudge) |
| [`demo-bug.tape`](demo-bug.tape) | `demo-bug.gif` | `bugreport/` — the `....++-.+` spider, now correct |
| [`demo-tui.tape`](demo-tui.tape) | `demo-tui.gif` | README interactive-TUI section (drives the TUI with keystrokes) |

## One-time setup

```sh
brew install vhs            # also pulls ttyd and ffmpeg
# or: go install github.com/charmbracelet/vhs@latest  (plus ttyd + ffmpeg)
```

## Regenerate

From the repository root:

```sh
vhs docs/demo.tape             # writes docs/demo.gif
vhs docs/demo-coroutines.tape  # writes docs/demo-coroutines.gif
vhs docs/demo-bug.tape         # writes docs/demo-bug.gif
vhs docs/demo-tui.tape         # writes docs/demo-tui.gif
git add docs/*.gif && git commit -m "Update demo GIFs"
```

## Tweaking

Edit [`demo.tape`](demo.tape): `Set Width/Height/FontSize/Theme` control the
look, `Type`/`Sleep`/`Enter` the script. The tape runs `go run .` so it always
records the current build. See `go run . -h` for other demos to feature
(`-coro poke|bump|nudge|active|loopless`, `-spider`, `-n`, `-steps`).
