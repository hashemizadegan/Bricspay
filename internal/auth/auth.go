package auth

import (
	"context"
	"errors"
	"fmt"
	"net/http"
	"os"
	"strconv"
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
	ErrInvalidToken = errors.New("invalid or expired token")
	ErrExpiredToken = errors.New("token has expired")

	jwtSecretMu sync.RWMutex
	jwtSecret   []byte
)

func init() {
	if s := os.Getenv("JWT_SECRET"); s != "" {
		jwtSecret = []byte(s)
	}
}

func SetJWTSecret(secret string) {
	jwtSecretMu.Lock()
	defer jwtSecretMu.Unlock()
	jwtSecret = []byte(secret)
}

func currentSecret() ([]byte, error) {
	jwtSecretMu.RLock()
	defer jwtSecretMu.RUnlock()
	if len(jwtSecret) == 0 {
		return nil, ErrEmptySecret
	}
	return jwtSecret, nil
}

type Claims struct {
	UserID   int64  `json:"user_id"`
	UserUUID string `json:"user_uuid,omitempty"`
	Role     string `json:"role"`
	jwt.RegisteredClaims
}

type Principal struct {
	UserID   int64
	UserUUID string
	Role     string
}

func HashPassword(password string) (string, error) {
	bytes, err := bcrypt.GenerateFromPassword([]byte(password), 12)
	return string(bytes), err
}

func CheckPassword(hash, password string) error {
	return bcrypt.CompareHashAndPassword([]byte(hash), []byte(password))
}

// IssueToken پشتیبانی همزمان از int64 و رشته (UUID)
func IssueToken(userID any, role string) (string, error) {
	sec, err := currentSecret()
	if err != nil {
		return "", err
	}

	var idInt int64
	var uuidStr string

	switch v := userID.(type) {
	case int64:
		idInt = v
	case int:
		idInt = int64(v)
	case string:
		uuidStr = v
		// در صورتی که رشته عددی بود تبدیل شود، در غیر این صورت 0 قرار می‌گیرد
		if parsed, err := strconv.ParseInt(v, 10, 64); err == nil {
			idInt = parsed
		}
	default:
		str := fmt.Sprintf("%v", v)
		if parsed, err := strconv.ParseInt(str, 10, 64); err == nil {
			idInt = parsed
		}
		uuidStr = str
	}

	if role == "" {
		role = "user"
	}

	now := time.Now()
	claims := Claims{
		UserID:   idInt,
		UserUUID: uuidStr,
		Role:     role,
		RegisteredClaims: jwt.RegisteredClaims{
			IssuedAt:  jwt.NewNumericDate(now),
			ExpiresAt: jwt.NewNumericDate(now.Add(24 * time.Hour)),
			Issuer:    "bricspay",
		},
	}

	token := jwt.NewWithClaims(jwt.SigningMethodHS256, claims)
	return token.SignedString(sec)
}

func ValidateToken(tokenStr string) (*Claims, error) {
	sec, err := currentSecret()
	if err != nil {
		return nil, err
	}

	token, err := jwt.ParseWithClaims(tokenStr, &Claims{}, func(t *jwt.Token) (interface{}, error) {
		if _, ok := t.Method.(*jwt.SigningMethodHMAC); !ok {
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

	claims, ok := token.Claims.(*Claims)
	if !ok || !token.Valid {
		return nil, ErrInvalidToken
	}

	return claims, nil
}

func Middleware(next http.Handler) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		authHeader := r.Header.Get("Authorization")
		if authHeader == "" || !strings.HasPrefix(authHeader, "Bearer ") {
			http.Error(w, `{"error":"unauthorized"}`, http.StatusUnauthorized)
			return
		}

		tokenStr := strings.TrimPrefix(authHeader, "Bearer ")
		claims, err := ValidateToken(tokenStr)
		if err != nil {
			http.Error(w, `{"error":"unauthorized"}`, http.StatusUnauthorized)
			return
		}

		ctx := context.WithValue(r.Context(), userContextKey, Principal{
			UserID:   claims.UserID,
			UserUUID: claims.UserUUID,
			Role:     claims.Role,
		})
		next.ServeHTTP(w, r.WithContext(ctx))
	})
}

func AdminOnly(next http.Handler) http.Handler {
	return Middleware(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		principal, ok := FromContext(r.Context())
		if !ok || principal.Role != "admin" {
			http.Error(w, `{"error":"forbidden"}`, http.StatusForbidden)
			return
		}
		next.ServeHTTP(w, r)
	}))
}

func FromContext(ctx context.Context) (Principal, bool) {
	p, ok := ctx.Value(userContextKey).(Principal)
	return p, ok
}

func UserFromContext(ctx context.Context) (Principal, bool) {
	return FromContext(ctx)
}
