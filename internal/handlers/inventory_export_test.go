package handlers

import (
	"archive/zip"
	"bytes"
	"easywms-demo-v3/internal/models"
	"encoding/xml"
	"io"
	"strings"
	"testing"
	"time"
)

func TestInventoryWorkbook(t *testing.T) {
	for _, rows := range [][]models.Inventory{nil, {{Product: models.Product{SKU: "00123", Name: "สินค้า & <test>", Unit: "ชิ้น", MinStock: 2}, Qty: 1.5}}} {
		data, err := inventoryWorkbook(rows)
		if err != nil {
			t.Fatal(err)
		}
		archive, err := zip.NewReader(bytes.NewReader(data), int64(len(data)))
		if err != nil {
			t.Fatal(err)
		}
		if len(archive.File) != 5 {
			t.Fatalf("unexpected parts: %d", len(archive.File))
		}
		for _, file := range archive.File {
			reader, err := file.Open()
			if err != nil {
				t.Fatal(err)
			}
			content, err := io.ReadAll(reader)
			reader.Close()
			if err != nil {
				t.Fatal(err)
			}
			decoder := xml.NewDecoder(bytes.NewReader(content))
			for {
				_, err := decoder.Token()
				if err == io.EOF {
					break
				}
				if err != nil {
					t.Fatalf("%s: %v", file.Name, err)
				}
			}
			if file.Name == "xl/worksheets/sheet1.xml" && len(rows) > 0 {
				for _, expected := range []string{`<c r="F2"><v>1.5</v></c>`, `00123`, `สินค้า &amp; &lt;test&gt;`, `LOW`, `A1:H2`} {
					if !strings.Contains(string(content), expected) {
						t.Errorf("missing %s", expected)
					}
				}
				if strings.Contains(string(content), "<f>") {
					t.Error("unexpected formula")
				}
			}
		}
	}
}

func TestMovementsWorkbook(t *testing.T) {
	data, err := movementsWorkbook([]models.StockMovement{{CreatedAt: time.Date(2026, 9, 17, 15, 8, 3, 0, time.UTC), Type: "IN", Qty: 2.5, Reference: "=1+1", Product: models.Product{SKU: "00123"}}})
	if err != nil {
		t.Fatal(err)
	}
	archive, err := zip.NewReader(bytes.NewReader(data), int64(len(data)))
	if err != nil {
		t.Fatal(err)
	}
	for _, file := range archive.File {
		if file.Name != "xl/worksheets/sheet1.xml" {
			continue
		}
		reader, err := file.Open()
		if err != nil {
			t.Fatal(err)
		}
		content, err := io.ReadAll(reader)
		reader.Close()
		if err != nil {
			t.Fatal(err)
		}
		for _, expected := range []string{`2026-09-17 22:08:03`, `<c r="E2"><v>2.5</v></c>`, `A1:I2`, `<c r="H2" t="inlineStr"><is><t xml:space="preserve">=1+1</t></is></c>`} {
			if !strings.Contains(string(content), expected) {
				t.Errorf("missing %s", expected)
			}
		}
		return
	}
	t.Fatal("missing worksheet")
}
