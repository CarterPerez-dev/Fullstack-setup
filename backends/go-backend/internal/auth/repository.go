// AngelaMos | 2026
// repository.go

package auth

import (
	"context"
	"database/sql"
	"errors"
	"fmt"
	"time"

	"github.com/carterperez-dev/templates/go-backend/internal/core"
)

type Repository interface {
	Create(ctx context.Context, token *RefreshToken) error
	FindByHash(ctx context.Context, tokenHash string) (*RefreshToken, error)
	FindByID(ctx context.Context, id string) (*RefreshToken, error)
	MarkAsUsed(ctx context.Context, id, replacedByID string) error
	RevokeByID(ctx context.Context, id string) error
	RevokeByFamilyID(ctx context.Context, familyID string) error
	RevokeAllForUser(ctx context.Context, userID string) error
	GetActiveSessionsForUser(
		ctx context.Context,
		userID string,
	) ([]RefreshToken, error)
	DeleteExpired(ctx context.Context) (int64, error)
}

type repository struct {
	db core.DBTX
}

func NewRepository(db core.DBTX) Repository {
	return &repository{db: db}
}

func (r *repository) Create(ctx context.Context, token *RefreshToken) error {
	query := `
		INSERT INTO refresh_tokens (
			id, user_id, token_hash, family_id, expires_at,
			user_agent, ip_address
		) VALUES (
			$1, $2, $3, $4, $5, $6, $7
		)
		RETURNING created_at`

	err := r.db.GetContext(ctx, &token.CreatedAt, query,
		token.ID,
		token.UserID,
		token.TokenHash,
		token.FamilyID,
		token.ExpiresAt,
		token.UserAgent,
		token.IPAddress,
	)
	if err != nil {
		return fmt.Errorf("create refresh token: %w", err)
	}

	return nil
}

func (r *repository) FindByHash(
	ctx context.Context,
	tokenHash string,
) (*RefreshToken, error) {
	query := `
		SELECT
			id, user_id, token_hash, family_id, expires_at, created_at,
			is_used, used_at, revoked_at, replaced_by_id, user_agent, ip_address
		FROM refresh_tokens
		WHERE token_hash = $1`

	var token RefreshToken
	err := r.db.GetContext(ctx, &token, query, tokenHash)
	if errors.Is(err, sql.ErrNoRows) {
		return nil, fmt.Errorf("find refresh token: %w", core.ErrNotFound)
	}
	if err != nil {
		return nil, fmt.Errorf("find refresh token: %w", err)
	}

	return &token, nil
}

func (r *repository) FindByID(
	ctx context.Context,
	id string,
) (*RefreshToken, error) {
	query := `
		SELECT
			id, user_id, token_hash, family_id, expires_at, created_at,
			is_used, used_at, revoked_at, replaced_by_id, user_agent, ip_address
		FROM refresh_tokens
		WHERE id = $1`

	var token RefreshToken
	err := r.db.GetContext(ctx, &token, query, id)
	if errors.Is(err, sql.ErrNoRows) {
		return nil, fmt.Errorf("find refresh token: %w", core.ErrNotFound)
	}
	if err != nil {
		return nil, fmt.Errorf("find refresh token: %w", err)
	}

	return &token, nil
}

func (r *repository) MarkAsUsed(
	ctx context.Context,
	id, replacedByID string,
) error {
	query := `
		UPDATE refresh_tokens
		SET is_used = true, used_at = NOW(), replaced_by_id = $2
		WHERE id = $1 AND is_used = false`

	result, err := r.db.ExecContext(ctx, query, id, replacedByID)
	if err != nil {
		return fmt.Errorf("mark refresh token as used: %w", err)
	}

	rows, err := result.RowsAffected()
	if err != nil {
		return fmt.Errorf("mark refresh token as used: %w", err)
	}

	if rows == 0 {
		return fmt.Errorf("mark refresh token as used: %w", core.ErrNotFound)
	}

	return nil
}

func (r *repository) RevokeByID(ctx context.Context, id string) error {
	query := `
		UPDATE refresh_tokens
		SET revoked_at = NOW()
		WHERE id = $1 AND revoked_at IS NULL`

	result, err := r.db.ExecContext(ctx, query, id)
	if err != nil {
		return fmt.Errorf("revoke refresh token: %w", err)
	}

	rows, err := result.RowsAffected()
	if err != nil {
		return fmt.Errorf("revoke refresh token: %w", err)
	}

	if rows == 0 {
		return fmt.Errorf("revoke refresh token: %w", core.ErrNotFound)
	}

	return nil
}

func (r *repository) RevokeByFamilyID(
	ctx context.Context,
	familyID string,
) error {
	query := `
		UPDATE refresh_tokens
		SET revoked_at = NOW()
		WHERE family_id = $1 AND revoked_at IS NULL`

	_, err := r.db.ExecContext(ctx, query, familyID)
	if err != nil {
		return fmt.Errorf("revoke token family: %w", err)
	}

	return nil
}

func (r *repository) RevokeAllForUser(
	ctx context.Context,
	userID string,
) error {
	query := `
		UPDATE refresh_tokens
		SET revoked_at = NOW()
		WHERE user_id = $1 AND revoked_at IS NULL`

	_, err := r.db.ExecContext(ctx, query, userID)
	if err != nil {
		return fmt.Errorf("revoke all user tokens: %w", err)
	}

	return nil
}

func (r *repository) GetActiveSessionsForUser(
	ctx context.Context,
	userID string,
) ([]RefreshToken, error) {
	query := `
		SELECT
			id, user_id, token_hash, family_id, expires_at, created_at,
			is_used, used_at, revoked_at, replaced_by_id, user_agent, ip_address
		FROM refresh_tokens
		WHERE user_id = $1
			AND revoked_at IS NULL
			AND is_used = false
			AND expires_at > NOW()
		ORDER BY created_at DESC`

	var tokens []RefreshToken
	err := r.db.SelectContext(ctx, &tokens, query, userID)
	if err != nil {
		return nil, fmt.Errorf("get active sessions: %w", err)
	}

	return tokens, nil
}

func (r *repository) DeleteExpired(ctx context.Context) (int64, error) {
	query := `
		DELETE FROM refresh_tokens
		WHERE expires_at < $1`

	cutoff := time.Now().Add(-24 * time.Hour)

	result, err := r.db.ExecContext(ctx, query, cutoff)
	if err != nil {
		return 0, fmt.Errorf("delete expired tokens: %w", err)
	}

	rows, err := result.RowsAffected()
	if err != nil {
		return 0, fmt.Errorf("delete expired tokens: %w", err)
	}

	return rows, nil
}
