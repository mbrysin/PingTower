package handler

import (
	"context"
	"errors"
	"net/http"
	"net/http/httptest"
	"strings"
	"testing"

	"github.com/gin-gonic/gin"
	"github.com/mbrysin/PingTower/services/auth/model"
	"github.com/mbrysin/PingTower/services/auth/repo"
	"github.com/mbrysin/PingTower/services/auth/service"
)

type fakeTokenManager struct{}

func (fakeTokenManager) Issue(userID int) (string, error) {
	return "test-token", nil
}

func (fakeTokenManager) Verify(raw string) (int, error) {
	return 1, nil
}

type fakeAuthService struct {
	err error
}

func (f fakeAuthService) Register(context.Context, string, string) (*model.User, error) {
	if f.err != nil {
		return nil, f.err
	}
	return &model.User{Id: 1, Username: "alice"}, nil
}

func (f fakeAuthService) Authenticate(context.Context, string, string) (*model.User, error) {
	if f.err != nil {
		return nil, f.err
	}
	return &model.User{Id: 1, Username: "alice"}, nil
}

func TestRegister(t *testing.T) {
	gin.SetMode(gin.TestMode)

	tests := []struct {
		name       string
		body       string
		serviceErr error
		wantStatus int
	}{
		{"success", `{"username":"alice","password":"password123"}`, nil, http.StatusCreated},
		{"invalid input", `{"username":"","password":"short"}`, service.ErrInvalidInput, http.StatusBadRequest},
		{"username taken", `{"username":"alice","password":"password123"}`, repo.ErrUsernameTaken, http.StatusConflict},
		{"internal error", `{"username":"alice","password":"password123"}`, errors.New("db error"), http.StatusInternalServerError},
		{"invalid JSON", `{`, nil, http.StatusBadRequest},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			router := gin.New()
			router.POST("/auth/register", New(fakeAuthService{err: tt.serviceErr}, fakeTokenManager{}).Register)

			req := httptest.NewRequest(http.MethodPost, "/auth/register", strings.NewReader(tt.body))
			req.Header.Set("Content-Type", "application/json")
			rec := httptest.NewRecorder()

			router.ServeHTTP(rec, req)

			if rec.Code != tt.wantStatus {
				t.Errorf("status = %d, want %d; body: %s", rec.Code, tt.wantStatus, rec.Body.String())
			}
		})
	}
}

func TestLogin(t *testing.T) {
	gin.SetMode(gin.TestMode)

	tests := []struct {
		name       string
		body       string
		serviceErr error
		wantStatus int
	}{
		{"success", `{"username":"alice","password":"password123"}`, nil, http.StatusOK},
		{"invalid credentials", `{"username":"alice","password":"wrong"}`, service.ErrInvalidCredentials, http.StatusUnauthorized},
		{"invalid JSON", `{`, nil, http.StatusBadRequest},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			router := gin.New()
			router.POST("/auth/login", New(fakeAuthService{err: tt.serviceErr}, fakeTokenManager{}).Login)

			req := httptest.NewRequest(http.MethodPost, "/auth/login", strings.NewReader(tt.body))
			req.Header.Set("Content-Type", "application/json")
			rec := httptest.NewRecorder()

			router.ServeHTTP(rec, req)

			if rec.Code != tt.wantStatus {
				t.Errorf("status = %d, want %d; body: %s", rec.Code, tt.wantStatus, rec.Body.String())
			}
		})
	}
}
