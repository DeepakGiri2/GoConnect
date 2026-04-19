package utils

import (
	"crypto/hmac"
	"crypto/sha256"
	"encoding/binary"
	"fmt"
	"math"
	"time"
)

const (
	otpDigits     = 6
	otpTimeWindow = 30
)

func GenerateOTP(userID, secret string) (string, error) {
	timeCounter := time.Now().Unix() / otpTimeWindow
	return generateTOTP(userID, secret, timeCounter)
}

func ValidateOTP(userID, otp, secret string) bool {
	currentTime := time.Now().Unix() / otpTimeWindow
	
	for i := -1; i <= 1; i++ {
		timeCounter := currentTime + int64(i)
		generatedOTP, err := generateTOTP(userID, secret, timeCounter)
		if err != nil {
			return false
		}
		if otp == generatedOTP {
			return true
		}
	}
	
	return false
}

func generateTOTP(userID, secret string, timeCounter int64) (string, error) {
	message := fmt.Sprintf("%s:%d", userID, timeCounter)
	
	h := hmac.New(sha256.New, []byte(secret))
	h.Write([]byte(message))
	hash := h.Sum(nil)
	
	offset := hash[len(hash)-1] & 0x0F
	
	truncatedHash := binary.BigEndian.Uint32(hash[offset:offset+4]) & 0x7FFFFFFF
	
	otp := truncatedHash % uint32(math.Pow10(otpDigits))
	
	return fmt.Sprintf("%0*d", otpDigits, otp), nil
}
