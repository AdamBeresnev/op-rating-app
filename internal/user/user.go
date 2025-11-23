package users

import (
	"time"

	"github.com/google/uuid"
)

type User struct {
	ID        uuid.UUID `db:"id"`
	Email     string    `db:"email"`
	Username  string    `db:"username"`
	CreatedAt time.Time `db:"time"`
}
