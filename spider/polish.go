package spider

import "fmt"

// Parse builds a spider from the right-Polish description used by Knuth's
// SPIDERS program. Reading left to right:
//
//   - '.' pushes a new vertex onto a stack;
//   - '+' draws the arc x <- y and '-' draws the arc x -> y between the top two
//     stack vertices x (second from top) and y (top), then pops y.
//
// So '+' makes y a negative child of x (a_y <= a_x) and '-' makes y a positive
// child (a_x <= a_y). Vertices are numbered 1..n in the order their dots appear,
// which is exactly the preorder New requires. Any vertices still on the stack at
// the end are component roots.
//
// For example, the Section-4 running example is "....+-.--..+-..-+".
func Parse(polish string) (*Spider, error) {
	n := 0
	var stack []int
	parentOf := map[int]int{}
	posOf := map[int]bool{}
	for i, c := range polish {
		switch c {
		case '.':
			n++
			stack = append(stack, n)
		case '+', '-':
			if len(stack) < 2 {
				return nil, fmt.Errorf("spider.Parse: %q at position %d needs two vertices on the stack", c, i)
			}
			y := stack[len(stack)-1]
			x := stack[len(stack)-2]
			stack = stack[:len(stack)-1] // pop y; x stays
			parentOf[y] = x
			posOf[y] = c == '-' // '-' => positive child, '+' => negative child
		default:
			return nil, fmt.Errorf("spider.Parse: unexpected %q at position %d (want '.', '+', or '-')", c, i)
		}
	}
	if n == 0 {
		return nil, fmt.Errorf("spider.Parse: empty description")
	}
	parent := make([]int, n+1)
	positive := make([]bool, n+1)
	for k := 1; k <= n; k++ {
		parent[k] = parentOf[k] // 0 if k is a root
		if p, ok := posOf[k]; ok {
			positive[k] = p
		} else {
			positive[k] = true // a root is a positive child of the virtual vertex 0
		}
	}
	return New(n, parent, positive), nil
}
