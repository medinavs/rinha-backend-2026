package http

import (
	"encoding/json"
	"net/http"
	"time"

	"github.com/medinavs/rinha-backend-2026/internal/application"
	"github.com/medinavs/rinha-backend-2026/internal/domain"
)

type Handler struct {
	FraudSvc *application.FraudDetectionService
}

type transactionField struct {
	Amount       float64 `json:"amount"`
	Installments int     `json:"installments"`
	RequestedAt  string  `json:"requested_at"`
}

type customerField struct {
	AvgAmount      float64  `json:"avg_amount"`
	TxCount24h     int      `json:"tx_count_24h"`
	KnownMerchants []string `json:"known_merchants"`
}

type merchantField struct {
	ID        string  `json:"id"`
	MCC       string  `json:"mcc"`
	AvgAmount float64 `json:"avg_amount"`
}

type terminalField struct {
	IsOnline    bool    `json:"is_online"`
	CardPresent bool    `json:"card_present"`
	KmFromHome  float64 `json:"km_from_home"`
}

type lastTransactionField struct {
	Timestamp     string  `json:"timestamp"`
	KmFromCurrent float64 `json:"km_from_current"`
}

type fraudScoreRequest struct {
	ID              string                `json:"id"`
	Transaction     transactionField      `json:"transaction"`
	Customer        customerField         `json:"customer"`
	Merchant        merchantField         `json:"merchant"`
	Terminal        terminalField         `json:"terminal"`
	LastTransaction *lastTransactionField `json:"last_transaction"`
}

type fraudScoreResponse struct {
	Approved   bool    `json:"approved"`
	FraudScore float64 `json:"fraud_score"`
}

func (h *Handler) HandleFraudScore(w http.ResponseWriter, r *http.Request) {
	var req fraudScoreRequest
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		http.Error(w, "invalid body", http.StatusBadRequest)
		return
	}

	requestedAt, err := time.Parse(time.RFC3339, req.Transaction.RequestedAt)
	if err != nil {
		http.Error(w, "invalid requested_at", http.StatusBadRequest)
		return
	}

	tx := domain.Transaction{
		ID:           req.ID,
		Amount:       req.Transaction.Amount,
		Installments: req.Transaction.Installments,
		RequestedAt:  requestedAt,
		Customer: domain.Customer{
			AvgAmount:      req.Customer.AvgAmount,
			TxCount24h:     req.Customer.TxCount24h,
			KnownMerchants: req.Customer.KnownMerchants,
		},
		Merchant: domain.Merchant{
			ID:        req.Merchant.ID,
			MCC:       req.Merchant.MCC,
			AvgAmount: req.Merchant.AvgAmount,
		},
		Terminal: domain.Terminal{
			IsOnline:    req.Terminal.IsOnline,
			CardPresent: req.Terminal.CardPresent,
			KmFromHome:  req.Terminal.KmFromHome,
		},
	}

	if req.LastTransaction != nil {
		ts, err := time.Parse(time.RFC3339, req.LastTransaction.Timestamp)
		if err == nil {
			tx.LastTransaction = &domain.LastTransaction{
				Timestamp:     ts,
				KmFromCurrent: req.LastTransaction.KmFromCurrent,
			}
		}
	}

	score, err := h.FraudSvc.Detect(r.Context(), tx)
	if err != nil {
		http.Error(w, "detection failed", http.StatusInternalServerError)
		return
	}

	resp := fraudScoreResponse{Approved: score.Approved, FraudScore: score.Score}
	w.Header().Set("Content-Type", "application/json")
	_ = json.NewEncoder(w).Encode(resp)
}

func (h *Handler) HandleReady(w http.ResponseWriter, r *http.Request) {
	w.WriteHeader(http.StatusOK)
	_, _ = w.Write([]byte("ok"))
}
