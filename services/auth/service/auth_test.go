package service

import (
	"context"
	"errors"
	"testing"

	"golang.org/x/crypto/bcrypt"

	"github.com/mbrysin/PingTower/services/auth/model"
	"github.com/mbrysin/PingTower/services/auth/repo"
)

var errRepo = errors.New("repository error")

type fakeUserRepository struct {
	users  map[int]model.User
	nextID int

	createErr         error
	getErr            error
	getByUsernameErr  error
	updateUsernameErr error
	updatePasswordErr error
	deleteErr         error
	listErr           error
	lastUsername      string
	lastPasswordHash  string
}

func newFakeUserRepository() *fakeUserRepository {
	return &fakeUserRepository{
		users:  make(map[int]model.User),
		nextID: 1,
	}
}

func (r *fakeUserRepository) Create(_ context.Context, user *model.User) (*model.User, error) {
	if r.createErr != nil {
		return nil, r.createErr
	}

	user.Id = r.nextID
	r.nextID++
	r.users[user.Id] = *user
	return user, nil
}

func (r *fakeUserRepository) Get(_ context.Context, id int) (*model.User, error) {
	if r.getErr != nil {
		return nil, r.getErr
	}

	user, ok := r.users[id]
	if !ok {
		return nil, repo.ErrUserNotFound
	}

	userCopy := user
	return &userCopy, nil
}

func (r *fakeUserRepository) GetByUsername(_ context.Context, username string) (*model.User, error) {
	if r.getByUsernameErr != nil {
		return nil, r.getByUsernameErr
	}

	for _, user := range r.users {
		if user.Username == username {
			userCopy := user
			return &userCopy, nil
		}
	}

	return nil, repo.ErrUserNotFound
}

func (r *fakeUserRepository) UpdateUsername(_ context.Context, id int, username string) (*model.User, error) {
	if r.updateUsernameErr != nil {
		return nil, r.updateUsernameErr
	}

	user, ok := r.users[id]
	if !ok {
		return nil, repo.ErrUserNotFound
	}

	user.Username = username
	r.users[id] = user
	r.lastUsername = username
	return &user, nil
}

func (r *fakeUserRepository) UpdatePassword(_ context.Context, id int, passwordHash string) (*model.User, error) {
	if r.updatePasswordErr != nil {
		return nil, r.updatePasswordErr
	}

	user, ok := r.users[id]
	if !ok {
		return nil, repo.ErrUserNotFound
	}

	user.PasswordHash = passwordHash
	r.users[id] = user
	r.lastPasswordHash = passwordHash
	return &user, nil
}

func (r *fakeUserRepository) Delete(_ context.Context, id int) error {
	if r.deleteErr != nil {
		return r.deleteErr
	}

	if _, ok := r.users[id]; !ok {
		return repo.ErrUserNotFound
	}

	delete(r.users, id)
	return nil
}

func (r *fakeUserRepository) List(_ context.Context) ([]model.User, error) {
	if r.listErr != nil {
		return nil, r.listErr
	}

	users := make([]model.User, 0, len(r.users))
	for _, user := range r.users {
		users = append(users, user)
	}

	return users, nil
}

func newTestService(repo *fakeUserRepository) *AuthService {
	svc := NewAuthService(repo)
	svc.bcryptCost = bcrypt.MinCost
	return svc
}

func TestRegister(t *testing.T) {
	t.Run("success", func(t *testing.T) {
		repo := newFakeUserRepository()
		svc := newTestService(repo)

		user, err := svc.Register(context.Background(), "  alice  ", "password123")
		if err != nil {
			t.Fatalf("Register() error = %v", err)
		}

		if user.Username != "alice" {
			t.Errorf("Username = %q, want %q", user.Username, "alice")
		}
		if user.Id == 0 {
			t.Error("expected repository to assign user ID")
		}
		if user.PasswordHash == "" || user.PasswordHash == "password123" {
			t.Error("expected password to be stored as a hash")
		}
		if err := bcrypt.CompareHashAndPassword(
			[]byte(user.PasswordHash),
			[]byte("password123"),
		); err != nil {
			t.Errorf("password hash does not match: %v", err)
		}
	})

	t.Run("rejects empty username", func(t *testing.T) {
		svc := newTestService(newFakeUserRepository())

		_, err := svc.Register(context.Background(), "   ", "password123")
		if !errors.Is(err, ErrInvalidInput) {
			t.Fatalf("Register() error = %v, want %v", err, ErrInvalidInput)
		}
	})

	t.Run("rejects short password", func(t *testing.T) {
		svc := newTestService(newFakeUserRepository())

		_, err := svc.Register(context.Background(), "alice", "short")
		if !errors.Is(err, ErrInvalidInput) {
			t.Fatalf("Register() error = %v, want %v", err, ErrInvalidInput)
		}
	})

	t.Run("returns repository error", func(t *testing.T) {
		repo := newFakeUserRepository()
		repo.createErr = errRepo
		svc := newTestService(repo)

		_, err := svc.Register(context.Background(), "alice", "password123")
		if !errors.Is(err, errRepo) {
			t.Fatalf("Register() error = %v, want %v", err, errRepo)
		}
	})
}

func TestAuthenticate(t *testing.T) {
	repo := newFakeUserRepository()
	svc := newTestService(repo)

	createdUser, err := svc.Register(context.Background(), "alice", "password123")
	if err != nil {
		t.Fatalf("Register() error = %v", err)
	}

	t.Run("success", func(t *testing.T) {
		user, err := svc.Authenticate(context.Background(), "  alice  ", "password123")
		if err != nil {
			t.Fatalf("Authenticate() error = %v", err)
		}
		if user.Id != createdUser.Id {
			t.Errorf("user ID = %d, want %d", user.Id, createdUser.Id)
		}
	})

	t.Run("rejects wrong password", func(t *testing.T) {
		_, err := svc.Authenticate(context.Background(), "alice", "wrongpassword")
		if !errors.Is(err, ErrInvalidCredentials) {
			t.Fatalf("Authenticate() error = %v, want %v", err, ErrInvalidCredentials)
		}
	})

	t.Run("rejects unknown username", func(t *testing.T) {
		_, err := svc.Authenticate(context.Background(), "unknown", "password123")
		if !errors.Is(err, ErrInvalidCredentials) {
			t.Fatalf("Authenticate() error = %v, want %v", err, ErrInvalidCredentials)
		}
	})

	t.Run("preserves repository error", func(t *testing.T) {
		failingRepo := newFakeUserRepository()
		failingRepo.getByUsernameErr = errRepo
		failingService := newTestService(failingRepo)

		_, err := failingService.Authenticate(context.Background(), "alice", "password123")
		if !errors.Is(err, errRepo) {
			t.Fatalf("Authenticate() error = %v, want wrapped %v", err, errRepo)
		}
	})
}
