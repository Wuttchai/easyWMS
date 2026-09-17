package services

import (
	"easywms-demo-v3/internal/models"
	"errors"
	"github.com/google/uuid"
	"gorm.io/gorm"
	"gorm.io/gorm/clause"
	"math"
)

func validQuantity(q float64) bool { return !math.IsNaN(q) && !math.IsInf(q, 0) && q > 0 }

// All stock writers lock the product first. This also serializes creation of a
// previously absent inventory row and keeps opposite-direction transfers safe.
func lockProduct(tx *gorm.DB, id uuid.UUID) error {
	var product models.Product
	return tx.Clauses(clause.Locking{Strength: "UPDATE"}).First(&product, "id=?", id).Error
}

func untrackedStock(tx *gorm.DB, inv *models.Inventory) (float64, error) {
	var total float64
	if err := tx.Model(&models.InventoryLot{}).Where("product_id=? AND location_id=?", inv.ProductID, inv.LocationID).Select("COALESCE(SUM(qty),0)").Scan(&total).Error; err != nil {
		return 0, err
	}
	if total > inv.Qty+0.0000001 {
		return 0, errors.New("lot balance exceeds inventory; reconcile stock before continuing")
	}
	return math.Max(0, inv.Qty-total), nil
}

// Transfer named lots in expiry order, followed by stock without a lot.
func transferLots(tx *gorm.DB, pid, from, to uuid.UUID, qty float64, reference, reason, user string) error {
	var lots []models.InventoryLot
	if err := tx.Clauses(clause.Locking{Strength: "UPDATE"}).Where("product_id=? AND location_id=? AND qty>0", pid, from).Order("exp_date ASC NULLS LAST, lot_no ASC").Find(&lots).Error; err != nil {
		return err
	}
	remaining := qty
	writeMovement := func(amount float64, lot string) error {
		for _, m := range []models.StockMovement{
			{ID: uuid.New(), ProductID: pid, LocationID: from, Type: "TRANSFER_OUT", Qty: amount, LotNo: lot, Reference: reference, ReasonCode: reason, CreatedBy: user},
			{ID: uuid.New(), ProductID: pid, LocationID: to, Type: "TRANSFER_IN", Qty: amount, LotNo: lot, Reference: reference, ReasonCode: reason, CreatedBy: user},
		} {
			var location models.Location
			other := to
			prefix := "To "
			if m.Type == "TRANSFER_IN" {
				other = from
				prefix = "From "
			}
			if err := tx.First(&location, "id=?", other).Error; err != nil {
				return err
			}
			m.Note = prefix + location.Code
			if err := tx.Create(&m).Error; err != nil {
				return err
			}
		}
		return nil
	}
	for _, lot := range lots {
		if remaining <= 0 {
			break
		}
		amount := math.Min(remaining, lot.Qty)
		dest, err := getLotForUpdate(tx, pid, to, lot.LotNo)
		if err != nil {
			return err
		}
		sameDate := func(a, b *models.InventoryLot) bool {
			return datesEqual(a.MfgDate, b.MfgDate) && datesEqual(a.ExpDate, b.ExpDate)
		}
		if dest.Qty > 0 && !sameDate(dest, &lot) {
			return errors.New("destination lot dates differ from source")
		}
		dest.MfgDate = lot.MfgDate
		dest.ExpDate = lot.ExpDate
		dest.Qty += amount
		lot.Qty -= amount
		if err := tx.Save(dest).Error; err != nil {
			return err
		}
		if err := tx.Save(&lot).Error; err != nil {
			return err
		}
		if err := writeMovement(amount, lot.LotNo); err != nil {
			return err
		}
		remaining -= amount
	}
	if remaining > 0 {
		return writeMovement(remaining, "")
	}
	return nil
}
