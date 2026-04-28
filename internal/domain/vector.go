package domain

const VectorDims = 14

type Vector [VectorDims]float32

type LabeledVector struct {
	Vector Vector
	Fraud  bool
}
