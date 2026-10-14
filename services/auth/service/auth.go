package service

import (
	"context"
	"errors"
	"fmt"
	"regexp"
	"strings"

	"golang.org/x/crypto/bcrypt"

	"github.com/mbrysin/PingTower/services/auth/model"
	"github.com/mbrysin/PingTower/services/auth/repo"
)

var (
	ErrInvalidInput       = errors.New("invalid input")
	ErrInvalidCredentials = errors.New("invalid username or password")
)

var usernamePattern = regexp.MustCompile(`^[A-Za-z_]{3,30}$`)

func validUsername(username string) bool {
	return usernamePattern.MatchString(username)
}

type UserRepository interface {
	Create(ctx context.Context, user *model.User) (*model.User, error)
	GetByUsername(ctx context.Context, username string) (*model.User, error)
}

type AuthService struct {
	users      UserRepository
	bcryptCost int
}

func NewAuthService(users UserRepository) *AuthService {
	return &AuthService{
		users:      users,
		bcryptCost: bcrypt.DefaultCost,
	}
}

func (s *AuthService) Register(ctx context.Context, username string, password string) (*model.User, error) {
	username = strings.TrimSpace(username)

	if !validUsername(username) || len(password) < 8 || len(password) > 72 {
		return nil, ErrInvalidInput
	}

	passwordHash, err := bcrypt.GenerateFromPassword(
		[]byte(password),
		s.bcryptCost,
	)
	if err != nil {
		return nil, err
	}

	user := &model.User{
		Username:     username,
		PasswordHash: string(passwordHash),
	}

	createdUser, err := s.users.Create(ctx, user)
	if err != nil {
		return nil, err
	}

	return createdUser, nil
}

func (s *AuthService) Authenticate(ctx context.Context, username string, password string) (*model.User, error) {
	user, err := s.users.GetByUsername(ctx, strings.TrimSpace(username))
	if errors.Is(err, repo.ErrUserNotFound) {
		return nil, ErrInvalidCredentials
	}
	if err != nil {
		return nil, fmt.Errorf("get user by username: %w", err)
	}

	if err := bcrypt.CompareHashAndPassword(
		[]byte(user.PasswordHash), []byte(password),
	); err != nil {
		return nil, ErrInvalidCredentials
	}

	return user, nil
}
