package utils

import (
	"fmt"

	"golang.org/x/crypto/bcrypt"
)

func GenerateEmployeeId(department string, position string, id uint) string {
	employeeID := fmt.Sprintf("%v-%v-%v", department, GetStringIntials(position), id)

	return employeeID
}

func GenerateUserPassword(password string) (string, error) {
	var hashedpass string

	hash, err := bcrypt.GenerateFromPassword([]byte(password), 10)
	if err != nil {
		return hashedpass, err
	}

	hashedpass = string(hash)

	return hashedpass, nil
}
