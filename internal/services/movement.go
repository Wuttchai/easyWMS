package services

import (
	"easywms-demo-v3/internal/models"
	"errors"
	"github.com/google/uuid"
	"gorm.io/gorm"
)

func ApplyMovement(db *gorm.DB, typ string, req MovementRequest) error {
	return db.Transaction(func(tx *gorm.DB) error {
		p, l, err := findProductLocation(tx, req.SKU, req.LocationCode)
		if err != nil {
			return err
		}
		inv, err := getInventoryForUpdate(tx, p.ID, l.ID)
		if err != nil {
			return err
		}
		if typ == "OUT" {
			if inv.Qty < req.Qty {
				return errors.New("insufficient stock")
			}
			inv.Qty -= req.Qty
		} else {
			inv.Qty += req.Qty
		}
		if err := tx.Save(inv).Error; err != nil {
			return err
		}
		if req.LotNo != "" {
			lot, err := getLotForUpdate(tx, p.ID, l.ID, req.LotNo)
			if err != nil {
				return err
			}
			if typ == "OUT" {
				if lot.Qty < req.Qty {
					return errors.New("insufficient lot stock")
				}
				lot.Qty -= req.Qty
			} else {
				lot.Qty += req.Qty
				lot.MfgDate = req.MfgDate
				lot.ExpDate = req.ExpDate
			}
			if err := tx.Save(lot).Error; err != nil {
				return err
			}
		}
		return tx.Create(&models.StockMovement{ID: uuid.New(), ProductID: p.ID, LocationID: l.ID, Type: typ, Qty: req.Qty, LotNo: req.LotNo, Reference: req.Reference, ReasonCode: req.ReasonCode, Note: req.Note, CreatedBy: req.CreatedBy}).Error
	})
}
