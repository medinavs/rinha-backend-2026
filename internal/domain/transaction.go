package domain

import "time"

type Transaction struct {
	ID              string
	Amount          float64
	Installments    int
	RequestedAt     time.Time
	Customer        Customer
	Merchant        Merchant
	Terminal        Terminal
	LastTransaction *LastTransaction
	MerchantKnown   bool
}

type Customer struct {
	AvgAmount  float64
	TxCount24h int
}

type Merchant struct {
	ID        string
	MCC       string
	AvgAmount float64
}

type Terminal struct {
	IsOnline    bool
	CardPresent bool
	KmFromHome  float64
}

type LastTransaction struct {
	Timestamp     time.Time
	KmFromCurrent float64
}
