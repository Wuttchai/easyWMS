# EasyWMS Demo V3

Mini WMS สำหรับ Demo ลูกค้า SME / โรงงาน พัฒนาด้วย **Go + Gin + GORM + PostgreSQL + Vanilla JS** ไม่ต้องติดตั้ง Node/NPM สำหรับการรันระบบ

## V3 Highlights

- Login + signed token (JWT-style HMAC token) และ Role
  - `ADMIN`
  - `SUPERVISOR`
  - `WAREHOUSE`
- Master Data
  - Product + Barcode
  - Warehouse / Zone / Location
  - Unit / Category / Storage Type / Reason Code
  - Supplier / Customer
  - User / Employee
- Receive / Issue / Stock Transfer
- Inventory by Location
- **Lot / Batch / MFG Date / Expiry Date** ตอน Receive/Issue
- Lot Inventory
- Stock Count
- **Adjustment Approval Workflow**
  - Stock Count พบส่วนต่าง -> สร้าง Adjustment `PENDING`
  - Stock ยังไม่เปลี่ยน
  - ADMIN/SUPERVISOR Approve -> Stock เปลี่ยน + Movement ถูกสร้าง
  - Reject -> Stock ไม่เปลี่ยน
- Manual Adjustment Request
- Stock Movement / Audit Trail + Created By
- Product CSV Import / Export
- Product Label Print
- Mobile Camera Barcode / QR demo ผ่าน Browser `BarcodeDetector` (ถ้า Browser รองรับ)
- Demo Seed Data

## Demo Accounts

| Username | Password | Role |
|---|---|---|
| `admin` | `admin123` | ADMIN |
| `supervisor` | `supervisor123` | SUPERVISOR |
| `operator` | `operator123` | WAREHOUSE |

> Password/token implementation นี้ทำไว้สำหรับ Demo เท่านั้น ก่อนใช้ Production ควรเปลี่ยนเป็น bcrypt/argon2, refresh token, secure cookie หรือ auth provider ที่เหมาะสม

## Run on Windows / VS Code

### 1) Start PostgreSQL

```bash
docker compose up -d
```

### 2) Download Go packages

```bash
go mod tidy
```

### 3) Run

```bash
go run ./cmd/server
```

เปิด:

```text
http://localhost:8080
```

Default DB:

```text
DB_HOST=localhost
DB_PORT=5432
DB_NAME=easywms
DB_USER=easywms
DB_PASSWORD=easywms123
DB_SSLMODE=disable
JWT_SECRET=easywms-demo-secret-change-me
```

### Reset Demo Database

ถ้าเคยรัน V1/V2 ด้วย volume เดิม และต้องการ Demo Data ใหม่ทั้งหมด:

```bash
docker compose down -v
docker compose up -d
go run ./cmd/server
```

> `docker compose down -v` จะลบข้อมูล Demo PostgreSQL เดิมทั้งหมด

## Demo Flow แนะนำ

### Flow 1 — Receive + Lot

Login `operator` แล้ว Receive:

```text
SKU       RM-STEEL-001
Location  A-01-01
Qty       10
Lot       LOT-260914
Reference GR-DEMO-001
Reason    RCV
```

ใส่ MFG/EXP ได้ แล้วดูผลที่ `Inventory`, `Lot / Expiry`, `Movements`

### Flow 2 — Stock Count + Approval

1. Login `operator`
2. ไป `Stock Count`
3. นับยอดจริงให้ต่างจาก System Qty
4. ระบบสร้าง Adjustment = `PENDING`
5. Inventory **ยังไม่เปลี่ยน**
6. Logout แล้ว Login `supervisor`
7. ไป `Adjustment Approval`
8. กด Approve
9. Inventory เปลี่ยน และเกิด `ADJUST_IN` หรือ `ADJUST_OUT` ใน Movement

จุดนี้เหมาะสำหรับ Demo เรื่อง Audit / Internal Control ให้โรงงาน

### Flow 3 — Manual Adjustment

Operator สามารถสร้าง Adjustment Request เช่น:

```text
Direction   OUT
Qty         2
Reason      DMG
Note        Damaged during handling
```

จากนั้น Supervisor/Admin เป็นผู้ Approve/Reject

### Flow 4 — Product CSV

ใน `Master Data > Product`:

- Export CSV
- แก้ไฟล์ CSV
- Import กลับเข้าระบบ

Header ที่รองรับ:

```text
sku,name,barcode,unit,category_code,storage_type,min_stock
```

### Flow 5 — Label

ใน `Product Master` กด `Print` เพื่อเปิด Product Label สำหรับพิมพ์

## Main APIs

Public:

```text
POST /api/login
```

Authenticated:

```text
GET  /api/dashboard
GET  /api/inventory
GET  /api/inventory-lots
GET  /api/movements
GET  /api/stock-counts
GET  /api/adjustments
POST /api/receive
POST /api/issue
POST /api/transfer
POST /api/stock-count
POST /api/adjustments
POST /api/adjustments/:id/approve
POST /api/adjustments/:id/reject
```

Product:

```text
GET  /api/products
POST /api/products
GET  /api/products/export
POST /api/products/import
GET  /api/products/:sku/label
```

## Important Demo Notes

- Camera scanner ต้องได้รับสิทธิ์กล้องจาก Browser
- Browser บางรุ่นไม่มี `BarcodeDetector`; สามารถใช้ USB/Bluetooth scanner ที่ยิงค่าเข้า input ได้เหมือน Keyboard
- Product CSV ใช้ CSV เพื่อให้ Demo V3 ไม่ต้องเพิ่ม library Excel ภายนอกอีกตัว ถ้าจะใช้จริงสามารถเปลี่ยนเป็น `.xlsx` ด้วย `excelize`
- V3 เป็น sales/demo prototype ยังไม่ควรเปิด Public Internet โดยไม่เพิ่ม Production security, validation, backup, observability และ HTTPS
