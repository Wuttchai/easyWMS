package services

import (
	"easywms-demo-v3/internal/models"
	"errors"
	"fmt"
	"github.com/google/uuid"
	"gorm.io/gorm"
	"gorm.io/gorm/clause"
	"time"
)

func CreateAdjustment(db *gorm.DB, req AdjustmentRequest) (*models.Adjustment, error) {
	p, l, err := findProductLocation(db, req.SKU, req.LocationCode)
	if err != nil {
		return nil, err
	}
	dir := req.Direction
	if dir != "IN" && dir != "OUT" {
		return nil, errors.New("direction must be IN or OUT")
	}
	doc := req.Reference
	if doc == "" {
		doc = fmt.Sprintf("ADJ-%s", time.Now().Format("20060102-150405.000"))
	}
	row := models.Adjustment{ID: uuid.New(), DocumentNo: doc, ProductID: p.ID, LocationID: l.ID, Qty: req.Qty, Direction: dir, ReasonCode: req.ReasonCode, Note: req.Note, Status: "PENDING", RequestedBy: req.RequestedBy}
	return &row, db.Create(&row).Error
}

func ApproveAdjustment(db *gorm.DB, id uuid.UUID, user string) (*models.Adjustment, error) {
	var out models.Adjustment
	err := db.Transaction(func(tx *gorm.DB) error {
		if err := tx.Clauses(clause.Locking{Strength: "UPDATE"}).First(&out, "id=?", id).Error; err != nil {
			return err
		}
		if out.Status != "PENDING" {
			return errors.New("adjustment is not pending")
		}
		inv, err := getInventoryForUpdate(tx, out.ProductID, out.LocationID)
		if err != nil {
			return err
		}
		typ := "ADJUST_IN"
		if out.Direction == "OUT" {
			if inv.Qty < out.Qty {
				return errors.New("insufficient stock for adjustment")
			}
			inv.Qty -= out.Qty
			typ = "ADJUST_OUT"
		} else {
			inv.Qty += out.Qty
		}
		if err := tx.Save(inv).Error; err != nil {
			return err
		}
		now := time.Now()
		out.Status = "APPROVED"
		out.ApprovedBy = user
		out.ApprovedAt = &now
		if err := tx.Save(&out).Error; err != nil {
			return err
		}
		return tx.Create(&models.StockMovement{ID: uuid.New(), ProductID: out.ProductID, LocationID: out.LocationID, Type: typ, Qty: out.Qty, Reference: out.DocumentNo, ReasonCode: out.ReasonCode, Note: out.Note, CreatedBy: user}).Error
	})
	return &out, err
}

func RejectAdjustment(db *gorm.DB, id uuid.UUID, user string) (*models.Adjustment, error) {
	var row models.Adjustment
	if err := db.First(&row, "id=?", id).Error; err != nil {
		return nil, err
	}
	if row.Status != "PENDING" {
		return nil, errors.New("adjustment is not pending")
	}
	now := time.Now()
	row.Status = "REJECTED"
	row.ApprovedBy = user
	row.ApprovedAt = &now
	return &row, db.Save(&row).Error
}
