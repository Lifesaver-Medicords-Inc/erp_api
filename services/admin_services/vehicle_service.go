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

func (v *VehicleService) CreateVehicleService(vehicle *models.VehicleModel, at models.At) (*models.VehicleModel, int, error) {
	tx := initializers.DB.Begin()

	if tx.Error != nil {
		return &models.VehicleModel{}, fiber.StatusInternalServerError, errors.New("failed to start DB transaction")
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

	// 4.4.4: every vehicle also exists as an OUTBOUND zone in the warehouse it is
	// homed to. Made here, inside the same transaction, so a vehicle can never
	// exist without the zone that 10.5's negative stock has to land on.
	if err := SyncVehicleZone(tx, vehicle, at); err != nil {
		tx.Rollback()
		return vehicle, fiber.StatusInternalServerError, err
	}

	if err := tx.Commit().Error; err != nil {
		tx.Rollback()
		return vehicle, fiber.StatusInternalServerError, errors.New("failed to commit transaction")
	}

	return vehicle, fiber.StatusOK, nil
}

func (v *VehicleService) GetVehicleService(conditions map[string]interface{}) (*models.VehicleModel, int, error) {
	tx := initializers.DB.Begin()

	var vehicle = &models.VehicleModel{}

	if tx.Error != nil {
		return vehicle, fiber.StatusInternalServerError, errors.New("failed to start DB transaction")
	}

	if err := tx.Preload("Files", func(db *gorm.DB) *gorm.DB {
		return db.Select("ID", "VehicleId", "FileName", "OriginalName", "FilePath", "Type", "Size")
	}).Where(conditions).First(vehicle).Error; err != nil {
		return vehicle, fiber.StatusNotFound, errors.New("failed getting vehicle")
	}

	return vehicle, 0, nil
}

func (v *VehicleService) GetVehiclesService(conditions map[string]interface{}) (*[]models.VehicleModel, int, error) {
	tx := initializers.DB.Begin()

	var vehicles = &[]models.VehicleModel{}
	if tx.Error != nil {
		return vehicles, fiber.StatusInternalServerError, errors.New("failed to start DB transaction")
	}

	if err := tx.Preload("Files", func(db *gorm.DB) *gorm.DB {
		return db.Select("ID", "VehicleId", "FileName", "OriginalName", "FilePath", "Type", "Size")
	}).Where(conditions).Find(vehicles).Error; err != nil {
		return vehicles, fiber.StatusNotFound, errors.New("failed getting vehicles")
	}
	return vehicles, fiber.StatusOK, nil
}

func (v *VehicleService) UpdateVehicleService(vehicle *models.VehicleModel, conditions map[string]interface{}, at models.At) (*models.VehicleModel, int, error) {
	tx := initializers.DB.Begin()

	if tx.Error != nil {
		return &models.VehicleModel{}, fiber.StatusInternalServerError, errors.New("failed to start DB transaction")
	}

	if err := services.DbUpdate(tx, &vehicle, conditions); err != nil {
		return vehicle, fiber.StatusInternalServerError, errors.New("failed updating vehicle")
	}

	atdata := models.VehicleAt{RefId: vehicle.ID, At: at}

	if err := services.DbInsert(tx, &atdata); err != nil {
		tx.Rollback()
		return vehicle, fiber.StatusInternalServerError, errors.New("failed creating vehicleat")
	}

	// Keep the zone in step. A renamed or re-homed vehicle must move its existing
	// zone, not leave the old one behind still holding stock.
	if err := SyncVehicleZone(tx, vehicle, at); err != nil {
		tx.Rollback()
		return vehicle, fiber.StatusInternalServerError, err
	}

	if err := tx.Commit().Error; err != nil {
		tx.Rollback()
		return vehicle, fiber.StatusInternalServerError, errors.New("failed to commit transaction")
	}

	return vehicle, fiber.StatusOK, nil
}

func (v *VehicleService) DeleteVehicleService(conditions map[string]interface{}, at models.At) (*models.VehicleModel, int, error) {
	tx := initializers.DB.Begin()

	if tx.Error != nil {
		return &models.VehicleModel{}, fiber.StatusInternalServerError, errors.New("failed to start DB transaction")
	}

	vehicle, status, err := v.GetVehicleService(conditions)

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

	return vehicle, fiber.StatusOK, nil
}
