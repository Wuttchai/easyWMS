package services

import (
	"easywms-demo-v3/internal/models"
	"errors"
	"fmt"
	"github.com/google/uuid"
	"gorm.io/gorm"
	"math"
	"time"
)

func ApplyStockCount(db *gorm.DB, req StockCountRequest) (*models.StockCount, error) {
	if math.IsNaN(req.CountedQty) || math.IsInf(req.CountedQty, 0) || req.CountedQty < 0 {
		return nil, errors.New("counted quantity must be zero or greater")
	}
	var result models.StockCount
	err := db.Transaction(func(tx *gorm.DB) error {
		p, l, err := findProductLocation(tx, req.SKU, req.LocationCode)
		if err != nil {
			return err
		}
		if err := lockProduct(tx, p.ID); err != nil {
			return err
		}
		inv, err := getInventoryForUpdate(tx, p.ID, l.ID)
		if err != nil {
			return err
		}
		diff := req.CountedQty - inv.Qty
		doc := req.Reference
		if doc == "" {
			doc = fmt.Sprintf("COUNT-%s-%s", time.Now().Format("20060102-150405"), uuid.NewString())
		}
		status := "COMPLETED"
		result = models.StockCount{ID: uuid.New(), DocumentNo: doc, ProductID: p.ID, LocationID: l.ID, SystemQty: inv.Qty, CountedQty: req.CountedQty, Difference: diff, Status: status, CreatedBy: req.CreatedBy}
		if diff != 0 {
			dir := "IN"
			qty := diff
			if diff < 0 {
				dir = "OUT"
				qty = -diff
			}
			adj := models.Adjustment{ID: uuid.New(), DocumentNo: "ADJ-" + doc, ProductID: p.ID, LocationID: l.ID, Qty: qty, Direction: dir, ReasonCode: req.ReasonCode, Note: fmt.Sprintf("Stock count %.2f -> %.2f", inv.Qty, req.CountedQty), Status: "PENDING", RequestedBy: req.CreatedBy}
			if err := tx.Create(&adj).Error; err != nil {
				return err
			}
			result.Status = "PENDING_APPROVAL"
			result.AdjustmentID = &adj.ID
		}
		return tx.Create(&result).Error
	})
	return &result, err
}
