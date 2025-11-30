package users

import (
	"time"

	"github.com/google/uuid"
)

type ContextKey string

const UserKey ContextKey = "user"

type User struct {
	ID         uuid.UUID `db:"id"`
	Email      string    `db:"email"`
	Username   string    `db:"username"`
	CreatedAt  time.Time `db:"created_at"`
	Provider   *string   `db:"provider"`
	ProviderID *string   `db:"provider_id"`
	AvatarURL  *string   `db:"avatar_url"`
}
