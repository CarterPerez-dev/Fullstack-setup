// AngelaMos | 2026
// service.go

package user

import (
	"context"
	"fmt"
	"strings"

	"github.com/google/uuid"

	"github.com/carterperez-dev/templates/go-backend/internal/auth"
	"github.com/carterperez-dev/templates/go-backend/internal/core"
)

type Service struct {
	repo Repository
}

func NewService(repo Repository) *Service {
	return &Service{repo: repo}
}

func (s *Service) GetByID(
	ctx context.Context,
	id string,
) (*auth.UserInfo, error) {
	user, err := s.repo.GetByID(ctx, id)
	if err != nil {
		return nil, err
	}

	return toUserInfo(user), nil
}

func (s *Service) GetByEmail(
	ctx context.Context,
	email string,
) (*auth.UserInfo, error) {
	user, err := s.repo.GetByEmail(ctx, strings.ToLower(email))
	if err != nil {
		return nil, err
	}

	return toUserInfo(user), nil
}

func (s *Service) Create(
	ctx context.Context,
	email, passwordHash, name string,
) (*auth.UserInfo, error) {
	user := &User{
		ID:           uuid.New().String(),
		Email:        strings.ToLower(email),
		PasswordHash: passwordHash,
		Name:         name,
		Role:         RoleUser,
		Tier:         TierFree,
	}

	if err := s.repo.Create(ctx, user); err != nil {
		return nil, err
	}

	return toUserInfo(user), nil
}

func (s *Service) IncrementTokenVersion(
	ctx context.Context,
	userID string,
) error {
	return s.repo.IncrementTokenVersion(ctx, userID)
}

func (s *Service) UpdatePassword(
	ctx context.Context,
	userID, passwordHash string,
) error {
	return s.repo.UpdatePassword(ctx, userID, passwordHash)
}

func (s *Service) GetUser(ctx context.Context, id string) (*User, error) {
	return s.repo.GetByID(ctx, id)
}

func (s *Service) UpdateUser(
	ctx context.Context,
	id string,
	req UpdateUserRequest,
) (*User, error) {
	user, err := s.repo.GetByID(ctx, id)
	if err != nil {
		return nil, err
	}

	if req.Name != nil {
		user.Name = *req.Name
	}

	if err := s.repo.Update(ctx, user); err != nil {
		return nil, err
	}

	return user, nil
}

func (s *Service) UpdateUserRole(
	ctx context.Context,
	id, role string,
) (*User, error) {
	if role != RoleUser && role != RoleAdmin {
		return nil, fmt.Errorf(
			"update role: invalid role %q: %w",
			role,
			core.ErrInvalidInput,
		)
	}

	user, err := s.repo.GetByID(ctx, id)
	if err != nil {
		return nil, err
	}

	user.Role = role

	if err := s.repo.Update(ctx, user); err != nil {
		return nil, err
	}

	return user, nil
}

func (s *Service) UpdateUserTier(
	ctx context.Context,
	id, tier string,
) (*User, error) {
	if tier != TierFree && tier != TierPro && tier != TierEnterprise {
		return nil, fmt.Errorf(
			"update tier: invalid tier %q: %w",
			tier,
			core.ErrInvalidInput,
		)
	}

	user, err := s.repo.GetByID(ctx, id)
	if err != nil {
		return nil, err
	}

	user.Tier = tier

	if err := s.repo.Update(ctx, user); err != nil {
		return nil, err
	}

	return user, nil
}

func (s *Service) DeleteUser(ctx context.Context, id string) error {
	return s.repo.SoftDelete(ctx, id)
}

func (s *Service) ListUsers(
	ctx context.Context,
	params ListUsersParams,
) ([]User, int, error) {
	return s.repo.List(ctx, params)
}

func (s *Service) GetMe(ctx context.Context, userID string) (*User, error) {
	if userID == "" {
		return nil, fmt.Errorf("get me: %w", core.ErrUnauthorized)
	}

	user, err := s.repo.GetByID(ctx, userID)
	if err != nil {
		return nil, err
	}

	return user, nil
}

func (s *Service) UpdateMe(
	ctx context.Context,
	userID string,
	req UpdateUserRequest,
) (*User, error) {
	if userID == "" {
		return nil, fmt.Errorf("update me: %w", core.ErrUnauthorized)
	}

	return s.UpdateUser(ctx, userID, req)
}

func (s *Service) DeleteMe(ctx context.Context, userID string) error {
	if userID == "" {
		return fmt.Errorf("delete me: %w", core.ErrUnauthorized)
	}

	return s.repo.SoftDelete(ctx, userID)
}

func (s *Service) EmailExists(
	ctx context.Context,
	email string,
) (bool, error) {
	exists, err := s.repo.ExistsByEmail(ctx, email)
	if err != nil {
		return false, err
	}
	return exists, nil
}

func (s *Service) CanDeleteUser(
	ctx context.Context,
	requesterID, targetID string,
) error {
	if requesterID == targetID {
		return nil
	}

	requester, err := s.repo.GetByID(ctx, requesterID)
	if err != nil {
		return err
	}

	if !requester.IsAdmin() {
		return fmt.Errorf("delete user: %w", core.ErrForbidden)
	}

	target, err := s.repo.GetByID(ctx, targetID)
	if err != nil {
		return err
	}

	if target.IsAdmin() {
		return fmt.Errorf("cannot delete admin users: %w", core.ErrForbidden)
	}

	return nil
}

func toUserInfo(u *User) *auth.UserInfo {
	return &auth.UserInfo{
		ID:           u.ID,
		Email:        u.Email,
		Name:         u.Name,
		PasswordHash: u.PasswordHash,
		Role:         u.Role,
		Tier:         u.Tier,
		TokenVersion: u.TokenVersion,
	}
}

var _ auth.UserProvider = (*Service)(nil)
