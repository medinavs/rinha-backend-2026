package config

import "os"

type Config struct {
	ListenAddr        string
	ReferencesPath    string
	NormalizationPath string
	MCCRiskPath       string
}

func Load() Config {
	return Config{
		ListenAddr:        getenv("LISTEN_ADDR", ":9999"),
		ReferencesPath:    getenv("REFERENCES_PATH", "/app/resources/references.bin"),
		NormalizationPath: getenv("NORMALIZATION_PATH", "/app/resources/normalization.json"),
		MCCRiskPath:       getenv("MCC_RISK_PATH", "/app/resources/mcc_risk.json"),
	}
}

func getenv(key, fallback string) string {
	if v := os.Getenv(key); v != "" {
		return v
	}
	return fallback
}
