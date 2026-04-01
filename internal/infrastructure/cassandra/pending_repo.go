package cassandra

import (
	"fmt"

	"github.com/gocql/gocql"
	"github.com/google/uuid"

	"rootrevolution-api/internal/domain/pending"
)

type pendingRepository struct {
	session *gocql.Session
}

func NewPendingRepository(session *gocql.Session) pending.Repository {
	return &pendingRepository{session: session}
}

func (r *pendingRepository) Save(p *pending.PendingUpdate) error {
	return r.session.Query(`INSERT INTO pending_updates
		(auth_token, product_id, org, update_data, requested_by, created_at, expires_at, status, action)
		VALUES (?,?,?,?,?,?,?,?,?)`,
		gocql.UUID(p.Token), p.ProductID, p.Org, p.UpdateData, p.RequestedBy,
		p.CreatedAt, p.ExpiresAt, string(p.Status), p.Action,
	).Exec()
}

func (r *pendingRepository) FindByToken(token uuid.UUID) (*pending.PendingUpdate, error) {
	var p pending.PendingUpdate
	var statusStr string
	var gid gocql.UUID

	err := r.session.Query(`SELECT auth_token, product_id, org, update_data, requested_by,
		created_at, expires_at, status, action
		FROM pending_updates WHERE auth_token = ?`, gocql.UUID(token)).Scan(
		&gid, &p.ProductID, &p.Org, &p.UpdateData, &p.RequestedBy,
		&p.CreatedAt, &p.ExpiresAt, &statusStr, &p.Action,
	)

	if err == gocql.ErrNotFound {
		return nil, nil
	}
	if err != nil {
		return nil, fmt.Errorf("finding pending update: %w", err)
	}

	p.Token = uuid.UUID(gid)
	p.Status = pending.UpdateStatus(statusStr)
	return &p, nil
}

func (r *pendingRepository) UpdateStatus(token uuid.UUID, status pending.UpdateStatus) error {
	return r.session.Query(`UPDATE pending_updates SET status = ? WHERE auth_token = ?`,
		string(status), gocql.UUID(token)).Exec()
}

func (r *pendingRepository) CleanExpired() error {
	// Cassandra TTL handles this; alternatively select and delete expired ones
	return nil
}
