package domain

const VectorDims = 14

type Vector [VectorDims]float64

type LabeledVector struct {
	Vector Vector
	Fraud  bool
}

type Neighbor struct {
	Distance float64
	Fraud    bool
}
