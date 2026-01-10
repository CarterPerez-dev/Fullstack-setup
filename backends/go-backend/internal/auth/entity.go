// AngelaMos | 2026
// entity.go

package auth

import (
	"time"
)

type RefreshToken struct {
	ID           string     `db:"id"`
	UserID       string     `db:"user_id"`
	TokenHash    string     `db:"token_hash"`
	FamilyID     string     `db:"family_id"`
	ExpiresAt    time.Time  `db:"expires_at"`
	CreatedAt    time.Time  `db:"created_at"`
	IsUsed       bool       `db:"is_used"`
	UsedAt       *time.Time `db:"used_at"`
	RevokedAt    *time.Time `db:"revoked_at"`
	ReplacedByID *string    `db:"replaced_by_id"`
	UserAgent    string     `db:"user_agent"`
	IPAddress    string     `db:"ip_address"`
}

func (t *RefreshToken) IsExpired() bool {
	return time.Now().After(t.ExpiresAt)
}

func (t *RefreshToken) IsRevoked() bool {
	return t.RevokedAt != nil
}

func (t *RefreshToken) IsValid() bool {
	return !t.IsExpired() && !t.IsRevoked() && !t.IsUsed
}

func (t *RefreshToken) MarkAsUsed(replacedByID string) {
	now := time.Now()
	t.IsUsed = true
	t.UsedAt = &now
	t.ReplacedByID = &replacedByID
}

func (t *RefreshToken) Revoke() {
	now := time.Now()
	t.RevokedAt = &now
}
