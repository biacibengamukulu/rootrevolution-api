package pending

import "github.com/google/uuid"

type Repository interface {
	Save(p *PendingUpdate) error
	FindByToken(token uuid.UUID) (*PendingUpdate, error)
	FindAllPending(org string) ([]*PendingUpdate, error)
	UpdateStatus(token uuid.UUID, status UpdateStatus) error
	CleanExpired() error
}
