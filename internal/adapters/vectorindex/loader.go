package vectorindex

import (
	"path/filepath"
	"strings"

	"github.com/medinavs/rinha-backend-2026/internal/adapters/reference"
)

func Load(path string) ([]float32, []byte, int, error) {
	if strings.HasSuffix(strings.ToLower(filepath.Ext(path)), "bin") {
		return reference.LoadBinary(path)
	}
	return reference.LoadJSONGZ(path)
}
