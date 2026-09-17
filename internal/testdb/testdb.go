// Package testdb creates a disposable schema inside an explicitly configured test database.
package testdb

import (
	"easywms-demo-v3/internal/models"
	"github.com/google/uuid"
	"gorm.io/driver/postgres"
	"gorm.io/gorm"
	"gorm.io/gorm/logger"
	"os"
	"strings"
	"testing"
)

func Open(t *testing.T) *gorm.DB {
	t.Helper()
	dsn := os.Getenv("TEST_DATABASE_DSN")
	if dsn == "" {
		t.Skip("TEST_DATABASE_DSN is required for PostgreSQL integration tests")
	}
	base, err := gorm.Open(postgres.Open(dsn), &gorm.Config{Logger: logger.Default.LogMode(logger.Silent)})
	if err != nil {
		t.Fatal(err)
	}
	schema := "test_" + strings.ReplaceAll(uuid.NewString(), "-", "")
	if err := base.Exec("CREATE SCHEMA " + schema).Error; err != nil {
		t.Fatal(err)
	}
	db, err := gorm.Open(postgres.Open(dsn+" search_path="+schema), &gorm.Config{Logger: logger.Default.LogMode(logger.Silent)})
	if err != nil {
		t.Fatal(err)
	}
	t.Cleanup(func() {
		sqlDB, _ := db.DB()
		sqlDB.Close()
		base.Exec("DROP SCHEMA " + schema + " CASCADE")
		sqlBase, _ := base.DB()
		sqlBase.Close()
	})
	if err := db.AutoMigrate(&models.Product{}, &models.Location{}, &models.Employee{}, &models.Inventory{}, &models.InventoryLot{}, &models.StockMovement{}, &models.StockCount{}, &models.Adjustment{}, &models.MutationReceipt{}); err != nil {
		t.Fatal(err)
	}
	return db
}
