package utils

import (
	"github.com/google/uuid"
)

func GenerateGUID() string {
	return uuid.New().String()
}

func IsValidGUID(guid string) bool {
	_, err := uuid.Parse(guid)
	return err == nil
}
