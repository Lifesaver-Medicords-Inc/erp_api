package utils

import (
	"golang.org/x/crypto/bcrypt"
)

func CompareUserPassword(userPassword string, requestPassword string) error {
	return bcrypt.CompareHashAndPassword([]byte(userPassword), []byte(requestPassword))
}
