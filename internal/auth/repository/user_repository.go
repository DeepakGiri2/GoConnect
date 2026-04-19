package repository

import (
	"database/sql"
	"errors"

	"github.com/goconnect/pkg/config"
	"github.com/goconnect/pkg/models"
	"github.com/goconnect/pkg/retry"
)

type UserRepository struct {
	db         *sql.DB
	retryConfig config.RetryConfig
}

func NewUserRepository(db *sql.DB, retryConfig config.RetryConfig) *UserRepository {
	return &UserRepository{
		db:         db,
		retryConfig: retryConfig,
	}
}

func (r *UserRepository) CreateUser(user *models.User) error {
	query := `
		INSERT INTO users (id, username, email, password_hash, is_verified, is_active, email_verified_at, totp_secret, totp_enabled)
		VALUES ($1, $2, $3, $4, $5, $6, $7, $8, $9)
		RETURNING created_at, updated_at
	`
	return r.db.QueryRow(
		query,
		user.ID,
		user.Username,
		user.Email,
		user.PasswordHash,
		user.IsVerified,
		user.IsActive,
		user.EmailVerifiedAt,
		user.TOTPSecret,
		user.TOTPEnabled,
	).Scan(&user.CreatedAt, &user.UpdatedAt)
}

func (r *UserRepository) GetUserByUsername(username string) (*models.User, error) {
	user := &models.User{}
	query := `
		SELECT id, username, email, password_hash, created_at, updated_at, is_active, is_verified,
		       email_verified_at, totp_secret, totp_enabled, totp_verified_at
		FROM users WHERE username = $1
	`
	err := r.db.QueryRow(query, username).Scan(
		&user.ID,
		&user.Username,
		&user.Email,
		&user.PasswordHash,
		&user.CreatedAt,
		&user.UpdatedAt,
		&user.IsActive,
		&user.IsVerified,
		&user.EmailVerifiedAt,
		&user.TOTPSecret,
		&user.TOTPEnabled,
		&user.TOTPVerifiedAt,
	)
	if err == sql.ErrNoRows {
		return nil, errors.New("user not found")
	}
	return user, err
}

func (r *UserRepository) GetUserByEmail(email string) (*models.User, error) {
	user := &models.User{}
	query := `
		SELECT id, username, email, password_hash, created_at, updated_at, is_active, is_verified,
		       email_verified_at, totp_secret, totp_enabled, totp_verified_at
		FROM users WHERE email = $1
	`
	err := r.db.QueryRow(query, email).Scan(
		&user.ID,
		&user.Username,
		&user.Email,
		&user.PasswordHash,
		&user.CreatedAt,
		&user.UpdatedAt,
		&user.IsActive,
		&user.IsVerified,
		&user.EmailVerifiedAt,
		&user.TOTPSecret,
		&user.TOTPEnabled,
		&user.TOTPVerifiedAt,
	)
	if err == sql.ErrNoRows {
		return nil, errors.New("user not found")
	}
	return user, err
}

func (r *UserRepository) GetUserByID(id string) (*models.User, error) {
	user := &models.User{}
	query := `
		SELECT id, username, email, password_hash, created_at, updated_at, is_active, is_verified,
		       email_verified_at, totp_secret, totp_enabled, totp_verified_at
		FROM users WHERE id = $1
	`
	err := r.db.QueryRow(query, id).Scan(
		&user.ID,
		&user.Username,
		&user.Email,
		&user.PasswordHash,
		&user.CreatedAt,
		&user.UpdatedAt,
		&user.IsActive,
		&user.IsVerified,
		&user.EmailVerifiedAt,
		&user.TOTPSecret,
		&user.TOTPEnabled,
		&user.TOTPVerifiedAt,
	)
	if err == sql.ErrNoRows {
		return nil, errors.New("user not found")
	}
	return user, err
}

func (r *UserRepository) UpdatePassword(userID, passwordHash string) error {
	return retry.WithRetry(func() error {
		query := `
			UPDATE users 
			SET password_hash = $1, updated_at = NOW() 
			WHERE id = $2 AND (password_hash != $1 OR password_hash IS NULL)
		`
		result, err := r.db.Exec(query, passwordHash, userID)
		if err != nil {
			return err
		}
		
		rowsAffected, err := result.RowsAffected()
		if err != nil {
			return err
		}
		
		if rowsAffected == 0 {
			var exists bool
			existsQuery := `SELECT EXISTS(SELECT 1 FROM users WHERE id = $1)`
			if err := r.db.QueryRow(existsQuery, userID).Scan(&exists); err != nil {
				return err
			}
			if !exists {
				return errors.New("user not found")
			}
		}
		
		return nil
	}, r.retryConfig)
}

func (r *UserRepository) VerifyUser(userID string) error {
	return retry.WithRetry(func() error {
		query := `
			UPDATE users 
			SET is_verified = true, email_verified_at = NOW(), updated_at = NOW() 
			WHERE id = $1 AND is_verified = false
		`
		result, err := r.db.Exec(query, userID)
		if err != nil {
			return err
		}
		
		rowsAffected, err := result.RowsAffected()
		if err != nil {
			return err
		}
		
		if rowsAffected == 0 {
			var exists bool
			existsQuery := `SELECT EXISTS(SELECT 1 FROM users WHERE id = $1)`
			if err := r.db.QueryRow(existsQuery, userID).Scan(&exists); err != nil {
				return err
			}
			if !exists {
				return errors.New("user not found")
			}
		}
		
		return nil
	}, r.retryConfig)
}

func (r *UserRepository) UsernameExists(username string) (bool, error) {
	var exists bool
	query := `SELECT EXISTS(SELECT 1 FROM users WHERE username = $1)`
	err := r.db.QueryRow(query, username).Scan(&exists)
	return exists, err
}

func (r *UserRepository) EmailExists(email string) (bool, error) {
	var exists bool
	query := `SELECT EXISTS(SELECT 1 FROM users WHERE email = $1)`
	err := r.db.QueryRow(query, email).Scan(&exists)
	return exists, err
}

func (r *UserRepository) GetAllUsernames() ([]string, error) {
	query := `SELECT username FROM users`
	rows, err := r.db.Query(query)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	var usernames []string
	for rows.Next() {
		var username string
		if err := rows.Scan(&username); err != nil {
			return nil, err
		}
		usernames = append(usernames, username)
	}
	return usernames, rows.Err()
}

func (r *UserRepository) UpdateTOTPSecret(userID, encryptedSecret string) error {
	return retry.WithRetry(func() error {
		query := `
			UPDATE users 
			SET totp_secret = $1, updated_at = NOW() 
			WHERE id = $2
		`
		result, err := r.db.Exec(query, encryptedSecret, userID)
		if err != nil {
			return err
		}
		
		rowsAffected, err := result.RowsAffected()
		if err != nil {
			return err
		}
		
		if rowsAffected == 0 {
			return errors.New("user not found")
		}
		
		return nil
	}, r.retryConfig)
}

func (r *UserRepository) EnableTOTP(userID string) error {
	return retry.WithRetry(func() error {
		query := `
			UPDATE users 
			SET totp_enabled = true, totp_verified_at = NOW(), updated_at = NOW() 
			WHERE id = $1 AND totp_enabled = false
		`
		result, err := r.db.Exec(query, userID)
		if err != nil {
			return err
		}
		
		rowsAffected, err := result.RowsAffected()
		if err != nil {
			return err
		}
		
		if rowsAffected == 0 {
			var exists bool
			existsQuery := `SELECT EXISTS(SELECT 1 FROM users WHERE id = $1)`
			if err := r.db.QueryRow(existsQuery, userID).Scan(&exists); err != nil {
				return err
			}
			if !exists {
				return errors.New("user not found")
			}
		}
		
		return nil
	}, r.retryConfig)
}

func (r *UserRepository) DisableTOTP(userID string) error {
	return retry.WithRetry(func() error {
		query := `
			UPDATE users 
			SET totp_enabled = false, totp_secret = NULL, totp_verified_at = NULL, updated_at = NOW() 
			WHERE id = $1 AND totp_enabled = true
		`
		result, err := r.db.Exec(query, userID)
		if err != nil {
			return err
		}
		
		rowsAffected, err := result.RowsAffected()
		if err != nil {
			return err
		}
		
		if rowsAffected == 0 {
			var exists bool
			existsQuery := `SELECT EXISTS(SELECT 1 FROM users WHERE id = $1)`
			if err := r.db.QueryRow(existsQuery, userID).Scan(&exists); err != nil {
				return err
			}
			if !exists {
				return errors.New("user not found")
			}
		}
		
		return nil
	}, r.retryConfig)
}
