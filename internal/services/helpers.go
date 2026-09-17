package services

import (
	"easywms-demo-v3/internal/models"
	"errors"
	"github.com/google/uuid"
	"gorm.io/gorm"
	"gorm.io/gorm/clause"
	"time"
)

func findProductLocation(tx *gorm.DB, sku, locationCode string) (models.Product, models.Location, error) {
	var p models.Product
	if err := tx.Where("sku = ? OR barcode = ?", sku, sku).First(&p).Error; err != nil {
		return p, models.Location{}, errors.New("product not found")
	}
	var l models.Location
	if err := tx.Where("code = ?", locationCode).First(&l).Error; err != nil {
		return p, l, errors.New("location not found")
	}
	return p, l, nil
}

func datesEqual(a, b *time.Time) bool {
	if a == nil || b == nil {
		return a == nil && b == nil
	}
	return a.Equal(*b)
}

func getInventoryForUpdate(tx *gorm.DB, pid, lid uuid.UUID) (*models.Inventory, error) {
	var v models.Inventory
	err := tx.Clauses(clause.Locking{Strength: "UPDATE"}).Where("product_id=? AND location_id=?", pid, lid).First(&v).Error
	if errors.Is(err, gorm.ErrRecordNotFound) {
		v = models.Inventory{ID: uuid.New(), ProductID: pid, LocationID: lid}
		if err := tx.Create(&v).Error; err != nil {
			return nil, err
		}
		return &v, nil
	}
	return &v, err
}

func getLotForUpdate(tx *gorm.DB, pid, lid uuid.UUID, lot string) (*models.InventoryLot, error) {
	var v models.InventoryLot
	err := tx.Clauses(clause.Locking{Strength: "UPDATE"}).Where("product_id=? AND location_id=? AND lot_no=?", pid, lid, lot).First(&v).Error
	if errors.Is(err, gorm.ErrRecordNotFound) {
		v = models.InventoryLot{ID: uuid.New(), ProductID: pid, LocationID: lid, LotNo: lot}
		if err := tx.Create(&v).Error; err != nil {
			return nil, err
		}
		return &v, nil
	}
	return &v, err
}
