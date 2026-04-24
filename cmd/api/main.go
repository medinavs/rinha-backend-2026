package main

import (
	"github.com/medinavs/rinha-backend-2026/internal/adapters/http"
	"github.com/medinavs/rinha-backend-2026/internal/config"
)

func main() {
	cfg := config.Load()
	http.StartServer(cfg)
}
