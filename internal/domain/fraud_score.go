package domain

type FraudScore struct {
	TransactionID string
	Score         float64
	Approved      bool
}

const FraudThreshold = 0.6

func NewFraudScore(transactionID string, score float64) FraudScore {
	return FraudScore{
		TransactionID: transactionID,
		Score:         score,
		Approved:      score < FraudThreshold,
	}
}
