package middleware

import (
	"context"
	"net/http"

	"github.com/google/uuid"
)

type ContextKey string

const UserIDKey ContextKey = "userID"
const SuperUserID = "00000000-0000-0000-0000-000000000001"

// Mock auth for now
func RequireAuth(next http.Handler) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		superUserID := uuid.MustParse(SuperUserID)
		ctx := context.WithValue(r.Context(), UserIDKey, superUserID)
		next.ServeHTTP(w, r.WithContext(ctx))
	})
}

func GetUserIDFromContext(ctx context.Context) (uuid.UUID, bool) {
	val := ctx.Value(UserIDKey)
	if val == nil {
		return uuid.Nil, false
	}

	id, ok := val.(uuid.UUID)
	return id, ok
}
