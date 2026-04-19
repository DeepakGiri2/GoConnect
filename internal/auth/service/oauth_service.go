package service

import (
	"context"
	"encoding/json"
	"errors"
	"fmt"
	"io"
	"net/http"
	"time"

	"golang.org/x/oauth2"
	"golang.org/x/oauth2/facebook"
	"golang.org/x/oauth2/github"
	"golang.org/x/oauth2/google"

	"github.com/goconnect/internal/auth/repository"
	"github.com/goconnect/pkg/config"
	"github.com/goconnect/pkg/models"
	"github.com/goconnect/pkg/utils"
)

type OAuthService struct {
	userRepo      *repository.UserRepository
	oauthRepo     *repository.OAuthRepository
	tokenRepo     *repository.TokenRepository
	bloomFilter   *BloomFilterService
	jwtConfig     config.JWTConfig
	googleConfig  *oauth2.Config
	facebookConfig *oauth2.Config
	githubConfig  *oauth2.Config
}

func NewOAuthService(
	userRepo *repository.UserRepository,
	oauthRepo *repository.OAuthRepository,
	tokenRepo *repository.TokenRepository,
	bloomFilter *BloomFilterService,
	jwtConfig config.JWTConfig,
	oauthConfig config.OAuthConfig,
) *OAuthService {
	return &OAuthService{
		userRepo:    userRepo,
		oauthRepo:   oauthRepo,
		tokenRepo:   tokenRepo,
		bloomFilter: bloomFilter,
		jwtConfig:   jwtConfig,
		googleConfig: &oauth2.Config{
			ClientID:     oauthConfig.Google.ClientID,
			ClientSecret: oauthConfig.Google.ClientSecret,
			RedirectURL:  oauthConfig.Google.RedirectURL,
			Scopes:       []string{"https://www.googleapis.com/auth/userinfo.email", "https://www.googleapis.com/auth/userinfo.profile"},
			Endpoint:     google.Endpoint,
		},
		facebookConfig: &oauth2.Config{
			ClientID:     oauthConfig.Facebook.ClientID,
			ClientSecret: oauthConfig.Facebook.ClientSecret,
			RedirectURL:  oauthConfig.Facebook.RedirectURL,
			Scopes:       []string{"email", "public_profile"},
			Endpoint:     facebook.Endpoint,
		},
		githubConfig: &oauth2.Config{
			ClientID:     oauthConfig.GitHub.ClientID,
			ClientSecret: oauthConfig.GitHub.ClientSecret,
			RedirectURL:  oauthConfig.GitHub.RedirectURL,
			Scopes:       []string{"user:email"},
			Endpoint:     github.Endpoint,
		},
	}
}

func (s *OAuthService) GetAuthURL(provider, state string) (string, error) {
	var config *oauth2.Config
	
	switch provider {
	case "google":
		config = s.googleConfig
	case "facebook":
		config = s.facebookConfig
	case "github":
		config = s.githubConfig
	default:
		return "", errors.New("unsupported OAuth provider")
	}
	
	return config.AuthCodeURL(state, oauth2.AccessTypeOffline), nil
}

func (s *OAuthService) HandleOAuthCallback(ctx context.Context, provider, code string) (*models.User, *utils.TokenPair, error) {
	var config *oauth2.Config
	
	switch provider {
	case "google":
		config = s.googleConfig
	case "facebook":
		config = s.facebookConfig
	case "github":
		config = s.githubConfig
	default:
		return nil, nil, errors.New("unsupported OAuth provider")
	}
	
	token, err := config.Exchange(ctx, code)
	if err != nil {
		return nil, nil, fmt.Errorf("failed to exchange code: %w", err)
	}
	
	userInfo, err := s.fetchUserInfo(ctx, provider, token.AccessToken)
	if err != nil {
		return nil, nil, fmt.Errorf("failed to fetch user info: %w", err)
	}
	
	oauthAccount, err := s.oauthRepo.GetOAuthAccount(provider, userInfo.ProviderUserID)
	
	var user *models.User
	
	if err != nil {
		user, err = s.createUserFromOAuth(userInfo)
		if err != nil {
			return nil, nil, err
		}
		
		oauthAccount = &models.OAuthAccount{
			UserID:         user.ID,
			Provider:       provider,
			ProviderUserID: userInfo.ProviderUserID,
			AccessToken:    token.AccessToken,
			RefreshToken:   token.RefreshToken,
			ExpiresAt:      token.Expiry,
		}
		
		if err := s.oauthRepo.CreateOAuthAccount(oauthAccount); err != nil {
			return nil, nil, err
		}
	} else {
		user, err = s.userRepo.GetUserByID(oauthAccount.UserID)
		if err != nil {
			return nil, nil, err
		}
		
		oauthAccount.AccessToken = token.AccessToken
		oauthAccount.RefreshToken = token.RefreshToken
		oauthAccount.ExpiresAt = token.Expiry
		
		if err := s.oauthRepo.UpdateOAuthAccount(oauthAccount); err != nil {
			return nil, nil, err
		}
	}
	
	tokens, err := s.generateTokenPair(user)
	if err != nil {
		return nil, nil, err
	}
	
	return user, tokens, nil
}

type OAuthUserInfo struct {
	ProviderUserID string
	Email          string
	Username       string
}

func (s *OAuthService) fetchUserInfo(ctx context.Context, provider, accessToken string) (*OAuthUserInfo, error) {
	var url string
	
	switch provider {
	case "google":
		url = "https://www.googleapis.com/oauth2/v2/userinfo"
	case "facebook":
		url = "https://graph.facebook.com/me?fields=id,email,name"
	case "github":
		url = "https://api.github.com/user"
	default:
		return nil, errors.New("unsupported provider")
	}
	
	req, err := http.NewRequestWithContext(ctx, "GET", url, nil)
	if err != nil {
		return nil, err
	}
	
	req.Header.Set("Authorization", "Bearer "+accessToken)
	
	client := &http.Client{Timeout: 10 * time.Second}
	resp, err := client.Do(req)
	if err != nil {
		return nil, err
	}
	defer resp.Body.Close()
	
	if resp.StatusCode != http.StatusOK {
		return nil, fmt.Errorf("failed to fetch user info: status %d", resp.StatusCode)
	}
	
	body, err := io.ReadAll(resp.Body)
	if err != nil {
		return nil, err
	}
	
	return s.parseUserInfo(provider, body)
}

func (s *OAuthService) parseUserInfo(provider string, data []byte) (*OAuthUserInfo, error) {
	var result map[string]interface{}
	if err := json.Unmarshal(data, &result); err != nil {
		return nil, err
	}
	
	userInfo := &OAuthUserInfo{}
	
	switch provider {
	case "google":
		userInfo.ProviderUserID = result["id"].(string)
		userInfo.Email = result["email"].(string)
		if name, ok := result["name"].(string); ok {
			userInfo.Username = name
		}
	case "facebook":
		userInfo.ProviderUserID = result["id"].(string)
		if email, ok := result["email"].(string); ok {
			userInfo.Email = email
		}
		if name, ok := result["name"].(string); ok {
			userInfo.Username = name
		}
	case "github":
		userInfo.ProviderUserID = fmt.Sprintf("%v", result["id"])
		if email, ok := result["email"].(string); ok && email != "" {
			userInfo.Email = email
		}
		if login, ok := result["login"].(string); ok {
			userInfo.Username = login
		}
	}
	
	return userInfo, nil
}

func (s *OAuthService) createUserFromOAuth(userInfo *OAuthUserInfo) (*models.User, error) {
	username := s.generateUniqueUsername(userInfo.Username)
	
	user := &models.User{
		ID:         utils.GenerateGUID(),
		Username:   username,
		Email:      userInfo.Email,
		IsActive:   true,
		IsVerified: true,
	}
	
	if err := s.userRepo.CreateUser(user); err != nil {
		return nil, err
	}
	
	s.bloomFilter.AddUsername(username)
	
	return user, nil
}

func (s *OAuthService) generateUniqueUsername(baseUsername string) string {
	username := baseUsername
	counter := 1
	
	for {
		exists, _ := s.userRepo.UsernameExists(username)
		if !exists {
			return username
		}
		username = fmt.Sprintf("%s%d", baseUsername, counter)
		counter++
	}
}

func (s *OAuthService) generateTokenPair(user *models.User) (*utils.TokenPair, error) {
	accessToken, err := utils.GenerateAccessToken(
		user.ID,
		user.Username,
		user.Email,
		s.jwtConfig.Secret,
		s.jwtConfig.AccessExpiry,
	)
	if err != nil {
		return nil, err
	}
	
	refreshToken, err := utils.GenerateRefreshToken(
		user.ID,
		s.jwtConfig.Secret,
		s.jwtConfig.RefreshExpiry,
	)
	if err != nil {
		return nil, err
	}
	
	tokenRecord := &models.RefreshToken{
		UserID:    user.ID,
		Token:     refreshToken,
		ExpiresAt: time.Now().Add(s.jwtConfig.RefreshExpiry),
		IsRevoked: false,
	}
	
	if err := s.tokenRepo.CreateRefreshToken(tokenRecord); err != nil {
		return nil, err
	}
	
	return &utils.TokenPair{
		AccessToken:  accessToken,
		RefreshToken: refreshToken,
	}, nil
}
