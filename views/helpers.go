package views

import (
	"context"

	"github.com/AdamBeresnev/op-rating-app/internal/middleware"
	users "github.com/AdamBeresnev/op-rating-app/internal/user"
)

func GetUser(ctx context.Context) *users.User {
	return middleware.GetAuthenticatedUser(ctx)
}
