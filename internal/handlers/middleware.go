package handlers

import (
	"context"
	"forum/internal/auth"
	"net/http"
)

type contextKey string

const userIDKey contextKey = "userID"

func (h *Handler) Middleware(next http.HandlerFunc) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		cookie, err := r.Cookie("access_token")
		if err != nil {
			writeJSONError(w, "Access token not available", http.StatusUnauthorized)
			return
		}

		userID, valid := auth.ValidateToken(r.Context(), h.conn, cookie.Value, "access")
		if !valid {
			writeJSONError(w, "invalid or expired access token", http.StatusUnauthorized)
			return
		}

		ctx := context.WithValue(r.Context(), userIDKey, userID)
		next(w, r.WithContext(ctx))
	}
}
