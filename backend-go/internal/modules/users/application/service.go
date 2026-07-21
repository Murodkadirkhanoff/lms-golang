package application

import (
	"context"
	"errors"
	"time"

	"github.com/chashma/lms/internal/modules/users/contract"
	"github.com/chashma/lms/internal/modules/users/domain"
	"golang.org/x/crypto/bcrypt"
)

// bcryptCost matches the Java/Go services ($2a$ cost 12) so existing password
// hashes keep verifying after the migration.
const bcryptCost = 12

// resetTokenTTL is how long a password-reset token stays valid.
const resetTokenTTL = 45 * time.Minute

// ErrInvalidResetToken is returned when a reset token is unknown or expired.
var ErrInvalidResetToken = errors.New("invalid or expired password reset token")

// Service implements the users use cases and the UserDirectory contract.
type Service struct {
	repo   Repository
	tokens TokenRepository
	issuer TokenIssuer
	mailer Mailer
}

// NewService wires the users service.
func NewService(repo Repository, tokens TokenRepository, issuer TokenIssuer, mailer Mailer) *Service {
	return &Service{repo: repo, tokens: tokens, issuer: issuer, mailer: mailer}
}

var _ contract.UserDirectory = (*Service)(nil)

// Register creates a new student account and returns it with a fresh token.
func (s *Service) Register(ctx context.Context, name, email, password string) (*domain.User, string, error) {
	hash, err := bcrypt.GenerateFromPassword([]byte(password), bcryptCost)
	if err != nil {
		return nil, "", err
	}
	u := &domain.User{Name: name, Email: email, Role: domain.RoleStudent, PasswordHash: hash}
	if err := s.repo.Insert(ctx, u); err != nil {
		return nil, "", err
	}
	token, err := s.issuer.New(u.ID, u.Role)
	if err != nil {
		return nil, "", err
	}
	return u, token, nil
}

// Authenticate verifies credentials and returns the user with a fresh token.
func (s *Service) Authenticate(ctx context.Context, email, password string) (*domain.User, string, error) {
	u, err := s.repo.FindByEmail(ctx, email)
	if err != nil {
		if errors.Is(err, domain.ErrNotFound) {
			return nil, "", domain.ErrInvalidCredentials
		}
		return nil, "", err
	}
	if bcrypt.CompareHashAndPassword(u.PasswordHash, []byte(password)) != nil {
		return nil, "", domain.ErrInvalidCredentials
	}
	token, err := s.issuer.New(u.ID, u.Role)
	if err != nil {
		return nil, "", err
	}
	return u, token, nil
}

// ForgotPassword issues a reset token and emails the link. It never reveals
// whether the email exists.
func (s *Service) ForgotPassword(ctx context.Context, email string) error {
	u, err := s.repo.FindByEmail(ctx, email)
	if err != nil {
		if errors.Is(err, domain.ErrNotFound) {
			return nil
		}
		return err
	}
	token, err := s.tokens.Create(ctx, u.ID, resetTokenTTL)
	if err != nil {
		return err
	}
	s.mailer.SendPasswordReset(u.Email, token)
	return nil
}

// ResetPassword sets a new password given a valid reset token.
func (s *Service) ResetPassword(ctx context.Context, password, token string) error {
	userID, err := s.tokens.UserIDForToken(ctx, token)
	if err != nil {
		if errors.Is(err, domain.ErrNotFound) {
			return ErrInvalidResetToken
		}
		return err
	}
	u, err := s.repo.FindByID(ctx, userID)
	if err != nil {
		return err
	}
	hash, err := bcrypt.GenerateFromPassword([]byte(password), bcryptCost)
	if err != nil {
		return err
	}
	u.PasswordHash = hash
	if err := s.repo.UpdatePassword(ctx, u); err != nil {
		return err
	}
	return s.tokens.DeleteAllForUser(ctx, u.ID)
}

// Me returns the current user.
func (s *Service) Me(ctx context.Context, id int64) (*domain.User, error) {
	return s.repo.FindByID(ctx, id)
}

// UpdateProfile changes the display name.
func (s *Service) UpdateProfile(ctx context.Context, id int64, name string) (*domain.User, error) {
	u, err := s.repo.FindByID(ctx, id)
	if err != nil {
		return nil, err
	}
	u.Name = name
	if err := s.repo.UpdateName(ctx, u); err != nil {
		return nil, err
	}
	return u, nil
}

// ChangePassword updates the password after confirming the current one.
func (s *Service) ChangePassword(ctx context.Context, id int64, current, next string) error {
	u, err := s.repo.FindByID(ctx, id)
	if err != nil {
		return err
	}
	if bcrypt.CompareHashAndPassword(u.PasswordHash, []byte(current)) != nil {
		return domain.ErrInvalidCredentials
	}
	hash, err := bcrypt.GenerateFromPassword([]byte(next), bcryptCost)
	if err != nil {
		return err
	}
	u.PasswordHash = hash
	return s.repo.UpdatePassword(ctx, u)
}

// ListUsers returns a page of users and the total count.
func (s *Service) ListUsers(ctx context.Context, page, pageSize int) ([]domain.User, int, error) {
	return s.repo.List(ctx, page, pageSize)
}

// UpdateRole changes a user's role (admin panel).
func (s *Service) UpdateRole(ctx context.Context, id int64, role string) error {
	return s.repo.UpdateRole(ctx, id, role)
}

// CountUsers returns the number of active users.
func (s *Service) CountUsers(ctx context.Context) (int, error) {
	return s.repo.Count(ctx)
}

// FindByIDs implements contract.UserDirectory.
func (s *Service) FindByIDs(ctx context.Context, ids []int64) ([]contract.UserSummary, error) {
	if len(ids) == 0 {
		return []contract.UserSummary{}, nil
	}
	users, err := s.repo.FindByIDs(ctx, ids)
	if err != nil {
		return nil, err
	}
	out := make([]contract.UserSummary, 0, len(users))
	for _, u := range users {
		out = append(out, contract.UserSummary{
			ID: u.ID, CreatedAt: u.CreatedAt, Name: u.Name, Email: u.Email, Role: u.Role,
		})
	}
	return out, nil
}
