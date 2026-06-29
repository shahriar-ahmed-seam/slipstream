package metrics

// P2Quantile implements the P² algorithm for online quantile estimation
// with O(1) time and O(1) memory per update. See:
//
//	Jain, R. and Chlamtac, I. (1985). The P² algorithm for dynamic
//	calculation of quantiles and histograms without storing observations.
//
// The estimator is approximate; it converges to the true quantile after
// the first five samples and tracks it thereafter.
type P2Quantile struct {
	p     float64
	n     [5]int     // actual number of obs <= q[i]
	q     [5]float64 // marker heights (sorted)
	ns    [5]int     // desired marker positions
	count int
	stage int // number of init samples seen
	init  [5]float64
}

// NewP2Quantile constructs an estimator for the supplied probability.
func NewP2Quantile(p float64) *P2Quantile {
	return &P2Quantile{p: p}
}

// Add feeds a new sample to the estimator.
func (q *P2Quantile) Add(x float64) {
	if q == nil {
		return
	}
	if q.stage < 5 {
		q.init[q.stage] = x
		q.stage++
		if q.stage == 5 {
			insertionSort5(q.init[:])
			for i := 0; i < 5; i++ {
				q.q[i] = q.init[i]
				q.n[i] = i + 1
			}
			q.count = 5
			q.computeDesired()
		}
		return
	}

	// Find the cell k in which x falls.
	k := 0
	if x < q.q[0] {
		k = 0
	} else if x >= q.q[4] {
		k = 4
	} else {
		for i := 1; i < 5; i++ {
			if x < q.q[i] {
				k = i
				break
			}
		}
	}
	// Increment counts for cells k+1..4 (cells to the right of x).
	for i := k + 1; i < 5; i++ {
		q.n[i]++
	}
	q.count++

	if x < q.q[0] {
		q.q[0] = x
	}
	if x > q.q[4] {
		q.q[4] = x
	}

	q.computeDesired()

	// Adjust interior markers.
	for i := 1; i <= 3; i++ {
		di := q.ns[i] - q.n[i]
		if (di >= 1 && q.n[i+1]-q.n[i] > 1) || (di <= -1 && q.n[i-1]-q.n[i] < -1) {
			d := 1
			if di < 0 {
				d = -1
			}
			par := parabolic(q.q[i-1], q.q[i], q.q[i+1],
				float64(q.n[i-1]), float64(q.n[i]), float64(q.n[i+1]), float64(d))
			if par < q.q[i-1] || par > q.q[i+1] {
				par = linear(q.q[i], q.q[i+d], float64(q.n[i]), float64(q.n[i+d]), d)
			}
			q.q[i] = par
			q.n[i] += d
		}
	}
}

// Quantile returns the current quantile estimate.
func (q *P2Quantile) Quantile() float64 {
	if q == nil {
		return 0
	}
	if q.stage < 5 {
		return 0
	}
	return q.q[2]
}

func (q *P2Quantile) computeDesired() {
	q.ns[0] = 0
	q.ns[1] = int(float64(q.count) * (q.p / 2))
	q.ns[2] = int(float64(q.count) * q.p)
	q.ns[3] = int(float64(q.count) * ((1 + q.p) / 2))
	q.ns[4] = q.count
}

func parabolic(qim1, qi, qip1, nim1, ni, nip1, d float64) float64 {
	n := d * (ni - nim1 + d) * (qip1 - qi) / (nip1 - nim1)
	n += d * (nip1 - ni - d) * (qi - qim1) / (nip1 - nim1)
	return qi + n/(ni-nim1+d)*(nip1-nim1)
}

func linear(qi, qj, ni, nj float64, d int) float64 {
	return qi + float64(d)*(qj-qi)/(nj-ni)
}

func insertionSort5(a []float64) {
	for i := 1; i < len(a); i++ {
		v := a[i]
		j := i
		for j > 0 && a[j-1] > v {
			a[j] = a[j-1]
			j--
		}
		a[j] = v
	}
}
