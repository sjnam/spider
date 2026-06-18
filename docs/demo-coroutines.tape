# VHS tape for the coroutine-trio GIF (poke / bump / nudge).
# Regenerate (from the repo root):  vhs docs/demo-coroutines.tape
Output docs/demo-coroutines.gif

Require go

Set Shell bash
Set FontSize 16
Set Width 900
Set Height 420
Set Padding 20
Set TypingSpeed 45ms
Set Theme "Dracula"

Type "# §1 poke — standard reflected Gray code"
Sleep 600ms
Enter
Type "go run . -coro poke -n 2"
Enter
Sleep 2.5s

Hide
Type "clear"
Enter
Show

Type "# §2 bump — chain 0 <= a1 <= a2 <= a3"
Sleep 600ms
Enter
Type "go run . -coro bump -n 3"
Enter
Sleep 2.5s

Hide
Type "clear"
Enter
Show

Type "# §3 nudge — fence a1 <= a2 >= a3 <= a4 (starts at 0001)"
Sleep 600ms
Enter
Type "go run . -coro nudge -n 4"
Enter
Sleep 3.5s
