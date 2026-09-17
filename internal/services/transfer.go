package services

import (
	"easywms-demo-v3/internal/models"
	"errors"
	"gorm.io/gorm"
)

func TransferStock(db *gorm.DB, req TransferRequest) error {
	if !validQuantity(req.Qty) {
		return errors.New("quantity must be greater than zero")
	}
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
		if err := lockProduct(tx, p.ID); err != nil {
			return err
		}
		a, err := getInventoryForUpdate(tx, p.ID, from.ID)
		if err != nil {
			return err
		}
		if a.Qty < req.Qty {
			return errors.New("insufficient stock")
		}
		if _, err := untrackedStock(tx, a); err != nil {
			return err
		}
		b, err := getInventoryForUpdate(tx, p.ID, to.ID)
		if err != nil {
			return err
		}
		if _, err := untrackedStock(tx, b); err != nil {
			return err
		}
		if err := transferLots(tx, p.ID, from.ID, to.ID, req.Qty, req.Reference, req.ReasonCode, req.CreatedBy); err != nil {
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
		return nil
	})
}
