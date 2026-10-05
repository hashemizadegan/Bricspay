package auth

import (
	"context"
	"errors"
	"net/http"
	"os"
	"strings"
	"sync"
	"time"

	"github.com/golang-jwt/jwt/v5"
	"golang.org/x/crypto/bcrypt"
)

type contextKey string

const userContextKey contextKey = "bricspay_user"

var (
	ErrEmptySecret  = errors.New("JWT_SECRET is not set")
	ErrInvalidToken = errors.New("invalid token")
	ErrExpiredToken = errors.New("token expired")
)

// Claims uses a string UserID so UUID primary keys are supported.
type Claims struct {
	UserID string `json:"user_id"`
	Role   string `json:"role"`
	jwt.RegisteredClaims
}

var (
	jwtMu     sync.RWMutex
	jwtSecret []byte
)

func init() {
	if s := strings.TrimSpace(os.Getenv("JWT_SECRET")); s != "" {
		jwtSecret = []byte(s)
	}
}

func SetJWTSecret(secret string) {
	s := strings.TrimSpace(secret)
	if s == "" {
		panic("auth.SetJWTSecret: empty secret")
	}
	jwtMu.Lock()
	defer jwtMu.Unlock()
	jwtSecret = []byte(s)
}

func currentSecret() ([]byte, error) {
	jwtMu.RLock()
	defer jwtMu.RUnlock()
	if len(jwtSecret) == 0 {
		return nil, ErrEmptySecret
	}
	return jwtSecret, nil
}

func HashPassword(password string) (string, error) {
	b, err := bcrypt.GenerateFromPassword([]byte(password), 12)
	return string(b), err
}

func CheckPassword(hash, password string) error {
	return bcrypt.CompareHashAndPassword([]byte(hash), []byte(password))
}

func IssueToken(userID string, role string) (string, error) {
	sec, err := currentSecret()
	if err != nil {
		return "", err
	}
	if role == "" {
		role = "user"
	}
	now := time.Now()
	t := jwt.NewWithClaims(jwt.SigningMethodHS256, Claims{
		UserID: userID,
		Role:   role,
		RegisteredClaims: jwt.RegisteredClaims{
			IssuedAt:  jwt.NewNumericDate(now),
			ExpiresAt: jwt.NewNumericDate(now.Add(24 * time.Hour)),
			Issuer:    "bricspay",
		},
	})
	return t.SignedString(sec)
}

func ValidateToken(tokenStr string) (*Claims, error) {
	sec, err := currentSecret()
	if err != nil {
		return nil, err
	}
	parsed, err := jwt.ParseWithClaims(tokenStr, &Claims{}, func(t *jwt.Token) (interface{}, error) {
		if t.Method != jwt.SigningMethodHS256 {
			return nil, ErrInvalidToken
		}
		return sec, nil
	})
	if err != nil {
		if errors.Is(err, jwt.ErrTokenExpired) {
			return nil, ErrExpiredToken
		}
		return nil, ErrInvalidToken
	}
	if !parsed.Valid {
		return nil, ErrInvalidToken
	}
	c, ok := parsed.Claims.(*Claims)
	if !ok {
		return nil, ErrInvalidToken
	}
	return c, nil
}

type Principal struct {
	UserID string
	Role   string
}

func Middleware(next http.Handler) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		h := r.Header.Get("Authorization")
		if !strings.HasPrefix(strings.ToLower(h), "bearer ") {
			http.Error(w, `{"error":"unauthorized"}`, http.StatusUnauthorized)
			return
		}
		c, err := ValidateToken(strings.TrimSpace(h[7:]))
		if err != nil {
			http.Error(w, `{"error":"unauthorized"}`, http.StatusUnauthorized)
			return
		}
		ctx := context.WithValue(r.Context(), userContextKey, Principal{UserID: c.UserID, Role: c.Role})
		next.ServeHTTP(w, r.WithContext(ctx))
	})
}

func AdminOnly(next http.Handler) http.Handler {
	return Middleware(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		p, ok := UserFromContext(r.Context())
		if !ok || p.Role != "admin" {
			http.Error(w, `{"error":"forbidden"}`, http.StatusForbidden)
			return
		}
		next.ServeHTTP(w, r)
	}))
}

func UserFromContext(ctx context.Context) (Principal, bool) {
	p, ok := ctx.Value(userContextKey).(Principal)
	return p, ok
}
