package repository

import (
	"database/sql"
	"errors"

	"github.com/goconnect/pkg/config"
	"github.com/goconnect/pkg/models"
	"github.com/goconnect/pkg/retry"
)

type OAuthRepository struct {
	db         *sql.DB
	retryConfig config.RetryConfig
}

func NewOAuthRepository(db *sql.DB, retryConfig config.RetryConfig) *OAuthRepository {
	return &OAuthRepository{
		db:         db,
		retryConfig: retryConfig,
	}
}

func (r *OAuthRepository) CreateOAuthAccount(account *models.OAuthAccount) error {
	query := `
		INSERT INTO oauth_accounts (user_id, provider, provider_user_id, access_token, refresh_token, expires_at)
		VALUES ($1, $2, $3, $4, $5, $6)
		RETURNING id, created_at, updated_at
	`
	return r.db.QueryRow(
		query,
		account.UserID,
		account.Provider,
		account.ProviderUserID,
		account.AccessToken,
		account.RefreshToken,
		account.ExpiresAt,
	).Scan(&account.ID, &account.CreatedAt, &account.UpdatedAt)
}

func (r *OAuthRepository) GetOAuthAccount(provider, providerUserID string) (*models.OAuthAccount, error) {
	account := &models.OAuthAccount{}
	query := `
		SELECT id, user_id, provider, provider_user_id, access_token, refresh_token, expires_at, created_at, updated_at
		FROM oauth_accounts WHERE provider = $1 AND provider_user_id = $2
	`
	err := r.db.QueryRow(query, provider, providerUserID).Scan(
		&account.ID,
		&account.UserID,
		&account.Provider,
		&account.ProviderUserID,
		&account.AccessToken,
		&account.RefreshToken,
		&account.ExpiresAt,
		&account.CreatedAt,
		&account.UpdatedAt,
	)
	if err == sql.ErrNoRows {
		return nil, errors.New("oauth account not found")
	}
	return account, err
}

func (r *OAuthRepository) UpdateOAuthAccount(account *models.OAuthAccount) error {
	return retry.WithRetry(func() error {
		query := `
			UPDATE oauth_accounts 
			SET access_token = $1, refresh_token = $2, expires_at = $3, updated_at = NOW()
			WHERE id = $4 AND (
				access_token != $1 OR 
				refresh_token != $2 OR 
				expires_at != $3 OR
				access_token IS NULL OR
				refresh_token IS NULL OR
				expires_at IS NULL
			)
		`
		result, err := r.db.Exec(query, account.AccessToken, account.RefreshToken, account.ExpiresAt, account.ID)
		if err != nil {
			return err
		}
		
		rowsAffected, err := result.RowsAffected()
		if err != nil {
			return err
		}
		
		if rowsAffected == 0 {
			var exists bool
			existsQuery := `SELECT EXISTS(SELECT 1 FROM oauth_accounts WHERE id = $1)`
			if err := r.db.QueryRow(existsQuery, account.ID).Scan(&exists); err != nil {
				return err
			}
			if !exists {
				return errors.New("oauth account not found")
			}
		}
		
		return nil
	}, r.retryConfig)
}
