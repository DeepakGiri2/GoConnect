package utils

import (
	"fmt"
	
	"github.com/pquerna/otp"
	"github.com/pquerna/otp/totp"
)

type TOTPSetupData struct {
	Secret      string `json:"secret"`
	QRCodeURL   string `json:"qr_code_url"`
	QRCode      string `json:"qr_code"`
	Issuer      string `json:"issuer"`
	AccountName string `json:"account_name"`
}

func GenerateTOTPSecret(issuer, accountName string) (*TOTPSetupData, error) {
	key, err := totp.Generate(totp.GenerateOpts{
		Issuer:      issuer,
		AccountName: accountName,
		Period:      30,
		Digits:      otp.DigitsSix,
		Algorithm:   otp.AlgorithmSHA1,
	})
	if err != nil {
		return nil, fmt.Errorf("failed to generate TOTP key: %w", err)
	}

	secret := key.Secret()
	qrURL := key.URL()

	return &TOTPSetupData{
		Secret:      secret,
		QRCodeURL:   qrURL,
		QRCode:      qrURL, // URL can be used to generate QR code on client
		Issuer:      issuer,
		AccountName: accountName,
	}, nil
}

func VerifyTOTPCode(secret, code string) bool {
	return totp.Validate(code, secret)
}
