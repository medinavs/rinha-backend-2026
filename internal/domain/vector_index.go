package domain

type VectorIndex interface {
	SearchTopK(query Vector, k int) (fraudVotes, considered int)
	Size() int
}
