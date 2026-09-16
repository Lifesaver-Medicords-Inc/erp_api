package adminservices

import (
	"errors"
	"fmt"
	"strings"

	"github.com/gofiber/fiber/v2"
	"github.com/pierceperado/smpc/initializers"
	"github.com/pierceperado/smpc/models"
	"github.com/pierceperado/smpc/services"
	"github.com/pierceperado/smpc/utils"
)

type UserService struct {
}

func NewUserService() *UserService {
	return &UserService{}
}

func (u *UserService) GetUsersService(conditions map[string]interface{}) (*[]models.User, int, error) {
	tx := initializers.DB.Begin()

	var users = &[]models.User{}

	if tx.Error != nil {
		return users, fiber.StatusInternalServerError, errors.New("failed to start DB transaction")
	}

	if err := tx.Where(conditions).Preload("Permissions").Preload("Position").Find(users).Error; err != nil {
		return users, fiber.StatusNotFound, errors.New("failed getting users")
	}

	for i := range *users {
		(*users)[i].Password = ""
	}

	return users, fiber.StatusOK, nil
}

func (u *UserService) GetUserService(conditions map[string]interface{}) (*models.User, int, error) {
	var user = &models.User{}

	tx := initializers.DB.Begin()

	if tx.Error != nil {
		return user, fiber.StatusInternalServerError, errors.New("failed to start DB transaction")
	}

	if err := tx.Where(conditions).Preload("Permissions").Preload("Position").First(user).Error; err != nil {
		return user, fiber.StatusNotFound, errors.New("failed getting users")
	}

	user.Password = ""

	return user, fiber.StatusOK, nil
}

func (u *UserService) UpdateUserService(user *models.User, conditions map[string]interface{}, at models.At) (*models.User, int, error) {
	// A password is never written through here. It used to be: whatever "password" the
	// body carried went into tbl_setup_users as-is - unhashed, so that user could no
	// longer log in - with a plain copy in the audit table. ChangePasswordService is the
	// one way to set a password now, and it hashes. Cleared here rather than in the
	// handlers so every caller is covered; DbUpdate skips empty fields, so the stored
	// hash is left alone.
	user.Password = ""

	tx := initializers.DB.Begin()

	if tx.Error != nil {
		return &models.User{}, fiber.StatusInternalServerError, errors.New("failed to start DB transaction")
	}

	if err := services.DbUpdate(tx, &user, conditions); err != nil {
		tx.Rollback()
		return user, fiber.StatusInternalServerError, errors.New("failed updating user")
	}

	atdata := models.UserAt{RefId: user.ID, UserContent: models.UserContent{
		FirstName:  user.FirstName,
		LastName:   user.LastName,
		PositionId: user.PositionId,
	}, At: at}

	if err := services.DbInsert(tx, &atdata); err != nil {
		tx.Rollback()
		return user, fiber.StatusInternalServerError, errors.New("failed creating userat")
	}

	if err := tx.Commit().Error; err != nil {
		tx.Rollback()
		return user, fiber.StatusInternalServerError, errors.New("failed to commit transaction")
	}

	return user, fiber.StatusOK, nil
}

func (u *UserService) DeleteUserService(conditions map[string]interface{}, at models.At) (*models.User, int, error) {
	tx := initializers.DB.Begin()

	if tx.Error != nil {
		return &models.User{}, fiber.StatusInternalServerError, errors.New("failed to start DB transaction")
	}

	user, _, err := u.GetUserService(conditions)

	if err != nil {
		return user, fiber.StatusInternalServerError, errors.New("user not found")
	}

	if err := services.DbDelete(tx, &user, conditions); err != nil {
		tx.Rollback()
		return user, fiber.StatusInternalServerError, errors.New("failed deleting user")
	}

	atdata := models.UserAt{RefId: user.ID, UserContent: user.UserContent, At: at}

	if err := services.DbInsert(tx, &atdata); err != nil {
		tx.Rollback()
		return user, fiber.StatusInternalServerError, errors.New("failed creating userat")
	}

	if err := tx.Commit().Error; err != nil {
		tx.Rollback()
		return user, fiber.StatusInternalServerError, errors.New("failed to commit transaction")
	}

	return user, fiber.StatusOK, nil
}

// MinPasswordLength is the shortest password ChangePasswordService accepts.
const MinPasswordLength = 8

// bcrypt reads at most 72 bytes, and golang.org/x/crypto refuses a longer password
// rather than silently ignoring the rest of it.
const maxPasswordBytes = 72

// ValidateNewPassword holds the rules for a password an administrator sets. It may not be
// the employee id: that is every account's initial password, so resetting to it would
// undo the reset. Leading or trailing spaces are refused because they are almost always a
// copy-paste accident that would leave the user unable to type their own password.
func ValidateNewPassword(password string, employeeId string) error {
	if len([]rune(password)) < MinPasswordLength {
		return fmt.Errorf("password must be at least %d characters", MinPasswordLength)
	}
	if len(password) > maxPasswordBytes {
		return fmt.Errorf("password must be at most %d bytes", maxPasswordBytes)
	}
	if strings.TrimSpace(password) != password {
		return errors.New("password must not start or end with a space")
	}
	if strings.EqualFold(password, strings.TrimSpace(employeeId)) {
		return errors.New("password must not be the employee id")
	}
	return nil
}

// ChangePasswordService sets a user's password on an administrator's say-so, from the
// Admin app's User Details. The route is gated on ADMIN USERS. The user's existing
// sessions are not ended by it - tokens are revoked one at a time, at logout.
func (u *UserService) ChangePasswordService(userId uint, password string, at models.At) (int, error) {
	var user models.User
	if err := initializers.DB.Where("id = ?", userId).First(&user).Error; err != nil {
		return fiber.StatusNotFound, errors.New("user not found")
	}

	if err := ValidateNewPassword(password, user.EmployeeId); err != nil {
		return fiber.StatusBadRequest, err
	}

	hash, err := utils.GenerateUserPassword(password)
	if err != nil {
		return fiber.StatusInternalServerError, errors.New("failed securing password")
	}

	tx := initializers.DB.Begin()
	if tx.Error != nil {
		return fiber.StatusInternalServerError, errors.New("failed to start DB transaction")
	}

	conditions := map[string]interface{}{"id": user.ID}
	if err := services.DbUpdate(tx, &models.User{UserContent: models.UserContent{Password: hash}}, conditions); err != nil {
		tx.Rollback()
		return fiber.StatusInternalServerError, errors.New("failed changing password")
	}

	// The audit row records that the password changed, by whom and when - never the
	// password or its hash.
	atdata := models.UserAt{RefId: user.ID, EmployeeId: user.EmployeeId, UserContent: models.UserContent{
		FirstName:  user.FirstName,
		LastName:   user.LastName,
		PositionId: user.PositionId,
		Department: user.Department,
	}, At: at}

	if err := services.DbInsert(tx, &atdata); err != nil {
		tx.Rollback()
		return fiber.StatusInternalServerError, errors.New("failed creating userat")
	}

	if err := tx.Commit().Error; err != nil {
		tx.Rollback()
		return fiber.StatusInternalServerError, errors.New("failed to commit transaction")
	}

	return fiber.StatusOK, nil
}
