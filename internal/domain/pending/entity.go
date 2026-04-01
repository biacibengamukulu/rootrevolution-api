package pending

import (
	"time"

	"github.com/google/uuid"
)

type UpdateStatus string

const (
	StatusPending  UpdateStatus = "pending"
	StatusApproved UpdateStatus = "approved"
	StatusExpired  UpdateStatus = "expired"
	StatusRejected UpdateStatus = "rejected"
)

type PendingUpdate struct {
	Token       uuid.UUID    `json:"token"`
	ProductID   int          `json:"product_id"`
	Org         string       `json:"org"`
	UpdateData  string       `json:"update_data"` // JSON-encoded product fields
	RequestedBy string       `json:"requested_by"`
	CreatedAt   time.Time    `json:"created_at"`
	ExpiresAt   time.Time    `json:"expires_at"`
	Status      UpdateStatus `json:"status"`
	Action      string       `json:"action"` // "create", "update", "delete"
}

func NewPendingUpdate(productID int, org, updateData, requestedBy, action string) *PendingUpdate {
	now := time.Now()
	return &PendingUpdate{
		Token:       uuid.New(),
		ProductID:   productID,
		Org:         org,
		UpdateData:  updateData,
		RequestedBy: requestedBy,
		CreatedAt:   now,
		ExpiresAt:   now.Add(24 * time.Hour),
		Status:      StatusPending,
		Action:      action,
	}
}

func (p *PendingUpdate) IsExpired() bool {
	return time.Now().After(p.ExpiresAt)
}

func (p *PendingUpdate) IsValid() bool {
	return p.Status == StatusPending && !p.IsExpired()
}
