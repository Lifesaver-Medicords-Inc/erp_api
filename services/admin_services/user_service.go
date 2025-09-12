package adminservices

import (
	"errors"

	"github.com/gofiber/fiber/v2"
	"github.com/pierceperado/smpc/initializers"
	"github.com/pierceperado/smpc/models"
	"github.com/pierceperado/smpc/services"
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

	return users, 0, nil
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

	return user, 0, nil
}

func (u *UserService) UpdateUserService(user *models.User, conditions map[string]interface{}, at models.At) (*models.User, int, error) {

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
		Password:   user.Password,
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

	return user, 0, nil
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

	return user, 0, nil
}
