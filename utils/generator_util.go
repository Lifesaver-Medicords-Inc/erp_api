package utils

import (
	"fmt"
)

func GenerateEmployeeId(department string, position string, id uint) string {
	employeeID := fmt.Sprintf("%v-%v-%v", department, GetStringIntials(position), id)

	return employeeID
}
