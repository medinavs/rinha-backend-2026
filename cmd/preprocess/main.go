package main

import (
	"flag"
	"log"
	"os"
	"path/filepath"

	"github.com/medinavs/rinha-backend-2026/internal/adapters/reference"
	"github.com/medinavs/rinha-backend-2026/internal/adapters/vectorindex"
)

func main() {
	in := flag.String("in", envOrDefault("REFERENCES_JSON_GZ", "resources/references.json.gz"), "input gzipped JSON references")
	out := flag.String("out", envOrDefault("REFERENCES_BIN", "resources/index.ivf8192.bin"), "output IVF binary index")
	clusters := flag.Int("clusters", 8192, "IVF cluster count; must be a power of two")
	nprobe := flag.Int("nprobe", 8, "default IVF nprobe stored in the index")
	ambiguousNProbe := flag.Int("ambiguous-nprobe", 32, "default IVF nprobe for ambiguous results")
	repair := flag.Bool("repair", true, "enable IVF bbox repair for ambiguous results")
	flag.Parse()

	if err := os.MkdirAll(filepath.Dir(*out), 0o755); err != nil {
		log.Fatalf("create output dir: %v", err)
	}

	refs, err := reference.LoadReferences(*in)
	if err != nil {
		log.Fatalf("load references: %v", err)
	}

	if err := vectorindex.WriteIVFBinaryIndex(*out, refs, vectorindex.IVFBuildOptions{
		Clusters:        *clusters,
		NProbe:          *nprobe,
		AmbiguousNProbe: *ambiguousNProbe,
		Repair:          *repair,
	}); err != nil {
		log.Fatalf("write IVF index: %v", err)
	}

	log.Printf("wrote %d references to %s (ivf, clusters=%d nprobe=%d ambiguous=%d repair=%v)",
		len(refs), *out, *clusters, *nprobe, *ambiguousNProbe, *repair)
}

func envOrDefault(key, fallback string) string {
	if v := os.Getenv(key); v != "" {
		return v
	}
	return fallback
}
