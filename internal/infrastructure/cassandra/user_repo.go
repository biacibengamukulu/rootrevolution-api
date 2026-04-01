package cassandra

import (
	"fmt"

	"github.com/gocql/gocql"
	"github.com/google/uuid"

	"rootrevolution-api/internal/domain/user"
)

type userRepository struct {
	session *gocql.Session
}

func NewUserRepository(session *gocql.Session) user.Repository {
	return &userRepository{session: session}
}

// toGocqlUUID converts google/uuid.UUID → gocql.UUID (both are [16]byte)
func toGocqlUUID(id uuid.UUID) gocql.UUID {
	return gocql.UUID(id)
}

// fromGocqlUUID converts gocql.UUID → google/uuid.UUID
func fromGocqlUUID(id gocql.UUID) uuid.UUID {
	return uuid.UUID(id)
}

func (r *userRepository) FindByEmail(email string) (*user.User, error) {
	var u user.User
	var roleStr string
	var gid gocql.UUID

	err := r.session.Query(`SELECT id, name, surname, email, password, role, status, created_at
		FROM users WHERE email = ?`, email).Scan(
		&gid, &u.Name, &u.Surname, &u.Email, &u.Password, &roleStr, &u.Status, &u.CreatedAt,
	)

	if err == gocql.ErrNotFound {
		return nil, nil
	}
	if err != nil {
		return nil, fmt.Errorf("finding user by email: %w", err)
	}

	u.ID = fromGocqlUUID(gid)
	u.Role = user.Role(roleStr)
	return &u, nil
}

func (r *userRepository) FindByID(id uuid.UUID) (*user.User, error) {
	var email string
	err := r.session.Query(`SELECT email FROM users_by_id WHERE id = ?`,
		toGocqlUUID(id)).Scan(&email)
	if err == gocql.ErrNotFound {
		return nil, nil
	}
	if err != nil {
		return nil, fmt.Errorf("finding user id mapping: %w", err)
	}

	return r.FindByEmail(email)
}

func (r *userRepository) FindAll() ([]user.User, error) {
	iter := r.session.Query(`SELECT id, name, surname, email, role, status, created_at FROM users`).Iter()

	var users []user.User
	var roleStr string
	var gid gocql.UUID
	var u user.User

	for iter.Scan(&gid, &u.Name, &u.Surname, &u.Email, &roleStr, &u.Status, &u.CreatedAt) {
		u.ID = fromGocqlUUID(gid)
		u.Role = user.Role(roleStr)
		users = append(users, u)
		u = user.User{}
		roleStr = ""
	}

	if err := iter.Close(); err != nil {
		return nil, fmt.Errorf("querying users: %w", err)
	}

	return users, nil
}

func (r *userRepository) Save(u *user.User) error {
	gid := toGocqlUUID(u.ID)

	if err := r.session.Query(`INSERT INTO users
		(id, name, surname, email, password, role, status, created_at)
		VALUES (?,?,?,?,?,?,?,?)`,
		gid, u.Name, u.Surname, u.Email, u.Password, string(u.Role), u.Status, u.CreatedAt,
	).Exec(); err != nil {
		return err
	}

	return r.session.Query(`INSERT INTO users_by_id (id, email) VALUES (?,?)`,
		gid, u.Email).Exec()
}

func (r *userRepository) Update(u *user.User) error {
	return r.session.Query(`UPDATE users SET
		name = ?, surname = ?, role = ?, status = ?
		WHERE email = ?`,
		u.Name, u.Surname, string(u.Role), u.Status, u.Email,
	).Exec()
}

func (r *userRepository) Delete(id uuid.UUID) error {
	u, err := r.FindByID(id)
	if err != nil || u == nil {
		return err
	}

	if err := r.session.Query(`DELETE FROM users WHERE email = ?`, u.Email).Exec(); err != nil {
		return err
	}

	return r.session.Query(`DELETE FROM users_by_id WHERE id = ?`, toGocqlUUID(id)).Exec()
}

func (r *userRepository) Exists(email string) (bool, error) {
	var count int
	err := r.session.Query(`SELECT COUNT(*) FROM users WHERE email = ?`, email).Scan(&count)
	if err != nil {
		return false, err
	}
	return count > 0, nil
}
