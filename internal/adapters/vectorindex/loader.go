package vectorindex

import (
	"compress/gzip"
	"encoding/json"
	"fmt"
	"os"

	"github.com/medinavs/rinha-backend-2026/internal/domain"
)

type rawRef struct {
	Vector [domain.VectorDims]float64 `json:"vector"`
	Label  string                     `json:"label"`
}

func LoadReferences(path string) ([]domain.LabeledVector, error) {
	f, err := os.Open(path)
	if err != nil {
		return nil, fmt.Errorf("open references: %w", err)
	}
	defer f.Close()

	gz, err := gzip.NewReader(f)
	if err != nil {
		return nil, fmt.Errorf("gzip reader: %w", err)
	}
	defer gz.Close()

	dec := json.NewDecoder(gz)

	tok, err := dec.Token()
	if err != nil {
		return nil, fmt.Errorf("read opening token: %w", err)
	}
	if d, ok := tok.(json.Delim); !ok || d != '[' {
		return nil, fmt.Errorf("expected json array, got %v", tok)
	}

	out := make([]domain.LabeledVector, 0, 100000)
	for dec.More() {
		var r rawRef
		if err := dec.Decode(&r); err != nil {
			return nil, fmt.Errorf("decode ref: %w", err)
		}
		out = append(out, domain.LabeledVector{
			Vector: r.Vector,
			Fraud:  r.Label == "fraud",
		})
	}
	return out, nil
}
