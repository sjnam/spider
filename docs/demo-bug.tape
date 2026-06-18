# VHS tape for the bug-case GIF (the spider that exposed Knuth's SPIDERS bug).
# Regenerate (from the repo root):  vhs docs/demo-bug.tape
Output docs/demo-bug.gif

Require go

Set Shell bash
Set FontSize 16
Set Width 860
Set Height 460
Set Padding 20
Set TypingSpeed 45ms
Set Theme "Dracula"

Type "# ....++-.+  — the 5-vertex spider that exposed the SPIDERS bug,"
Sleep 700ms
Enter
Type "# now generating all 10 order ideals correctly:"
Sleep 700ms
Enter
Type "go run . -coro active -spider '....++-.+'"
Enter
Sleep 4s
