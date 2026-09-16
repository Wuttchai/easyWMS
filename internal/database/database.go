package database

import (
	"easywms-demo-v3/internal/auth"
	"easywms-demo-v3/internal/config"
	"easywms-demo-v3/internal/models"
	"fmt"
	"github.com/google/uuid"
	"gorm.io/driver/postgres"
	"gorm.io/gorm"
)

func Connect(cfg config.Config) (*gorm.DB, error) {
	dsn := fmt.Sprintf("host=%s user=%s password=%s dbname=%s port=%s sslmode=%s TimeZone=Asia/Bangkok", cfg.DBHost, cfg.DBUser, cfg.DBPassword, cfg.DBName, cfg.DBPort, cfg.DBSSLMode)
	return gorm.Open(postgres.Open(dsn), &gorm.Config{})
}
func MigrateAndSeed(db *gorm.DB) error {
	if err := db.AutoMigrate(&models.Product{}, &models.Warehouse{}, &models.Zone{}, &models.Location{}, &models.Unit{}, &models.Supplier{}, &models.Customer{}, &models.Employee{}, &models.ProductCategory{}, &models.StorageType{}, &models.ReasonCode{}, &models.Inventory{}, &models.InventoryLot{}, &models.StockMovement{}, &models.StockCount{}, &models.Adjustment{}); err != nil {
		return err
	}
	if err := SyncStockCountStatuses(db); err != nil { return err }
	seed := func(model any, rows any) error {
		var n int64
		db.Model(model).Count(&n)
		if n == 0 {
			return db.Create(rows).Error
		}
		return nil
	}
	if err := seed(&models.Warehouse{}, &[]models.Warehouse{{ID: uuid.New(), Code: "WH01", Name: "Main Warehouse", Address: "Ayutthaya"}, {ID: uuid.New(), Code: "WH02", Name: "Finished Goods Warehouse", Address: "Ayutthaya"}}); err != nil {
		return err
	}
	if err := seed(&models.Unit{}, &[]models.Unit{{ID: uuid.New(), Code: "PCS", Name: "Piece"}, {ID: uuid.New(), Code: "BOX", Name: "Box"}, {ID: uuid.New(), Code: "KG", Name: "Kilogram"}}); err != nil {
		return err
	}
	if err := seed(&models.Zone{}, &[]models.Zone{{ID: uuid.New(), WarehouseCode: "WH01", Code: "RAW", Name: "Raw Material Zone"}, {ID: uuid.New(), WarehouseCode: "WH02", Code: "FG", Name: "Finished Goods Zone"}}); err != nil {
		return err
	}
	if err := seed(&models.ProductCategory{}, &[]models.ProductCategory{{ID: uuid.New(), Code: "RAW", Name: "Raw Material"}, {ID: uuid.New(), Code: "FG", Name: "Finished Goods"}}); err != nil {
		return err
	}
	if err := seed(&models.StorageType{}, &[]models.StorageType{{ID: uuid.New(), Code: "NORMAL", Name: "Normal Storage"}, {ID: uuid.New(), Code: "QC", Name: "Quality Hold"}}); err != nil {
		return err
	}
	if err := seed(&models.ReasonCode{}, &[]models.ReasonCode{{ID: uuid.New(), Code: "RCV", Name: "Receive", MovementType: "IN"}, {ID: uuid.New(), Code: "ISS", Name: "Issue", MovementType: "OUT"}, {ID: uuid.New(), Code: "TRF", Name: "Transfer", MovementType: "TRANSFER"}, {ID: uuid.New(), Code: "CNT", Name: "Stock Count", MovementType: "ADJUST"}, {ID: uuid.New(), Code: "DMG", Name: "Damaged", MovementType: "ADJUST"}, {ID: uuid.New(), Code: "LOST", Name: "Lost", MovementType: "ADJUST"}}); err != nil {
		return err
	}
	if err := seed(&models.Supplier{}, &[]models.Supplier{{ID: uuid.New(), Code: "SUP-001", Name: "Ayutthaya Steel Supply", Phone: "035-000-001", Email: "sales@example.com"}}); err != nil {
		return err
	}
	if err := seed(&models.Customer{}, &[]models.Customer{{ID: uuid.New(), Code: "CUS-001", Name: "Demo Factory A", Phone: "081-000-0001", Email: "factory@example.com"}}); err != nil {
		return err
	}
	var empCount int64
	db.Model(&models.Employee{}).Count(&empCount)
	if empCount == 0 {
		rows := []models.Employee{{ID: uuid.New(), Code: "EMP-001", Name: "Demo Admin", Username: "admin", PasswordHash: auth.HashPassword("admin123"), Role: "ADMIN", Email: "admin@easywms.local", Active: true}, {ID: uuid.New(), Code: "EMP-002", Name: "Warehouse Operator", Username: "operator", PasswordHash: auth.HashPassword("operator123"), Role: "WAREHOUSE", Email: "operator@easywms.local", Active: true}, {ID: uuid.New(), Code: "EMP-003", Name: "Demo Supervisor", Username: "supervisor", PasswordHash: auth.HashPassword("supervisor123"), Role: "SUPERVISOR", Email: "supervisor@easywms.local", Active: true}}
		if err := db.Create(&rows).Error; err != nil {
			return err
		}
	}
	if err := seed(&models.Product{}, &[]models.Product{{ID: uuid.New(), SKU: "RM-STEEL-001", Name: "เหล็กแผ่น 2.0 mm", Barcode: "885000000001", Unit: "PCS", CategoryCode: "RAW", StorageType: "NORMAL", MinStock: 20}, {ID: uuid.New(), SKU: "RM-BOLT-001", Name: "Bolt M8", Barcode: "885000000002", Unit: "PCS", CategoryCode: "RAW", StorageType: "NORMAL", MinStock: 100}, {ID: uuid.New(), SKU: "FG-BOX-001", Name: "กล่องสินค้าสำเร็จรูป", Barcode: "885000000003", Unit: "BOX", CategoryCode: "FG", StorageType: "NORMAL", MinStock: 10}}); err != nil {
		return err
	}
	if err := seed(&models.Location{}, &[]models.Location{{ID: uuid.New(), WarehouseCode: "WH01", ZoneCode: "RAW", Code: "A-01-01", Name: "Rack A / Bay 01 / Level 01"}, {ID: uuid.New(), WarehouseCode: "WH01", ZoneCode: "RAW", Code: "A-01-02", Name: "Rack A / Bay 01 / Level 02"}, {ID: uuid.New(), WarehouseCode: "WH02", ZoneCode: "FG", Code: "B-01-01", Name: "Rack B / Bay 01 / Level 01"}}); err != nil {
		return err
	}
	var invCount int64
	db.Model(&models.Inventory{}).Count(&invCount)
	if invCount == 0 {
		var ps []models.Product
		var ls []models.Location
		db.Find(&ps)
		db.Find(&ls)
		for i, p := range ps {
			l := ls[i%len(ls)]
			q := []float64{75, 250, 8}[i%3]
			inv := models.Inventory{ID: uuid.New(), ProductID: p.ID, LocationID: l.ID, Qty: q}
			if err := db.Create(&inv).Error; err != nil {
				return err
			}
			mv := models.StockMovement{ID: uuid.New(), ProductID: p.ID, LocationID: l.ID, Type: "IN", Qty: q, Reference: "OPENING", Note: "Demo opening balance", CreatedBy: "SYSTEM"}
			if err := db.Create(&mv).Error; err != nil {
				return err
			}
		}
	}
	return nil
}
