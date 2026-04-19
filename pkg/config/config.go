package config

import (
	"os"
	"time"

	"github.com/spf13/viper"
)

type Config struct {
	Database    DatabaseConfig
	Redis       RedisConfig
	JWT         JWTConfig
	OTP         OTPConfig
	RateLimit   RateLimitConfig
	OAuth       OAuthConfig
	Server      ServerConfig
	AuthService AuthServiceConfig
	Email       EmailConfig
	Retry       RetryConfig
}

type DatabaseConfig struct {
	Host            string
	Port            int
	Name            string
	User            string
	Password        string
	MaxOpenConns    int
	MaxIdleConns    int
	ConnMaxLifetime time.Duration
	ConnMaxIdleTime time.Duration
}

type RedisConfig struct {
	Host     string
	Port     int
	Password string
}

type JWTConfig struct {
	Secret        string
	AccessExpiry  time.Duration
	RefreshExpiry time.Duration
}

type OTPConfig struct {
	Secret            string
	Expiry            time.Duration
	Length            int
	MaxVerifyAttempts int
	MaxResendAttempts int
	ResendCooldown    time.Duration
	BlockDuration     time.Duration
	EncryptionKey     string
}

type RateLimitConfig struct {
	RegistrationIP        int
	RegistrationEmail     int
	UsernameCheck         int
	PendingRegTTL         time.Duration
	UnverifiedUserCleanup time.Duration
}

type OAuthConfig struct {
	Google   OAuthProvider
	Facebook OAuthProvider
	GitHub   OAuthProvider
}

type OAuthProvider struct {
	ClientID     string
	ClientSecret string
	RedirectURL  string
}

type ServerConfig struct {
	Port     int
	Host     string
	LogLevel string
	Env      string
}

type AuthServiceConfig struct {
	Host string
	Port int
}

type EmailConfig struct {
	SMTPHost     string
	SMTPPort     string
	SMTPUsername string
	SMTPPassword string
	FromEmail    string
	FromName     string
}

type RetryConfig struct {
	MaxRetries     int
	InitialBackoff time.Duration
	MaxBackoff     time.Duration
}

func LoadConfig() (*Config, error) {
	viper.SetConfigFile(".env")
	viper.AutomaticEnv()

	// Try to read .env file, but don't fail if it doesn't exist
	// Environment variables can be loaded via docker-compose or system env
	if err := viper.ReadInConfig(); err != nil {
		// File not found is acceptable - we'll use environment variables
		if _, ok := err.(viper.ConfigFileNotFoundError); !ok && !os.IsNotExist(err) {
			// Return error only if it's not a "file not found" error
			return nil, err
		}
	}

	viper.SetDefault("DB_MAX_OPEN_CONNS", 25)
	viper.SetDefault("DB_MAX_IDLE_CONNS", 10)
	viper.SetDefault("DB_CONN_MAX_LIFETIME", 5*time.Minute)
	viper.SetDefault("DB_CONN_MAX_IDLE_TIME", 10*time.Minute)
	viper.SetDefault("RETRY_MAX_RETRIES", 3)
	viper.SetDefault("RETRY_INITIAL_BACKOFF", 100*time.Millisecond)
	viper.SetDefault("RETRY_MAX_BACKOFF", 5*time.Second)

	config := &Config{
		Database: DatabaseConfig{
			Host:            viper.GetString("DATABASE_HOST"),
			Port:            viper.GetInt("DATABASE_PORT"),
			Name:            viper.GetString("DATABASE_NAME"),
			User:            viper.GetString("DATABASE_USER"),
			Password:        viper.GetString("DATABASE_PASSWORD"),
			MaxOpenConns:    viper.GetInt("DB_MAX_OPEN_CONNS"),
			MaxIdleConns:    viper.GetInt("DB_MAX_IDLE_CONNS"),
			ConnMaxLifetime: viper.GetDuration("DB_CONN_MAX_LIFETIME"),
			ConnMaxIdleTime: viper.GetDuration("DB_CONN_MAX_IDLE_TIME"),
		},
		Redis: RedisConfig{
			Host:     viper.GetString("REDIS_HOST"),
			Port:     viper.GetInt("REDIS_PORT"),
			Password: viper.GetString("REDIS_PASSWORD"),
		},
		JWT: JWTConfig{
			Secret:        viper.GetString("JWT_SECRET"),
			AccessExpiry:  viper.GetDuration("JWT_ACCESS_EXPIRY"),
			RefreshExpiry: viper.GetDuration("JWT_REFRESH_EXPIRY"),
		},
		OTP: OTPConfig{
			Secret:            viper.GetString("OTP_SECRET"),
			Expiry:            viper.GetDuration("OTP_EXPIRY"),
			Length:            viper.GetInt("OTP_LENGTH"),
			MaxVerifyAttempts: viper.GetInt("OTP_MAX_VERIFY_ATTEMPTS"),
			MaxResendAttempts: viper.GetInt("OTP_MAX_RESEND_ATTEMPTS"),
			ResendCooldown:    viper.GetDuration("OTP_RESEND_COOLDOWN"),
			BlockDuration:     viper.GetDuration("OTP_BLOCK_DURATION"),
			EncryptionKey:     viper.GetString("TOTP_ENCRYPTION_KEY"),
		},
		RateLimit: RateLimitConfig{
			RegistrationIP:        viper.GetInt("RATE_LIMIT_REGISTRATION_IP"),
			RegistrationEmail:     viper.GetInt("RATE_LIMIT_REGISTRATION_EMAIL"),
			UsernameCheck:         viper.GetInt("RATE_LIMIT_USERNAME_CHECK"),
			PendingRegTTL:         viper.GetDuration("PENDING_REGISTRATION_TTL"),
			UnverifiedUserCleanup: viper.GetDuration("UNVERIFIED_USER_CLEANUP_TTL"),
		},
		OAuth: OAuthConfig{
			Google: OAuthProvider{
				ClientID:     viper.GetString("GOOGLE_CLIENT_ID"),
				ClientSecret: viper.GetString("GOOGLE_CLIENT_SECRET"),
				RedirectURL:  viper.GetString("GOOGLE_REDIRECT_URL"),
			},
			Facebook: OAuthProvider{
				ClientID:     viper.GetString("FACEBOOK_CLIENT_ID"),
				ClientSecret: viper.GetString("FACEBOOK_CLIENT_SECRET"),
				RedirectURL:  viper.GetString("FACEBOOK_REDIRECT_URL"),
			},
			GitHub: OAuthProvider{
				ClientID:     viper.GetString("GITHUB_CLIENT_ID"),
				ClientSecret: viper.GetString("GITHUB_CLIENT_SECRET"),
				RedirectURL:  viper.GetString("GITHUB_REDIRECT_URL"),
			},
		},
		Server: ServerConfig{
			Host:     viper.GetString("SERVER_HOST"),
			Port:     viper.GetInt("SERVER_PORT"),
			LogLevel: viper.GetString("LOG_LEVEL"),
			Env:      viper.GetString("ENV"),
		},
		AuthService: AuthServiceConfig{
			Host: viper.GetString("AUTH_SERVICE_HOST"),
			Port: viper.GetInt("AUTH_SERVICE_PORT"),
		},
		Email: EmailConfig{
			SMTPHost:     viper.GetString("SMTP_HOST"),
			SMTPPort:     viper.GetString("SMTP_PORT"),
			SMTPUsername: viper.GetString("SMTP_USERNAME"),
			SMTPPassword: viper.GetString("SMTP_PASSWORD"),
			FromEmail:    viper.GetString("FROM_EMAIL"),
			FromName:     viper.GetString("FROM_NAME"),
		},
		Retry: RetryConfig{
			MaxRetries:     viper.GetInt("RETRY_MAX_RETRIES"),
			InitialBackoff: viper.GetDuration("RETRY_INITIAL_BACKOFF"),
			MaxBackoff:     viper.GetDuration("RETRY_MAX_BACKOFF"),
		},
	}

	return config, nil
}
