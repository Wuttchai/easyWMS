package models

import (
	"github.com/google/uuid"
	"gorm.io/gorm"
	"time"
)

func ensureID(id *uuid.UUID) {
	if *id == uuid.Nil {
		*id = uuid.New()
	}
}

type Product struct {
	ID           uuid.UUID `gorm:"type:uuid;primaryKey" json:"id"`
	SKU          string    `gorm:"uniqueIndex;not null" json:"sku"`
	Name         string    `gorm:"not null" json:"name"`
	Barcode      string    `gorm:"index" json:"barcode"`
	Unit         string    `gorm:"not null;default:PCS" json:"unit"`
	CategoryCode string    `gorm:"index" json:"category_code"`
	StorageType  string    `gorm:"index" json:"storage_type"`
	MinStock     float64   `gorm:"not null;default:0" json:"min_stock"`
	CreatedAt    time.Time `json:"created_at"`
	UpdatedAt    time.Time `json:"updated_at"`
}
type Warehouse struct {
	ID        uuid.UUID `gorm:"type:uuid;primaryKey" json:"id"`
	Code      string    `gorm:"uniqueIndex;not null" json:"code"`
	Name      string    `gorm:"not null" json:"name"`
	Address   string    `json:"address"`
	CreatedAt time.Time `json:"created_at"`
}
type Zone struct {
	ID            uuid.UUID `gorm:"type:uuid;primaryKey" json:"id"`
	WarehouseCode string    `gorm:"index;not null" json:"warehouse_code"`
	Code          string    `gorm:"uniqueIndex;not null" json:"code"`
	Name          string    `gorm:"not null" json:"name"`
	CreatedAt     time.Time `json:"created_at"`
}
type Location struct {
	ID            uuid.UUID `gorm:"type:uuid;primaryKey" json:"id"`
	WarehouseCode string    `gorm:"index" json:"warehouse_code"`
	ZoneCode      string    `gorm:"index" json:"zone_code"`
	Code          string    `gorm:"uniqueIndex;not null" json:"code"`
	Name          string    `json:"name"`
	CreatedAt     time.Time `json:"created_at"`
}
type Unit struct {
	ID        uuid.UUID `gorm:"type:uuid;primaryKey" json:"id"`
	Code      string    `gorm:"uniqueIndex;not null" json:"code"`
	Name      string    `gorm:"not null" json:"name"`
	CreatedAt time.Time `json:"created_at"`
}
type Supplier struct {
	ID        uuid.UUID `gorm:"type:uuid;primaryKey" json:"id"`
	Code      string    `gorm:"uniqueIndex;not null" json:"code"`
	Name      string    `gorm:"not null" json:"name"`
	Phone     string    `json:"phone"`
	Email     string    `json:"email"`
	CreatedAt time.Time `json:"created_at"`
}
type Customer struct {
	ID        uuid.UUID `gorm:"type:uuid;primaryKey" json:"id"`
	Code      string    `gorm:"uniqueIndex;not null" json:"code"`
	Name      string    `gorm:"not null" json:"name"`
	Phone     string    `json:"phone"`
	Email     string    `json:"email"`
	CreatedAt time.Time `json:"created_at"`
}
type Employee struct {
	ID           uuid.UUID `gorm:"type:uuid;primaryKey" json:"id"`
	Code         string    `gorm:"uniqueIndex;not null" json:"code"`
	Name         string    `gorm:"not null" json:"name"`
	Username     string    `gorm:"uniqueIndex;not null" json:"username"`
	PasswordHash string    `json:"-"`
	Role         string    `gorm:"not null" json:"role"`
	Email        string    `json:"email"`
	Active       bool      `gorm:"not null;default:true" json:"active"`
	CreatedAt    time.Time `json:"created_at"`
}
type ProductCategory struct {
	ID        uuid.UUID `gorm:"type:uuid;primaryKey" json:"id"`
	Code      string    `gorm:"uniqueIndex;not null" json:"code"`
	Name      string    `gorm:"not null" json:"name"`
	CreatedAt time.Time `json:"created_at"`
}
type StorageType struct {
	ID        uuid.UUID `gorm:"type:uuid;primaryKey" json:"id"`
	Code      string    `gorm:"uniqueIndex;not null" json:"code"`
	Name      string    `gorm:"not null" json:"name"`
	CreatedAt time.Time `json:"created_at"`
}
type ReasonCode struct {
	ID           uuid.UUID `gorm:"type:uuid;primaryKey" json:"id"`
	Code         string    `gorm:"uniqueIndex;not null" json:"code"`
	Name         string    `gorm:"not null" json:"name"`
	MovementType string    `gorm:"not null" json:"movement_type"`
	CreatedAt    time.Time `json:"created_at"`
}

type Inventory struct {
	ID         uuid.UUID `gorm:"type:uuid;primaryKey" json:"id"`
	ProductID  uuid.UUID `gorm:"type:uuid;index:idx_inventory_unique,unique" json:"product_id"`
	LocationID uuid.UUID `gorm:"type:uuid;index:idx_inventory_unique,unique" json:"location_id"`
	Qty        float64   `gorm:"not null;default:0" json:"qty"`
	Product    Product   `json:"product"`
	Location   Location  `json:"location"`
	UpdatedAt  time.Time `json:"updated_at"`
}
type InventoryLot struct {
	ID         uuid.UUID  `gorm:"type:uuid;primaryKey" json:"id"`
	ProductID  uuid.UUID  `gorm:"type:uuid;index:idx_lot_unique,unique" json:"product_id"`
	LocationID uuid.UUID  `gorm:"type:uuid;index:idx_lot_unique,unique" json:"location_id"`
	LotNo      string     `gorm:"index:idx_lot_unique,unique" json:"lot_no"`
	MfgDate    *time.Time `json:"mfg_date"`
	ExpDate    *time.Time `json:"exp_date"`
	Qty        float64    `gorm:"not null;default:0" json:"qty"`
	Product    Product    `json:"product"`
	Location   Location   `json:"location"`
	UpdatedAt  time.Time  `json:"updated_at"`
}
type StockMovement struct {
	ID         uuid.UUID `gorm:"type:uuid;primaryKey" json:"id"`
	ProductID  uuid.UUID `gorm:"type:uuid;index" json:"product_id"`
	LocationID uuid.UUID `gorm:"type:uuid;index" json:"location_id"`
	Type       string    `gorm:"not null" json:"type"`
	Qty        float64   `gorm:"not null" json:"qty"`
	LotNo      string    `json:"lot_no"`
	Reference  string    `json:"reference"`
	ReasonCode string    `json:"reason_code"`
	Note       string    `json:"note"`
	CreatedBy  string    `json:"created_by"`
	CreatedAt  time.Time `json:"created_at"`
	Product    Product   `json:"product"`
	Location   Location  `json:"location"`
}
type StockCount struct {
	ID           uuid.UUID  `gorm:"type:uuid;primaryKey" json:"id"`
	DocumentNo   string     `gorm:"uniqueIndex;not null" json:"document_no"`
	ProductID    uuid.UUID  `gorm:"type:uuid;index" json:"product_id"`
	LocationID   uuid.UUID  `gorm:"type:uuid;index" json:"location_id"`
	SystemQty    float64    `json:"system_qty"`
	CountedQty   float64    `json:"counted_qty"`
	Difference   float64    `json:"difference"`
	Status       string     `gorm:"not null" json:"status"`
	AdjustmentID *uuid.UUID `gorm:"type:uuid" json:"adjustment_id"`
	CreatedBy    string     `json:"created_by"`
	CreatedAt    time.Time  `json:"created_at"`
	Product      Product    `json:"product"`
	Location     Location   `json:"location"`
}
type Adjustment struct {
	ID          uuid.UUID  `gorm:"type:uuid;primaryKey" json:"id"`
	DocumentNo  string     `gorm:"uniqueIndex;not null" json:"document_no"`
	ProductID   uuid.UUID  `gorm:"type:uuid;index" json:"product_id"`
	LocationID  uuid.UUID  `gorm:"type:uuid;index" json:"location_id"`
	Qty         float64    `json:"qty"`
	Direction   string     `gorm:"not null" json:"direction"`
	ReasonCode  string     `json:"reason_code"`
	Note        string     `json:"note"`
	Status      string     `gorm:"not null;index" json:"status"`
	RequestedBy string     `json:"requested_by"`
	ApprovedBy  string     `json:"approved_by"`
	ApprovedAt  *time.Time `json:"approved_at"`
	CreatedAt   time.Time  `json:"created_at"`
	Product     Product    `json:"product"`
	Location    Location   `json:"location"`
}

func (p *Product) BeforeCreate(tx *gorm.DB) error         { ensureID(&p.ID); return nil }
func (m *Warehouse) BeforeCreate(tx *gorm.DB) error       { ensureID(&m.ID); return nil }
func (m *Zone) BeforeCreate(tx *gorm.DB) error            { ensureID(&m.ID); return nil }
func (m *Location) BeforeCreate(tx *gorm.DB) error        { ensureID(&m.ID); return nil }
func (m *Unit) BeforeCreate(tx *gorm.DB) error            { ensureID(&m.ID); return nil }
func (m *Supplier) BeforeCreate(tx *gorm.DB) error        { ensureID(&m.ID); return nil }
func (m *Customer) BeforeCreate(tx *gorm.DB) error        { ensureID(&m.ID); return nil }
func (m *Employee) BeforeCreate(tx *gorm.DB) error        { ensureID(&m.ID); return nil }
func (m *ProductCategory) BeforeCreate(tx *gorm.DB) error { ensureID(&m.ID); return nil }
func (m *StorageType) BeforeCreate(tx *gorm.DB) error     { ensureID(&m.ID); return nil }
func (m *ReasonCode) BeforeCreate(tx *gorm.DB) error      { ensureID(&m.ID); return nil }
func (m *Inventory) BeforeCreate(tx *gorm.DB) error       { ensureID(&m.ID); return nil }
func (m *InventoryLot) BeforeCreate(tx *gorm.DB) error    { ensureID(&m.ID); return nil }
func (m *StockMovement) BeforeCreate(tx *gorm.DB) error   { ensureID(&m.ID); return nil }
func (m *StockCount) BeforeCreate(tx *gorm.DB) error      { ensureID(&m.ID); return nil }
func (m *Adjustment) BeforeCreate(tx *gorm.DB) error      { ensureID(&m.ID); return nil }
