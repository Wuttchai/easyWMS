package services

import (
	"easywms-demo-v3/internal/models"
	"easywms-demo-v3/internal/testdb"
	"gorm.io/gorm"
	"sync"
	"sync/atomic"
	"testing"
	"time"
)

func fixture(t *testing.T) (*gorm.DB, models.Product, models.Location, models.Location) {
	t.Helper()
	db := testdb.Open(t)
	p := models.Product{SKU: "SKU-1", Name: "Test", Unit: "PCS"}
	a := models.Location{Code: "A"}
	b := models.Location{Code: "B"}
	for _, row := range []any{&p, &a, &b} {
		if err := db.Create(row).Error; err != nil {
			t.Fatal(err)
		}
	}
	return db, p, a, b
}
func TestTransferPreservesLotsAndExpiry(t *testing.T) {
	db, p, a, b := fixture(t)
	expiry := time.Date(2027, 1, 1, 0, 0, 0, 0, time.UTC)
	if err := ApplyMovement(db, "IN", MovementRequest{SKU: p.SKU, LocationCode: a.Code, Qty: 10, LotNo: "L1", ExpDate: &expiry}); err != nil {
		t.Fatal(err)
	}
	if err := TransferStock(db, TransferRequest{SKU: p.SKU, FromLocation: a.Code, ToLocation: b.Code, Qty: 4}); err != nil {
		t.Fatal(err)
	}
	for _, test := range []struct {
		loc models.Location
		qty float64
	}{{a, 6}, {b, 4}} {
		var inv models.Inventory
		var lot models.InventoryLot
		if err := db.First(&inv, "product_id=? AND location_id=?", p.ID, test.loc.ID).Error; err != nil {
			t.Fatal(err)
		}
		if err := db.First(&lot, "product_id=? AND location_id=?", p.ID, test.loc.ID).Error; err != nil {
			t.Fatal(err)
		}
		if inv.Qty != test.qty || lot.Qty != test.qty || !lot.ExpDate.Equal(expiry) {
			t.Fatalf("inventory/lot mismatch: %+v %+v", inv, lot)
		}
	}
	if err := ApplyMovement(db, "OUT", MovementRequest{SKU: p.SKU, LocationCode: b.Code, Qty: 1}); err == nil {
		t.Fatal("untracked issue consumed tracked lot")
	}
	if err := ApplyMovement(db, "OUT", MovementRequest{SKU: p.SKU, LocationCode: b.Code, Qty: 5, LotNo: "L1"}); err == nil {
		t.Fatal("over-issue succeeded")
	}
	var inv models.Inventory
	db.First(&inv, "product_id=? AND location_id=?", p.ID, b.ID)
	if inv.Qty != 4 {
		t.Fatal("failed request changed stock")
	}
}
func TestConcurrentIssuesCannotOversell(t *testing.T) {
	db, p, a, _ := fixture(t)
	if err := ApplyMovement(db, "IN", MovementRequest{SKU: p.SKU, LocationCode: a.Code, Qty: 10}); err != nil {
		t.Fatal(err)
	}
	var wg sync.WaitGroup
	var success atomic.Int32
	for i := 0; i < 20; i++ {
		wg.Add(1)
		go func() {
			defer wg.Done()
			if ApplyMovement(db, "OUT", MovementRequest{SKU: p.SKU, LocationCode: a.Code, Qty: 1}) == nil {
				success.Add(1)
			}
		}()
	}
	wg.Wait()
	var inv models.Inventory
	db.First(&inv, "product_id=? AND location_id=?", p.ID, a.ID)
	if success.Load() != 10 || inv.Qty != 0 {
		t.Fatalf("success=%d qty=%v", success.Load(), inv.Qty)
	}
}
func TestConcurrentReceiptsCreateOneBalance(t *testing.T) {
	db, p, a, _ := fixture(t)
	var wg sync.WaitGroup
	errs := make(chan error, 10)
	for i := 0; i < 10; i++ {
		wg.Add(1)
		go func() {
			defer wg.Done()
			errs <- ApplyMovement(db, "IN", MovementRequest{SKU: p.SKU, LocationCode: a.Code, Qty: 1})
		}()
	}
	wg.Wait()
	close(errs)
	for err := range errs {
		if err != nil {
			t.Fatal(err)
		}
	}
	var inv models.Inventory
	db.First(&inv, "product_id=? AND location_id=?", p.ID, a.ID)
	if inv.Qty != 10 {
		t.Fatal(inv.Qty)
	}
}
func TestApprovalOnceAndStaleCount(t *testing.T) {
	db, p, a, _ := fixture(t)
	if err := ApplyMovement(db, "IN", MovementRequest{SKU: p.SKU, LocationCode: a.Code, Qty: 10}); err != nil {
		t.Fatal(err)
	}
	adj, err := CreateAdjustment(db, AdjustmentRequest{SKU: p.SKU, LocationCode: a.Code, Qty: 2, Direction: "OUT"})
	if err != nil {
		t.Fatal(err)
	}
	var wg sync.WaitGroup
	var success atomic.Int32
	for i := 0; i < 5; i++ {
		wg.Add(1)
		go func() {
			defer wg.Done()
			if _, err := ApproveAdjustment(db, adj.ID, "admin"); err == nil {
				success.Add(1)
			}
		}()
	}
	wg.Wait()
	if success.Load() != 1 {
		t.Fatal("approval repeated", success.Load())
	}
	count, err := ApplyStockCount(db, StockCountRequest{SKU: p.SKU, LocationCode: a.Code, CountedQty: 7})
	if err != nil {
		t.Fatal(err)
	}
	if err := ApplyMovement(db, "IN", MovementRequest{SKU: p.SKU, LocationCode: a.Code, Qty: 1}); err != nil {
		t.Fatal(err)
	}
	if _, err := ApproveAdjustment(db, *count.AdjustmentID, "admin"); err == nil {
		t.Fatal("stale count was approved")
	}
}
