package domain

type VectorIndex interface {
	Search(query Vector, k int) []Neighbor
	Size() int
}
