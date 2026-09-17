package handlers

import (
	"easywms-demo-v3/internal/models"
	"github.com/gin-gonic/gin"
	"strconv"
	"time"
)

func (h Handler) ExportMovementsExcel(c *gin.Context) {
	var rows []models.StockMovement
	if err := h.DB.Preload("Product").Preload("Location").Order("created_at desc").Find(&rows).Error; err != nil {
		c.JSON(500, gin.H{"error": "Unable to export movements"})
		return
	}
	data, err := movementsWorkbook(rows)
	if err != nil {
		c.JSON(500, gin.H{"error": "Unable to create Excel file"})
		return
	}
	filename := "movements_" + time.Now().In(time.FixedZone("ICT", 7*60*60)).Format("20060102_150405") + ".xlsx"
	c.Header("Content-Disposition", `attachment; filename="`+filename+`"`)
	c.Data(200, "application/vnd.openxmlformats-officedocument.spreadsheetml.sheet", data)
}

func movementsWorkbook(rows []models.StockMovement) ([]byte, error) {
	values := make([][]string, 0, len(rows))
	for _, row := range rows {
		values = append(values, []string{row.CreatedAt.In(time.FixedZone("ICT", 7*60*60)).Format("2006-01-02 15:04:05"), row.Type, row.Product.SKU, row.Location.Code, strconv.FormatFloat(row.Qty, 'f', -1, 64), row.LotNo, row.ReasonCode, row.Reference, row.CreatedBy})
	}
	return excelWorkbook("Movements", []string{"Date (UTC+07:00)", "Type", "SKU", "Location", "Qty", "Lot", "Reason", "Reference", "User"}, values, 4)
}
