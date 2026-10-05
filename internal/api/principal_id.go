package api

import (
	"net/http"

	"bricspay/internal/auth"
)

func principalUserID64(r *http.Request) (int64, bool) {
	principal, ok := auth.FromContext(r.Context())
	if !ok {
		return 0, false
	}

	return principal.UserID, true
}
