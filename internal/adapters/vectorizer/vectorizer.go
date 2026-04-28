package vectorizer

import (
	"encoding/json"
	"fmt"
	"os"

	"github.com/medinavs/rinha-backend-2026/internal/domain"
)

type Normalization struct {
	MaxAmount            float32 `json:"max_amount"`
	MaxInstallments      float32 `json:"max_installments"`
	AmountVsAvgRatio     float32 `json:"amount_vs_avg_ratio"`
	MaxMinutes           float32 `json:"max_minutes"`
	MaxKm                float32 `json:"max_km"`
	MaxTxCount24h        float32 `json:"max_tx_count_24h"`
	MaxMerchantAvgAmount float32 `json:"max_merchant_avg_amount"`
}

const DefaultMCCRisk float32 = 0.5

type Vectorizer struct {
	Norm    Normalization
	MCCRisk map[string]float32
}

func Load(normalizationPath, mccRiskPath string) (*Vectorizer, error) {
	normBytes, err := os.ReadFile(normalizationPath)
	if err != nil {
		return nil, fmt.Errorf("read normalization.json: %w", err)
	}
	var norm Normalization
	if err := json.Unmarshal(normBytes, &norm); err != nil {
		return nil, fmt.Errorf("parse normalization.json: %w", err)
	}

	mccBytes, err := os.ReadFile(mccRiskPath)
	if err != nil {
		return nil, fmt.Errorf("read mcc_risk.json: %w", err)
	}
	mccRisk := make(map[string]float32)
	if err := json.Unmarshal(mccBytes, &mccRisk); err != nil {
		return nil, fmt.Errorf("parse mcc_risk.json: %w", err)
	}

	return &Vectorizer{Norm: norm, MCCRisk: mccRisk}, nil
}

func (v *Vectorizer) Vectorize(tx domain.Transaction) domain.Vector {
	var out domain.Vector

	out[0] = clamp(float32(tx.Amount) / v.Norm.MaxAmount)
	out[1] = clamp(float32(tx.Installments) / v.Norm.MaxInstallments)

	if tx.Customer.AvgAmount > 0 {
		out[2] = clamp((float32(tx.Amount) / float32(tx.Customer.AvgAmount)) / v.Norm.AmountVsAvgRatio)
	} else {
		out[2] = 1.0
	}

	requestedAt := tx.RequestedAt.UTC()
	out[3] = float32(requestedAt.Hour()) / 23.0

	// go weekday: sunday=0..saturday=6; spec: mon=0..sun=6
	wd := int(requestedAt.Weekday())
	out[4] = float32((wd+6)%7) / 6.0

	if tx.LastTransaction != nil {
		minutes := float32(tx.RequestedAt.Sub(tx.LastTransaction.Timestamp).Minutes())
		if minutes < 0 {
			minutes = 0
		}
		out[5] = clamp(minutes / v.Norm.MaxMinutes)
		out[6] = clamp(float32(tx.LastTransaction.KmFromCurrent) / v.Norm.MaxKm)
	} else {
		out[5] = -1
		out[6] = -1
	}

	out[7] = clamp(float32(tx.Terminal.KmFromHome) / v.Norm.MaxKm)
	out[8] = clamp(float32(tx.Customer.TxCount24h) / v.Norm.MaxTxCount24h)

	if tx.Terminal.IsOnline {
		out[9] = 1
	}
	if tx.Terminal.CardPresent {
		out[10] = 1
	}

	if !tx.MerchantKnown {
		out[11] = 1
	}

	if r, ok := v.MCCRisk[tx.Merchant.MCC]; ok {
		out[12] = r
	} else {
		out[12] = DefaultMCCRisk
	}

	out[13] = clamp(float32(tx.Merchant.AvgAmount) / v.Norm.MaxMerchantAvgAmount)

	return out
}

func clamp(x float32) float32 {
	if x < 0 {
		return 0
	}
	if x > 1 {
		return 1
	}
	return x
}
