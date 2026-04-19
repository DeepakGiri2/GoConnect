package repository

import (
	"database/sql"
	"errors"

	"github.com/goconnect/pkg/models"
)

type TokenRepository struct {
	db *sql.DB
}

func NewTokenRepository(db *sql.DB) *TokenRepository {
	return &TokenRepository{db: db}
}

func (r *TokenRepository) CreateRefreshToken(token *models.RefreshToken) error {
	query := `
		INSERT INTO refresh_tokens (user_id, token, expires_at)
		VALUES ($1, $2, $3)
		RETURNING id, created_at
	`
	return r.db.QueryRow(query, token.UserID, token.Token, token.ExpiresAt).Scan(&token.ID, &token.CreatedAt)
}

func (r *TokenRepository) GetRefreshToken(token string) (*models.RefreshToken, error) {
	rt := &models.RefreshToken{}
	query := `
		SELECT id, user_id, token, expires_at, created_at, is_revoked
		FROM refresh_tokens WHERE token = $1
	`
	err := r.db.QueryRow(query, token).Scan(
		&rt.ID,
		&rt.UserID,
		&rt.Token,
		&rt.ExpiresAt,
		&rt.CreatedAt,
		&rt.IsRevoked,
	)
	if err == sql.ErrNoRows {
		return nil, errors.New("refresh token not found")
	}
	return rt, err
}

func (r *TokenRepository) RevokeRefreshToken(token string) error {
	query := `UPDATE refresh_tokens SET is_revoked = true WHERE token = $1`
	_, err := r.db.Exec(query, token)
	return err
}

func (r *TokenRepository) RevokeAllUserTokens(userID string) error {
	query := `UPDATE refresh_tokens SET is_revoked = true WHERE user_id = $1`
	_, err := r.db.Exec(query, userID)
	return err
}

func (r *TokenRepository) CleanupExpiredTokens() error {
	query := `DELETE FROM refresh_tokens WHERE expires_at < NOW()`
	_, err := r.db.Exec(query)
	return err
}
