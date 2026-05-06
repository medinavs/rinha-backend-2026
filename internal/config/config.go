package config

import (
	"os"
	"strconv"
	"strings"
)

type Config struct {
	ListenAddr        string
	IndexPath         string
	NormalizationPath string
	MCCRiskPath       string
	ANNNProbe         int
	ANNAmbiguousProbe int
	ANNRepair         bool
}

func Load() Config {
	return Config{
		ListenAddr:        getenv("LISTEN_ADDR", ":8080"),
		IndexPath:         getenv("INDEX_PATH", "/app/resources/index.ivf8192.bin"),
		NormalizationPath: getenv("NORMALIZATION_PATH", "/app/resources/normalization.json"),
		MCCRiskPath:       getenv("MCC_RISK_PATH", "/app/resources/mcc_risk.json"),
		ANNNProbe:         getenvInt("ANN_NPROBE", 8),
		ANNAmbiguousProbe: getenvInt("ANN_AMBIGUOUS_NPROBE", 24),
		ANNRepair:         getenvBool("ANN_REPAIR", true),
	}
}

func getenv(key, fallback string) string {
	if v := os.Getenv(key); v != "" {
		return v
	}
	return fallback
}

func getenvInt(key string, fallback int) int {
	v := os.Getenv(key)
	if v == "" {
		return fallback
	}
	parsed, err := strconv.Atoi(v)
	if err != nil {
		return fallback
	}
	return parsed
}

func getenvBool(key string, fallback bool) bool {
	v := os.Getenv(key)
	if v == "" {
		return fallback
	}
	switch strings.ToLower(v) {
	case "1", "true", "yes", "y", "on":
		return true
	case "0", "false", "no", "n", "off":
		return false
	default:
		return fallback
	}
}
