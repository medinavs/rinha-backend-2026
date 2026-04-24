package domain

type Vectorizer interface {
	Vectorize(tx Transaction) Vector
}
