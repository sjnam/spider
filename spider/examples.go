package spider

// NoArcs returns the spider with n vertices and no arcs — n isolated component
// roots. Its ideals are all 2^n bit patterns (the poke / unconstrained case).
func NoArcs(n int) *Spider {
	parent := make([]int, n+1)
	positive := make([]bool, n+1)
	for k := 1; k <= n; k++ {
		positive[k] = true // root: positive child of vertex 0
	}
	return New(n, parent, positive)
}

// Chain returns the oriented path 1 -> 2 -> … -> n, whose ideals are the
// n-tuples with 0 <= a_1 <= … <= a_n <= 1 (the bump case). Each k>1 is a
// positive child of k-1.
func Chain(n int) *Spider {
	parent := make([]int, n+1)
	positive := make([]bool, n+1)
	positive[1] = true
	for k := 2; k <= n; k++ {
		parent[k] = k - 1
		positive[k] = true
	}
	return New(n, parent, positive)
}

// Fence returns the fence 1 -> 2 <- 3 -> 4 <- …, whose ideals satisfy
// a_1 <= a_2 >= a_3 <= … (the nudge case). The tree is the path 1..n; each
// even k is a positive child of k-1 and each odd k>1 is a negative child.
func Fence(n int) *Spider {
	parent := make([]int, n+1)
	positive := make([]bool, n+1)
	positive[1] = true
	for k := 2; k <= n; k++ {
		parent[k] = k - 1
		positive[k] = k%2 == 0
	}
	return New(n, parent, positive)
}

// Example returns the 9-vertex running example of Section 4 (page 12):
//
//	    1
//	 ┌──┼──┐
//	 2  6  8        positive vertices: 2,3,5,6,9
//	┌┴┐ │  │        negative vertices: 4,7,8
//	3 5 7  9
//	│
//	4
//
// with U_1={2,6,9}, V_1={4,7,8}, and 60 order ideals.
func Example() *Spider {
	//          1  2  3  4  5  6  7  8  9
	parent := []int{0, 0, 1, 2, 3, 2, 1, 6, 1, 8}
	positive := []bool{false, true, true, true, false, true, true, false, false, true}
	return New(9, parent, positive)
}
