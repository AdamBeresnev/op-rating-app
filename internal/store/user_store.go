package store

import (
	"context"

	users "github.com/AdamBeresnev/op-rating-app/internal/user"
	"github.com/jmoiron/sqlx"
)

type UserStore struct {
	db *sqlx.DB
}

const (
	getUserQuery           = "SELECT * FROM users WHERE id = ?"
	getUserByProviderQuery = `
        SELECT * FROM users 
        WHERE provider = ? 
        AND provider_id = ?
    `
	createUserQuery = `
		INSERT INTO users (id, email, username, provider, provider_id, avatar_url) VALUES
		(:id, :email, :username, :provider, :provider_id, :avatar_url)
	`
	updateUserNameAndAvatarQuery = `
		UPDATE users SET
		username = :username,
		avatar_url = :avatar_url,
		WHERE id = :id
	`
)

func NewUserStore(db *sqlx.DB) *UserStore {
	return &UserStore{db: db}
}

func (s *UserStore) GetUserByProvider(ctx context.Context, provider string, providerID string) (*users.User, error) {
	var user users.User
	err := s.db.GetContext(ctx, &user, getUserByProviderQuery, provider, providerID)
	if err != nil {
		return nil, err
	}

	return &user, nil
}

func (s *UserStore) GetUser(ctx context.Context, id interface{}) (*users.User, error) {
	var user users.User
	err := s.db.GetContext(ctx, &user, getUserQuery, id)
	if err != nil {
		return nil, err
	}
	return &user, nil
}

func (s *UserStore) CreateUser(ctx context.Context, user *users.User) error {
	_, err := s.db.NamedExecContext(ctx, createUserQuery, user)
	return err
}

func (s *UserStore) UpdateUserNameAndAvatar(ctx context.Context, user *users.User) error {
	_, err := s.db.NamedExecContext(ctx, updateUserNameAndAvatarQuery, user)
	return err
}
