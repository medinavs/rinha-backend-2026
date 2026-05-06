package vectorindex

import "github.com/medinavs/rinha-backend-2026/internal/domain"

// Size returns the number of indexed vectors.
func (idx *QuantizedIndex) Size() int { return len(idx.Labels) }

// SearchTopK adapts QuantizedIndex.Search5 to the domain.VectorIndex contract.
// k is currently ignored (the IVF kernel is hard-coded to k=5) and reported
// back as the considered count once results are available.
func (idx *QuantizedIndex) SearchTopK(query domain.Vector, k int) (int, int) {
	if k <= 0 || idx.Size() == 0 {
		return 0, 0
	}
	frauds := idx.Search5(query)
	return frauds, 5
}
