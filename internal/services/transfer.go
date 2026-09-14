package services

import (
	"easywms-demo-v3/internal/models"
	"errors"
	"github.com/google/uuid"
	"gorm.io/gorm"
)

func TransferStock(db *gorm.DB, req TransferRequest) error {
	if req.FromLocation == req.ToLocation {
		return errors.New("source and destination must differ")
	}
	return db.Transaction(func(tx *gorm.DB) error {
		p, from, err := findProductLocation(tx, req.SKU, req.FromLocation)
		if err != nil {
			return err
		}
		var to models.Location
		if err := tx.Where("code=?", req.ToLocation).First(&to).Error; err != nil {
			return errors.New("destination not found")
		}
		a, err := getInventoryForUpdate(tx, p.ID, from.ID)
		if err != nil {
			return err
		}
		if a.Qty < req.Qty {
			return errors.New("insufficient stock")
		}
		b, err := getInventoryForUpdate(tx, p.ID, to.ID)
		if err != nil {
			return err
		}
		a.Qty -= req.Qty
		b.Qty += req.Qty
		if err := tx.Save(a).Error; err != nil {
			return err
		}
		if err := tx.Save(b).Error; err != nil {
			return err
		}
		if err := tx.Create(&models.StockMovement{ID: uuid.New(), ProductID: p.ID, LocationID: from.ID, Type: "TRANSFER_OUT", Qty: req.Qty, Reference: req.Reference, ReasonCode: req.ReasonCode, Note: "To " + to.Code, CreatedBy: req.CreatedBy}).Error; err != nil {
			return err
		}
		return tx.Create(&models.StockMovement{ID: uuid.New(), ProductID: p.ID, LocationID: to.ID, Type: "TRANSFER_IN", Qty: req.Qty, Reference: req.Reference, ReasonCode: req.ReasonCode, Note: "From " + from.Code, CreatedBy: req.CreatedBy}).Error
	})
}
