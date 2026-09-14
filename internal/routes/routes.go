package routes

import (
	"easywms-demo-v3/internal/handlers"
	"net/http"

	"github.com/gin-gonic/gin"
)

func Register(r *gin.Engine, h handlers.Handler) {
	r.POST("/api/login", h.Login)

	api := r.Group("/api", h.Auth())
	{
		api.GET("/dashboard", h.Dashboard)
		api.GET("/products", h.ListProducts)
		api.POST("/products", handlers.RequireRoles("ADMIN", "SUPERVISOR"), h.CreateProduct)
		api.GET("/products/export", h.ExportProductsCSV)
		api.POST("/products/import", handlers.RequireRoles("ADMIN", "SUPERVISOR"), h.ImportProductsCSV)
		api.GET("/products/:sku/label", h.ProductLabel)
		api.GET("/warehouses", h.ListWarehouses)
		api.POST("/warehouses", handlers.RequireRoles("ADMIN"), h.CreateWarehouse)
		api.GET("/zones", h.ListZones)
		api.POST("/zones", handlers.RequireRoles("ADMIN"), h.CreateZone)
		api.GET("/locations", h.ListLocations)
		api.POST("/locations", handlers.RequireRoles("ADMIN"), h.CreateLocation)
		api.DELETE("/:resource/:id", handlers.RequireRoles("ADMIN"), h.DeleteMaster)
		api.GET("/units", h.ListUnits)
		api.POST("/units", handlers.RequireRoles("ADMIN"), h.CreateUnit)
		api.GET("/suppliers", h.ListSuppliers)
		api.POST("/suppliers", handlers.RequireRoles("ADMIN", "SUPERVISOR"), h.CreateSupplier)
		api.GET("/customers", h.ListCustomers)
		api.POST("/customers", handlers.RequireRoles("ADMIN", "SUPERVISOR"), h.CreateCustomer)
		api.GET("/employees", handlers.RequireRoles("ADMIN"), h.ListEmployees)
		api.POST("/employees", handlers.RequireRoles("ADMIN"), h.CreateEmployee)
		api.GET("/categories", h.ListCategories)
		api.POST("/categories", handlers.RequireRoles("ADMIN"), h.CreateCategory)
		api.GET("/storage-types", h.ListStorageTypes)
		api.POST("/storage-types", handlers.RequireRoles("ADMIN"), h.CreateStorageType)
		api.GET("/reason-codes", h.ListReasonCodes)
		api.POST("/reason-codes", handlers.RequireRoles("ADMIN"), h.CreateReasonCode)
		api.GET("/inventory", h.ListInventory)
		api.GET("/inventory-lots", h.ListLotInventory)
		api.GET("/movements", h.ListMovements)
		api.GET("/stock-counts", h.ListStockCounts)
		api.GET("/adjustments", h.ListAdjustments)
		api.POST("/receive", h.Receive)
		api.POST("/issue", h.Issue)
		api.POST("/transfer", h.Transfer)
		api.POST("/stock-count", h.StockCount)
		api.POST("/adjustments", h.CreateAdjustment)
		api.POST("/adjustments/:id/approve", handlers.RequireRoles("ADMIN", "SUPERVISOR"), h.ApproveAdjustment)
		api.POST("/adjustments/:id/reject", handlers.RequireRoles("ADMIN", "SUPERVISOR"), h.RejectAdjustment)
	}

	r.StaticFile("/", "./web/index.html")
	r.StaticFile("/index.html", "./web/index.html")
	r.StaticFile("/app.js", "./web/app.js")
	r.StaticFile("/styles.css", "./web/styles.css")
	r.StaticFile("/js/core.js", "./web/js/core.js")
	r.StaticFile("/js/inventory.js", "./web/js/inventory.js")
	r.StaticFile("/js/master.js", "./web/js/master.js")
	r.StaticFile("/js/transactions.js", "./web/js/transactions.js")
	r.StaticFile("/js/scanner.js", "./web/js/scanner.js")
	r.NoRoute(func(c *gin.Context) {
		if c.Request.Method == http.MethodGet {
			c.File("./web/index.html")
			return
		}
		c.JSON(404, gin.H{"error": "not found"})
	})
}
