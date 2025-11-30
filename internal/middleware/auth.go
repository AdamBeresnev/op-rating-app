package middleware

import (
	"context"
	"net/http"
	"os"

	"github.com/AdamBeresnev/op-rating-app/internal/store"
	users "github.com/AdamBeresnev/op-rating-app/internal/user"
	"github.com/alexedwards/scs/v2"
	"github.com/google/uuid"
	"github.com/markbates/goth"
	"github.com/markbates/goth/providers/discord"
	"github.com/markbates/goth/providers/google"
)

type ContextKey string

const UserIDKey ContextKey = "userID"
const SuperUserID = "00000000-0000-0000-0000-000000000001"

func InitAuth() {
	discordKey := os.Getenv("DISCORD_KEY")
	discordSecret := os.Getenv("DISCORD_SECRET")
	discordCallbackURL := os.Getenv("DISCORD_CALLBACK_URL")

	googleKey := os.Getenv("GOOGLE_KEY")
	googleSecret := os.Getenv("GOOGLE_SECRET")
	googleCallbackURL := os.Getenv("GOOGLE_CALLBACK_URL")

	goth.UseProviders(
		discord.New(discordKey, discordSecret, discordCallbackURL, discord.ScopeIdentify, discord.ScopeEmail),
		google.New(googleKey, googleSecret, googleCallbackURL, "email", "profile"),
	)
}

func RequireAuth(sessionManager *scs.SessionManager, userStore *store.UserStore) func(http.Handler) http.Handler {
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

			// Add the user to context so that we can easily get it whenever we want
			user, err := userStore.GetUser(ctx, userID)
			if err == nil {
				ctx = context.WithValue(ctx, users.UserKey, user)
			}

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

func GetAuthenticatedUser(ctx context.Context) *users.User {
	val := ctx.Value(users.UserKey)
	if val == nil {
		return nil
	}
	user, ok := val.(*users.User)
	if !ok {
		return nil
	}
	return user
}
