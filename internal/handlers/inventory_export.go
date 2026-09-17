package handlers

import (
	"archive/zip"
	"bytes"
	"easywms-demo-v3/internal/models"
	"encoding/xml"
	"fmt"
	"github.com/gin-gonic/gin"
	"strconv"
	"time"
)

func (h Handler) ExportInventoryExcel(c *gin.Context) {
	var rows []models.Inventory
	if err := h.DB.Preload("Product").Preload("Location").Order("updated_at desc").Find(&rows).Error; err != nil {
		c.JSON(500, gin.H{"error": "Unable to export inventory"})
		return
	}
	data, err := inventoryWorkbook(rows)
	if err != nil {
		c.JSON(500, gin.H{"error": "Unable to create Excel file"})
		return
	}
	filename := "inventory_" + time.Now().In(time.FixedZone("ICT", 7*60*60)).Format("20060102_150405") + ".xlsx"
	c.Header("Content-Disposition", `attachment; filename="`+filename+`"`)
	c.Data(200, "application/vnd.openxmlformats-officedocument.spreadsheetml.sheet", data)
}

func inventoryWorkbook(rows []models.Inventory) ([]byte, error) {
	values := make([][]string, 0, len(rows))
	for _, row := range rows {
		status := "OK"
		if row.Qty <= row.Product.MinStock {
			status = "LOW"
		}
		values = append(values, []string{row.Product.SKU, row.Product.Name, row.Location.WarehouseCode, row.Location.ZoneCode, row.Location.Code, strconv.FormatFloat(row.Qty, 'f', -1, 64), row.Product.Unit, status})
	}
	return excelWorkbook("Inventory", []string{"SKU", "Product", "Warehouse", "Zone", "Location", "Qty", "Unit", "Status"}, values, 5)
}

func excelWorkbook(name string, headers []string, rows [][]string, numeric int) ([]byte, error) {
	var sheet bytes.Buffer
	fmt.Fprintf(&sheet, `<worksheet xmlns="http://schemas.openxmlformats.org/spreadsheetml/2006/main"><sheetViews><sheetView workbookViewId="0"><pane ySplit="1" topLeftCell="A2" activePane="bottomLeft" state="frozen"/></sheetView></sheetViews><cols><col min="1" max="%d" width="22" customWidth="1"/></cols><sheetData>`, len(headers))
	writeRow := func(number int, values []string, numeric int) {
		fmt.Fprintf(&sheet, `<row r="%d">`, number)
		for i, value := range values {
			ref := fmt.Sprintf("%c%d", 'A'+i, number)
			if i == numeric {
				fmt.Fprintf(&sheet, `<c r="%s"><v>%s</v></c>`, ref, value)
			} else {
				fmt.Fprintf(&sheet, `<c r="%s" t="inlineStr"><is><t xml:space="preserve">`, ref)
				_ = xml.EscapeText(&sheet, []byte(value))
				sheet.WriteString(`</t></is></c>`)
			}
		}
		sheet.WriteString(`</row>`)
	}
	writeRow(1, headers, -1)
	for i, row := range rows {
		writeRow(i+2, row, numeric)
	}
	fmt.Fprintf(&sheet, `</sheetData><autoFilter ref="A1:%c%d"/></worksheet>`, 'A'+len(headers)-1, len(rows)+1)
	parts := []struct{ name, content string }{
		{"[Content_Types].xml", `<Types xmlns="http://schemas.openxmlformats.org/package/2006/content-types"><Default Extension="rels" ContentType="application/vnd.openxmlformats-package.relationships+xml"/><Default Extension="xml" ContentType="application/xml"/><Override PartName="/xl/workbook.xml" ContentType="application/vnd.openxmlformats-officedocument.spreadsheetml.sheet.main+xml"/><Override PartName="/xl/worksheets/sheet1.xml" ContentType="application/vnd.openxmlformats-officedocument.spreadsheetml.worksheet+xml"/></Types>`},
		{"_rels/.rels", `<Relationships xmlns="http://schemas.openxmlformats.org/package/2006/relationships"><Relationship Id="rId1" Type="http://schemas.openxmlformats.org/officeDocument/2006/relationships/officeDocument" Target="xl/workbook.xml"/></Relationships>`},
		{"xl/workbook.xml", `<workbook xmlns="http://schemas.openxmlformats.org/spreadsheetml/2006/main" xmlns:r="http://schemas.openxmlformats.org/officeDocument/2006/relationships"><sheets><sheet name="` + name + `" sheetId="1" r:id="rId1"/></sheets></workbook>`},
		{"xl/_rels/workbook.xml.rels", `<Relationships xmlns="http://schemas.openxmlformats.org/package/2006/relationships"><Relationship Id="rId1" Type="http://schemas.openxmlformats.org/officeDocument/2006/relationships/worksheet" Target="worksheets/sheet1.xml"/></Relationships>`},
		{"xl/worksheets/sheet1.xml", sheet.String()},
	}
	var output bytes.Buffer
	archive := zip.NewWriter(&output)
	for _, part := range parts {
		writer, err := archive.Create(part.name)
		if err != nil {
			return nil, err
		}
		if _, err = writer.Write([]byte(part.content)); err != nil {
			return nil, err
		}
	}
	if err := archive.Close(); err != nil {
		return nil, err
	}
	return output.Bytes(), nil
}
