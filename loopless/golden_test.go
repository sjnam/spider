package loopless

import (
	"strings"
	"testing"

	"github.com/sjnam/spider/spider"
)

// golden holds forward listings produced by the provided spiders.c (compiled
// and run as `spiders <polish> 0`, taking the lines up to "...N so far"). These
// pin the loopless port to that program's exact output. Two of the three cases
// (".....---..+++.+..++" and "..+....+--.+..-+-.+") are shapes the provided
// source mishandles — its listings there include non-ideal labelings — so these
// goldens also document that the port faithfully reproduces the source's bug
// rather than papering over it. See the package comment.
var golden = map[string]string{
	// The §2 example 1 -> 2 <- 3.
	"...+-": "000 010 011 111 110",

	// A shape where SPIDERS (and loopless) diverge from active's sorted order:
	// after 1000000000 the next flip is bit 5, not bit 9.
	".....---..+++.+..++": "0011100000 0001100000 0000100000 0000000000 1000000000 " +
		"1000100000 1000100010 1000000010 1000000011 1000100011 1000100111 1000000111 " +
		"1000000110 1000100110 1000100100 1000000100 1001000100 1001000110 1001000111 " +
		"1001000011 1001000010 1001000000 1011000000 1011000010 1011000011 1011000111 " +
		"1011000110 1011000100 1111000100 1111000110 1111000111 1111000011 1111000010 " +
		"1111000000 1111010000 1111010010 1111010011 1111010111 1111010110 1111010100 " +
		"1111011100 1111011110 1111011111 1111011011 1111011010 1111011000",

	"..+....+--.+..-+-.+": "0000000010 0000000000 0000100000 0000100010 0000110010 " +
		"0000110000 0001110000 0001110010 0001100010 0001100000 0011100000 0011100010 " +
		"0011100110 0011101110 0011101010 0011101000 0011111000 0011111010 0011111110 " +
		"0011110110 0011110010 0011110000 1011110000 1011110010 1011110011 1011110001 " +
		"1011110101 1011110100 1011111100 1011111101 1011111001 1011111000 1011111010 " +
		"1011111011 1011101011 1011101010 1011101000 1011101001 1011101101 1011101100 " +
		"1011100100 1011100101 1011100001 1011100000 1011100010 1011100011 1111100011 " +
		"1111100010 1111100000 1111100001 1111100101 1111100100 1111101100 1111101101 " +
		"1111101001 1111101000 1111101010 1111101011 1111111011 1111111010 1111111000 " +
		"1111111001 1111111101 1111111100 1111110100 1111110101 1111110001 1111110000 " +
		"1111110010 1111110011",
}

func render(p []int) string {
	var b strings.Builder
	for _, v := range p {
		b.WriteByte(byte('0' + v))
	}
	return b.String()
}

// TestGoldenAgainstSpidersProgram checks the loopless forward listing matches,
// pattern for pattern, the output of the real SPIDERS C program.
func TestGoldenAgainstSpidersProgram(t *testing.T) {
	for polish, want := range golden {
		s, err := spider.Parse(polish)
		if err != nil {
			t.Fatalf("Parse(%q): %v", polish, err)
		}
		wantPats := strings.Fields(want)
		var got []string
		for p := range Sequence(s) {
			got = append(got, render(p))
		}
		if len(got) != len(wantPats) {
			t.Errorf("%s: got %d patterns, want %d", polish, len(got), len(wantPats))
			continue
		}
		for i := range wantPats {
			if got[i] != wantPats[i] {
				t.Errorf("%s: step %d got %s, want %s", polish, i, got[i], wantPats[i])
				break
			}
		}
	}
}
