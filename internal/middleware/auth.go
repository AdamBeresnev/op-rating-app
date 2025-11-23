package middleware

import (
	"context"
	"net/http"

	"github.com/google/uuid"
)

type ContextKey string

const UserIDKey ContextKey = "userID"

// Mock auth for now
func RequireAuth(next http.Handler) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		superUserID := uuid.MustParse("00000000-0000-0000-0000-000000000001")
		ctx := context.WithValue(r.Context(), UserIDKey, superUserID)
		next.ServeHTTP(w, r.WithContext(ctx))
	})
}
