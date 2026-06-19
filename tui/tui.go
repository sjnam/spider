// Package tui is a small dependency-free terminal visualizer for spider
// squishing. It precomputes the Gray listing of a spider's order ideals with
// package active, then lets you step through it — watching the bit string, the
// spider tree (each node coloured by its active-list state), and the active list
// itself change one bit at a time. Rendering is raw ANSI; input uses cbreak mode
// via stty, so it needs a real terminal (macOS/Linux).
package tui

import (
	"fmt"
	"os"
	"os/exec"
	"os/signal"
	"strings"
	"syscall"
	"time"

	"github.com/sjnam/spider/active"
	"github.com/sjnam/spider/spider"
)

// frame is one snapshot of the generation: the labeling, each node's
// active-list state, and which node's bit changed to reach this frame.
type frame struct {
	bits    []int
	inList  []bool // indexed by node 1..n
	asleep  []bool // indexed by node 1..n
	changed int    // node whose bit just flipped (0 for the initial frame)
}

func buildFrames(s *spider.Spider) []frame {
	n := s.N()
	g := active.New(s)
	snap := func(changed int) frame {
		f := frame{bits: g.Bits(), inList: make([]bool, n+1), asleep: make([]bool, n+1), changed: changed}
		nodes, asl := g.Active()
		for i, k := range nodes {
			f.inList[k], f.asleep[k] = true, asl[i]
		}
		return f
	}
	frames := []frame{snap(0)}
	prev := frames[0].bits
	for g.Next() {
		cur := g.Bits()
		ch := 0
		for i := range cur {
			if cur[i] != prev[i] {
				ch = i + 1
				break
			}
		}
		frames = append(frames, snap(ch))
		prev = cur
	}
	return frames
}

// ANSI styling helpers.
func sgr(code, s string) string { return "\x1b[" + code + "m" + s + "\x1b[0m" }

const (
	bold      = "1"
	dim       = "2"
	underline = "4"
	invert    = "7"
	green     = "32"
	yellow    = "33"
	cyan      = "36"
)

type model struct {
	s      *spider.Spider
	name   string
	depth  []int // depth[k] for indentation
	frames []frame
	i      int
}

func newModel(s *spider.Spider, name string) *model {
	n := s.N()
	depth := make([]int, n+1)
	var setDepth func(k, d int)
	setDepth = func(k, d int) {
		depth[k] = d
		for _, c := range s.Children(k) {
			setDepth(c, d+1)
		}
	}
	for _, r := range s.Roots() {
		setDepth(r, 0)
	}
	return &model{s: s, name: name, depth: depth, frames: buildFrames(s), i: 0}
}

func (m *model) view() string {
	f := m.frames[m.i]
	var b strings.Builder
	fmt.Fprintf(&b, "  %s   ideals: %d   step %s/%d\r\n\r\n",
		sgr(bold, "spider squishing — "+m.name), m.s.Total(),
		sgr(cyan, fmt.Sprint(m.i+1)), len(m.frames))

	// bit string
	b.WriteString("  a = ")
	for k := 1; k <= m.s.N(); k++ {
		b.WriteString(m.styleBit(f, k))
	}
	b.WriteString("\r\n\r\n")

	// spider tree, in preorder
	for k := 1; k <= m.s.N(); k++ {
		b.WriteString(m.styleNode(f, k))
		b.WriteString("\r\n")
	}

	// active list
	b.WriteString("\r\n  active list: ")
	first := true
	for k := 1; k <= m.s.N(); k++ {
		if !f.inList[k] {
			continue
		}
		if !first {
			b.WriteString(" ")
		}
		first = false
		s := fmt.Sprint(k)
		if f.asleep[k] {
			s = sgr(dim, sgr(underline, s))
		} else {
			s = sgr(green, s)
		}
		b.WriteString(s)
	}
	b.WriteString("\r\n\r\n  ")
	b.WriteString(sgr(dim, "[space/→] next   [←] prev   [a] auto   [r] reset   [q] quit"))
	b.WriteString("\r\n")
	return b.String()
}

// styleBit renders a_k in the bit string, highlighting the bit that just flipped
// and dimming the bits of asleep, listed nodes (the paper's underline column).
func (m *model) styleBit(f frame, k int) string {
	s := fmt.Sprint(f.bits[k-1])
	switch {
	case k == f.changed:
		return sgr(invert, sgr(yellow, s))
	case f.inList[k] && f.asleep[k]:
		return sgr(underline, s)
	default:
		return s
	}
}

// styleNode renders one tree line: indentation, the arc sign, the node id, and
// its bit, coloured by active-list state.
func (m *model) styleNode(f frame, k int) string {
	indent := strings.Repeat("│  ", m.depth[k])
	sign := sgr(dim, "●") // root
	if p := m.s.Parent(k); p != 0 {
		if m.s.IsPositive(k) {
			sign = sgr(cyan, "↑") // positive child: a_parent ≤ a_k
		} else {
			sign = sgr(yellow, "↓") // negative child: a_k ≤ a_parent
		}
	}
	label := fmt.Sprintf("%d:%d", k, f.bits[k-1])
	switch {
	case k == f.changed:
		label = sgr(invert, label)
	case f.inList[k] && !f.asleep[k]:
		label = sgr(green, sgr(bold, label)) // active, awake
	case f.inList[k] && f.asleep[k]:
		label = sgr(dim, sgr(underline, label)) // active, asleep
	default:
		label = sgr(dim, label) // not on the active list
	}
	return "  " + indent + sign + " " + label
}

func (m *model) advance(d int) {
	m.i += d
	if m.i < 0 {
		m.i = 0
	}
	if m.i >= len(m.frames) {
		m.i = len(m.frames) - 1
	}
}

// Run starts the interactive visualizer for s (name is shown in the header).
func Run(s *spider.Spider, name string) error {
	if fi, _ := os.Stdin.Stat(); fi.Mode()&os.ModeCharDevice == 0 {
		return fmt.Errorf("tui needs a terminal (stdin is not a tty)")
	}
	m := newModel(s, name)

	restore := rawMode()
	defer restore()
	sig := make(chan os.Signal, 1)
	signal.Notify(sig, syscall.SIGINT, syscall.SIGTERM)
	go func() { <-sig; restore(); os.Exit(0) }()

	fmt.Print("\x1b[?25l")       // hide cursor
	defer fmt.Print("\x1b[?25h") // show cursor

	keys := make(chan rune, 8)
	go readKeys(keys)
	ticker := time.NewTicker(200 * time.Millisecond)
	ticker.Stop()
	auto := false

	draw := func() { fmt.Print("\x1b[2J\x1b[H" + m.view()) }
	draw()
	for {
		select {
		case k, ok := <-keys:
			if !ok || k == 'q' || k == 27 || k == 3 {
				return nil
			}
			switch k {
			case ' ', 'l', 'n', 'C': // C = right arrow (final byte)
				m.advance(+1)
			case 'h', 'p', 'D': // D = left arrow
				m.advance(-1)
			case 'r':
				m.i = 0
			case 'a':
				auto = !auto
				if auto {
					ticker.Reset(200 * time.Millisecond)
				} else {
					ticker.Stop()
				}
			}
			draw()
		case <-ticker.C:
			if auto {
				if m.i == len(m.frames)-1 {
					auto = false
					ticker.Stop()
				} else {
					m.advance(+1)
				}
				draw()
			}
		}
	}
}

// rawMode puts the terminal in cbreak/no-echo mode and returns a restore func.
func rawMode() func() {
	run := func(args ...string) {
		c := exec.Command("stty", args...)
		c.Stdin = os.Stdin
		_ = c.Run()
	}
	run("cbreak", "-echo")
	return func() { run("sane") }
}

// readKeys reads keystrokes (decoding arrow-key escape sequences to their final
// byte) and forwards them as runes, closing the channel on EOF.
func readKeys(out chan<- rune) {
	defer close(out)
	buf := make([]byte, 4)
	for {
		n, err := os.Stdin.Read(buf)
		if n == 0 || err != nil {
			return
		}
		if n >= 3 && buf[0] == 27 && buf[1] == '[' {
			out <- rune(buf[2]) // C/D for right/left arrows
			continue
		}
		out <- rune(buf[0])
	}
}
