package services

import (
	"easywms-demo-v3/internal/models"
	"errors"
	"github.com/google/uuid"
	"gorm.io/gorm"
)

func ApplyMovement(db *gorm.DB, typ string, req MovementRequest) error {
	if !validQuantity(req.Qty) {
		return errors.New("quantity must be greater than zero")
	}
	if typ != "IN" && typ != "OUT" {
		return errors.New("invalid movement type")
	}
	if req.MfgDate != nil && req.ExpDate != nil && req.ExpDate.Before(*req.MfgDate) {
		return errors.New("expiry date must not precede manufacturing date")
	}
	return db.Transaction(func(tx *gorm.DB) error {
		var customer models.Customer
		var supplier models.Supplier
		if req.SupplierCode != "" {
			if typ != "IN" {
				return errors.New("supplier is only allowed for receiving goods")
			}
			if err := tx.Where("code = ?", req.SupplierCode).First(&supplier).Error; err != nil {
				return errors.New("supplier not found")
			}
		}
		if typ == "OUT" {
			if req.IssuePurpose == "" {
				req.IssuePurpose = "INTERNAL"
			}
			switch req.IssuePurpose {
			case "CUSTOMER":
				if req.CustomerCode == "" {
					return errors.New("customer is required for customer delivery")
				}
				if err := tx.Where("code = ?", req.CustomerCode).First(&customer).Error; err != nil {
					return errors.New("customer not found")
				}
			case "INTERNAL":
				if req.CustomerCode != "" {
					return errors.New("customer is only allowed for customer delivery")
				}
			default:
				return errors.New("invalid issue purpose")
			}
		} else {
			req.IssuePurpose = ""
		}
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
		if typ == "OUT" {
			untracked, err := untrackedStock(tx, inv)
			if err != nil {
				return err
			}
			if req.LotNo == "" && req.Qty > untracked+0.0000001 {
				return errors.New("select a lot; untracked stock is insufficient")
			}
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
				if lot.Qty > 0 && (!datesEqual(lot.MfgDate, req.MfgDate) || !datesEqual(lot.ExpDate, req.ExpDate)) {
					return errors.New("lot dates differ from existing stock")
				}
				lot.Qty += req.Qty
				lot.MfgDate = req.MfgDate
				lot.ExpDate = req.ExpDate
			}
			if err := tx.Save(lot).Error; err != nil {
				return err
			}
		}
		return tx.Create(&models.StockMovement{ID: uuid.New(), ProductID: p.ID, LocationID: l.ID, Type: typ, Qty: req.Qty, LotNo: req.LotNo, Reference: req.Reference, ReasonCode: req.ReasonCode, Note: req.Note, CreatedBy: req.CreatedBy, IssuePurpose: req.IssuePurpose, CustomerCode: customer.Code, CustomerName: customer.Name, SupplierCode: supplier.Code, SupplierName: supplier.Name}).Error
	})
}
