package vectorindex

import "github.com/medinavs/rinha-backend-2026/internal/domain"

type BruteForceIndex struct {
	data   []float32 // N*14, row-major
	labels []uint8   // 1 = fraud, 0 = legit
	n      int
}

func NewBruteForceIndex(refs []domain.LabeledVector) *BruteForceIndex {
	n := len(refs)
	data := make([]float32, n*domain.VectorDims)
	labels := make([]uint8, n)
	for i := range refs {
		base := i * domain.VectorDims
		row := data[base : base+domain.VectorDims : base+domain.VectorDims]
		src := &refs[i].Vector
		row[0] = float32(src[0])
		row[1] = float32(src[1])
		row[2] = float32(src[2])
		row[3] = float32(src[3])
		row[4] = float32(src[4])
		row[5] = float32(src[5])
		row[6] = float32(src[6])
		row[7] = float32(src[7])
		row[8] = float32(src[8])
		row[9] = float32(src[9])
		row[10] = float32(src[10])
		row[11] = float32(src[11])
		row[12] = float32(src[12])
		row[13] = float32(src[13])
		if refs[i].Fraud {
			labels[i] = 1
		}
	}
	return &BruteForceIndex{data: data, labels: labels, n: n}
}

func (b *BruteForceIndex) Size() int { return b.n }

func (b *BruteForceIndex) Search(q domain.Vector, k int) []domain.Neighbor {
	n := b.n
	if k <= 0 || n == 0 {
		return nil
	}
	if k > n {
		k = n
	}

	q0 := float32(q[0])
	q1 := float32(q[1])
	q2 := float32(q[2])
	q3 := float32(q[3])
	q4 := float32(q[4])
	q5 := float32(q[5])
	q6 := float32(q[6])
	q7 := float32(q[7])
	q8 := float32(q[8])
	q9 := float32(q[9])
	q10 := float32(q[10])
	q11 := float32(q[11])
	q12 := float32(q[12])
	q13 := float32(q[13])

	data := b.data
	labels := b.labels
	top := make([]domain.Neighbor, 0, k)

	for i := range n {
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
		d := float64(sum)

		isFraud := labels[i] == 1

		if len(top) < k {
			top = append(top, domain.Neighbor{Distance: d, Fraud: isFraud})
			if len(top) == k {
				siftDownAll(top)
			}
			continue
		}

		if d < top[0].Distance {
			top[0].Distance = d
			top[0].Fraud = isFraud
			siftDown(top, 0)
		}
	}

	return top
}

// max-heap over Distance so top[0] is the current worst kept neighbor.
func siftDownAll(h []domain.Neighbor) {
	for i := len(h)/2 - 1; i >= 0; i-- {
		siftDown(h, i)
	}
}

func siftDown(h []domain.Neighbor, i int) {
	n := len(h)
	for {
		l := 2*i + 1
		r := 2*i + 2
		largest := i
		if l < n && h[l].Distance > h[largest].Distance {
			largest = l
		}
		if r < n && h[r].Distance > h[largest].Distance {
			largest = r
		}
		if largest == i {
			return
		}
		h[i], h[largest] = h[largest], h[i]
		i = largest
	}
}
