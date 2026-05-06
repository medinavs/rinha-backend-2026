package reference

import (
	"compress/gzip"
	"encoding/json"
	"errors"
	"fmt"
	"io"
	"os"

	"github.com/medinavs/rinha-backend-2026/internal/adapters/vectorindex"
	"github.com/medinavs/rinha-backend-2026/internal/domain"
)

type rawRecord struct {
	Vector [domain.VectorDims]float32 `json:"vector"`
	Label  string                     `json:"label"`
}

func openMaybeGzip(path string) (io.ReadCloser, *os.File, error) {
	f, err := os.Open(path)
	if err != nil {
		return nil, nil, fmt.Errorf("open references: %w", err)
	}
	gz, err := gzip.NewReader(f)
	if err != nil {
		f.Close()
		return nil, nil, fmt.Errorf("gzip reader: %w", err)
	}
	return gz, f, nil
}

func LoadReferences(path string) ([]vectorindex.Reference, error) {
	gz, f, err := openMaybeGzip(path)
	if err != nil {
		return nil, err
	}
	defer gz.Close()
	defer f.Close()

	dec := json.NewDecoder(gz)
	tok, err := dec.Token()
	if err != nil {
		return nil, fmt.Errorf("read opening token: %w", err)
	}
	if d, ok := tok.(json.Delim); !ok || d != '[' {
		return nil, errors.New("references must be a JSON array")
	}

	refs := make([]vectorindex.Reference, 0, 200000)
	for dec.More() {
		var r rawRecord
		if err := dec.Decode(&r); err != nil {
			return nil, fmt.Errorf("decode record %d: %w", len(refs), err)
		}
		refs = append(refs, vectorindex.Reference{Vector: r.Vector, Label: r.Label})
	}
	if len(refs) == 0 {
		return nil, errors.New("reference dataset is empty")
	}
	return refs, nil
}
