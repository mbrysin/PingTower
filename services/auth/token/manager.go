package token

import (
	"fmt"
	"strconv"
	"time"

	"github.com/golang-jwt/jwt/v5"
)

type Manager struct {
	secret []byte
}

func NewManager(secret string) (*Manager, error) {
	if len(secret) < 32 {
		return nil, fmt.Errorf("JWT secret must be at least 32 bytes")
	}
	return &Manager{secret: []byte(secret)}, nil
}

func (m *Manager) Issue(userID int) (string, error) {
	claims := jwt.RegisteredClaims{
		Subject:   strconv.Itoa(userID),
		ExpiresAt: jwt.NewNumericDate(time.Now().Add(15 * time.Minute)),
		IssuedAt:  jwt.NewNumericDate(time.Now()),
	}
	return jwt.NewWithClaims(jwt.SigningMethodHS256, claims).SignedString(m.secret)
}

func (m *Manager) Verify(raw string) (int, error) {
	claims := &jwt.RegisteredClaims{}
	_, err := jwt.ParseWithClaims(raw, claims, func(t *jwt.Token) (any, error) {
		return m.secret, nil
	}, jwt.WithValidMethods([]string{"HS256"}))
	if err != nil {
		return 0, err
	}

	return strconv.Atoi(claims.Subject)
}
