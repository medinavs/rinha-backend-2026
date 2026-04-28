package http

import (
	"errors"
	"sync"
	"time"

	"github.com/medinavs/rinha-backend-2026/internal/application"
	"github.com/medinavs/rinha-backend-2026/internal/domain"
	"github.com/valyala/fasthttp"
	"github.com/valyala/fastjson"
)

type Handler struct {
	FraudSvc *application.FraudDetectionService
}

// 6 possible responses indexed by fraud votes (0..5). Pre-rendered so the
// hot path never marshals JSON.
var prebuiltResponses = [6][]byte{
	[]byte(`{"approved":true,"fraud_score":0}`),
	[]byte(`{"approved":true,"fraud_score":0.2}`),
	[]byte(`{"approved":true,"fraud_score":0.4}`),
	[]byte(`{"approved":false,"fraud_score":0.6}`),
	[]byte(`{"approved":false,"fraud_score":0.8}`),
	[]byte(`{"approved":false,"fraud_score":1}`),
}

var contentTypeJSON = []byte("application/json")

// fastjson parsers are not goroutine-safe; pool one per request.
var parserPool sync.Pool = sync.Pool{
	New: func() any { return new(fastjson.Parser) },
}

func (h *Handler) HandleFraudScore(ctx *fasthttp.RequestCtx) {
	parser := parserPool.Get().(*fastjson.Parser)
	defer parserPool.Put(parser)

	v, err := parser.ParseBytes(ctx.PostBody())
	if err != nil {
		ctx.Error("invalid body", fasthttp.StatusBadRequest)
		return
	}

	tx, err := buildTransaction(v)
	if err != nil {
		ctx.Error(err.Error(), fasthttp.StatusBadRequest)
		return
	}

	frauds, considered := h.FraudSvc.Detect(tx)
	if considered == 0 {
		ctx.Error("no references", fasthttp.StatusInternalServerError)
		return
	}

	// considered is always 5 (K) at runtime, so frauds maps directly into the
	// prebuilt slot. Guard the bound just in case.
	idx := frauds
	if idx < 0 {
		idx = 0
	}
	if idx >= len(prebuiltResponses) {
		idx = len(prebuiltResponses) - 1
	}

	ctx.SetContentTypeBytes(contentTypeJSON)
	ctx.SetStatusCode(fasthttp.StatusOK)
	ctx.SetBody(prebuiltResponses[idx])
}

func (h *Handler) HandleReady(ctx *fasthttp.RequestCtx) {
	ctx.SetStatusCode(fasthttp.StatusOK)
	ctx.SetBodyString("ok")
}

// buildTransaction reads only the fields the vectorizer needs from the parsed
// JSON value — no intermediate request struct, no []string for known merchants.
func buildTransaction(v *fastjson.Value) (domain.Transaction, error) {
	var tx domain.Transaction

	tx.ID = string(v.GetStringBytes("id"))

	transactionVal := v.Get("transaction")
	if transactionVal == nil {
		return tx, errors.New("missing transaction")
	}
	tx.Amount = transactionVal.GetFloat64("amount")
	tx.Installments = transactionVal.GetInt("installments")

	requestedAtRaw := transactionVal.GetStringBytes("requested_at")
	requestedAt, ok := parseRFC3339Z(requestedAtRaw)
	if !ok {
		return tx, errors.New("invalid requested_at")
	}
	tx.RequestedAt = requestedAt

	customerVal := v.Get("customer")
	if customerVal != nil {
		tx.Customer.AvgAmount = customerVal.GetFloat64("avg_amount")
		tx.Customer.TxCount24h = customerVal.GetInt("tx_count_24h")
	}

	merchantVal := v.Get("merchant")
	if merchantVal != nil {
		tx.Merchant.ID = string(merchantVal.GetStringBytes("id"))
		tx.Merchant.MCC = string(merchantVal.GetStringBytes("mcc"))
		tx.Merchant.AvgAmount = merchantVal.GetFloat64("avg_amount")
	}

	terminalVal := v.Get("terminal")
	if terminalVal != nil {
		tx.Terminal.IsOnline = terminalVal.GetBool("is_online")
		tx.Terminal.CardPresent = terminalVal.GetBool("card_present")
		tx.Terminal.KmFromHome = terminalVal.GetFloat64("km_from_home")
	}

	if customerVal != nil {
		known := customerVal.GetArray("known_merchants")
		merchantIDBytes := []byte(tx.Merchant.ID)
		for _, k := range known {
			if bytesEqualString(k.GetStringBytes(), merchantIDBytes) {
				tx.MerchantKnown = true
				break
			}
		}
	}

	if last := v.Get("last_transaction"); last != nil && last.Type() != fastjson.TypeNull {
		tsBytes := last.GetStringBytes("timestamp")
		if ts, ok := parseRFC3339Z(tsBytes); ok {
			tx.LastTransaction = &domain.LastTransaction{
				Timestamp:     ts,
				KmFromCurrent: last.GetFloat64("km_from_current"),
			}
		}
	}

	return tx, nil
}

func bytesEqualString(a, b []byte) bool {
	if len(a) != len(b) {
		return false
	}
	for i := range a {
		if a[i] != b[i] {
			return false
		}
	}
	return true
}

// parseRFC3339Z parses the fixed format "2006-01-02T15:04:05Z" used by the
// challenge — much cheaper than time.Parse with full RFC3339 grammar.
func parseRFC3339Z(b []byte) (time.Time, bool) {
	if len(b) != 20 {
		return fallbackParse(b)
	}
	if b[4] != '-' || b[7] != '-' || b[10] != 'T' ||
		b[13] != ':' || b[16] != ':' || b[19] != 'Z' {
		return fallbackParse(b)
	}
	y, ok1 := atoi4(b[0:4])
	mo, ok2 := atoi2(b[5:7])
	d, ok3 := atoi2(b[8:10])
	h, ok4 := atoi2(b[11:13])
	mi, ok5 := atoi2(b[14:16])
	s, ok6 := atoi2(b[17:19])
	if !(ok1 && ok2 && ok3 && ok4 && ok5 && ok6) {
		return fallbackParse(b)
	}
	return time.Date(y, time.Month(mo), d, h, mi, s, 0, time.UTC), true
}

func fallbackParse(b []byte) (time.Time, bool) {
	t, err := time.Parse(time.RFC3339, string(b))
	if err != nil {
		return time.Time{}, false
	}
	return t, true
}

func atoi4(b []byte) (int, bool) {
	if len(b) != 4 {
		return 0, false
	}
	a := int(b[0] - '0')
	b1 := int(b[1] - '0')
	c := int(b[2] - '0')
	d := int(b[3] - '0')
	if a < 0 || a > 9 || b1 < 0 || b1 > 9 || c < 0 || c > 9 || d < 0 || d > 9 {
		return 0, false
	}
	return a*1000 + b1*100 + c*10 + d, true
}

func atoi2(b []byte) (int, bool) {
	if len(b) != 2 {
		return 0, false
	}
	a := int(b[0] - '0')
	c := int(b[1] - '0')
	if a < 0 || a > 9 || c < 0 || c > 9 {
		return 0, false
	}
	return a*10 + c, true
}
