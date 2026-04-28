package vectorindex

import (
	"math"

	"github.com/medinavs/rinha-backend-2026/internal/domain"
)

type BruteForceIndex struct {
	data   []float32 // n*VectorDims
	labels []byte    // 1 = fraud, 0 = legit
	n      int
}

func NewBruteForceIndex(data []float32, labels []byte, count int) *BruteForceIndex {
	return &BruteForceIndex{data: data, labels: labels, n: count}
}

func (b *BruteForceIndex) Size() int { return b.n }

const maxK = 5

func (b *BruteForceIndex) SearchTopK(q domain.Vector, k int) (int, int) {
	if k <= 0 || b.n == 0 {
		return 0, 0
	}
	if k > maxK {
		k = maxK
	}
	if k > b.n {
		k = b.n
	}

	q0 := q[0]
	q1 := q[1]
	q2 := q[2]
	q3 := q[3]
	q4 := q[4]
	q5 := q[5]
	q6 := q[6]
	q7 := q[7]
	q8 := q[8]
	q9 := q[9]
	q10 := q[10]
	q11 := q[11]
	q12 := q[12]
	q13 := q[13]

	inf := float32(math.Inf(1))
	var bestDist [maxK]float32
	var bestFraud [maxK]bool
	for i := 0; i < maxK; i++ {
		bestDist[i] = inf
	}

	data := b.data
	labels := b.labels
	n := b.n
	worstIdx := k - 1

	for i := 0; i < n; i++ {
		base := i * domain.VectorDims
		r := data[base : base+domain.VectorDims : base+domain.VectorDims]

		d0 := q0 - r[0]
		d1 := q1 - r[1]
		d2 := q2 - r[2]
		d3 := q3 - r[3]
		d4 := q4 - r[4]
		d5 := q5 - r[5]
		d6 := q6 - r[6]
		d7 := q7 - r[7]
		d8 := q8 - r[8]
		d9 := q9 - r[9]
		d10 := q10 - r[10]
		d11 := q11 - r[11]
		d12 := q12 - r[12]
		d13 := q13 - r[13]

		sum := d0*d0 + d1*d1 + d2*d2 + d3*d3 +
			d4*d4 + d5*d5 + d6*d6 + d7*d7 +
			d8*d8 + d9*d9 + d10*d10 + d11*d11 +
			d12*d12 + d13*d13

		if sum >= bestDist[worstIdx] {
			continue
		}

		isFraud := labels[i] == 1
		insertAt := worstIdx
		for insertAt > 0 && sum < bestDist[insertAt-1] {
			bestDist[insertAt] = bestDist[insertAt-1]
			bestFraud[insertAt] = bestFraud[insertAt-1]
			insertAt--
		}
		bestDist[insertAt] = sum
		bestFraud[insertAt] = isFraud
	}

	frauds := 0
	for i := 0; i < k; i++ {
		if bestFraud[i] {
			frauds++
		}
	}
	return frauds, k
}
