package main

import (
	"flag"
	"log"
	"os"
	"path/filepath"

	"github.com/medinavs/rinha-backend-2026/internal/adapters/reference"
)

func main() {
	in := flag.String("in", envOrDefault("REFERENCES_JSON_GZ", "resources/references.json.gz"), "input gzipped JSON references")
	out := flag.String("out", envOrDefault("REFERENCES_BIN", "resources/references.bin"), "output binary references")
	flag.Parse()

	if err := os.MkdirAll(filepath.Dir(*out), 0o755); err != nil {
		log.Fatalf("create output dir: %v", err)
	}

	vectors, labels, count, err := reference.LoadJSONGZ(*in)
	if err != nil {
		log.Fatalf("load json.gz: %v", err)
	}

	if err := reference.WriteBinary(*out, vectors, labels, count); err != nil {
		log.Fatalf("write binary: %v", err)
	}

	log.Printf("wrote %d references to %s", count, *out)
}

func envOrDefault(key, fallback string) string {
	if v := os.Getenv(key); v != "" {
		return v
	}
	return fallback
}
