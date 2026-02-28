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

type PositionService struct {
}

func NewPositionService() *PositionService {
	return &PositionService{}
}

func (p *PositionService) GetPositionsService(conditions map[string]interface{}) (*[]models.PositionModel, int, error) {

	tx := initializers.DB.Begin()

	var positions = &[]models.PositionModel{}

	if tx.Error != nil {
		return positions, fiber.StatusInternalServerError, errors.New("failed to start DB transaction")
	}

	if err := tx.Where(conditions).Preload("Access").Find(positions).Error; err != nil {
		fmt.Println("ERROR:", err)
		return positions, fiber.StatusNotFound, errors.New("failed getting positions")
	}
	return positions, fiber.StatusOK, nil
}

func (p *PositionService) GetPositionService(conditions map[string]interface{}) (*models.PositionModel, int, error) {
	tx := initializers.DB.Begin()

	var position = &models.PositionModel{}
	if tx.Error != nil {
		return position, fiber.StatusInternalServerError, errors.New("failed to start DB transaction")
	}

	if err := tx.Where(conditions).Preload("Access").First(position).Error; err != nil {
		return position, fiber.StatusNotFound, errors.New("failed getting position")
	}

	return position, fiber.StatusOK, nil
}

func (p *PositionService) CreatePositionService(position *models.PositionModel, at models.At) (*models.PositionModel, int, error) {

	tx := initializers.DB.Begin()

	if tx.Error != nil {
		return position, fiber.StatusInternalServerError, errors.New("failed to start DB transaction")
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

	return position, fiber.StatusCreated, nil
}

func (p *PositionService) UpdatePositionService(position *models.PositionModel, conditions map[string]interface{}, at models.At) (*models.PositionModel, int, error) {
	tx := initializers.DB.Begin()

	if tx.Error != nil {
		return position, fiber.StatusInternalServerError, errors.New("failed to start DB transaction")
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

	return position, fiber.StatusOK, nil
}

func (p *PositionService) DeletePositionService(conditions map[string]interface{}, at models.At) (*models.PositionModel, int, error) {

	tx := initializers.DB.Begin()

	if tx.Error != nil {
		return &models.PositionModel{}, fiber.StatusInternalServerError, errors.New("failed to start DB transaction")
	}

	position, status, err := p.GetPositionService(conditions)

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

	return position, fiber.StatusOK, nil
}
