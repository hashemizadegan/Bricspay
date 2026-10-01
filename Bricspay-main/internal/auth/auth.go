package auth

import (
	"errors"
	"time"

	"github.com/golang-jwt/jwt/v5"
	"golang.org/x/crypto/bcrypt"
)

const (
	tokenTTL     = 24 * time.Hour
	bcryptCost   = 12
)

// Role represents a user's permission level.
type Role string

const (
	RoleUser  Role = "user"
	RoleAdmin Role = "admin"
)

// Claims holds the JWT payload.
type Claims struct {
	UserID int64  `json:"uid"`
	Email  string `json:"email"`
	Role   Role   `json:"role"`
	jwt.RegisteredClaims
}

// Service handles token creation, validation, and password hashing.
type Service struct {
	secret []byte
}

// New creates an auth Service. Panics if secret is empty (caught at startup).
func New(secret string) *Service {
	if secret == "" {
		panic("auth: JWT_SECRET must not be empty")
	}
	return &Service{secret: []byte(secret)}
}

// HashPassword returns a bcrypt hash of the plain-text password.
func (s *Service) HashPassword(plain string) (string, error) {
	if len(plain) < 8 {
		return "", errors.New("password must be at least 8 characters")
	}
	b, err := bcrypt.GenerateFromPassword([]byte(plain), bcryptCost)
	return string(b), err
}

// CheckPassword reports whether plain matches the stored hash.
func (s *Service) CheckPassword(hash, plain string) bool {
	return bcrypt.CompareHashAndPassword([]byte(hash), []byte(plain)) == nil
}

// IssueToken creates a signed JWT for the given user.
func (s *Service) IssueToken(userID int64, email string, role Role) (string, error) {
	now := time.Now()
	claims := Claims{
		UserID: userID,
		Email:  email,
		Role:   role,
		RegisteredClaims: jwt.RegisteredClaims{
			IssuedAt:  jwt.NewNumericDate(now),
			ExpiresAt: jwt.NewNumericDate(now.Add(tokenTTL)),
		},
	}
	token := jwt.NewWithClaims(jwt.SigningMethodHS256, claims)
	return token.SignedString(s.secret)
}

// ValidateToken parses and validates a JWT, returning its claims.
func (s *Service) ValidateToken(tokenStr string) (*Claims, error) {
	token, err := jwt.ParseWithClaims(tokenStr, &Claims{}, func(t *jwt.Token) (any, error) {
		if _, ok := t.Method.(*jwt.SigningMethodHMAC); !ok {
			return nil, errors.New("unexpected signing method")
		}
		return s.secret, nil
	})
	if err != nil {
		return nil, err
	}
	claims, ok := token.Claims.(*Claims)
	if !ok || !token.Valid {
		return nil, errors.New("invalid token")
	}
	return claims, nil
}
