package repository

import (
	"database/sql"
	"errors"
	"time"

	"github.com/goconnect/pkg/crypto"
)

type TOTPRepository struct {
	db        *sql.DB
	encryptor *crypto.Encryptor
}

func NewTOTPRepository(db *sql.DB, encryptionKey string) (*TOTPRepository, error) {
	encryptor, err := crypto.NewEncryptor(encryptionKey)
	if err != nil {
		return nil, err
	}
	return &TOTPRepository{
		db:        db,
		encryptor: encryptor,
	}, nil
}

func (r *TOTPRepository) SaveTOTPSecret(userID, secret string) error {
	encryptedSecret, err := r.encryptor.EncryptTOTPSecret(secret)
	if err != nil {
		return err
	}

	query := `
		UPDATE users 
		SET totp_secret = $1, totp_enabled = true, updated_at = NOW()
		WHERE id = $2
	`
	_, err = r.db.Exec(query, encryptedSecret, userID)
	return err
}

func (r *TOTPRepository) GetTOTPSecret(userID string) (string, error) {
	var encryptedSecret sql.NullString
	query := `SELECT totp_secret FROM users WHERE id = $1 AND totp_enabled = true`
	
	err := r.db.QueryRow(query, userID).Scan(&encryptedSecret)
	if err == sql.ErrNoRows {
		return "", errors.New("TOTP not enabled for user")
	}
	if err != nil {
		return "", err
	}

	if !encryptedSecret.Valid {
		return "", errors.New("TOTP secret is null")
	}

	return r.encryptor.DecryptTOTPSecret(encryptedSecret.String)
}

func (r *TOTPRepository) DisableTOTP(userID string) error {
	query := `
		UPDATE users 
		SET totp_secret = NULL, totp_enabled = false, updated_at = NOW()
		WHERE id = $1
	`
	_, err := r.db.Exec(query, userID)
	return err
}

func (r *TOTPRepository) IsTOTPEnabled(userID string) (bool, error) {
	var enabled bool
	query := `SELECT totp_enabled FROM users WHERE id = $1`
	err := r.db.QueryRow(query, userID).Scan(&enabled)
	if err == sql.ErrNoRows {
		return false, errors.New("user not found")
	}
	return enabled, err
}

func (r *TOTPRepository) MarkTOTPVerified(userID string) error {
	query := `
		UPDATE users 
		SET totp_verified_at = $1, updated_at = NOW()
		WHERE id = $2
	`
	_, err := r.db.Exec(query, time.Now(), userID)
	return err
}
