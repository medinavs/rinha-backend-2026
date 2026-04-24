package http

import (
	"log"
	"net/http"

	"github.com/medinavs/rinha-backend-2026/internal/adapters/vectorindex"
	"github.com/medinavs/rinha-backend-2026/internal/adapters/vectorizer"
	"github.com/medinavs/rinha-backend-2026/internal/application"
	"github.com/medinavs/rinha-backend-2026/internal/config"
)

func StartServer(cfg config.Config) {
	vec, err := vectorizer.Load(cfg.NormalizationPath, cfg.MCCRiskPath)
	if err != nil {
		log.Fatalf("load vectorizer: %v", err)
	}

	log.Printf("loading references from %s ...", cfg.ReferencesPath)
	refs, err := vectorindex.LoadReferences(cfg.ReferencesPath)
	if err != nil {
		log.Fatalf("load references: %v", err)
	}
	log.Printf("loaded %d reference vectors", len(refs))
	index := vectorindex.NewBruteForceIndex(refs)

	fraudSvc := application.NewFraudDetectionService(vec, index)
	handler := &Handler{FraudSvc: fraudSvc}

	mux := http.NewServeMux()
	mux.HandleFunc("POST /fraud-score", handler.HandleFraudScore)
	mux.HandleFunc("GET /ready", handler.HandleReady)

	log.Printf("listening on %s", cfg.ListenAddr)
	if err := http.ListenAndServe(cfg.ListenAddr, mux); err != nil {
		log.Fatalf("server error: %v", err)
	}
}
