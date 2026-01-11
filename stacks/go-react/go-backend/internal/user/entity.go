// AngelaMos | 2026
// entity.go

package user

import (
	"time"
)

type User struct {
	ID           string     `db:"id"`
	Email        string     `db:"email"`
	PasswordHash string     `db:"password_hash"`
	Name         string     `db:"name"`
	Role         string     `db:"role"`
	Tier         string     `db:"tier"`
	TokenVersion int        `db:"token_version"`
	CreatedAt    time.Time  `db:"created_at"`
	UpdatedAt    time.Time  `db:"updated_at"`
	DeletedAt    *time.Time `db:"deleted_at"`
}

func (u *User) IsDeleted() bool {
	return u.DeletedAt != nil
}

func (u *User) IsAdmin() bool {
	return u.Role == RoleAdmin
}

const (
	RoleUser  = "user"
	RoleAdmin = "admin"
)

const (
	TierFree       = "free"
	TierPro        = "pro"
	TierEnterprise = "enterprise"
)
