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

type WarehouseService struct {
}

func NewWarehouseService() *WarehouseService {
	return &WarehouseService{}
}

func (w *WarehouseService) CreateWarehouseService(warehouse *models.WarehouseName, at models.At) (*models.WarehouseName, int, error) {
	tx := initializers.DB.Begin()

	if tx.Error != nil {
		return &models.WarehouseName{}, fiber.StatusInternalServerError, errors.New("failed to start DB transaction")
	}

	if err := services.DbInsert(tx, &warehouse); err != nil {
		if strings.Contains(err.Error(), "duplicate key") {
			err = errors.New("duplicate record error")
		} else {
			err = errors.New("failed creating warehouse")
		}
		tx.Rollback()
		return warehouse, fiber.StatusInternalServerError, err
	}

	atdata := models.WarehouseNameAt{RefId: warehouse.ID, At: at}
	if err := services.DbInsert(tx, &atdata); err != nil {
		tx.Rollback()
		return warehouse, fiber.StatusInternalServerError, errors.New("failed creating warehouseat")
	}

	if err := tx.Commit().Error; err != nil {
		tx.Rollback()
		return warehouse, fiber.StatusInternalServerError, errors.New("failed to commit transaction")
	}

	return warehouse, fiber.StatusOK, nil
}

func (w *WarehouseService) GetWarehousesService(conditions map[string]interface{}) (*[]models.WarehouseName, int, error) {
	tx := initializers.DB.Begin()

	var warehouse = &[]models.WarehouseName{}

	if tx.Error != nil {
		return warehouse, fiber.StatusInternalServerError, errors.New("failed to start DB transaction")
	}

	if err := tx.Where(conditions).Find(&warehouse).Error; err != nil {
		fmt.Println("ERROR:", err)
		return warehouse, fiber.StatusNotFound, errors.New("failed getting warehouse")
	}
	return warehouse, fiber.StatusOK, nil
}

func (w *WarehouseService) GetWarehouseService(conditions map[string]interface{}) (*models.WarehouseName, int, error) {
	tx := initializers.DB.Begin()

	var warehouse = &models.WarehouseName{}
	if tx.Error != nil {
		return warehouse, fiber.StatusInternalServerError, errors.New("failed to start DB transaction")
	}

	if err := tx.Where(conditions).Preload("Access").First(&warehouse).Error; err != nil {
		return warehouse, fiber.StatusNotFound, errors.New("failed getting warehouse")
	}

	return warehouse, fiber.StatusOK, nil
}
