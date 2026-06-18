# Regenerating the demo GIF

The animation in the top-level README is produced with
[VHS](https://github.com/charmbracelet/vhs) from [`demo.tape`](demo.tape) — a
declarative script, so the GIF is fully reproducible (no live screen recording).

## One-time setup

```sh
brew install vhs            # also pulls ttyd and ffmpeg
# or: go install github.com/charmbracelet/vhs@latest  (plus ttyd + ffmpeg)
```

## Regenerate

From the repository root:

```sh
vhs docs/demo.tape          # writes docs/demo.gif
git add docs/demo.gif && git commit -m "Update demo GIF"
```

## Tweaking

Edit [`demo.tape`](demo.tape): `Set Width/Height/FontSize/Theme` control the
look, `Type`/`Sleep`/`Enter` the script. The tape runs `go run .` so it always
records the current build. See `go run . -h` for other demos to feature
(`-coro poke|bump|nudge|active|loopless`, `-spider`, `-n`, `-steps`).
