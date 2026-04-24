package application

import (
	"context"

	"github.com/medinavs/rinha-backend-2026/internal/domain"
)

const K = 5

type FraudDetectionService struct {
	Vectorizer domain.Vectorizer
	Index      domain.VectorIndex
}

func NewFraudDetectionService(v domain.Vectorizer, idx domain.VectorIndex) *FraudDetectionService {
	return &FraudDetectionService{
		Vectorizer: v,
		Index:      idx,
	}
}

func (s *FraudDetectionService) Detect(ctx context.Context, tx domain.Transaction) (domain.FraudScore, error) {
	vec := s.Vectorizer.Vectorize(tx)
	neighbors := s.Index.Search(vec, K)

	frauds := 0
	for _, n := range neighbors {
		if n.Fraud {
			frauds++
		}
	}

	var score float64
	if len(neighbors) > 0 {
		score = float64(frauds) / float64(len(neighbors))
	}

	return domain.NewFraudScore(tx.ID, score), nil
}
