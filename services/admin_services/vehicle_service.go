package adminservices

import (
	"errors"
	"strings"

	"github.com/gofiber/fiber/v2"
	"github.com/pierceperado/smpc/initializers"
	"github.com/pierceperado/smpc/models"
	"github.com/pierceperado/smpc/services"
	"gorm.io/gorm"
)

type VehicleService struct {
}

func NewVehicleService() *VehicleService {
	return &VehicleService{}
}

func (v *VehicleService) CreateVehicle(vehicle models.VehicleModel, at models.At) (models.VehicleModel, int, error) {
	tx := initializers.DB.Begin()

	if tx.Error != nil {
		return models.VehicleModel{}, fiber.StatusInternalServerError, errors.New("failed to start DB transaction")
	}

	if err := services.DbInsert(tx, &vehicle); err != nil {
		if strings.Contains(err.Error(), "duplicate key") {

			err = errors.New("duplicate record error")
		} else {

			err = errors.New("failed creating vehicle")
		}
		tx.Rollback()
		return vehicle, fiber.StatusInternalServerError, err
	}

	atdata := models.VehicleAt{RefId: vehicle.ID, At: at}
	if err := services.DbInsert(tx, &atdata); err != nil {
		tx.Rollback()
		return vehicle, fiber.StatusInternalServerError, errors.New("failed creating vehicleat")
	}

	if err := tx.Commit().Error; err != nil {
		tx.Rollback()
		return vehicle, fiber.StatusInternalServerError, errors.New("failed to commit transaction")
	}

	return vehicle, 0, nil
}

func (v *VehicleService) GetVehicle(conditions map[string]interface{}) (models.VehicleModel, int, error) {
	tx := initializers.DB.Begin()

	if tx.Error != nil {
		return models.VehicleModel{}, fiber.StatusInternalServerError, errors.New("failed to start DB transaction")
	}
	var vehicle models.VehicleModel

	if err := tx.Preload("Files", func(db *gorm.DB) *gorm.DB {
		return db.Select("ID", "VehicleId", "FileName", "OriginalName", "FilePath", "Type", "Size")
	}).Where(conditions).First(&vehicle).Error; err != nil {
		return vehicle, fiber.StatusNotFound, errors.New("failed getting vehicle")
	}

	return vehicle, 0, nil
}

func (v *VehicleService) GetVehicles(conditions map[string]interface{}) ([]models.VehicleModel, int, error) {
	tx := initializers.DB.Begin()

	if tx.Error != nil {
		return []models.VehicleModel{}, fiber.StatusInternalServerError, errors.New("failed to start DB transaction")
	}

	var vehicles []models.VehicleModel

	if err := tx.Preload("Files", func(db *gorm.DB) *gorm.DB {
		return db.Select("ID", "VehicleId", "FileName", "OriginalName", "FilePath", "Type", "Size")
	}).Where(conditions).Find(&vehicles).Error; err != nil {
		return vehicles, fiber.StatusNotFound, errors.New("failed getting vehicles")
	}
	return vehicles, 0, nil
}

func (v *VehicleService) UpdateVehicle(vehicle models.VehicleModel, conditions map[string]interface{}, at models.At) (models.VehicleModel, int, error) {
	tx := initializers.DB.Begin()

	if tx.Error != nil {
		return models.VehicleModel{}, fiber.StatusInternalServerError, errors.New("failed to start DB transaction")
	}

	if err := services.DbUpdate(tx, &vehicle, conditions); err != nil {
		return vehicle, fiber.StatusInternalServerError, errors.New("failed updating vehicle")
	}

	atdata := models.VehicleAt{RefId: vehicle.ID, At: at}

	if err := services.DbInsert(tx, &atdata); err != nil {
		tx.Rollback()
		return vehicle, fiber.StatusInternalServerError, errors.New("failed creating vehicleat")
	}

	if err := tx.Commit().Error; err != nil {
		tx.Rollback()
		return vehicle, fiber.StatusInternalServerError, errors.New("failed to commit transaction")
	}

	return vehicle, 0, nil
}

func (v *VehicleService) DeleteVehicle(conditions map[string]interface{}, at models.At) (models.VehicleModel, int, error) {
	tx := initializers.DB.Begin()

	if tx.Error != nil {
		return models.VehicleModel{}, fiber.StatusInternalServerError, errors.New("failed to start DB transaction")
	}

	vehicle, status, err := v.GetVehicle(conditions)

	if err != nil {
		return vehicle, status, errors.New("vehicle not found")
	}

	if err := services.DbDelete(tx, &vehicle, conditions); err != nil {
		return vehicle, fiber.StatusInternalServerError, errors.New("failed deleting vehicle")
	}

	atdata := models.VehicleAt{RefId: vehicle.ID, At: at}

	if err := services.DbInsert(tx, &atdata); err != nil {
		tx.Rollback()
		return vehicle, fiber.StatusInternalServerError, errors.New("failed creating vehicleat")
	}

	if err := tx.Commit().Error; err != nil {
		tx.Rollback()
		return vehicle, fiber.StatusInternalServerError, errors.New("failed to commit transaction")
	}

	return vehicle, 0, nil
}
