# VHS tape for the interactive TUI GIF.
# Regenerate (from the repo root):  vhs docs/demo-tui.tape
Output docs/demo-tui.gif

Require go

Set Shell bash
Set FontSize 16
Set Width 720
Set Height 600
Set Padding 16
Set TypingSpeed 40ms
Set Theme "Dracula"

# Build off-screen so the recording starts clean.
Hide
Type "go build -o /tmp/sptui ." Enter
Sleep 2.5s
Type "clear" Enter
Show

Type "/tmp/sptui -coro tui -spider example"
Enter
Sleep 1.5s

# Step forward a few times by hand…
Right
Sleep 800ms
Right
Sleep 800ms
Right
Sleep 800ms
Right
Sleep 800ms

# …then let it play.
Type "a"
Sleep 6s
Type "q"
Sleep 500ms
