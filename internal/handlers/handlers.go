package handlers

import (
	"encoding/csv"
	"errors"
	"fmt"
	"io"

	"strconv"
	"strings"
	"time"

	"easywms-demo-v3/internal/auth"
	"easywms-demo-v3/internal/models"
	"easywms-demo-v3/internal/services"

	"github.com/gin-gonic/gin"
	"github.com/google/uuid"
	"gorm.io/gorm"
)

type Handler struct {
	DB        *gorm.DB
	JWTSecret string
}

type loginRequest struct {
	Username string `json:"username"`
	Password string `json:"password"`
}

func (h Handler) Login(c *gin.Context) {
	var req loginRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(400, gin.H{"error": err.Error()})
		return
	}
	var u models.Employee
	if err := h.DB.Where("username=? AND active=?", req.Username, true).First(&u).Error; err != nil || !auth.VerifyPassword(u.PasswordHash, req.Password) {
		c.JSON(401, gin.H{"error": "invalid username or password"})
		return
	}
	if !strings.HasPrefix(u.PasswordHash, "$2") {
		if err := h.DB.Model(&u).Update("password_hash", auth.HashPassword(req.Password)).Error; err != nil {
			c.JSON(500, gin.H{"error": "unable to upgrade password"})
			return
		}
	}
	token, _ := auth.Sign(h.JWTSecret, auth.Claims{UserID: u.ID.String(), Username: u.Username, Role: u.Role, Exp: auth.ExpiresIn(12)})
	c.JSON(200, gin.H{"token": token, "user": gin.H{"id": u.ID, "name": u.Name, "username": u.Username, "role": u.Role}})
}
func (h Handler) Auth() gin.HandlerFunc {
	return func(c *gin.Context) {
		token := auth.Bearer(c.GetHeader("Authorization"))
		cl, err := auth.Parse(h.JWTSecret, token)
		if err != nil {
			c.AbortWithStatusJSON(401, gin.H{"error": "unauthorized"})
			return
		}
		var account models.Employee
		if err := h.DB.Where("id = ? AND active = ?", cl.UserID, true).First(&account).Error; err != nil {
			c.AbortWithStatusJSON(401, gin.H{"error": "unauthorized"})
			return
		}
		c.Set("username", account.Username)
		c.Set("role", account.Role)
		c.Set("user_id", cl.UserID)
		c.Next()
	}
}
func RequireRoles(roles ...string) gin.HandlerFunc {
	set := map[string]bool{}
	for _, r := range roles {
		set[r] = true
	}
	return func(c *gin.Context) {
		role, _ := c.Get("role")
		if !set[fmt.Sprint(role)] {
			c.AbortWithStatusJSON(403, gin.H{"error": "permission denied"})
			return
		}
		c.Next()
	}
}
func user(c *gin.Context) string { v, _ := c.Get("username"); return fmt.Sprint(v) }

func (h Handler) Dashboard(c *gin.Context) {
	var products, locations, warehouses, employees, pending int64
	var inv []models.Inventory
	var movements []models.StockMovement
	h.DB.Model(&models.Product{}).Count(&products)
	h.DB.Model(&models.Location{}).Count(&locations)
	h.DB.Model(&models.Warehouse{}).Count(&warehouses)
	h.DB.Model(&models.Employee{}).Where("active=?", true).Count(&employees)
	h.DB.Model(&models.Adjustment{}).Where("status=?", "PENDING").Count(&pending)
	h.DB.Preload("Product").Preload("Location").Find(&inv)
	h.DB.Preload("Product").Preload("Location").Order("created_at desc").Limit(8).Find(&movements)
	total := 0.0
	low := 0
	for _, x := range inv {
		total += x.Qty
		if x.Qty <= x.Product.MinStock {
			low++
		}
	}
	c.JSON(200, gin.H{"products": products, "locations": locations, "warehouses": warehouses, "employees": employees, "total_qty": total, "low_stock": low, "pending_adjustments": pending, "recent_movements": movements})
}

func (h Handler) bindCreate(c *gin.Context, v any) bool {
	if err := c.ShouldBindJSON(v); err != nil {
		c.JSON(400, gin.H{"error": err.Error()})
		return false
	}
	return true
}
func (h Handler) create(c *gin.Context, v any) {
	if err := h.DB.Create(v).Error; err != nil {
		c.JSON(400, gin.H{"error": err.Error()})
		return
	}
	c.JSON(201, v)
}
func (h Handler) ListProducts(c *gin.Context) {
	var r []models.Product
	h.DB.Order("sku").Find(&r)
	c.JSON(200, r)
}
func (h Handler) CreateProduct(c *gin.Context) {
	var r models.Product
	if !h.bindCreate(c, &r) {
		return
	}
	if r.Unit == "" {
		r.Unit = "PCS"
	}
	h.create(c, &r)
}
func (h Handler) ListWarehouses(c *gin.Context) {
	var r []models.Warehouse
	h.DB.Order("code").Find(&r)
	c.JSON(200, r)
}
func (h Handler) CreateWarehouse(c *gin.Context) {
	var r models.Warehouse
	if h.bindCreate(c, &r) {
		h.create(c, &r)
	}
}
func (h Handler) ListZones(c *gin.Context) {
	var r []models.Zone
	h.DB.Order("code").Find(&r)
	c.JSON(200, r)
}
func (h Handler) CreateZone(c *gin.Context) {
	var r models.Zone
	if h.bindCreate(c, &r) {
		h.create(c, &r)
	}
}
func (h Handler) ListLocations(c *gin.Context) {
	var r []models.Location
	h.DB.Order("code").Find(&r)
	c.JSON(200, r)
}
func (h Handler) CreateLocation(c *gin.Context) {
	var r models.Location
	if h.bindCreate(c, &r) {
		r.WarehouseCode = strings.TrimSpace(r.WarehouseCode)
		r.ZoneCode = strings.TrimSpace(r.ZoneCode)
		r.Code = strings.TrimSpace(r.Code)
		r.Name = strings.TrimSpace(r.Name)
		if r.WarehouseCode == "" || r.ZoneCode == "" || r.Code == "" || r.Name == "" {
			c.JSON(400, gin.H{"error": "warehouse, zone, code and name are required"})
			return
		}
		var warehouse models.Warehouse
		if err := h.DB.Where("code = ?", r.WarehouseCode).First(&warehouse).Error; err != nil {
			c.JSON(400, gin.H{"error": "warehouse not found"})
			return
		}
		var zone models.Zone
		if err := h.DB.Where("code = ? AND warehouse_code = ?", r.ZoneCode, r.WarehouseCode).First(&zone).Error; err != nil {
			c.JSON(400, gin.H{"error": "zone does not belong to the selected warehouse"})
			return
		}
		h.create(c, &r)
	}
}
func (h Handler) DeleteLocation(c *gin.Context) {
	id, err := uuid.Parse(c.Param("id"))
	if err != nil {
		c.JSON(400, gin.H{"error": "invalid location id"})
		return
	}
	var location models.Location
	if err := h.DB.First(&location, "id = ?", id).Error; err != nil {
		c.JSON(404, gin.H{"error": "location not found"})
		return
	}
	var qty float64
	h.DB.Model(&models.Inventory{}).Where("location_id = ?", id).Select("COALESCE(SUM(qty), 0)").Scan(&qty)
	if qty > 0 {
		c.JSON(400, gin.H{"error": "cannot delete a location with stock"})
		return
	}
	if err := h.DB.Delete(&location).Error; err != nil {
		c.JSON(400, gin.H{"error": err.Error()})
		return
	}
	c.JSON(200, gin.H{"message": "location deleted"})
}
func (h Handler) DeleteMaster(c *gin.Context) {
	id, err := uuid.Parse(c.Param("id"))
	if err != nil {
		c.JSON(400, gin.H{"error": "invalid master id"})
		return
	}
	resource := c.Param("resource")
	var code string
	var record any
	switch resource {
	case "products":
		var row models.Product
		record, code = &row, "product"
		if err = h.DB.First(&row, "id = ?", id).Error; err == nil {
			var count int64
			h.DB.Model(&models.Inventory{}).Where("product_id = ?", id).Count(&count)
			if count == 0 {
				h.DB.Model(&models.InventoryLot{}).Where("product_id = ?", id).Count(&count)
			}
			if count == 0 {
				h.DB.Model(&models.StockMovement{}).Where("product_id = ?", id).Count(&count)
			}
			if count == 0 {
				h.DB.Model(&models.StockCount{}).Where("product_id = ?", id).Count(&count)
			}
			if count == 0 {
				h.DB.Model(&models.Adjustment{}).Where("product_id = ?", id).Count(&count)
			}
			if count > 0 {
				c.JSON(400, gin.H{"error": "cannot delete product because it is already used"})
				return
			}
		}
	case "warehouses":
		var row models.Warehouse
		record, code = &row, "warehouse"
		if err = h.DB.First(&row, "id = ?", id).Error; err == nil {
			var count int64
			h.DB.Model(&models.Zone{}).Where("warehouse_code = ?", row.Code).Count(&count)
			if count == 0 {
				h.DB.Model(&models.Location{}).Where("warehouse_code = ?", row.Code).Count(&count)
			}
			if count > 0 {
				c.JSON(400, gin.H{"error": "cannot delete warehouse because it has zones or locations"})
				return
			}
		}
	case "zones":
		var row models.Zone
		record, code = &row, "zone"
		if err = h.DB.First(&row, "id = ?", id).Error; err == nil {
			var count int64
			h.DB.Model(&models.Location{}).Where("zone_code = ? AND warehouse_code = ?", row.Code, row.WarehouseCode).Count(&count)
			if count > 0 {
				c.JSON(400, gin.H{"error": "cannot delete zone because it has locations"})
				return
			}
		}
	case "locations":
		var row models.Location
		record, code = &row, "location"
		if err = h.DB.First(&row, "id = ?", id).Error; err == nil {
			var count int64
			h.DB.Model(&models.Inventory{}).Where("location_id = ?", id).Count(&count)
			if count == 0 {
				h.DB.Model(&models.InventoryLot{}).Where("location_id = ?", id).Count(&count)
			}
			if count == 0 {
				h.DB.Model(&models.StockMovement{}).Where("location_id = ?", id).Count(&count)
			}
			if count == 0 {
				h.DB.Model(&models.StockCount{}).Where("location_id = ?", id).Count(&count)
			}
			if count == 0 {
				h.DB.Model(&models.Adjustment{}).Where("location_id = ?", id).Count(&count)
			}
			if count > 0 {
				c.JSON(400, gin.H{"error": "cannot delete location because it is already used"})
				return
			}
		}
	case "units":
		var row models.Unit
		record, code = &row, "unit"
		if err = h.DB.First(&row, "id = ?", id).Error; err == nil {
			var count int64
			h.DB.Model(&models.Product{}).Where("unit = ?", row.Code).Count(&count)
			if count > 0 {
				c.JSON(400, gin.H{"error": "cannot delete unit because it is used by products"})
				return
			}
		}
	case "categories":
		var row models.ProductCategory
		record, code = &row, "category"
		if err = h.DB.First(&row, "id = ?", id).Error; err == nil {
			var count int64
			h.DB.Model(&models.Product{}).Where("category_code = ?", row.Code).Count(&count)
			if count > 0 {
				c.JSON(400, gin.H{"error": "cannot delete category because it is used by products"})
				return
			}
		}
	case "storage-types":
		var row models.StorageType
		record, code = &row, "storage type"
		if err = h.DB.First(&row, "id = ?", id).Error; err == nil {
			var count int64
			h.DB.Model(&models.Product{}).Where("storage_type = ?", row.Code).Count(&count)
			if count > 0 {
				c.JSON(400, gin.H{"error": "cannot delete storage type because it is used by products"})
				return
			}
		}
	case "reason-codes":
		var row models.ReasonCode
		record, code = &row, "reason code"
		if err = h.DB.First(&row, "id = ?", id).Error; err == nil {
			var count int64
			h.DB.Model(&models.StockMovement{}).Where("reason_code = ?", row.Code).Count(&count)
			if count == 0 {
				h.DB.Model(&models.Adjustment{}).Where("reason_code = ?", row.Code).Count(&count)
			}
			if count > 0 {
				c.JSON(400, gin.H{"error": "cannot delete reason code because it is already used"})
				return
			}
		}
	case "employees":
		var row models.Employee
		record, code = &row, "employee"
		if err = h.DB.First(&row, "id = ?", id).Error; err == nil {
			if row.Username == "admin" {
				c.JSON(400, gin.H{"error": "cannot delete the demo admin account"})
				return
			}
			var count int64
			h.DB.Model(&models.StockMovement{}).Where("created_by = ?", row.Username).Count(&count)
			if count == 0 {
				h.DB.Model(&models.StockCount{}).Where("created_by = ?", row.Username).Count(&count)
			}
			if count == 0 {
				h.DB.Model(&models.Adjustment{}).Where("requested_by = ? OR approved_by = ?", row.Username, row.Username).Count(&count)
			}
			if count > 0 {
				c.JSON(400, gin.H{"error": "cannot delete employee because the account is used in audit history"})
				return
			}
		}
	case "suppliers":
		var row models.Supplier
		record, code = &row, "supplier"
		err = h.DB.First(&row, "id = ?", id).Error
	case "customers":
		var row models.Customer
		record, code = &row, "customer"
		err = h.DB.First(&row, "id = ?", id).Error
	default:
		c.JSON(404, gin.H{"error": "master delete is not supported"})
		return
	}
	if err != nil {
		if errors.Is(err, gorm.ErrRecordNotFound) {
			c.JSON(404, gin.H{"error": code + " not found"})
			return
		}
		c.JSON(400, gin.H{"error": err.Error()})
		return
	}
	if err := h.DB.Delete(record).Error; err != nil {
		c.JSON(400, gin.H{"error": err.Error()})
		return
	}
	c.JSON(200, gin.H{"message": code + " deleted"})
}
func (h Handler) ListUnits(c *gin.Context) {
	var r []models.Unit
	h.DB.Order("code").Find(&r)
	c.JSON(200, r)
}
func (h Handler) CreateUnit(c *gin.Context) {
	var r models.Unit
	if h.bindCreate(c, &r) {
		h.create(c, &r)
	}
}
func (h Handler) ListSuppliers(c *gin.Context) {
	var r []models.Supplier
	h.DB.Order("code").Find(&r)
	c.JSON(200, r)
}
func (h Handler) CreateSupplier(c *gin.Context) {
	var r models.Supplier
	if h.bindCreate(c, &r) {
		h.create(c, &r)
	}
}
func (h Handler) ListCustomers(c *gin.Context) {
	var r []models.Customer
	h.DB.Order("code").Find(&r)
	c.JSON(200, r)
}
func (h Handler) CreateCustomer(c *gin.Context) {
	var r models.Customer
	if h.bindCreate(c, &r) {
		h.create(c, &r)
	}
}
func (h Handler) ListCategories(c *gin.Context) {
	var r []models.ProductCategory
	h.DB.Order("code").Find(&r)
	c.JSON(200, r)
}
func (h Handler) CreateCategory(c *gin.Context) {
	var r models.ProductCategory
	if h.bindCreate(c, &r) {
		h.create(c, &r)
	}
}
func (h Handler) ListStorageTypes(c *gin.Context) {
	var r []models.StorageType
	h.DB.Order("code").Find(&r)
	c.JSON(200, r)
}
func (h Handler) CreateStorageType(c *gin.Context) {
	var r models.StorageType
	if h.bindCreate(c, &r) {
		h.create(c, &r)
	}
}
func (h Handler) ListReasonCodes(c *gin.Context) {
	var r []models.ReasonCode
	h.DB.Order("code").Find(&r)
	c.JSON(200, r)
}
func (h Handler) CreateReasonCode(c *gin.Context) {
	var r models.ReasonCode
	if h.bindCreate(c, &r) {
		h.create(c, &r)
	}
}
func (h Handler) ListEmployees(c *gin.Context) {
	var r []models.Employee
	h.DB.Order("code").Find(&r)
	c.JSON(200, r)
}
func (h Handler) CreateEmployee(c *gin.Context) {
	var req struct{ Code, Name, Username, Password, Role, Email string }
	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(400, gin.H{"error": err.Error()})
		return
	}
	if len(req.Password) < 12 || len(req.Password) > 72 || strings.TrimSpace(req.Username) == "" || strings.TrimSpace(req.Code) == "" || strings.TrimSpace(req.Name) == "" {
		c.JSON(400, gin.H{"error": "name, code and username are required; password must be 12 to 72 bytes"})
		return
	}
	if req.Role != "ADMIN" && req.Role != "SUPERVISOR" && req.Role != "WAREHOUSE" {
		c.JSON(400, gin.H{"error": "invalid role"})
		return
	}
	r := models.Employee{ID: uuid.New(), Code: req.Code, Name: req.Name, Username: req.Username, PasswordHash: auth.HashPassword(req.Password), Role: req.Role, Email: req.Email, Active: true}
	h.create(c, &r)
}

func (h Handler) ListInventory(c *gin.Context) {
	var r []models.Inventory
	h.DB.Preload("Product").Preload("Location").Order("updated_at desc").Find(&r)
	c.JSON(200, r)
}
func (h Handler) ListLotInventory(c *gin.Context) {
	var r []models.InventoryLot
	h.DB.Preload("Product").Preload("Location").Where("qty > 0").Order("updated_at desc").Find(&r)
	c.JSON(200, r)
}
func (h Handler) ListMovements(c *gin.Context) {
	var r []models.StockMovement
	query := h.DB.Preload("Product").Preload("Location")
	if typ := c.Query("type"); typ != "" {
		query = query.Where("type = ?", typ)
	}
	if err := query.Order("created_at desc").Limit(300).Find(&r).Error; err != nil {
		c.JSON(500, gin.H{"error": "unable to load movement history"})
		return
	}
	c.JSON(200, r)
}
func (h Handler) ListStockCounts(c *gin.Context) {
	var r []models.StockCount
	h.DB.Preload("Product").Preload("Location").Order("created_at desc").Limit(100).Find(&r)
	c.JSON(200, r)
}
func (h Handler) ListAdjustments(c *gin.Context) {
	var r []models.Adjustment
	h.DB.Preload("Product").Preload("Location").Order("created_at desc").Limit(200).Find(&r)
	c.JSON(200, r)
}

func (h Handler) Receive(c *gin.Context) { h.applyMovement(c, "IN", "receive completed") }
func (h Handler) Issue(c *gin.Context)   { h.applyMovement(c, "OUT", "issue completed") }
func (h Handler) applyMovement(c *gin.Context, typ, msg string) {
	var req services.MovementRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(400, gin.H{"error": err.Error()})
		return
	}
	req.CreatedBy = user(c)
	if err := services.ApplyMovement(h.DB, typ, req); err != nil {
		c.JSON(400, gin.H{"error": err.Error()})
		return
	}
	c.JSON(200, gin.H{"message": msg})
}
func (h Handler) Transfer(c *gin.Context) {
	var req services.TransferRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(400, gin.H{"error": err.Error()})
		return
	}
	req.CreatedBy = user(c)
	if err := services.TransferStock(h.DB, req); err != nil {
		c.JSON(400, gin.H{"error": err.Error()})
		return
	}
	c.JSON(200, gin.H{"message": "transfer completed"})
}
func (h Handler) StockCount(c *gin.Context) {
	var req services.StockCountRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(400, gin.H{"error": err.Error()})
		return
	}
	req.CreatedBy = user(c)
	row, err := services.ApplyStockCount(h.DB, req)
	if err != nil {
		c.JSON(400, gin.H{"error": err.Error()})
		return
	}
	c.JSON(200, gin.H{"message": "stock count saved; difference requires approval", "data": row})
}
func (h Handler) CreateAdjustment(c *gin.Context) {
	var req services.AdjustmentRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(400, gin.H{"error": err.Error()})
		return
	}
	req.RequestedBy = user(c)
	row, err := services.CreateAdjustment(h.DB, req)
	if err != nil {
		c.JSON(400, gin.H{"error": err.Error()})
		return
	}
	c.JSON(201, gin.H{"message": "adjustment submitted for approval", "data": row})
}
func (h Handler) ApproveAdjustment(c *gin.Context) {
	id, err := uuid.Parse(c.Param("id"))
	if err != nil {
		c.JSON(400, gin.H{"error": "invalid id"})
		return
	}
	row, err := services.ApproveAdjustment(h.DB, id, user(c))
	if err != nil {
		c.JSON(400, gin.H{"error": err.Error()})
		return
	}
	c.JSON(200, gin.H{"message": "adjustment approved", "data": row})
}
func (h Handler) RejectAdjustment(c *gin.Context) {
	id, err := uuid.Parse(c.Param("id"))
	if err != nil {
		c.JSON(400, gin.H{"error": "invalid id"})
		return
	}
	row, err := services.RejectAdjustment(h.DB, id, user(c))
	if err != nil {
		c.JSON(400, gin.H{"error": err.Error()})
		return
	}
	c.JSON(200, gin.H{"message": "adjustment rejected", "data": row})
}

func (h Handler) ExportProductsCSV(c *gin.Context) {
	var rows []models.Product
	h.DB.Order("sku").Find(&rows)
	c.Header("Content-Disposition", "attachment; filename=products.csv")
	c.Header("Content-Type", "text/csv; charset=utf-8")
	_, _ = c.Writer.Write([]byte("\xEF\xBB\xBF"))
	w := csv.NewWriter(c.Writer)
	_ = w.Write([]string{"sku", "name", "barcode", "unit", "category_code", "storage_type", "min_stock"})
	for _, x := range rows {
		_ = w.Write([]string{x.SKU, x.Name, x.Barcode, x.Unit, x.CategoryCode, x.StorageType, strconv.FormatFloat(x.MinStock, 'f', 2, 64)})
	}
	w.Flush()
}
func (h Handler) ImportProductsCSV(c *gin.Context) {
	f, _, err := c.Request.FormFile("file")
	if err != nil {
		c.JSON(400, gin.H{"error": "file is required"})
		return
	}
	defer f.Close()
	r := csv.NewReader(f)
	header, err := r.Read()
	if err != nil {
		c.JSON(400, gin.H{"error": "invalid csv"})
		return
	}
	for i := range header {
		header[i] = strings.TrimPrefix(strings.TrimSpace(header[i]), "\ufeff")
	}
	idx := map[string]int{}
	for i, v := range header {
		idx[v] = i
	}
	count := 0
	for {
		rec, err := r.Read()
		if err == io.EOF {
			break
		}
		if err != nil {
			c.JSON(400, gin.H{"error": err.Error()})
			return
		}
		get := func(k string) string {
			if n, ok := idx[k]; ok && n < len(rec) {
				return strings.TrimSpace(rec[n])
			}
			return ""
		}
		sku := get("sku")
		if sku == "" {
			continue
		}
		min, _ := strconv.ParseFloat(get("min_stock"), 64)
		var p models.Product
		err = h.DB.Where("sku=?", sku).First(&p).Error
		if err == gorm.ErrRecordNotFound {
			p = models.Product{ID: uuid.New(), SKU: sku}
		}
		p.Name = get("name")
		p.Barcode = get("barcode")
		p.Unit = get("unit")
		if p.Unit == "" {
			p.Unit = "PCS"
		}
		p.CategoryCode = get("category_code")
		p.StorageType = get("storage_type")
		p.MinStock = min
		if err == gorm.ErrRecordNotFound {
			err = h.DB.Create(&p).Error
		} else {
			err = h.DB.Save(&p).Error
		}
		if err != nil {
			c.JSON(400, gin.H{"error": err.Error()})
			return
		}
		count++
	}
	c.JSON(200, gin.H{"message": "import completed", "rows": count})
}

func (h Handler) ProductLabel(c *gin.Context) {
	sku := c.Param("sku")
	var p models.Product
	if err := h.DB.Where("sku=?", sku).First(&p).Error; err != nil {
		c.JSON(404, gin.H{"error": "product not found"})
		return
	}
	c.JSON(200, p)
}
func ParseDate(s string) *time.Time {
	if s == "" {
		return nil
	}
	t, err := time.Parse("2006-01-02", s)
	if err != nil {
		return nil
	}
	return &t
}
