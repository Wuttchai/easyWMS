package handlers

import (
	"bytes"
	"crypto/sha256"
	"easywms-demo-v3/internal/models"
	"encoding/binary"
	"encoding/hex"
	"errors"
	"github.com/gin-gonic/gin"
	"github.com/google/uuid"
	"gorm.io/gorm"
	"io"
	"net/http/httptest"
)

// Stock and its saved response are committed in the same database transaction.
func (h Handler) Mutation(action string) gin.HandlerFunc {
	return func(c *gin.Context) {
		key := c.GetHeader("Idempotency-Key")
		if _, err := uuid.Parse(key); err != nil {
			c.JSON(400, gin.H{"error": "valid Idempotency-Key UUID is required"})
			return
		}
		body, err := io.ReadAll(io.LimitReader(c.Request.Body, 1024*1024+1))
		if err != nil || len(body) > 1024*1024 {
			c.JSON(400, gin.H{"error": "invalid request body"})
			return
		}
		identity := sha256.Sum256([]byte(c.GetString("user_id") + ":" + key))
		fingerprint := sha256.Sum256(append([]byte(c.Request.Method+":"+c.Request.URL.RequestURI()+":"), body...))
		receipt := models.MutationReceipt{ID: hex.EncodeToString(identity[:]), Fingerprint: hex.EncodeToString(fingerprint[:])}
		rejected := errors.New("request rejected")
		err = h.DB.Transaction(func(tx *gorm.DB) error {
			if err := tx.Exec("SELECT pg_advisory_xact_lock(?)", int64(binary.BigEndian.Uint64(identity[:8]))).Error; err != nil {
				return err
			}
			var existing models.MutationReceipt
			err := tx.First(&existing, "id=?", receipt.ID).Error
			if err == nil {
				if existing.Fingerprint != receipt.Fingerprint {
					receipt.Status = 409
					receipt.Body = `{"error":"idempotency key already used for different input"}`
					return rejected
				}
				receipt = existing
				return nil
			}
			if !errors.Is(err, gorm.ErrRecordNotFound) {
				return err
			}
			recorder := httptest.NewRecorder()
			inner, _ := gin.CreateTestContext(recorder)
			inner.Request = c.Request.Clone(c.Request.Context())
			inner.Request.Body = io.NopCloser(bytes.NewReader(body))
			inner.Params = c.Params
			for k, v := range c.Keys {
				inner.Set(k, v)
			}
			handler := h
			handler.DB = tx
			switch action {
			case "receive":
				handler.Receive(inner)
			case "issue":
				handler.Issue(inner)
			case "transfer":
				handler.Transfer(inner)
			case "count":
				handler.StockCount(inner)
			case "adjustment":
				handler.CreateAdjustment(inner)
			case "approve":
				handler.ApproveAdjustment(inner)
			case "reject":
				handler.RejectAdjustment(inner)
			default:
				return errors.New("unknown mutation")
			}
			receipt.Status = recorder.Code
			receipt.Body = recorder.Body.String()
			if receipt.Status >= 400 {
				return rejected
			}
			return tx.Create(&receipt).Error
		})
		if err != nil && !errors.Is(err, rejected) {
			c.JSON(500, gin.H{"error": "unable to commit request; retry with the same idempotency key"})
			return
		}
		c.Data(receipt.Status, "application/json; charset=utf-8", []byte(receipt.Body))
	}
}
