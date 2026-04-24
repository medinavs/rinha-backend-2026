package config

import (
	"os"
	"strconv"
)

type Config struct {
	ListenAddr        string
	ReferencesPath    string
	NormalizationPath string
	MCCRiskPath       string
	HNSWIndexPath     string
	EfSearch          int
}

func Load() Config {
	return Config{
		ListenAddr:        getenv("LISTEN_ADDR", ":9999"),
		ReferencesPath:    getenv("REFERENCES_PATH", "/app/resources/references.json.gz"),
		NormalizationPath: getenv("NORMALIZATION_PATH", "/app/resources/normalization.json"),
		MCCRiskPath:       getenv("MCC_RISK_PATH", "/app/resources/mcc_risk.json"),
		HNSWIndexPath:     getenv("HNSW_INDEX_PATH", "/app/resources/hnsw.bin"),
		EfSearch:          getint("EF_SEARCH", 300),
	}
}

func getenv(key, fallback string) string {
	if v := os.Getenv(key); v != "" {
		return v
	}
	return fallback
}

func getint(key string, fallback int) int {
	if v := os.Getenv(key); v != "" {
		if n, err := strconv.Atoi(v); err == nil {
			return n
		}
	}
	return fallback
}
