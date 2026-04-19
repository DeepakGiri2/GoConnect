package repository

import (
	"database/sql"
	"fmt"

	"github.com/goconnect/pkg/models"
)

type UnverifiedUserRepository struct {
	db *sql.DB
}

func NewUnverifiedUserRepository(db *sql.DB) *UnverifiedUserRepository {
	return &UnverifiedUserRepository{db: db}
}

// Create creates a new unverified user
func (r *UnverifiedUserRepository) Create(user *models.UnverifiedUser) error {
	query := `
		INSERT INTO unverified_users (id, username, email, password_hash, expires_at)
		VALUES ($1, $2, $3, $4, $5)
	`
	
	_, err := r.db.Exec(query,
		user.ID,
		user.Username,
		user.Email,
		user.PasswordHash,
		user.ExpiresAt,
	)
	
	if err != nil {
		return fmt.Errorf("failed to create unverified user: %w", err)
	}
	
	return nil
}

// GetByEmail retrieves an unverified user by email
func (r *UnverifiedUserRepository) GetByEmail(email string) (*models.UnverifiedUser, error) {
	query := `
		SELECT id, username, email, password_hash, created_at, expires_at
		FROM unverified_users
		WHERE email = $1
	`
	
	user := &models.UnverifiedUser{}
	err := r.db.QueryRow(query, email).Scan(
		&user.ID,
		&user.Username,
		&user.Email,
		&user.PasswordHash,
		&user.CreatedAt,
		&user.ExpiresAt,
	)
	
	if err != nil {
		if err == sql.ErrNoRows {
			return nil, fmt.Errorf("unverified user not found")
		}
		return nil, fmt.Errorf("failed to get unverified user: %w", err)
	}
	
	return user, nil
}

// GetByUsername retrieves an unverified user by username
func (r *UnverifiedUserRepository) GetByUsername(username string) (*models.UnverifiedUser, error) {
	query := `
		SELECT id, username, email, password_hash, created_at, expires_at
		FROM unverified_users
		WHERE username = $1
	`
	
	user := &models.UnverifiedUser{}
	err := r.db.QueryRow(query, username).Scan(
		&user.ID,
		&user.Username,
		&user.Email,
		&user.PasswordHash,
		&user.CreatedAt,
		&user.ExpiresAt,
	)
	
	if err != nil {
		if err == sql.ErrNoRows {
			return nil, fmt.Errorf("unverified user not found")
		}
		return nil, fmt.Errorf("failed to get unverified user: %w", err)
	}
	
	return user, nil
}

// DeleteByEmail deletes an unverified user by email
func (r *UnverifiedUserRepository) DeleteByEmail(email string) error {
	query := `DELETE FROM unverified_users WHERE email = $1`
	
	_, err := r.db.Exec(query, email)
	if err != nil {
		return fmt.Errorf("failed to delete unverified user: %w", err)
	}
	
	return nil
}

// DeleteByUsername deletes an unverified user by username
func (r *UnverifiedUserRepository) DeleteByUsername(username string) error {
	query := `DELETE FROM unverified_users WHERE username = $1`
	
	_, err := r.db.Exec(query, username)
	if err != nil {
		return fmt.Errorf("failed to delete unverified user: %w", err)
	}
	
	return nil
}

// EmailExists checks if an email exists in unverified_users
func (r *UnverifiedUserRepository) EmailExists(email string) (bool, error) {
	query := `SELECT EXISTS(SELECT 1 FROM unverified_users WHERE email = $1)`
	
	var exists bool
	err := r.db.QueryRow(query, email).Scan(&exists)
	if err != nil {
		return false, fmt.Errorf("failed to check email existence: %w", err)
	}
	
	return exists, nil
}

// UsernameExists checks if a username exists in unverified_users
func (r *UnverifiedUserRepository) UsernameExists(username string) (bool, error) {
	query := `SELECT EXISTS(SELECT 1 FROM unverified_users WHERE username = $1)`
	
	var exists bool
	err := r.db.QueryRow(query, username).Scan(&exists)
	if err != nil {
		return false, fmt.Errorf("failed to check username existence: %w", err)
	}
	
	return exists, nil
}

// CleanupExpired deletes all expired unverified users
// Returns the number of deleted rows
func (r *UnverifiedUserRepository) CleanupExpired() (int64, error) {
	query := `DELETE FROM unverified_users WHERE expires_at < NOW()`
	
	result, err := r.db.Exec(query)
	if err != nil {
		return 0, fmt.Errorf("failed to cleanup expired users: %w", err)
	}
	
	rowsAffected, err := result.RowsAffected()
	if err != nil {
		return 0, fmt.Errorf("failed to get rows affected: %w", err)
	}
	
	return rowsAffected, nil
}

// Count returns the total number of unverified users
func (r *UnverifiedUserRepository) Count() (int64, error) {
	query := `SELECT COUNT(*) FROM unverified_users`
	
	var count int64
	err := r.db.QueryRow(query).Scan(&count)
	if err != nil {
		return 0, fmt.Errorf("failed to count unverified users: %w", err)
	}
	
	return count, nil
}
