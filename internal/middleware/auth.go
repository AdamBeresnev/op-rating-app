package middleware

import (
	"context"
	"net/http"
	"os"

	"github.com/alexedwards/scs/v2"
	"github.com/google/uuid"
	"github.com/markbates/goth"
	"github.com/markbates/goth/providers/discord"
)

type ContextKey string

const UserIDKey ContextKey = "userID"
const SuperUserID = "00000000-0000-0000-0000-000000000001"

func InitAuth() {
	discordKey := os.Getenv("DISCORD_KEY")
	discordSecret := os.Getenv("DISCORD_SECRET")

	callbackURL := os.Getenv("AUTH_CALLBACK_URL")

	goth.UseProviders(discord.New(discordKey, discordSecret, callbackURL, discord.ScopeIdentify, discord.ScopeEmail))
}

func RequireAuth(sessionManager *scs.SessionManager) func(http.Handler) http.Handler {
	return func(next http.Handler) http.Handler {
		return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
			userIDStr := sessionManager.GetString(r.Context(), "userID")
			if userIDStr == "" {
				http.Redirect(w, r, "/login", http.StatusFound)
				return
			}

			userID, err := uuid.Parse(userIDStr)
			if err != nil {
				sessionManager.Remove(r.Context(), "userID")
				http.Redirect(w, r, "/login", http.StatusFound)
				return
			}

			ctx := context.WithValue(r.Context(), UserIDKey, userID)
			next.ServeHTTP(w, r.WithContext(ctx))
		})
	}
}

func GetUserIDFromContext(ctx context.Context) (uuid.UUID, bool) {
	val := ctx.Value(UserIDKey)
	if val == nil {
		return uuid.Nil, false
	}

	id, ok := val.(uuid.UUID)
	return id, ok
}
