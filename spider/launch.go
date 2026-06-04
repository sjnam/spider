package spider

import "slices"

// computeLaunch fills alpha[k], omega[k], and tau[k] from the launch recursion
// of Section 6. It assumes every child's launch data is already computed (the
// caller iterates k from n down to 1).
//
//   - tau[k] over child j's subtree is omega_j if j is positive, alpha_j if j is
//     negative (the transition values); tau[k][k] = -1 marks the flipping bit.
//   - alpha[k] (first of P_k): the forced region is 0, and each block u in U_k is
//     set to its transition value when delta_uk is even, otherwise to the other
//     endpoint of G_u, where delta_uk = product of n_{u'} for u' in U_k before u.
//   - omega[k] (last of Q_k): dual, with the forced region 1 and blocks over V_k.
func (s *Spider) computeLaunch(k int) {
	sc := s.scope[k]
	tau := make([]int, s.n+1)
	alpha := make([]int, s.n+1)
	omega := make([]int, s.n+1)

	// tau: copy each child's first/last pattern; mark a_k as the flipping bit.
	tau[k] = -1
	for _, j := range s.children[k] {
		src := s.alpha[j]
		if s.positive[j] {
			src = s.omega[j]
		}
		copy(tau[j:s.scope[j]+1], src[j:s.scope[j]+1])
	}

	// alpha: baseline forced 0; lay down the U_k blocks per the delta parity.
	s.layBlocks(alpha, tau, s.u[k])

	// omega: baseline forced 1; lay down the V_k blocks.
	for i := k; i <= sc; i++ {
		omega[i] = 1
	}
	s.layBlocks(omega, tau, s.v[k])

	s.tau[k] = tau
	s.alpha[k] = alpha
	s.omega[k] = omega
}

// layBlocks writes each block's reflected-product endpoint into dst over the
// blocks' subtrees. For block b, the endpoint is tau's value when the running
// product (delta) of earlier block sizes is even, otherwise the opposite
// endpoint of G_b. Since tau over b's subtree always equals alpha_b or omega_b,
// "opposite" just means copying the other one.
func (s *Spider) layBlocks(dst, tau []int, blocks []int) {
	delta := 1
	for _, b := range blocks {
		lo, hi := b, s.scope[b]
		isOmega := slices.Equal(tau[lo:hi+1], s.omega[b][lo:hi+1])
		useOmega := isOmega
		if delta%2 == 1 { // delta odd: take the opposite endpoint
			useOmega = !useOmega
		}
		src := s.alpha[b]
		if useOmega {
			src = s.omega[b]
		}
		copy(dst[lo:hi+1], src[lo:hi+1])
		delta *= s.count[b]
	}
}

// sub returns p restricted to the subtree [k, scope(k)] as a fresh slice.
func (s *Spider) sub(p []int, k int) []int {
	return slices.Clone(p[k : s.scope[k]+1])
}

// Alpha returns alpha_k, the first pattern of the Gray path G_k, over the
// vertices [k, scope(k)].
func (s *Spider) Alpha(k int) []int { return s.sub(s.alpha[k], k) }

// Omega returns omega_k, the last pattern of G_k, over [k, scope(k)].
func (s *Spider) Omega(k int) []int { return s.sub(s.omega[k], k) }

// Tau returns tau_k, the transition pattern, over [k, scope(k)]. The first
// entry (vertex k) is -1, marking the bit that flips between P_k and Q_k.
func (s *Spider) Tau(k int) []int { return s.sub(s.tau[k], k) }

// Initial returns the spider's starting configuration a_1…a_n: the first
// pattern of the whole Gray path, formed by concatenating alpha_r over the
// component roots r (the delta_{j,0}=1 convention that keeps components
// independent). It is a valid ideal and the natural seed for the generators.
func (s *Spider) Initial() []int {
	a := make([]int, s.n)
	for _, r := range s.roots {
		for i := r; i <= s.scope[r]; i++ {
			a[i-1] = s.alpha[r][i]
		}
	}
	return a
}
