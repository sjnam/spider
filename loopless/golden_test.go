package loopless

import (
	"strings"
	"testing"

	"github.com/sjnam/spider/spider"
)

// TestGoldenAgainstSpidersProgram checks the loopless forward listing matches
// the provided spiders.c on a shape that source handles correctly: the §2
// example 1 -> 2 <- 3, whose listing (from `spiders "...+-" 0`) is below. On
// shapes where the source is buggy (a chain nested in a near-set) this package
// intentionally diverges — see the package comment and TestRandomizedAgainstActive.
func TestGoldenAgainstSpidersProgram(t *testing.T) {
	const polish = "...+-"
	want := strings.Fields("000 010 011 111 110")

	s, err := spider.Parse(polish)
	if err != nil {
		t.Fatalf("Parse(%q): %v", polish, err)
	}
	var got []string
	for p := range Sequence(s) {
		var b strings.Builder
		for _, v := range p {
			b.WriteByte(byte('0' + v))
		}
		got = append(got, b.String())
	}
	if len(got) != len(want) {
		t.Fatalf("%s: got %d patterns, want %d", polish, len(got), len(want))
	}
	for i := range want {
		if got[i] != want[i] {
			t.Errorf("%s: step %d got %s, want %s", polish, i, got[i], want[i])
		}
	}
}
