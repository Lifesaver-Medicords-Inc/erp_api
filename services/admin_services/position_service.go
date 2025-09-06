package adminservices

import (
	"errors"
	"fmt"
	"strings"

	"github.com/gofiber/fiber/v2"
	"github.com/pierceperado/smpc/initializers"
	"github.com/pierceperado/smpc/models"
	"github.com/pierceperado/smpc/services"
)

func GetPositions(conditions map[string]interface{}) ([]models.Position, int, error) {

	tx := initializers.DB.Begin()

	if tx.Error != nil {
		return []models.Position{}, fiber.StatusInternalServerError, errors.New("failed to start DB transaction")
	}

	var positions []models.Position

	if err := tx.Where(conditions).Preload("Access").Find(&positions).Error; err != nil {
		fmt.Println("ERROR:", err)
		return positions, fiber.StatusNotFound, errors.New("failed getting positions")
	}
	return positions, 0, nil
}

func GetPosition(conditions map[string]interface{}) (models.Position, int, error) {
	tx := initializers.DB.Begin()

	if tx.Error != nil {
		return models.Position{}, fiber.StatusInternalServerError, errors.New("failed to start DB transaction")
	}
	var position models.Position

	if err := tx.Where(conditions).Preload("Access").First(&position).Error; err != nil {
		return position, fiber.StatusNotFound, errors.New("failed getting position")
	}

	return position, 0, nil
}

func CreatePosition(position models.Position, at models.At) (models.Position, int, error) {

	tx := initializers.DB.Begin()

	if tx.Error != nil {
		return models.Position{}, fiber.StatusInternalServerError, errors.New("failed to start DB transaction")
	}

	if err := services.DbInsert(tx, &position); err != nil {
		if strings.Contains(err.Error(), "duplicate key") {

			err = errors.New("duplicate record error")
		} else {

			err = errors.New("failed creating position")
		}
		tx.Rollback()
		return position, fiber.StatusInternalServerError, err
	}

	atdata := models.PositionAt{RefId: position.ID, At: at}
	if err := services.DbInsert(tx, &atdata); err != nil {
		tx.Rollback()
		return position, fiber.StatusInternalServerError, errors.New("failed creating positionat")
	}

	if err := tx.Commit().Error; err != nil {
		tx.Rollback()
		return position, fiber.StatusInternalServerError, errors.New("failed to commit transaction")
	}

	return position, 0, nil
}

func UpdatePosition(position models.Position, conditions map[string]interface{}, at models.At) (models.Position, int, error) {
	tx := initializers.DB.Begin()

	if tx.Error != nil {
		return models.Position{}, fiber.StatusInternalServerError, errors.New("failed to start DB transaction")
	}

	if err := services.DbUpdate(tx, &position, conditions); err != nil {
		return position, fiber.StatusInternalServerError, errors.New("failed updating position")
	}

	atdata := models.PositionAt{RefId: position.ID, At: at}

	if err := services.DbInsert(tx, &atdata); err != nil {
		tx.Rollback()
		return position, fiber.StatusInternalServerError, errors.New("failed creating positionat")
	}

	if err := tx.Commit().Error; err != nil {
		tx.Rollback()
		return position, fiber.StatusInternalServerError, errors.New("failed to commit transaction")
	}

	return position, 0, nil
}

func DeletePosition(conditions map[string]interface{}, at models.At) (models.Position, int, error) {

	tx := initializers.DB.Begin()

	if tx.Error != nil {
		return models.Position{}, fiber.StatusInternalServerError, errors.New("failed to start DB transaction")
	}

	position, status, err := GetPosition(conditions)

	if err != nil {
		return position, status, errors.New("position not found")
	}

	if err := services.DbDelete(tx, &position, conditions); err != nil {
		return position, fiber.StatusInternalServerError, errors.New("failed deleting position")
	}

	atdata := models.PositionAt{RefId: position.ID, At: at}

	if err := services.DbInsert(tx, &atdata); err != nil {
		tx.Rollback()
		return position, fiber.StatusInternalServerError, errors.New("failed creating positionat")
	}

	if err := tx.Commit().Error; err != nil {
		tx.Rollback()
		return position, fiber.StatusInternalServerError, errors.New("failed to commit transaction")
	}

	return position, 0, nil
}
