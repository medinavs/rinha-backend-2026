package http

import (
	"log"
	"time"

	"github.com/medinavs/rinha-backend-2026/internal/adapters/vectorindex"
	"github.com/medinavs/rinha-backend-2026/internal/adapters/vectorizer"
	"github.com/medinavs/rinha-backend-2026/internal/application"
	"github.com/medinavs/rinha-backend-2026/internal/config"
	"github.com/valyala/fasthttp"
)

func StartServer(cfg config.Config) {
	vec, err := vectorizer.Load(cfg.NormalizationPath, cfg.MCCRiskPath)
	if err != nil {
		log.Fatalf("load vectorizer: %v", err)
	}

	log.Printf("loading IVF index from %s ...", cfg.IndexPath)
	t0 := time.Now()
	index, err := vectorindex.LoadBinaryIndex(cfg.IndexPath)
	if err != nil {
		log.Fatalf("load index: %v", err)
	}
	index.SetIVFSearch(cfg.ANNNProbe, cfg.ANNAmbiguousProbe, cfg.ANNRepair)
	log.Printf("loaded %d vectors (clusters=%d nprobe=%d ambiguous=%d repair=%v) in %s",
		index.Size(), index.IVF.Clusters, index.IVF.NProbe, index.IVF.AmbiguousNProbe, index.IVF.Repair, time.Since(t0))

	fraudSvc := application.NewFraudDetectionService(vec, index)
	handler := &Handler{FraudSvc: fraudSvc}

	requestHandler := func(ctx *fasthttp.RequestCtx) {
		path := ctx.Path()
		method := ctx.Method()
		switch {
		case len(method) == 4 && string(method) == "POST" && string(path) == "/fraud-score":
			handler.HandleFraudScore(ctx)
		case len(method) == 3 && string(method) == "GET" && string(path) == "/ready":
			handler.HandleReady(ctx)
		default:
			ctx.Error("not found", fasthttp.StatusNotFound)
		}
	}

	server := &fasthttp.Server{
		Handler:            requestHandler,
		Name:               "rinha-backend-2026",
		ReadBufferSize:     8192,
		WriteBufferSize:    8192,
		DisableKeepalive:   false,
		TCPKeepalive:       true,
		ReduceMemoryUsage:  false,
		MaxRequestBodySize: 1 << 20,
	}

	log.Printf("listening on %s", cfg.ListenAddr)
	if err := server.ListenAndServe(cfg.ListenAddr); err != nil {
		log.Fatalf("server error: %v", err)
	}
}
