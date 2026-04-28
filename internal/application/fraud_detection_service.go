package application

import (
	"github.com/medinavs/rinha-backend-2026/internal/domain"
)

const K = 5

type FraudDetectionService struct {
	Vectorizer domain.Vectorizer
	Index      domain.VectorIndex
}

func NewFraudDetectionService(v domain.Vectorizer, idx domain.VectorIndex) *FraudDetectionService {
	return &FraudDetectionService{Vectorizer: v, Index: idx}
}

func (s *FraudDetectionService) Detect(tx domain.Transaction) (frauds, considered int) {
	vec := s.Vectorizer.Vectorize(tx)
	return s.Index.SearchTopK(vec, K)
}
