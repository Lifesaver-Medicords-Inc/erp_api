package utils

import (
	"strconv"

	"github.com/gofiber/fiber/v2"
	"github.com/pierceperado/smpc/models"
)

func GetAtData(c *fiber.Ctx, at models.At) models.At {
	userSet := c.Locals("user")

	var userObj models.User
	if userSet != nil {
		userObj = userSet.(models.User)
	}

	method := c.Method()
	var action string

	switch method {
	case "POST":
		action = "INSERT"
	case "PUT":
		action = "UPDATE"
	case "PATCH":
		action = "UPDATE"
	case "DELETE":
		action = "DELETE"
	default:
		action = "UNLISTED"
	}

	clientIP := c.IP()

	atData := models.At{AtAction: action, IpAddress: clientIP, MotherboardSerialNo: at.MotherboardSerialNo, MachineName: at.MachineName, AtDate: GetCurrentDatetime(), AtUserId: strconv.Itoa(int(userObj.ID)), AtUser: userObj.EmployeeId}

	return atData
}
