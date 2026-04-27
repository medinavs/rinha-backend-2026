package http

import (
	"encoding/json"
	"time"

	"github.com/medinavs/rinha-backend-2026/internal/application"
	"github.com/medinavs/rinha-backend-2026/internal/domain"
	"github.com/valyala/fasthttp"
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

func (h *Handler) HandleFraudScore(ctx *fasthttp.RequestCtx) {
	var req fraudScoreRequest
	if err := json.Unmarshal(ctx.PostBody(), &req); err != nil {
		ctx.Error("invalid body", fasthttp.StatusBadRequest)
		return
	}

	requestedAt, err := time.Parse(time.RFC3339, req.Transaction.RequestedAt)
	if err != nil {
		ctx.Error("invalid requested_at", fasthttp.StatusBadRequest)
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

	score, err := h.FraudSvc.Detect(ctx, tx)
	if err != nil {
		ctx.Error("detection failed", fasthttp.StatusInternalServerError)
		return
	}

	resp := fraudScoreResponse{Approved: score.Approved, FraudScore: score.Score}
	body, err := json.Marshal(resp)
	if err != nil {
		ctx.Error("encode failed", fasthttp.StatusInternalServerError)
		return
	}
	ctx.SetContentType("application/json")
	ctx.SetStatusCode(fasthttp.StatusOK)
	ctx.SetBody(body)
}

func (h *Handler) HandleReady(ctx *fasthttp.RequestCtx) {
	ctx.SetStatusCode(fasthttp.StatusOK)
	ctx.SetBodyString("ok")
}
