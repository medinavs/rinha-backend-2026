package referenceio

import (
	"compress/gzip"
	"encoding/binary"
	"encoding/json"
	"errors"
	"fmt"
	"io"
	"os"

	"github.com/medinavs/rinha-backend-2026/internal/domain"
)

const (
	fileMagic   = "R26REF01"
	fileVersion = uint32(1)
)

type rawRecord struct {
	Vector [domain.VectorDims]float32 `json:"vector"`
	Label  string                     `json:"label"`
}

func LoadJSONGZ(path string) ([]float32, []byte, int, error) {
	f, err := os.Open(path)
	if err != nil {
		return nil, nil, 0, fmt.Errorf("open references: %w", err)
	}
	defer f.Close()

	gz, err := gzip.NewReader(f)
	if err != nil {
		return nil, nil, 0, fmt.Errorf("gzip reader: %w", err)
	}
	defer gz.Close()

	dec := json.NewDecoder(gz)
	tok, err := dec.Token()
	if err != nil {
		return nil, nil, 0, fmt.Errorf("read opening token: %w", err)
	}
	if d, ok := tok.(json.Delim); !ok || d != '[' {
		return nil, nil, 0, errors.New("references must be a JSON array")
	}

	vectors := make([]float32, 0, 200000*domain.VectorDims)
	labels := make([]byte, 0, 200000)
	count := 0

	for dec.More() {
		var r rawRecord
		if err := dec.Decode(&r); err != nil {
			return nil, nil, 0, fmt.Errorf("decode record %d: %w", count, err)
		}
		vectors = append(vectors, r.Vector[:]...)
		if r.Label == "fraud" {
			labels = append(labels, 1)
		} else {
			labels = append(labels, 0)
		}
		count++
	}

	return vectors, labels, count, nil
}

func WriteBinary(path string, vectors []float32, labels []byte, count int) error {
	if len(vectors) != count*domain.VectorDims {
		return fmt.Errorf("invalid vectors length: got %d want %d", len(vectors), count*domain.VectorDims)
	}
	if len(labels) != count {
		return fmt.Errorf("invalid labels length: got %d want %d", len(labels), count)
	}

	f, err := os.Create(path)
	if err != nil {
		return fmt.Errorf("create binary: %w", err)
	}
	defer f.Close()

	if _, err := f.Write([]byte(fileMagic)); err != nil {
		return err
	}
	if err := binary.Write(f, binary.LittleEndian, fileVersion); err != nil {
		return err
	}
	if err := binary.Write(f, binary.LittleEndian, uint32(count)); err != nil {
		return err
	}
	if err := binary.Write(f, binary.LittleEndian, vectors); err != nil {
		return err
	}
	if _, err := f.Write(labels); err != nil {
		return err
	}
	return nil
}

func LoadBinary(path string) ([]float32, []byte, int, error) {
	f, err := os.Open(path)
	if err != nil {
		return nil, nil, 0, fmt.Errorf("open binary: %w", err)
	}
	defer f.Close()

	header := make([]byte, len(fileMagic))
	if _, err := io.ReadFull(f, header); err != nil {
		return nil, nil, 0, fmt.Errorf("read magic: %w", err)
	}
	if string(header) != fileMagic {
		return nil, nil, 0, fmt.Errorf("invalid binary magic: %q", string(header))
	}

	var version uint32
	if err := binary.Read(f, binary.LittleEndian, &version); err != nil {
		return nil, nil, 0, fmt.Errorf("read version: %w", err)
	}
	if version != fileVersion {
		return nil, nil, 0, fmt.Errorf("unsupported binary version: %d", version)
	}

	var count32 uint32
	if err := binary.Read(f, binary.LittleEndian, &count32); err != nil {
		return nil, nil, 0, fmt.Errorf("read count: %w", err)
	}
	count := int(count32)

	vectors := make([]float32, count*domain.VectorDims)
	if err := binary.Read(f, binary.LittleEndian, vectors); err != nil {
		return nil, nil, 0, fmt.Errorf("read vectors: %w", err)
	}

	labels := make([]byte, count)
	if _, err := io.ReadFull(f, labels); err != nil {
		return nil, nil, 0, fmt.Errorf("read labels: %w", err)
	}

	return vectors, labels, count, nil
}
