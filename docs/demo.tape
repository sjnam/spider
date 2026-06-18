# VHS tape for the spider demo GIF.
# Regenerate (from the repo root):  vhs docs/demo.tape
# Requires: https://github.com/charmbracelet/vhs
Output docs/demo.gif

Require go

Set Shell bash
Set FontSize 16
Set Width 960
Set Height 600
Set Padding 20
Set TypingSpeed 45ms
Set Theme "Dracula"

Type "# poke / bump / nudge — the paper's 'troll' coroutines, as goroutines"
Sleep 800ms
Enter
Type "go run . -coro bump -n 3"
Enter
Sleep 3s

Hide
Type "clear"
Enter
Show

Type "# active list on the 9-vertex example spider (60 ideals; cf. page 21)"
Sleep 800ms
Enter
Type "go run . -coro active -spider example -steps 13"
Enter
Sleep 4s
