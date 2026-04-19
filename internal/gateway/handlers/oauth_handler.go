package handlers

import (
	"context"
	"crypto/rand"
	"encoding/base64"
	"net/http"

	"github.com/gin-gonic/gin"
	pb "github.com/goconnect/api/shared/proto_gen/api/shared/proto"
)

type OAuthHandler struct {
	authClient pb.AuthServiceClient
}

func NewOAuthHandler(authClient pb.AuthServiceClient) *OAuthHandler {
	return &OAuthHandler{authClient: authClient}
}

func (h *OAuthHandler) InitiateOAuth(c *gin.Context) {
	provider := c.Param("provider")
	
	if provider != "google" && provider != "facebook" && provider != "github" {
		c.JSON(http.StatusBadRequest, gin.H{"error": "unsupported provider"})
		return
	}

	state := generateState()
	c.SetCookie("oauth_state", state, 600, "/", "", false, true)

	var authURL string
	switch provider {
	case "google":
		authURL = "https://accounts.google.com/o/oauth2/v2/auth"
	case "facebook":
		authURL = "https://www.facebook.com/v12.0/dialog/oauth"
	case "github":
		authURL = "https://github.com/login/oauth/authorize"
	}

	c.JSON(http.StatusOK, gin.H{
		"auth_url": authURL,
		"state":    state,
	})
}

func (h *OAuthHandler) OAuthCallback(c *gin.Context) {
	provider := c.Param("provider")
	code := c.Query("code")
	state := c.Query("state")

	savedState, err := c.Cookie("oauth_state")
	if err != nil || savedState != state {
		c.JSON(http.StatusBadRequest, gin.H{"error": "invalid state parameter"})
		return
	}

	if code == "" {
		c.JSON(http.StatusBadRequest, gin.H{"error": "authorization code required"})
		return
	}

	resp, err := h.authClient.OAuthLogin(context.Background(), &pb.OAuthLoginRequest{
		Provider: provider,
		Code:     code,
	})

	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "internal server error"})
		return
	}

	if !resp.Success {
		c.JSON(http.StatusBadRequest, gin.H{"error": resp.Message})
		return
	}

	c.JSON(http.StatusOK, gin.H{
		"success": true,
		"message": resp.Message,
		"user": gin.H{
			"id":       resp.UserId,
			"username": resp.Username,
			"email":    resp.Email,
		},
		"access_token":  resp.AccessToken,
		"refresh_token": resp.RefreshToken,
	})
}

func generateState() string {
	b := make([]byte, 32)
	rand.Read(b)
	return base64.URLEncoding.EncodeToString(b)
}
