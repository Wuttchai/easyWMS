package models

import "time"

// Receipt and stock changes commit together; retrying a key cannot repeat stock changes.
type MutationReceipt struct {
	ID          string `gorm:"primaryKey;size:64"`
	Fingerprint string `gorm:"not null;size:64"`
	Status      int
	Body        string `gorm:"type:text"`
	CreatedAt   time.Time
}
