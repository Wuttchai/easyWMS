package services

import "time"

type MovementRequest struct {
	SKU          string     `json:"sku" binding:"required"`
	LocationCode string     `json:"location_code" binding:"required"`
	Qty          float64    `json:"qty" binding:"required,gt=0"`
	LotNo        string     `json:"lot_no"`
	MfgDate      *time.Time `json:"mfg_date"`
	ExpDate      *time.Time `json:"exp_date"`
	Reference    string     `json:"reference"`
	ReasonCode   string     `json:"reason_code"`
	Note         string     `json:"note"`
	CreatedBy    string     `json:"-"`
}

type TransferRequest struct {
	SKU          string  `json:"sku" binding:"required"`
	FromLocation string  `json:"from_location" binding:"required"`
	ToLocation   string  `json:"to_location" binding:"required"`
	Qty          float64 `json:"qty" binding:"required,gt=0"`
	Reference    string  `json:"reference"`
	ReasonCode   string  `json:"reason_code"`
	Note         string  `json:"note"`
	CreatedBy    string  `json:"-"`
}

type StockCountRequest struct {
	SKU          string  `json:"sku" binding:"required"`
	LocationCode string  `json:"location_code" binding:"required"`
	CountedQty   float64 `json:"counted_qty" binding:"gte=0"`
	Reference    string  `json:"reference"`
	ReasonCode   string  `json:"reason_code"`
	CreatedBy    string  `json:"-"`
}

type AdjustmentRequest struct {
	SKU          string  `json:"sku" binding:"required"`
	LocationCode string  `json:"location_code" binding:"required"`
	Qty          float64 `json:"qty" binding:"required,gt=0"`
	Direction    string  `json:"direction" binding:"required"`
	Reference    string  `json:"reference"`
	ReasonCode   string  `json:"reason_code"`
	Note         string  `json:"note"`
	RequestedBy  string  `json:"-"`
}
