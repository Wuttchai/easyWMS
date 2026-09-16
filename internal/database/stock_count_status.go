package database

import "gorm.io/gorm"

// SyncStockCountStatuses repairs legacy status links without changing inventory.
func SyncStockCountStatuses(db *gorm.DB) error {
	return db.Exec(`UPDATE stock_counts AS sc
		SET status = CASE a.status WHEN 'APPROVED' THEN 'COMPLETED' ELSE 'REJECTED' END
		FROM adjustments AS a
		WHERE sc.adjustment_id = a.id AND sc.status = 'PENDING_APPROVAL'
		AND a.status IN ('APPROVED', 'REJECTED')`).Error
}
