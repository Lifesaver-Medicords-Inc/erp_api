package public_services

import (
	"errors"
	"fmt"
	"strconv"
	"time"

	"github.com/gofiber/fiber/v2"
	"github.com/pierceperado/smpc/initializers"
	"github.com/pierceperado/smpc/models"
	"github.com/pierceperado/smpc/services"
	"github.com/pierceperado/smpc/utils"
	"gorm.io/gorm"
)

func CreateAccount(c *fiber.Ctx, tx *gorm.DB) (models.User, int, error) {
	var user models.User

	var body models.UserAt

	if err := c.BodyParser(&body); err != nil {
		return user, fiber.StatusBadRequest, errors.New("cannot bind request")
	}

	user = models.User{UserContent: body.UserContent}
	if err := services.DbInsert(tx, &user); err != nil {
		fmt.Println("Error:", err)
		return user, fiber.StatusInternalServerError, errors.New("failed creating user")
	}

	// Read the position's NAME from the database rather than off user.Position.
	// That association is never loaded here - the user was built from the
	// request body and inserted, and GORM does not populate a relation it was
	// not asked to preload - so user.Position.Name was always "", which is why
	// every existing account reads "Admin--2" / "Warehouse--3" with an empty
	// middle segment. Since the employee id doubles as the account's initial
	// password just below, that defect also made every initial credential
	// malformed and predictable.
	var positionName string
	if user.PositionId != 0 {
		var positions []models.PositionModel
		// Limit(1).Find, not First: a user whose position was deleted is a
		// recoverable case (the id simply omits the segment), not an error
		// worth logging red on every account creation.
		if err := tx.Where("id = ?", user.PositionId).Limit(1).Find(&positions).Error; err == nil && len(positions) > 0 {
			positionName = positions[0].Name
		}
	}

	employeeId := utils.GenerateEmployeeId(body.Department, positionName, user.ID)
	user.EmployeeId = employeeId

	password, err := utils.GenerateUserPassword(employeeId)
	if err != nil {
		return user, fiber.StatusInternalServerError, errors.New("failed generating password")
	}
	user.Password = password

	if err := services.DbUpdate(tx, &user, nil); err != nil {
		return user, fiber.StatusInternalServerError, errors.New("failed udpating user")
	}

	body.RefId = user.ID
	body.EmployeeId = employeeId
	body.Password = password
	body.At = utils.GetAtData(c, body.At)

	if err := services.DbInsert(tx, &body); err != nil {
		return user, fiber.StatusInternalServerError, errors.New("failed creating userat")
	}

	return user, 0, nil
}

func LoginAccount(c *fiber.Ctx) (models.User, int, error) {
	var user models.User

	var body models.UserAt
	if err := c.BodyParser(&body); err != nil {
		return user, fiber.StatusBadRequest, errors.New("cannot bind request")
	}

	conditions := map[string]interface{}{
		"employee_id": body.EmployeeId,
	}

	if err := initializers.DB.
		// Position.Access carries the tbl_position_access codes. Preloading only
		// "Position" left CurrentUser.position.access empty in every client, so no app
		// could gate a button on a granted code even though the whole catalog/grant
		// mechanism existed - the Sales Order approve/cancel strip (spec 3.3) is the
		// first thing to actually need it client-side.
		Preload("Position").
		Preload("Position.Access").
		Where(conditions).
		First(&user).Error; err != nil {
		return user, fiber.StatusUnauthorized, errors.New("Invalid user employee id")
	}

	if err := utils.CompareUserPassword(user.Password, body.Password); err != nil {
		return user, fiber.StatusUnauthorized, errors.New("Invalid user password")
	}

	body.AtUserId = strconv.Itoa(int(user.ID))
	if err := utils.CreateAuthToken(c, body.At, user.ID); err != nil {
		return user, fiber.StatusUnauthorized, err
	}
	fmt.Println("LOGIN:", user.Position.Name)

	return user, 0, nil
}

func LogoutAccount(c *fiber.Ctx) {
	c.Cookie(&fiber.Cookie{
		Name:     "Authorization",
		Value:    "",
		Expires:  time.Now().Add(-time.Hour),
		SameSite: fiber.CookieSameSiteLaxMode,
		HTTPOnly: true,
		Secure:   true,
	})
}
