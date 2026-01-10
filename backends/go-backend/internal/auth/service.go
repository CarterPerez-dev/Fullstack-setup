// AngelaMos | 2026
// service.go

package auth

import (
	"context"
	"errors"
	"fmt"
	"time"

	"github.com/google/uuid"
	"github.com/redis/go-redis/v9"

	"github.com/carterperez-dev/templates/go-backend/internal/core"
)

var (
	ErrInvalidCredentials = errors.New("invalid credentials")
	ErrTokenReuse         = errors.New("token reuse detected")
	ErrEmailExists        = errors.New("email already exists")
)

type UserInfo struct {
	ID           string
	Email        string
	Name         string
	PasswordHash string
	Role         string
	Tier         string
	TokenVersion int
}

type UserProvider interface {
	GetByEmail(ctx context.Context, email string) (*UserInfo, error)
	GetByID(ctx context.Context, id string) (*UserInfo, error)
	Create(
		ctx context.Context,
		email, passwordHash, name string,
	) (*UserInfo, error)
	IncrementTokenVersion(ctx context.Context, userID string) error
	UpdatePassword(ctx context.Context, userID, passwordHash string) error
}

type Service struct {
	repo         Repository
	jwt          *JWTManager
	userProvider UserProvider
	redis        *redis.Client
	blacklistTTL time.Duration
}

func NewService(
	repo Repository,
	jwt *JWTManager,
	userProvider UserProvider,
	redisClient *redis.Client,
) *Service {
	return &Service{
		repo:         repo,
		jwt:          jwt,
		userProvider: userProvider,
		redis:        redisClient,
		blacklistTTL: 15 * time.Minute,
	}
}

func (s *Service) Login(
	ctx context.Context,
	req LoginRequest,
	userAgent, ipAddress string,
) (*AuthResponse, error) {
	user, err := s.userProvider.GetByEmail(ctx, req.Email)
	if err != nil {
		if errors.Is(err, core.ErrNotFound) {
			//nolint:errcheck // timing attack prevention - always verify to prevent enumeration
			_, _, _ = core.VerifyPasswordTimingSafe(req.Password, nil)
			return nil, ErrInvalidCredentials
		}
		return nil, fmt.Errorf("get user: %w", err)
	}

	valid, newHash, err := core.VerifyPasswordTimingSafe(
		req.Password,
		&user.PasswordHash,
	)
	if err != nil {
		return nil, fmt.Errorf("verify password: %w", err)
	}

	if !valid {
		return nil, ErrInvalidCredentials
	}

	if newHash != "" {
		//nolint:errcheck // best-effort rehash upgrade
		_ = s.userProvider.UpdatePassword(ctx, user.ID, newHash)
	}

	return s.createAuthResponse(ctx, user, userAgent, ipAddress, "", nil)
}

func (s *Service) Register(
	ctx context.Context,
	req RegisterRequest,
	userAgent, ipAddress string,
) (*AuthResponse, error) {
	passwordHash, err := core.HashPassword(req.Password)
	if err != nil {
		return nil, fmt.Errorf("hash password: %w", err)
	}

	user, err := s.userProvider.Create(ctx, req.Email, passwordHash, req.Name)
	if err != nil {
		if errors.Is(err, core.ErrDuplicateKey) {
			return nil, ErrEmailExists
		}
		return nil, fmt.Errorf("create user: %w", err)
	}

	return s.createAuthResponse(ctx, user, userAgent, ipAddress, "", nil)
}

func (s *Service) Refresh(
	ctx context.Context,
	refreshToken, userAgent, ipAddress string,
) (*AuthResponse, error) {
	tokenHash := core.HashToken(refreshToken)

	storedToken, err := s.repo.FindByHash(ctx, tokenHash)
	if err != nil {
		if errors.Is(err, core.ErrNotFound) {
			return nil, fmt.Errorf("refresh: %w", core.ErrTokenInvalid)
		}
		return nil, fmt.Errorf("find token: %w", err)
	}

	if storedToken.IsUsed {
		//nolint:errcheck // security revocation continues regardless
		_ = s.repo.RevokeByFamilyID(ctx, storedToken.FamilyID)
		return nil, ErrTokenReuse
	}

	if !storedToken.IsValid() {
		if storedToken.IsRevoked() {
			return nil, fmt.Errorf("refresh: %w", core.ErrTokenRevoked)
		}
		return nil, fmt.Errorf("refresh: %w", core.ErrTokenExpired)
	}

	user, err := s.userProvider.GetByID(ctx, storedToken.UserID)
	if err != nil {
		return nil, fmt.Errorf("get user: %w", err)
	}

	return s.createAuthResponse(
		ctx,
		user,
		userAgent,
		ipAddress,
		storedToken.FamilyID,
		&storedToken.ID,
	)
}

func (s *Service) Logout(
	ctx context.Context,
	refreshToken, userID string,
) error {
	tokenHash := core.HashToken(refreshToken)

	storedToken, err := s.repo.FindByHash(ctx, tokenHash)
	if err != nil {
		if errors.Is(err, core.ErrNotFound) {
			return nil
		}
		return fmt.Errorf("find token: %w", err)
	}

	if storedToken.UserID != userID {
		return fmt.Errorf("logout: %w", core.ErrForbidden)
	}

	if err := s.repo.RevokeByID(ctx, storedToken.ID); err != nil &&
		!errors.Is(err, core.ErrNotFound) {
		return fmt.Errorf("revoke token: %w", err)
	}

	return nil
}

func (s *Service) LogoutAll(ctx context.Context, userID string) error {
	if err := s.repo.RevokeAllForUser(ctx, userID); err != nil {
		return fmt.Errorf("revoke all tokens: %w", err)
	}

	if err := s.userProvider.IncrementTokenVersion(ctx, userID); err != nil {
		return fmt.Errorf("increment token version: %w", err)
	}

	return nil
}

func (s *Service) RevokeAccessToken(
	ctx context.Context,
	jti string,
	expiresAt time.Time,
) error {
	key := "blacklist:" + jti
	ttl := time.Until(expiresAt)

	if ttl <= 0 {
		return nil
	}

	if err := s.redis.Set(ctx, key, "1", ttl).Err(); err != nil {
		return fmt.Errorf("blacklist token: %w", err)
	}

	return nil
}

func (s *Service) IsAccessTokenBlacklisted(
	ctx context.Context,
	jti string,
) (bool, error) {
	key := "blacklist:" + jti

	exists, err := s.redis.Exists(ctx, key).Result()
	if err != nil {
		return false, fmt.Errorf("check blacklist: %w", err)
	}

	return exists > 0, nil
}

func (s *Service) GetActiveSessions(
	ctx context.Context,
	userID string,
) ([]SessionInfo, error) {
	tokens, err := s.repo.GetActiveSessionsForUser(ctx, userID)
	if err != nil {
		return nil, fmt.Errorf("get sessions: %w", err)
	}

	sessions := make([]SessionInfo, 0, len(tokens))
	for _, t := range tokens {
		sessions = append(sessions, SessionInfo{
			ID:        t.ID,
			UserAgent: t.UserAgent,
			IPAddress: t.IPAddress,
			CreatedAt: t.CreatedAt,
			ExpiresAt: t.ExpiresAt,
		})
	}

	return sessions, nil
}

func (s *Service) RevokeSession(
	ctx context.Context,
	userID, sessionID string,
) error {
	token, err := s.repo.FindByID(ctx, sessionID)
	if err != nil {
		return fmt.Errorf("find session: %w", err)
	}

	if token.UserID != userID {
		return fmt.Errorf("revoke session: %w", core.ErrForbidden)
	}

	if err := s.repo.RevokeByID(ctx, sessionID); err != nil {
		return fmt.Errorf("revoke session: %w", err)
	}

	return nil
}

func (s *Service) ChangePassword(
	ctx context.Context,
	userID, currentPassword, newPassword string,
) error {
	user, err := s.userProvider.GetByID(ctx, userID)
	if err != nil {
		return fmt.Errorf("get user: %w", err)
	}

	valid, _, err := core.VerifyPasswordWithRehash(
		currentPassword,
		user.PasswordHash,
	)
	if err != nil {
		return fmt.Errorf("verify password: %w", err)
	}

	if !valid {
		return ErrInvalidCredentials
	}

	newHash, err := core.HashPassword(newPassword)
	if err != nil {
		return fmt.Errorf("hash password: %w", err)
	}

	if err := s.userProvider.UpdatePassword(ctx, userID, newHash); err != nil {
		return fmt.Errorf("update password: %w", err)
	}

	if err := s.LogoutAll(ctx, userID); err != nil {
		return fmt.Errorf("logout all: %w", err)
	}

	return nil
}

func (s *Service) ValidateTokenVersion(
	ctx context.Context,
	userID string,
	tokenVersion int,
) error {
	user, err := s.userProvider.GetByID(ctx, userID)
	if err != nil {
		return fmt.Errorf("get user: %w", err)
	}

	if tokenVersion < user.TokenVersion {
		return fmt.Errorf("validate token version: %w", core.ErrTokenRevoked)
	}

	return nil
}

func (s *Service) GetCurrentUser(
	ctx context.Context,
	userID string,
) (*UserResponse, error) {
	user, err := s.userProvider.GetByID(ctx, userID)
	if err != nil {
		return nil, err
	}

	return &UserResponse{
		ID:    user.ID,
		Email: user.Email,
		Name:  user.Name,
		Role:  user.Role,
		Tier:  user.Tier,
	}, nil
}

func (s *Service) createAuthResponse(
	ctx context.Context,
	user *UserInfo,
	userAgent, ipAddress, familyID string,
	oldTokenID *string,
) (*AuthResponse, error) {
	accessToken, err := s.jwt.CreateAccessToken(AccessTokenClaims{
		UserID:       user.ID,
		Role:         user.Role,
		Tier:         user.Tier,
		TokenVersion: user.TokenVersion,
	})
	if err != nil {
		return nil, fmt.Errorf("create access token: %w", err)
	}

	refreshData, err := s.jwt.CreateRefreshToken(user.ID, familyID)
	if err != nil {
		return nil, fmt.Errorf("create refresh token: %w", err)
	}

	newTokenID := uuid.New().String()

	refreshTokenEntity := &RefreshToken{
		ID:        newTokenID,
		UserID:    user.ID,
		TokenHash: refreshData.Hash,
		FamilyID:  refreshData.FamilyID,
		ExpiresAt: refreshData.ExpiresAt,
		UserAgent: userAgent,
		IPAddress: ipAddress,
	}

	if err := s.repo.Create(ctx, refreshTokenEntity); err != nil {
		return nil, fmt.Errorf("store refresh token: %w", err)
	}

	if oldTokenID != nil {
		//nolint:errcheck // best-effort token chain tracking
		_ = s.repo.MarkAsUsed(ctx, *oldTokenID, newTokenID)
	}

	return &AuthResponse{
		User: UserResponse{
			ID:        user.ID,
			Email:     user.Email,
			Name:      user.Name,
			Role:      user.Role,
			Tier:      user.Tier,
			CreatedAt: time.Now(),
		},
		Tokens: TokenResponse{
			AccessToken:  accessToken,
			RefreshToken: refreshData.Token,
			TokenType:    "Bearer",
			ExpiresIn:    int(15 * time.Minute / time.Second),
			ExpiresAt:    time.Now().Add(15 * time.Minute),
		},
	}, nil
}
