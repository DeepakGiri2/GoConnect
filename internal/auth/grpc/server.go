package grpc

import (
	"context"

	pb "github.com/goconnect/api/shared/proto_gen/api/shared/proto"
	"github.com/goconnect/internal/auth/service"
)

type AuthGRPCServer struct {
	pb.UnimplementedAuthServiceServer
	authService  *service.AuthService
	oauthService *service.OAuthService
	bloomFilter  *service.BloomFilterService
	otpService   interface {
		GetBlockStatus(ctx context.Context, email string) (bool, int64, int, error)
	}
}

func NewAuthGRPCServer(
	authService *service.AuthService,
	oauthService *service.OAuthService,
	bloomFilter *service.BloomFilterService,
	otpService interface {
		GetBlockStatus(ctx context.Context, email string) (bool, int64, int, error)
	},
) *AuthGRPCServer {
	return &AuthGRPCServer{
		authService:  authService,
		oauthService: oauthService,
		bloomFilter:  bloomFilter,
		otpService:   otpService,
	}
}

func (s *AuthGRPCServer) Register(ctx context.Context, req *pb.RegisterRequest) (*pb.RegisterResponse, error) {
	regResp, err := s.authService.Register(ctx, req.Username, req.Email, req.Password)
	if err != nil {
		return &pb.RegisterResponse{
			Success: false,
			Message: err.Error(),
		}, nil
	}

	return &pb.RegisterResponse{
		Success: regResp.Success,
		Message: regResp.Message,
		Email:   regResp.Email,
	}, nil
}

func (s *AuthGRPCServer) Login(ctx context.Context, req *pb.LoginRequest) (*pb.LoginResponse, error) {
	loginResp, err := s.authService.Login(req.Username, req.Password)
	if err != nil {
		return &pb.LoginResponse{
			Success: false,
			Message: err.Error(),
		}, nil
	}

	// User not verified
	if !loginResp.Success {
		return &pb.LoginResponse{
			Success:    false,
			Message:    loginResp.Message,
			IsVerified: false,
		}, nil
	}

	// TOTP required
	if loginResp.RequiresTOTP {
		return &pb.LoginResponse{
			Success:      true,
			Message:      "TOTP verification required",
			UserId:       loginResp.User.ID,
			RequiresTotp: true,
			IsVerified:   true,
		}, nil
	}

	return &pb.LoginResponse{
		Success:      true,
		Message:      "login successful",
		UserId:       loginResp.User.ID,
		Username:     loginResp.User.Username,
		Email:        loginResp.User.Email,
		AccessToken:  loginResp.Tokens.AccessToken,
		RefreshToken: loginResp.Tokens.RefreshToken,
		RequiresTotp: false,
		IsVerified:   true,
	}, nil
}

func (s *AuthGRPCServer) ValidateToken(ctx context.Context, req *pb.ValidateTokenRequest) (*pb.ValidateTokenResponse, error) {
	claims, err := s.authService.ValidateAccessToken(req.Token)
	if err != nil {
		return &pb.ValidateTokenResponse{
			Valid: false,
		}, nil
	}

	return &pb.ValidateTokenResponse{
		Valid:    true,
		UserId:   claims.UserID,
		Username: claims.Username,
		Email:    claims.Email,
	}, nil
}

func (s *AuthGRPCServer) RefreshToken(ctx context.Context, req *pb.RefreshTokenRequest) (*pb.RefreshTokenResponse, error) {
	tokens, err := s.authService.RefreshTokens(req.RefreshToken)
	if err != nil {
		return &pb.RefreshTokenResponse{
			Success: false,
		}, nil
	}

	return &pb.RefreshTokenResponse{
		Success:      true,
		AccessToken:  tokens.AccessToken,
		RefreshToken: tokens.RefreshToken,
	}, nil
}

func (s *AuthGRPCServer) GenerateOTP(ctx context.Context, req *pb.GenerateOTPRequest) (*pb.GenerateOTPResponse, error) {
	err := s.authService.GenerateOTP(req.Email)
	if err != nil {
		return &pb.GenerateOTPResponse{
			Success: false,
			Message: err.Error(),
		}, nil
	}

	return &pb.GenerateOTPResponse{
		Success: true,
		Message: "OTP sent successfully",
	}, nil
}

func (s *AuthGRPCServer) VerifyOTP(ctx context.Context, req *pb.VerifyOTPRequest) (*pb.VerifyOTPResponse, error) {
	user, err := s.authService.VerifyOTP(req.Email, req.Otp)
	if err != nil {
		return &pb.VerifyOTPResponse{
			Success: false,
			Message: err.Error(),
		}, nil
	}

	return &pb.VerifyOTPResponse{
		Success: true,
		Message: "OTP verified successfully",
		UserId:  user.ID,
	}, nil
}

func (s *AuthGRPCServer) ResetPassword(ctx context.Context, req *pb.ResetPasswordRequest) (*pb.ResetPasswordResponse, error) {
	err := s.authService.ResetPassword(req.Email, req.Otp, req.NewPassword)
	if err != nil {
		return &pb.ResetPasswordResponse{
			Success: false,
			Message: err.Error(),
		}, nil
	}

	return &pb.ResetPasswordResponse{
		Success: true,
		Message: "password reset successful",
	}, nil
}

func (s *AuthGRPCServer) OAuthLogin(ctx context.Context, req *pb.OAuthLoginRequest) (*pb.OAuthLoginResponse, error) {
	user, tokens, err := s.oauthService.HandleOAuthCallback(ctx, req.Provider, req.Code)
	if err != nil {
		return &pb.OAuthLoginResponse{
			Success: false,
			Message: err.Error(),
		}, nil
	}

	return &pb.OAuthLoginResponse{
		Success:      true,
		Message:      "OAuth login successful",
		UserId:       user.ID,
		Username:     user.Username,
		Email:        user.Email,
		AccessToken:  tokens.AccessToken,
		RefreshToken: tokens.RefreshToken,
	}, nil
}

func (s *AuthGRPCServer) CheckUsernameAvailability(ctx context.Context, req *pb.CheckUsernameRequest) (*pb.CheckUsernameResponse, error) {
	available, err := s.bloomFilter.CheckUsernameAvailability(ctx, req.Username)
	if err != nil {
		return &pb.CheckUsernameResponse{
			Available: false,
		}, nil
	}

	return &pb.CheckUsernameResponse{
		Available: available,
	}, nil
}

// VerifyEmail verifies the email with OTP and creates the user account
func (s *AuthGRPCServer) VerifyEmail(ctx context.Context, req *pb.VerifyEmailRequest) (*pb.VerifyEmailResponse, error) {
	verifyResp, err := s.authService.VerifyEmailAndCreateUser(ctx, req.Email, req.Otp)
	if err != nil {
		return &pb.VerifyEmailResponse{
			Success: false,
			Message: err.Error(),
		}, nil
	}

	return &pb.VerifyEmailResponse{
		Success:      verifyResp.Success,
		Message:      verifyResp.Message,
		UserId:       verifyResp.User.ID,
		Username:     verifyResp.User.Username,
		Email:        verifyResp.User.Email,
		AccessToken:  verifyResp.Tokens.AccessToken,
		RefreshToken: verifyResp.Tokens.RefreshToken,
	}, nil
}

// ResendVerificationOTP resends the email verification OTP
func (s *AuthGRPCServer) ResendVerificationOTP(ctx context.Context, req *pb.ResendOTPRequest) (*pb.ResendOTPResponse, error) {
	err := s.authService.ResendVerificationOTP(ctx, req.Email)
	if err != nil {
		return &pb.ResendOTPResponse{
			Success: false,
			Message: err.Error(),
		}, nil
	}

	return &pb.ResendOTPResponse{
		Success: true,
		Message: "Verification OTP resent successfully. Please check your email.",
	}, nil
}

// GetBlockStatus checks if an email is blocked and returns remaining time/attempts
func (s *AuthGRPCServer) GetBlockStatus(ctx context.Context, req *pb.GetBlockStatusRequest) (*pb.GetBlockStatusResponse, error) {
	isBlocked, remainingSeconds, remainingAttempts, err := s.otpService.GetBlockStatus(ctx, req.Email)
	if err != nil {
		return &pb.GetBlockStatusResponse{
			IsBlocked:         false,
			RemainingSeconds:  0,
			RemainingAttempts: 3,
		}, nil
	}

	return &pb.GetBlockStatusResponse{
		IsBlocked:         isBlocked,
		RemainingSeconds:  remainingSeconds,
		RemainingAttempts: int32(remainingAttempts),
	}, nil
}

// SetupTOTP sets up TOTP for a verified user (optional post-email-verification step)
func (s *AuthGRPCServer) SetupTOTP(ctx context.Context, req *pb.SetupTOTPRequest) (*pb.SetupTOTPResponse, error) {
	totpSetup, err := s.authService.SetupTOTP(req.UserId)
	if err != nil {
		return &pb.SetupTOTPResponse{
			Success: false,
			Message: err.Error(),
		}, nil
	}

	return &pb.SetupTOTPResponse{
		Success:     true,
		Message:     "TOTP setup successful. Scan the QR code with your authenticator app.",
		Secret:      totpSetup.Secret,
		QrCode:      totpSetup.QRCode,
		Issuer:      totpSetup.Issuer,
		AccountName: totpSetup.AccountName,
	}, nil
}

// VerifyTOTP verifies TOTP code during login and issues tokens
func (s *AuthGRPCServer) VerifyTOTP(ctx context.Context, req *pb.VerifyTOTPRequest) (*pb.VerifyTOTPResponse, error) {
	tokens, err := s.authService.VerifyTOTPAndLogin(req.Username, req.TotpCode)
	if err != nil {
		return &pb.VerifyTOTPResponse{
			Success: false,
			Message: err.Error(),
		}, nil
	}

	return &pb.VerifyTOTPResponse{
		Success:      true,
		Message:      "TOTP verification successful",
		AccessToken:  tokens.AccessToken,
		RefreshToken: tokens.RefreshToken,
	}, nil
}
