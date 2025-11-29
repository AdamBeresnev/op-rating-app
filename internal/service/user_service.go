package service

import (
	"context"
	"database/sql"

	"github.com/AdamBeresnev/op-rating-app/internal/store"
	users "github.com/AdamBeresnev/op-rating-app/internal/user"
	"github.com/AdamBeresnev/op-rating-app/internal/utils"
	"github.com/google/uuid"
	"github.com/jmoiron/sqlx"
	"github.com/markbates/goth"
)

type UserService struct {
	db    *sqlx.DB
	store *store.UserStore
}

func NewUserService(db *sqlx.DB, store *store.UserStore) *UserService {
	return &UserService{db: db, store: store}
}

func (s *UserService) FindOrCreateUserByProvider(ctx context.Context, gothUser goth.User) (*users.User, error) {
	user, err := s.store.GetUserByProvider(ctx, gothUser.Provider, gothUser.UserID)

	if err == nil {
		if utils.OrZero(user.AvatarURL) != gothUser.AvatarURL || user.Username != gothUser.NickName {
			*user.AvatarURL = gothUser.AvatarURL
			s.store.UpdateUserNameAndAvatar(ctx, user)
		}
		return user, nil
	}

	if err == sql.ErrNoRows {
		newUser := &users.User{
			ID:         uuid.New(),
			Email:      gothUser.Email,
			Username:   gothUser.Name,
			Provider:   &gothUser.Provider,
			ProviderID: &gothUser.UserID,
			AvatarURL:  &gothUser.AvatarURL,
		}
		err := s.store.CreateUser(ctx, newUser)
		return newUser, err
	}

	return nil, err
}

func (s *UserService) EnsureGuestUser(ctx context.Context) (*users.User, error) {
	guestID := uuid.MustParse("00000000-0000-0000-0000-000000000001")
	user, err := s.store.GetUser(ctx, guestID)
	if err == nil {
		return user, nil
	}

	if err == sql.ErrNoRows {
		guestUser := &users.User{
			ID:       guestID,
			Email:    "guest@op-rating.app",
			Username: "Guest User",
		}
		// We can reuse CreateUser if we handle pointers correctly, or we might need a slightly different logic
		// given the fields in CreateUser.
		// Let's check CreateUser implementation in store.
		err := s.store.CreateUser(ctx, guestUser)
		return guestUser, err
	}
	return nil, err
}
